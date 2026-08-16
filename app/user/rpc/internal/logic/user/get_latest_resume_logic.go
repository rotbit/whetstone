package userlogic

import (
	"context"
	"errors"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GetLatestResumeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLatestResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLatestResumeLogic {
	return &GetLatestResumeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询用户最近上传的一份简历。
func (l *GetLatestResumeLogic) GetLatestResume(in *pb.GetLatestResumeReq) (*pb.ResumeInfo, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 不正确")
	}

	resume, err := l.svcCtx.ResumesModel.FindLatestByUserID(l.ctx, uint64(in.GetUserId()))
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "简历不存在")
		}
		l.Errorf("find latest resume failed: %v", err)
		return nil, status.Error(codes.Internal, "查询简历失败")
	}

	parsedJSON := ""
	if resume.ParsedJSON.Valid {
		parsedJSON = resume.ParsedJSON.String
	}
	return &pb.ResumeInfo{
		ResumeId:   int64(resume.ID),
		ParsedJson: parsedJSON,
		ParseState: resume.ParseState,
	}, nil
}
