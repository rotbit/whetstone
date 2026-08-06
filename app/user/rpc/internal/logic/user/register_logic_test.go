package userlogic

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type fakeUsersModel struct {
	insert func(ctx context.Context, data *model.Users) (sql.Result, error)
}

func (m *fakeUsersModel) Insert(ctx context.Context, data *model.Users) (sql.Result, error) {
	return m.insert(ctx, data)
}

type fakeResult struct {
	id  int64
	err error
}

func (r fakeResult) LastInsertId() (int64, error) {
	return r.id, r.err
}

func (r fakeResult) RowsAffected() (int64, error) {
	return 1, nil
}

func TestRegisterSuccess(t *testing.T) {
	var inserted *model.Users
	svcCtx := &svc.ServiceContext{
		UsersModel: &fakeUsersModel{
			insert: func(_ context.Context, data *model.Users) (sql.Result, error) {
				inserted = data
				return fakeResult{id: 42}, nil
			},
		},
	}

	resp, err := NewRegisterLogic(context.Background(), svcCtx).Register(&pb.RegisterReq{
		Phone:    " 13800138000 ",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if resp.UserId != 42 || resp.Phone != "13800138000" || resp.Plan != defaultPlan {
		t.Fatalf("Register() response = %+v", resp)
	}
	if inserted == nil || inserted.Phone != resp.Phone || inserted.Plan != defaultPlan {
		t.Fatalf("inserted user = %+v", inserted)
	}
	if inserted.Password == "password123" {
		t.Fatal("password was stored as plain text")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(inserted.Password), []byte("password123")); err != nil {
		t.Fatalf("stored password hash is invalid: %v", err)
	}
}

func TestRegisterRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name string
		req  *pb.RegisterReq
	}{
		{name: "invalid phone", req: &pb.RegisterReq{Phone: "abc", Password: "password123"}},
		{name: "short password", req: &pb.RegisterReq{Phone: "13800138000", Password: "short"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			svcCtx := &svc.ServiceContext{
				UsersModel: &fakeUsersModel{
					insert: func(context.Context, *model.Users) (sql.Result, error) {
						t.Fatal("Insert() should not be called")
						return nil, nil
					},
				},
			}

			_, err := NewRegisterLogic(context.Background(), svcCtx).Register(test.req)
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("Register() code = %v, want %v", status.Code(err), codes.InvalidArgument)
			}
		})
	}
}

func TestRegisterReturnsAlreadyExists(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		UsersModel: &fakeUsersModel{
			insert: func(context.Context, *model.Users) (sql.Result, error) {
				return nil, &mysql.MySQLError{Number: 1062, Message: "duplicate phone"}
			},
		},
	}

	_, err := NewRegisterLogic(context.Background(), svcCtx).Register(&pb.RegisterReq{
		Phone:    "13800138000",
		Password: "password123",
	})
	if status.Code(err) != codes.AlreadyExists {
		t.Fatalf("Register() code = %v, want %v", status.Code(err), codes.AlreadyExists)
	}
}

func TestRegisterReturnsInternalForInsertFailure(t *testing.T) {
	svcCtx := &svc.ServiceContext{
		UsersModel: &fakeUsersModel{
			insert: func(context.Context, *model.Users) (sql.Result, error) {
				return nil, errors.New("database unavailable")
			},
		},
	}

	_, err := NewRegisterLogic(context.Background(), svcCtx).Register(&pb.RegisterReq{
		Phone:    "13800138000",
		Password: "password123",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("Register() code = %v, want %v", status.Code(err), codes.Internal)
	}
}
