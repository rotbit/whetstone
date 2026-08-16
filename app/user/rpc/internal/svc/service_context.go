package svc

import (
	"context"
	"database/sql"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UsersModel 只暴露用户 RPC Log
type UsersModel interface {
	Insert(ctx context.Context, data *model.Users) (sql.Result, error)
	FindOneByPhone(ctx context.Context, phone string) (*model.Users, error)
}

(*model.Users, error)
}

// R
type ResumesModel interface {
	Insert(ctx context.Context, data *model.Resume) (sql.Result, error)
	FindLatestByUserID(ctx context.Context, userID uint64) (*model.Resume, error)
ontext, userID uint64) (*model.Resume, error)
	FindOneByOSSURL(ctx context.
	FindOneByOSSURL(ctx context.Context, ossURL string) (*model.Resume, error)
}

。
type JDsModel interface
type JDsModel interface {
	Insert(ctx context.Context, data *model.JD) (sql.Result, error)
}

type ServiceContext struct {
	Config       config.Config
    config.Config
	Users
	UsersModel   UsersModel
	ResumesModel ResumesModel
	JDsModel     JDsModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:       c,
		UsersModel:   model.NewUsersModel(conn),
		ResumesModel: model.NewResumesModel(conn),
		JDsModel:     model.NewJDsModel(conn),
	}
}
