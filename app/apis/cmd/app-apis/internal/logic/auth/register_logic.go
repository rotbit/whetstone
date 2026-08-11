package auth

import (
	"context"
	"time"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (*types.TokenResp, error) {

	//网关层做快速校验，挡掉明显垃圾，节省rpc资源(针对输入)
	if !phonePattern.MatchString(req.Phone) {
		return nil, status.Error(codes.InvalidArgument, "手机号格式错误")
	}
	if len(req.Password) < 8 || len(req.Password) > 20 {
		return nil, status.Error(codes.InvalidArgument, "密码长度必须在8-20之间")
	}

	//调用UserRpc.Register 注册用户
	registered, err := l.svcCtx.UserRpc.Register(l.ctx, &userclient.RegisterReq{
		Phone:    req.Phone,
		Password: req.Password,
	})
	if err != nil {
		// 区分超时和其他错误
		if status.Code(err) == codes.DeadlineExceeded {
			l.Errorf("登录rpc超时:%v", err)
			return nil, status.Error(codes.Internal, "登录超时，请稍后重试")
		}
		return nil, err
	}

	now := time.Now().Unix()
	expireAt := now + l.svcCtx.Config.Auth.AccessExpire
	accessToken, err := createAccessToken(
		l.svcCtx.Config.Auth.AccessSecret,
		now,
		expireAt,
		registered.UserId,
	)
	if err != nil {
		l.Errorf("create access token failed: %v", err)
		return nil, status.Error(codes.Internal, "注册失败")
	}

	return &types.TokenResp{
		AccessToken: accessToken,
		ExpireAt:    expireAt,
	}, nil
}

func createAccessToken(secret string, issuedAt, expireAt, userID int64) (string, error) {
	claims := jwt.MapClaims{
		"iat": issuedAt,
		"exp": expireAt,
		"uid": userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
