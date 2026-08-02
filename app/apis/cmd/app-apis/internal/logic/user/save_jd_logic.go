package user

import (
	"context"

	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SaveJdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSaveJdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveJdLogic {
	return &SaveJdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SaveJdLogic) SaveJd(req *types.SaveJdReq) (resp *types.SaveJdResp, err error) {
	// todo: add your logic here and delete this line

	return
}
