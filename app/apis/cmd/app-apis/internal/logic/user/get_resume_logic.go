package user

import (
	"context"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"

	"github.com/zeromicro/go-zero/core/logx"
)

// GetResumeLogic 查询当前登录用户最近上传的一份简历。
type GetResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewGetResumeLogic 创建带请求上下文日志的简历查询逻辑。
func NewGetResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResumeLogic {
	return &GetResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// GetResume 从 JWT 上下文取得用户 ID，再由 user-rpc 查询 MySQL。
// API 只返回解析结果和状态，不向客户端暴露内部 OSS URL。
func (l *GetResumeLogic) GetResume() (resp *types.ResumeResp, err error) {
	userID, err := userIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}

	resume, err := l.svcCtx.UserRpc.GetLatestResume(l.ctx, &userclient.GetLatestResumeReq{UserId: userID})
	if err != nil {
		return nil, err
	}
	return &types.ResumeResp{
		ResumeId:   resume.ResumeId,
		ParsedJson: resume.ParsedJson,
		ParseState: resume.ParseState,
	}, nil
}
