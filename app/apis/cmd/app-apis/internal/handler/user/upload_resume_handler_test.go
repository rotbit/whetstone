package user

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"

	"google.golang.org/grpc"
)

type uploadHandlerUserRPC struct {
	userclient.User
	t *testing.T
}

func (f *uploadHandlerUserRPC) SaveResume(
	_ context.Context,
	req *userclient.SaveResumeReq,
	_ ...grpc.CallOption,
) (*userclient.SaveResumeResp, error) {
	if req.UserId != 42 || !strings.HasPrefix(req.OssUrl, "https://resume.example/resumes/42/") {
		f.t.Fatalf("SaveResume() request = %+v", req)
	}
	return &userclient.SaveResumeResp{ResumeId: 11, ParseState: "parsing"}, nil
}

type uploadHandlerStorage struct {
	t *testing.T
}

func (s *uploadHandlerStorage) Put(
	_ context.Context,
	key string,
	body io.Reader,
	size int64,
	contentType string,
) (string, error) {
	uploaded, err := io.ReadAll(body)
	if err != nil {
		s.t.Fatalf("read uploaded body: %v", err)
	}
	if string(uploaded) != "%PDF-1.7\nresume" || size != int64(len(uploaded)) || contentType != "application/pdf" {
		s.t.Fatalf("Put() key=%q body=%q size=%d contentType=%q", key, uploaded, size, contentType)
	}
	return "put-request", nil
}

func (s *uploadHandlerStorage) Delete(context.Context, string) (string, error) {
	s.t.Fatal("Delete() should not be called")
	return "", nil
}

func (s *uploadHandlerStorage) URL(key string) string {
	return "https://resume.example/" + key
}

func TestUploadResumeHandlerReadsMultipartFile(t *testing.T) {
	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)
	part, err := multipartWriter.CreateFormFile("file", "resume.pdf")
	if err != nil {
		t.Fatalf("CreateFormFile() error = %v", err)
	}
	if _, err := part.Write([]byte("%PDF-1.7\nresume")); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := multipartWriter.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume/upload", &requestBody)
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
	req = req.WithContext(context.WithValue(req.Context(), "uid", json.Number("42")))
	recorder := httptest.NewRecorder()
	svcCtx := &svc.ServiceContext{
		UserRpc:       &uploadHandlerUserRPC{t: t},
		ObjectStorage: &uploadHandlerStorage{t: t},
	}

	UploadResumeHandler(svcCtx).ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %q", recorder.Code, recorder.Body.String())
	}
	var resp types.UploadResumeResp
	if err := json.Unmarshal(recorder.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v, body = %q", err, recorder.Body.String())
	}
	if resp.ResumeId != 11 || resp.ParseState != "parsing" {
		t.Fatalf("response = %+v", resp)
	}
}

func TestUploadResumeHandlerRejectsNonMultipartRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/resume/upload", strings.NewReader("not multipart"))
	recorder := httptest.NewRecorder()

	UploadResumeHandler(&svc.ServiceContext{}).ServeHTTP(recorder, req)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %q", recorder.Code, http.StatusBadRequest, recorder.Body.String())
	}
}
