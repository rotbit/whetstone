// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package system

import (
	"net/http"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/logic/system"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func HealthHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := system.NewHealthLogic(r.Context(), svcCtx)
		resp, err := l.Health()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
