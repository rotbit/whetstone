package svc

import (
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
// OSS配置错误通过 logx.Must 立即终止启动，避免服务看似健康但上传接口始终失败。
func NewServiceContext(c config.Config) *ServiceContext {
	// 注册请求含明文密码， 禁止 zRPC 客户端日志记录该方法的请求正文
	zrpc.DontLogClientContentForMethod(userpb.User_Register_FullMethodName)

	objectStorage, err := storage.NewOSSStorage(storage.OSSConfig{
		Region:          c.OSS.Region,
		Endpoint:        c.OSS.Endpoint,
		Bucket:          c.OSS.Bucket,
		AccessKeyID:     c.OSS.AccessKeyID,
		AccessKeySecret: c.OSS.AccessKeySecret,
		ObjectURLPrefix: c.OSS.ObjectURLPrefix,
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
