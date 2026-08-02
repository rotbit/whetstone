package websocket

import (
	"net/http"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
)

func WebsocketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO(M1): 校验 JWT 和 sessionId，升级 WebSocket，并注册到
		// svcCtx.WsConnections。流式问答由 internal/ws/pipeline 编排。
		_ = svcCtx
		http.Error(w, "websocket upgrade is not implemented", http.StatusNotImplemented)
	}
}
