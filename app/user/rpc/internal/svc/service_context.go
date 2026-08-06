package svc

import (
	"context"
	"database/sql"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type UsersModel interface {
	Insert(ctx context.Context, data *model.Users) (sql.Result, error)
}

type ServiceContext struct {
	Config     config.Config
	UsersModel UsersModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		UsersModel: model.NewUsersModel(sqlx.NewMysql(c.Mysql.DataSource)),
	}
}
