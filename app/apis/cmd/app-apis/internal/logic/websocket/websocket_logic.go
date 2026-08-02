package websocket

import (
	"context"

	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type WebsocketLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWebsocketLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WebsocketLogic {
	return &WebsocketLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WebsocketLogic) Websocket() error {
	// todo: add your logic here and delete this line

	return nil
}
