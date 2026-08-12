package userlogic

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/rotbit/whetstone/app/user/model"
	"golang.org/x/crypto/bcrypt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *pb.LoginRep) (*pb.LoginResp, error) {

	phone := strings.TrimSpace(in.GetPhone()) //清理手机号首尾空格
	// 校验手机号格式
	if !phonePattern.MatchString(phone) {
		return nil, status.Error(codes.InvalidArgument, "手机号格式错误")
	}
	password := strings.TrimSpace(in.GetPassword()) //清理密码首尾空格
	if len(password) < 8 || len(password) > 20 {
		return nil, status.Error(codes.InvalidArgument, "密码长度必须在8-20之间")
	}

	//调用UsersMode1 .FindoneByPhone 查询用户。
	user, err := l.svcCtx.UsersModel.FindOneByPhone(l.ctx, phone)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			// 用户不存在 → 返回 404，用"手机号或密码错误"
			// 不能说"用户不存在"，否则会被撞库
			return nil, status.Error(codes.Unauthenticated, "手机号或密码错误")
		}
		//其他数据库错误 ->记日志 +返回 500
		l.Logger.Errorf("find user by phone failed: %v", err)
		return nil, status.Error(codes.Internal, "登录失败")
	}

	//密码校验
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "手机号或密码错误")
	}

	// 签发 JWT access token
	now := time.Now().Unix()
	expireAt := now + l.svcCtx.Config.TokenAuth.AccessExpire
	accessToken, err := createAccessToken(
		l.svcCtx.Config.TokenAuth.AccessSecret,
		now,
		expireAt,
		int64(user.Id),
	)
	if err != nil {
		l.Errorf("create access token failed: %v", err)
		return nil, status.Error(codes.Internal, "登录失败")
	}

	//登录成功
	return &pb.LoginResp{
		UserId:      int64(user.Id),
		Phone:       user.Phone,
		Plan:        user.Plan,
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
