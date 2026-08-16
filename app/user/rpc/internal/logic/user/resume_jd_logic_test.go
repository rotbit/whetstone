package userlogic

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/go-sql-driver/mysql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeResumesModel struct {
	insertFn       func(context.Context, *model.Resume) (sql.Result, error)
	findLatestFn   func(context.Context, uint64) (*model.Resume, error)
	findByOSSURLFn func(context.Context, string) (*model.Resume, error)
}

func (m *fakeResumesModel) Insert(ctx context.Context, data *model.Resume) (sql.Result, error) {
	return m.insertFn(ctx, data)
}

func (m *fakeResumesModel) FindLatestByUserID(ctx context.Context, userID uint64) (*model.Resume, error) {
	return m.findLatestFn(ctx, userID)
}

func (m *fakeResumesModel) FindOneByOSSURL(ctx context.Context, ossURL string) (*model.Resume, error) {
	return m.findByOSSURLFn(ctx, ossURL)
}

type fakeJDsModel struct {
	insertFn func(context.Context, *model.JD) (sql.Result, error)
}

func (m *fakeJDsModel) Insert(ctx context.Context, data *model.JD) (sql.Result, error) {
	return m.insertFn(ctx, data)
}

func TestSaveResumeInsertsMetadata(t *testing.T) {
	var inserted *model.Resume
	resumes := &fakeResumesModel{
		findByOSSURLFn: func(context.Context, string) (*model.Resume, error) {
			return nil, model.ErrNotFound
		},
		insertFn: func(_ context.Context, resume *model.Resume) (sql.Result, error) {
			inserted = resume
			return fakeResult{id: 23}, nil
		},
	}
	svcCtx := &svc.ServiceContext{ResumesModel: resumes}
	url := "https://resume.example/resumes/42/resume.pdf"

	resp, err := NewSaveResumeLogic(context.Background(), svcCtx).SaveResume(&pb.SaveResumeReq{
		UserId: 42,
		OssUrl: "  " + url + "  ",
	})
	if err != nil {
		t.Fatalf("SaveResume() error = %v", err)
	}
	if resp.ResumeId != 23 || resp.ParseState != defaultResumeParseState {
		t.Fatalf("SaveResume() response = %+v", resp)
	}
	if inserted == nil || inserted.UserID != 42 || inserted.OSSURL != url ||
		inserted.ParseState != defaultResumeParseState {
		t.Fatalf("inserted resume = %+v", inserted)
	}
}

func TestSaveResumeReturnsExistingMetadata(t *testing.T) {
	resumes := &fakeResumesModel{
		findByOSSURLFn: func(context.Context, string) (*model.Resume, error) {
			return &model.Resume{ID: 18, UserID: 42, ParseState: "done"}, nil
		},
		insertFn: func(context.Context, *model.Resume) (sql.Result, error) {
			t.Fatal("Insert() should not be called")
			return nil, nil
		},
	}

	resp, err := NewSaveResumeLogic(context.Background(), &svc.ServiceContext{
		ResumesModel: resumes,
	}).SaveResume(&pb.SaveResumeReq{UserId: 42, OssUrl: "https://resume.example/resume.pdf"})
	if err != nil {
		t.Fatalf("SaveResume() error = %v", err)
	}
	if resp.ResumeId != 18 || resp.ParseState != "done" {
		t.Fatalf("SaveResume() response = %+v", resp)
	}
}

func TestSaveResumeHandlesDuplicateInsertRace(t *testing.T) {
	findCalls := 0
	resumes := &fakeResumesModel{
		findByOSSURLFn: func(context.Context, string) (*model.Resume, error) {
			findCalls++
			if findCalls == 1 {
				return nil, model.ErrNotFound
			}
			return &model.Resume{ID: 31, UserID: 42, ParseState: "parsing"}, nil
		},
		insertFn: func(context.Context, *model.Resume) (sql.Result, error) {
			return nil, &mysql.MySQLError{Number: 1062, Message: "duplicate OSS URL"}
		},
	}

	resp, err := NewSaveResumeLogic(context.Background(), &svc.ServiceContext{
		ResumesModel: resumes,
	}).SaveResume(&pb.SaveResumeReq{UserId: 42, OssUrl: "https://resume.example/resume.pdf"})
	if err != nil {
		t.Fatalf("SaveResume() error = %v", err)
	}
	if findCalls != 2 || resp.ResumeId != 31 {
		t.Fatalf("find calls = %d, response = %+v", findCalls, resp)
	}
}

func TestSaveResumeRejectsInvalidOrForeignURL(t *testing.T) {
	logic := NewSaveResumeLogic(context.Background(), &svc.ServiceContext{})
	invalidURLs := []string{
		"",
		"http://resume.example/resume.pdf",
		"https://resume.example/resume.pdf?signature=secret",
		"https://resume.example/" + strings.Repeat("a", 500),
	}
	for _, rawURL := range invalidURLs {
		_, err := logic.SaveResume(&pb.SaveResumeReq{UserId: 42, OssUrl: rawURL})
		if status.Code(err) != codes.InvalidArgument {
			t.Fatalf("SaveResume(%q) code = %v, want %v", rawURL, status.Code(err), codes.InvalidArgument)
		}
	}

	resumes := &fakeResumesModel{
		findByOSSURLFn: func(context.Context, string) (*model.Resume, error) {
			return &model.Resume{ID: 5, UserID: 99, ParseState: "parsing"}, nil
		},
	}
	_, err := NewSaveResumeLogic(context.Background(), &svc.ServiceContext{
		ResumesModel: resumes,
	}).SaveResume(&pb.SaveResumeReq{UserId: 42, OssUrl: "https://resume.example/resume.pdf"})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("SaveResume() code = %v, want %v", status.Code(err), codes.AlreadyExists)
	}
}

func TestGetLatestResumeReturnsParsedData(t *testing.T) {
	resumes := &fakeResumesModel{
		findLatestFn: func(_ context.Context, userID uint64) (*model.Resume, error) {
			if userID != 42 {
				t.Fatalf("FindLatestByUserID() userID = %d, want 42", userID)
			}
			return &model.Resume{
				ID:         8,
				ParsedJSON: sql.NullString{String: `{"name":"Ada"}`, Valid: true},
				ParseState: "done",
			}, nil
		},
	}

	resp, err := NewGetLatestResumeLogic(context.Background(), &svc.ServiceContext{
		ResumesModel: resumes,
	}).GetLatestResume(&pb.GetLatestResumeReq{UserId: 42})
	if err != nil {
		t.Fatalf("GetLatestResume() error = %v", err)
	}
	if resp.ResumeId != 8 || resp.ParsedJson != `{"name":"Ada"}` || resp.ParseState != "done" {
		t.Fatalf("GetLatestResume() response = %+v", resp)
	}
}

func TestGetLatestResumeReturnsNotFound(t *testing.T) {
	resumes := &fakeResumesModel{
		findLatestFn: func(context.Context, uint64) (*model.Resume, error) {
			return nil, model.ErrNotFound
		},
	}

	_, err := NewGetLatestResumeLogic(context.Background(), &svc.ServiceContext{
		ResumesModel: resumes,
	}).GetLatestResume(&pb.GetLatestResumeReq{UserId: 42})
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetLatestResume() code = %v, want %v", status.Code(err), codes.NotFound)
	}
}

func TestSaveJdTrimsAndInsertsContent(t *testing.T) {
	var inserted *model.JD
	jds := &fakeJDsModel{
		insertFn: func(_ context.Context, jd *model.JD) (sql.Result, error) {
			inserted = jd
			return fakeResult{id: 17}, nil
		},
	}

	resp, err := NewSaveJdLogic(context.Background(), &svc.ServiceContext{JDsModel: jds}).SaveJd(&pb.SaveJdReq{
		UserId:  42,
		Title:   "  Go 工程师  ",
		Content: "\n负责 API 开发\n",
	})
	if err != nil {
		t.Fatalf("SaveJd() error = %v", err)
	}
	if resp.JdId != 17 || inserted == nil || inserted.UserID != 42 || inserted.Title != "Go 工程师" ||
		inserted.Content != "负责 API 开发" {
		t.Fatalf("SaveJd() response = %+v, inserted = %+v", resp, inserted)
	}
}

func TestSaveJdRejectsInvalidInputAndDatabaseErrors(t *testing.T) {
	logic := NewSaveJdLogic(context.Background(), &svc.ServiceContext{})
	_, err := logic.SaveJd(&pb.SaveJdReq{UserId: 42, Title: " ", Content: "content"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("SaveJd() invalid code = %v, want %v", status.Code(err), codes.InvalidArgument)
	}

	jds := &fakeJDsModel{
		insertFn: func(context.Context, *model.JD) (sql.Result, error) {
			return nil, errors.New("database unavailable")
		},
	}
	_, err = NewSaveJdLogic(context.Background(), &svc.ServiceContext{JDsModel: jds}).SaveJd(&pb.SaveJdReq{
		UserId:  42,
		Title:   "Go 工程师",
		Content: "content",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("SaveJd() database code = %v, want %v", status.Code(err), codes.Internal)
	}
}
