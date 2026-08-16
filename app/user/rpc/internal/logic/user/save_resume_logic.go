package userlogic

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultResumeParseState = "parsing"
	maxOSSURLLength         = 512
)

type SaveResumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveResumeLogic {
	return &SaveResumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存已上传到 OSS 的简历元数据；oss_url 用作幂等键。
func (l *SaveResumeLogic) SaveResume(in *pb.SaveResumeReq) (*pb.SaveResumeResp, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 不正确")
	}

	ossURL := strings.TrimSpace(in.GetOssUrl())
	if !isValidOSSURL(ossURL) {
		return nil, status.Error(codes.InvalidArgument, "简历文件地址不正确")
	}

	existing, err := l.svcCtx.ResumesModel.FindOneByOSSURL(l.ctx, ossURL)
	if err == nil {
		return resumeSaveResponse(existing, in.GetUserId())
	}
	if !errors.Is(err, model.ErrNotFound) {
		l.Errorf("find resume by OSS URL failed: %v", err)
		return nil, status.Error(codes.Internal, "保存简历失败")
	}

	result, err := l.svcCtx.ResumesModel.Insert(l.ctx, &model.Resume{
		UserID:     uint64(in.GetUserId()),
		OSSURL:     ossURL,
		ParseState: defaultResumeParseState,
	})
	if err != nil {
		if isDuplicateEntry(err) {
			return l.findDuplicateResume(ossURL, in.GetUserId())
		}
		l.Errorf("insert resume failed: %v", err)
		return nil, status.Error(codes.Internal, "保存简历失败")
	}

	resumeID, err := result.LastInsertId()
	if err != nil {
		l.Errorf("get inserted resume id failed: %v", err)
		return nil, status.Error(codes.Internal, "保存简历失败")
	}

	return &pb.SaveResumeResp{ResumeId: resumeID, ParseState: defaultResumeParseState}, nil
}

func (l *SaveResumeLogic) findDuplicateResume(ossURL string, userID int64) (*pb.SaveResumeResp, error) {
	resume, err := l.svcCtx.ResumesModel.FindOneByOSSURL(l.ctx, ossURL)
	if err != nil {
		l.Errorf("find duplicate resume failed: %v", err)
		return nil, status.Error(codes.Internal, "保存简历失败")
	}
	return resumeSaveResponse(resume, userID)
}

func resumeSaveResponse(resume *model.Resume, userID int64) (*pb.SaveResumeResp, error) {
	if resume.UserID != uint64(userID) {
		return nil, status.Error(codes.AlreadyExists, "简历文件地址已存在")
	}
	return &pb.SaveResumeResp{ResumeId: int64(resume.ID), ParseState: resume.ParseState}, nil
}

func isValidOSSURL(rawURL string) bool {
	if rawURL == "" || len(rawURL) > maxOSSURLLength {
		return false
	}
	parsedURL, err := url.ParseRequestURI(rawURL)
	return err == nil && parsedURL.Scheme == "https" && parsedURL.Host != "" && parsedURL.User == nil &&
		parsedURL.RawQuery == "" && parsedURL.Fragment == ""
}
