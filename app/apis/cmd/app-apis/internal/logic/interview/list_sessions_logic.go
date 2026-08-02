package interview

import (
	"context"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListSessionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListSessionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListSessionsLogic {
	return &ListSessionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListSessionsLogic) ListSessions(req *types.ListSessionsReq) (resp *types.ListSessionsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
