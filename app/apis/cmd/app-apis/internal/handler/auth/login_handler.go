package auth

import (
	"net/http"

	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/logic/auth"
	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func LoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := auth.NewLoginLogic(r.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
