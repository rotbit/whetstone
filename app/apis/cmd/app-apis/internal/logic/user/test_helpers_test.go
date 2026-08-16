package user

import (
	"context"
	"encoding/json"
	"io"

	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"

	"google.golang.org/grpc"
)

type fakeUserRPC struct {
	userclient.User
	saveResumeFn      func(context.Context, *userclient.SaveResumeReq) (*userclient.SaveResumeResp, error)
	getLatestResumeFn func(context.Context, *userclient.GetLatestResumeReq) (*userclient.ResumeInfo, error)
	saveJdFn          func(context.Context, *userclient.SaveJdReq) (*userclient.SaveJdResp, error)
}

func (f *fakeUserRPC) SaveResume(
	ctx context.Context,
	req *userclient.SaveResumeReq,
	_ ...grpc.CallOption,
) (*userclient.SaveResumeResp, error) {
	return f.saveResumeFn(ctx, req)
}

func (f *fakeUserRPC) GetLatestResume(
	ctx context.Context,
	req *userclient.GetLatestResumeReq,
	_ ...grpc.CallOption,
) (*userclient.ResumeInfo, error) {
	return f.getLatestResumeFn(ctx, req)
}

func (f *fakeUserRPC) SaveJd(
	ctx context.Context,
	req *userclient.SaveJdReq,
	_ ...grpc.CallOption,
) (*userclient.SaveJdResp, error) {
	return f.saveJdFn(ctx, req)
}

type fakeObjectStorage struct {
	putFn    func(context.Context, string, io.Reader, int64, string) (string, error)
	deleteFn func(context.Context, string) (string, error)
	urlFn    func(string) string
}

func (f *fakeObjectStorage) Put(
	ctx context.Context,
	key string,
	body io.Reader,
	size int64,
	contentType string,
) (string, error) {
	return f.putFn(ctx, key, body, size, contentType)
}

func (f *fakeObjectStorage) Delete(ctx context.Context, key string) (string, error) {
	return f.deleteFn(ctx, key)
}

func (f *fakeObjectStorage) URL(key string) string {
	return f.urlFn(key)
}

func contextWithUserID(userID string) context.Context {
	return context.WithValue(context.Background(), "uid", json.Number(userID))
}
