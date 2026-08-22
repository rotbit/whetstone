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
	FindOneByPhone(ctx context.Context, phone string) (*model.Users, error)
}

// ResumesModel 只暴露 SaveResume / 解析任务 / GetResume 实际需要的数据库方法。
// logic 新增数据库操作时，必须在此同步声明后编译期才会通过。
type ResumesModel interface {
	Insert(ctx context.Context, data *model.Resumes) (sql.Result, error)
	FindOneByOSSURL(ctx context.Context, ossUrl string) (*model.Resumes, error)
	FindLatestByUserID(ctx context.Context, userID uint64) (*model.Resumes, error)
}

type ServiceContext struct {
	Config       config.Config
	UsersModel   UsersModel
	ResumesModel ResumesModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.Mysql.DataSource)
	return &ServiceContext{
		Config:       c,
		UsersModel:   model.NewUsersModel(conn),
		ResumesModel: model.NewResumesModel(conn),
	}
}
