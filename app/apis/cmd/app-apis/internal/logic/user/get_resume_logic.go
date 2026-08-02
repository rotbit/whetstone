package user

import (
	"context"

	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetResumeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetResumeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetResumeLogic {
	return &GetResumeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetResumeLogic) GetResume() (resp *types.ResumeResp, err error) {
	// todo: add your logic here and delete this line

	return
}
