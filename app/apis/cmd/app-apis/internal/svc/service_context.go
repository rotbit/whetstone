package svc

import (
	"os"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/config"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/storage"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/ws/conn"
	interviewclient "github.com/rotbit/whetstone/app/interview/rpc/client/interview"
	questionclient "github.com/rotbit/whetstone/app/question/rpc/client/question"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"
	userpb "github.com/rotbit/whetstone/app/user/rpc/pb"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	UserRpc       userclient.User
	InterviewRpc  interviewclient.Interview
	QuestionRpc   questionclient.Question
	WsConnections *conn.Manager
	ObjectStorage storage.ObjectStorage
}

// NewServiceContext 注入外部依赖
// OSS配置通过环境变量直接读取，避免 go-zero 嵌套结构体解析问题。
func NewServiceContext(c config.Config) *ServiceContext {
	// 注册请求含明文密码， 禁止 zRPC 客户端日志记录该方法的请求正文
	zrpc.DontLogClientContentForMethod(userpb.User_Register_FullMethodName)

	// 从环境变量读取 OSS 配置
	objectStorage, err := storage.NewOSSStorage(storage.OSSConfig{
		Region:          os.Getenv("OSS_REGION"),
		Endpoint:        os.Getenv("OSS_ENDPOINT"),
		Bucket:          os.Getenv("OSS_BUCKET"),
		AccessKeyID:     os.Getenv("OSS_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("OSS_ACCESS_KEY_SECRET"),
		ObjectURLPrefix: os.Getenv("OSS_OBJECT_URL_PREFIX"),
	})
	logx.Must(err)
	return &ServiceContext{
		Config:        c,
		UserRpc:       userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		InterviewRpc:  interviewclient.NewInterview(zrpc.MustNewClient(c.InterviewRpc)),
		QuestionRpc:   questionclient.NewQuestion(zrpc.MustNewClient(c.QuestionRpc)),
		WsConnections: conn.NewManager(),
		ObjectStorage: objectStorage,
	}
}
