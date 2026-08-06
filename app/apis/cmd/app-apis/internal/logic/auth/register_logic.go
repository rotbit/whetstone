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
	registered, err := l.svcCtx.UserRpc.Register(l.ctx, &userclient.RegisterReq{
		Phone:    req.Phone,
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
