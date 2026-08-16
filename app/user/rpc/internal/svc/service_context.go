package svc

import (
	"context"
	"database/sql"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// UsersModel 只暴露用户 RPC Logic 实际使用的方法，便于单元测试注入 Fake。
type UsersModel interface {
	Insert(ctx context.Context, data *model.Users) (sql.Result, error)
	FindOneByPhone(ctx context.Context, phone string) (*model.Users, error)
}

// ResumesModel 只暴露保存、按用户查最新记录和按 OSS URL 查幂等记录的方法。
type ResumesModel interface {
	Insert(ctx context.Context, data *model.Resume) (sql.Result, error)
	FindLatestByUserID(ctx context.Context, userID uint64) (*model.Resume, error)
	FindOneByOSSURL(ctx context.Context, ossURL string) (*model.Resume, error)
}

// JDsModel 只暴露当前保存 JD 所需的写入方法。
type JDsModel interface {
	Insert(ctx context.Context, data *model.JD) (sql.Result, error)
}

// ServiceContext 保存 user-rpc 的配置和数据访问依赖。
type ServiceContext struct {
	Config       config.Config
	UsersModel   UsersModel
	ResumesModel ResumesModel
	JDsModel     JDsModel
}

// NewServiceContext 创建一个共享的 MySQL 连接，并由三个 Model 复用其连接池。
func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:       c,
		UsersModel:   model.NewUsersModel(conn),
		ResumesModel: model.NewResumesModel(conn),
		JDsModel:     model.NewJDsModel(conn),
	}
}
