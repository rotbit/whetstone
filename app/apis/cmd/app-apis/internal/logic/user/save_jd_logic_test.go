package user

import (
	"context"
	"testing"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSaveJdTrimsAndSavesContent(t *testing.T) {
	rpc := &fakeUserRPC{
		saveJdFn: func(_ context.Context, req *userclient.SaveJdReq) (*userclient.SaveJdResp, error) {
			if req.UserId != 42 || req.Title != "Go 工程师" || req.Content != "负责 API 开发" {
				t.Fatalf("SaveJd() request = %+v", req)
			}
			return &userclient.SaveJdResp{JdId: 15}, nil
		},
	}

	resp, err := NewSaveJdLogic(contextWithUserID("42"), &svc.ServiceContext{UserRpc: rpc}).SaveJd(
		&types.SaveJdReq{Title: "  Go 工程师  ", Content: "\n负责 API 开发\n"},
	)
	if err != nil {
		t.Fatalf("SaveJd() error = %v", err)
	}
	if resp.JdId != 15 {
		t.Fatalf("SaveJd() response = %+v", resp)
	}
}

func TestSaveJdRejectsInvalidInputBeforeRPC(t *testing.T) {
	rpcCalled := false
	rpc := &fakeUserRPC{
		saveJdFn: func(context.Context, *userclient.SaveJdReq) (*userclient.SaveJdResp, error) {
			rpcCalled = true
			return nil, nil
		},
	}

	_, err := NewSaveJdLogic(contextWithUserID("42"), &svc.ServiceContext{UserRpc: rpc}).SaveJd(
		&types.SaveJdReq{Title: " ", Content: "content"},
	)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("SaveJd() code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
	if rpcCalled {
		t.Fatal("SaveJd RPC should not be called")
	}
}
