package user

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUploadResumeSuccess(t *testing.T) {
	pdf := []byte("%PDF-1.7\nresume")
	var uploadedKey string
	storage := &fakeObjectStorage{
		putFn: func(_ context.Context, key string, body io.Reader, size int64, contentType string) (string, error) {
			uploaded, err := io.ReadAll(body)
			if err != nil {
				t.Fatalf("read uploaded body: %v", err)
			}
			if !bytes.Equal(uploaded, pdf) || size != int64(len(pdf)) || contentType != pdfContentType {
				t.Fatalf("uploaded body=%q size=%d contentType=%q", uploaded, size, contentType)
			}
			uploadedKey = key
			return "oss-request-1", nil
		},
		deleteFn: func(context.Context, string) (string, error) {
			t.Fatal("Delete() should not be called")
			return "", nil
		},
		urlFn: func(key string) string { return "https://resume.example/" + key },
	}
	rpc := &fakeUserRPC{
		saveResumeFn: func(_ context.Context, req *userclient.SaveResumeReq) (*userclient.SaveResumeResp, error) {
			if req.UserId != 42 || req.OssUrl != "https://resume.example/"+uploadedKey {
				t.Fatalf("SaveResume() request = %+v", req)
			}
			return &userclient.SaveResumeResp{ResumeId: 9, ParseState: "parsing"}, nil
		},
	}

	resp, err := NewUploadResumeLogic(contextWithUserID("42"), &svc.ServiceContext{
		UserRpc:       rpc,
		ObjectStorage: storage,
	}).UploadResume(bytes.NewReader(pdf), int64(len(pdf)))
	if err != nil {
		t.Fatalf("UploadResume() error = %v", err)
	}
	if resp.ResumeId != 9 || resp.ParseState != "parsing" {
		t.Fatalf("UploadResume() response = %+v", resp)
	}
	if !strings.HasPrefix(uploadedKey, "resumes/42/") || !strings.HasSuffix(uploadedKey, ".pdf") {
		t.Fatalf("uploaded key = %q", uploadedKey)
	}
}

func TestUploadResumeDeletesObjectAfterDefiniteRPCFailure(t *testing.T) {
	pdf := []byte("%PDF-1.7\nresume")
	var uploadedKey, deletedKey string
	storage := successfulStorage(&uploadedKey, &deletedKey)
	rpc := &fakeUserRPC{
		saveResumeFn: func(context.Context, *userclient.SaveResumeReq) (*userclient.SaveResumeResp, error) {
			return nil, status.Error(codes.InvalidArgument, "invalid URL")
		},
	}

	_, err := NewUploadResumeLogic(contextWithUserID("42"), &svc.ServiceContext{
		UserRpc:       rpc,
		ObjectStorage: storage,
	}).UploadResume(bytes.NewReader(pdf), int64(len(pdf)))
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("UploadResume() code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
	if deletedKey == "" || deletedKey != uploadedKey {
		t.Fatalf("deleted key = %q, uploaded key = %q", deletedKey, uploadedKey)
	}
}

func TestUploadResumeKeepsObjectWhenRPCResultIsUncertain(t *testing.T) {
	pdf := []byte("%PDF-1.7\nresume")
	var uploadedKey, deletedKey string
	callCount := 0
	rpc := &fakeUserRPC{
		saveResumeFn: func(context.Context, *userclient.SaveResumeReq) (*userclient.SaveResumeResp, error) {
			callCount++
			return nil, status.Error(codes.Unavailable, "rpc unavailable")
		},
	}

	_, err := NewUploadResumeLogic(contextWithUserID("42"), &svc.ServiceContext{
		UserRpc:       rpc,
		ObjectStorage: successfulStorage(&uploadedKey, &deletedKey),
	}).UploadResume(bytes.NewReader(pdf), int64(len(pdf)))
	if status.Code(err) != codes.Unavailable {
		t.Fatalf("UploadResume() code = %v, want %v", status.Code(err), codes.Unavailable)
	}
	if callCount != 2 {
		t.Fatalf("SaveResume() calls = %d, want 2", callCount)
	}
	if deletedKey != "" {
		t.Fatalf("Delete() key = %q, want no deletion", deletedKey)
	}
}

func TestUploadResumeRejectsInvalidPDFBeforeStorage(t *testing.T) {
	putCalled := false
	storage := &fakeObjectStorage{
		putFn: func(context.Context, string, io.Reader, int64, string) (string, error) {
			putCalled = true
			return "", errors.New("unexpected Put call")
		},
	}

	_, err := NewUploadResumeLogic(contextWithUserID("42"), &svc.ServiceContext{
		ObjectStorage: storage,
	}).UploadResume(bytes.NewReader([]byte("not a PDF")), 9)
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("UploadResume() code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}
	if putCalled {
		t.Fatal("Put() should not be called")
	}
}

func TestGetResumeMapsLatestResume(t *testing.T) {
	rpc := &fakeUserRPC{
		getLatestResumeFn: func(_ context.Context, req *userclient.GetLatestResumeReq) (*userclient.ResumeInfo, error) {
			if req.UserId != 42 {
				t.Fatalf("GetLatestResume() userId = %d, want 42", req.UserId)
			}
			return &userclient.ResumeInfo{ResumeId: 7, ParsedJson: `{"name":"Ada"}`, ParseState: "done"}, nil
		},
	}

	resp, err := NewGetResumeLogic(contextWithUserID("42"), &svc.ServiceContext{UserRpc: rpc}).GetResume()
	if err != nil {
		t.Fatalf("GetResume() error = %v", err)
	}
	if resp.ResumeId != 7 || resp.ParsedJson != `{"name":"Ada"}` || resp.ParseState != "done" {
		t.Fatalf("GetResume() response = %+v", resp)
	}
}

func TestResumeRPCErrorCertainty(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		uncertain bool
	}{
		{name: "invalid argument", err: status.Error(codes.InvalidArgument, "invalid"), uncertain: false},
		{name: "already exists", err: status.Error(codes.AlreadyExists, "exists"), uncertain: false},
		{name: "unavailable", err: status.Error(codes.Unavailable, "unavailable"), uncertain: true},
		{name: "canceled", err: status.Error(codes.Canceled, "canceled"), uncertain: true},
		{name: "plain error", err: errors.New("network failure"), uncertain: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isUncertainRPCError(test.err); got != test.uncertain {
				t.Fatalf("isUncertainRPCError() = %v, want %v", got, test.uncertain)
			}
		})
	}
}

func successfulStorage(uploadedKey, deletedKey *string) *fakeObjectStorage {
	return &fakeObjectStorage{
		putFn: func(_ context.Context, key string, _ io.Reader, _ int64, _ string) (string, error) {
			*uploadedKey = key
			return "put-request", nil
		},
		deleteFn: func(_ context.Context, key string) (string, error) {
			*deletedKey = key
			return "delete-request", nil
		},
		urlFn: func(key string) string { return "https://resume.example/" + key },
	}
}
