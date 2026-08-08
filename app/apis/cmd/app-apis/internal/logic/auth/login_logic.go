package auth

import (
	"context"
	"time"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"
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
	logined,err := l.svcCtx.UserRpc.Login(l.ctx,&userclient.LoginRep{
		Phone: req.Phone,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	expireAt := now + l.svcCtx.Config.Auth.AccessExpire
	accessToken, err := createAccessToken(
		l.svcCtx.Config.Auth.AccessSecret,
		now,
		expireAt,
		logined.UserId,
	)
	if err != nil {
		l.Errorf("create access token failed: %v", err)
		return nil, status.Error(codes.Internal, "登录失败")
	}
	
	l.Logger.Info("登录成功",logined)
	return &types.TokenResp{
		AccessToken: accessToken,
		ExpireAt:    expireAt,
	},nil
}