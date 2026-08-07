package auth

import (
	"context"
	"errors"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.TokenResp, err error) {
	// todo: add your logic here and delete this line
	if req.Phone == "19900008007" && req.Password == "TestPass123" {
		l.Logger.Info("登录成功")
		return &types.TokenResp{
			AccessToken: "1123456",
			ExpireAt:    1786000000,
		}, nil

	}
	l.Logger.Error("登录失败")
	return nil, errors.New("账号或者密码错误")

}
