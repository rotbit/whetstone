package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeUserRpc struct {
	registerReq  *userclient.RegisterReq
	registerResp *userclient.RegisterResp
	registerErr  error
}

func (f *fakeUserRpc) Register(
	_ context.Context,
	in *userclient.RegisterReq,
	_ ...grpc.CallOption,
) (*userclient.RegisterResp, error) {
	f.registerReq = in
	return f.registerResp, f.registerErr
}

func (f *fakeUserRpc) Login(
	context.Context,
	*userclient.LoginRep,
	...grpc.CallOption,
) (*userclient.LoginResp, error) {
	panic("unexpected Login call")
}

func (f *fakeUserRpc) GetUser(
	context.Context,
	*userclient.GetUserReq,
	...grpc.CallOption,
) (*userclient.UserInfo, error) {
	panic("unexpected GetUser call")
}

func (f *fakeUserRpc) DeductCredit(
	context.Context,
	*userclient.DeductCreditReq,
	...grpc.CallOption,
) (*userclient.DeductCreditResp, error) {
	panic("unexpected DeductCredit call")
}

func (f *fakeUserRpc) RefundCredit(
	context.Context,
	*userclient.RefundCreditReq,
	...grpc.CallOption,
) (*userclient.RefundCreditResp, error) {
	panic("unexpected RefundCredit call")
}

func TestRegisterReturnsSignedToken(t *testing.T) {
	fakeRpc := &fakeUserRpc{
		registerResp: &userclient.RegisterResp{UserId: 42, Phone: "13800138000", Plan: "free"},
	}
	svcCtx := &svc.ServiceContext{UserRpc: fakeRpc}
	svcCtx.Config.Auth.AccessSecret = "test-access-secret"
	svcCtx.Config.Auth.AccessExpire = 3600

	resp, err := NewRegisterLogic(context.Background(), svcCtx).Register(&types.RegisterReq{
		Phone:    "13800138000",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if fakeRpc.registerReq.Phone != "13800138000" || fakeRpc.registerReq.Password != "password123" {
		t.Fatalf("RPC request = %+v", fakeRpc.registerReq)
	}
	if resp.ExpireAt <= 0 || resp.AccessToken == "" {
		t.Fatalf("Register() response = %+v", resp)
	}

	claims := parseTokenClaims(t, resp.AccessToken, svcCtx.Config.Auth.AccessSecret)
	if claims["uid"] != float64(42) {
		t.Fatalf("token uid = %v, want 42", claims["uid"])
	}
	if claims["exp"] != float64(resp.ExpireAt) {
		t.Fatalf("token exp = %v, want %d", claims["exp"], resp.ExpireAt)
	}
}

func TestRegisterPropagatesRpcError(t *testing.T) {
	fakeRpc := &fakeUserRpc{registerErr: status.Error(codes.AlreadyExists, "手机号已注册")}
	svcCtx := &svc.ServiceContext{UserRpc: fakeRpc}

	_, err := NewRegisterLogic(context.Background(), svcCtx).Register(&types.RegisterReq{})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("Register() code = %v, want %v", status.Code(err), codes.AlreadyExists)
	}
}

func parseTokenClaims(t *testing.T, tokenString, secret string) jwt.MapClaims {
	t.Helper()
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		t.Fatal("token claims are invalid")
	}
	return claims
}
