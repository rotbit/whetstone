package user

import (
	"net/http"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/logic/user"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetResumeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := user.NewGetResumeLogic(r.Context(), svcCtx)
		resp, err := l.GetResume()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
