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
	// 新记录先进入 parsing，后续解析任务再更新为 done 或 failed。
	defaultResumeParseState = "parsing"
	// 与 resumes.oss_url VARCHAR(512) 保持一致，按 UTF-8 字节长度限制。
	maxOSSURLLength = 512
)

// SaveResumeLogic 保存 app-apis 已上传成功的 OSS 文件元数据。
type SaveResumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

// NewSaveResumeLogic 创建带 RPC 请求上下文日志的简历保存逻辑。
func NewSaveResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveResumeLogic {
	return &SaveResumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// SaveResume 保存已上传到 OSS 的简历元数据，并以 oss_url 作为幂等键。
// 应用层先查询可快速返回历史结果，数据库唯一索引负责解决并发请求之间的竞态。
func (l *SaveResumeLogic) SaveResume(in *pb.SaveResumeReq) (*pb.SaveResumeResp, error) {
	// user-rpc 不信任调用方传入的零值或负数用户 ID。
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 不正确")
	}

	// 只保存稳定 HTTPS URL，明确拒绝带签名查询参数的临时下载地址。
	ossURL := strings.TrimSpace(in.GetOssUrl())
	if !isValidOSSURL(ossURL) {
		return nil, status.Error(codes.InvalidArgument, "简历文件地址不正确")
	}

	// 常规重试在这里直接返回已有记录，不会重复创建简历。
	existing, err := l.svcCtx.ResumesModel.FindOneByOSSURL(l.ctx, ossURL)
	if err == nil {
		return resumeSaveResponse(existing, in.GetUserId())
	}
	if !errors.Is(err, model.ErrNotFound) {
		l.Errorf("find resume by OSS URL failed: %v", err)
		return nil, status.Error(codes.Internal, "保存简历失败")
	}

	// 只有明确未找到相同 URL 时才插入；解析字段暂不写入，使用数据库默认值。
	result, err := l.svcCtx.ResumesModel.Insert(l.ctx, &model.Resume{
		UserID:     uint64(in.GetUserId()),
		OSSURL:     ossURL,
		ParseState: defaultResumeParseState,
	})
	if err != nil {
		// 两个并发请求可能同时通过前置查询，唯一索引冲突后重新查询即可得到同一结果。
		if isDuplicateEntry(err) {
			return l.findDuplicateResume(ossURL, in.GetUserId())
		}
		l.Errorf("insert resume failed: %v", err)
		return nil, status.Error(codes.Internal, "保存简历失败")
	}

	// RPC 直接返回数据库生成的 ID，API 不需要再次查询。
	resumeID, err := result.LastInsertId()
	if err != nil {
		l.Errorf("get inserted resume id failed: %v", err)
		return nil, status.Error(codes.Internal, "保存简历失败")
	}

	return &pb.SaveResumeResp{ResumeId: resumeID, ParseState: defaultResumeParseState}, nil
}

// findDuplicateResume 处理并发插入触发的 MySQL 1062，并返回已经成功写入的记录。
func (l *SaveResumeLogic) findDuplicateResume(ossURL string, userID int64) (*pb.SaveResumeResp, error) {
	resume, err := l.svcCtx.ResumesModel.FindOneByOSSURL(l.ctx, ossURL)
	if err != nil {
		l.Errorf("find duplicate resume failed: %v", err)
		return nil, status.Error(codes.Internal, "保存简历失败")
	}
	return resumeSaveResponse(resume, userID)
}

// resumeSaveResponse 校验幂等记录归属，禁止把其他用户的简历 ID 返回给当前调用方。
func resumeSaveResponse(resume *model.Resume, userID int64) (*pb.SaveResumeResp, error) {
	if resume.UserID != uint64(userID) {
		return nil, status.Error(codes.AlreadyExists, "简历文件地址已存在")
	}
	return &pb.SaveResumeResp{ResumeId: int64(resume.ID), ParseState: resume.ParseState}, nil
}

// isValidOSSURL 校验数据库可存储的稳定 HTTPS URL。
// 禁止用户凭证、查询参数和片段，避免把 AccessKey、签名或其他临时信息持久化。
func isValidOSSURL(rawURL string) bool {
	if rawURL == "" || len(rawURL) > maxOSSURLLength {
		return false
	}
	parsedURL, err := url.ParseRequestURI(rawURL)
	return err == nil && parsedURL.Scheme == "https" && parsedURL.Host != "" && parsedURL.User == nil &&
		parsedURL.RawQuery == "" && parsedURL.Fragment == ""
}
