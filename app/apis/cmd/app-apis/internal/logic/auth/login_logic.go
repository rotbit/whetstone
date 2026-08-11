package auth

import (
	"context"
	"regexp"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

var phonePattern = regexp.MustCompile(`^\+?[0-9]{6,20}$`)

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.TokenResp, err error) {

	//网关层做快速校验，挡掉明显垃圾，节省rpc资源(针对输入)
	if !phonePattern.MatchString(req.Phone) {
		return nil, status.Error(codes.InvalidArgument, "手机号格式错误")
	}

	if len(req.Password) < 8 || len(req.Password) > 20 {
		return nil, status.Error(codes.InvalidArgument, "密码长度必须在8-20之间")
	}

	logined, err := l.svcCtx.UserRpc.Login(l.ctx, &userclient.LoginRep{
		Phone:    req.Phone,
		Password: req.Password,
	})

	if err != nil {
		return nil, err
	}

	l.Logger.Info("登录成功", logined)
	return &types.TokenResp{
		AccessToken: logined.AccessToken,
		ExpireAt:    logined.ExpireAt,
	}, nil
}
