package svc

import (
	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/config"
	"github.com/yourname/whetstone/app/apis/cmd/app-apis/internal/ws/conn"
	interviewclient "github.com/yourname/whetstone/app/interview/rpc/client/interview"
	questionclient "github.com/yourname/whetstone/app/question/rpc/client/question"
	userclient "github.com/yourname/whetstone/app/user/rpc/client/user"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	UserRpc       userclient.User
	InterviewRpc  interviewclient.Interview
	QuestionRpc   questionclient.Question
	WsConnections *conn.Manager
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:        c,
		UserRpc:       userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		InterviewRpc:  interviewclient.NewInterview(zrpc.MustNewClient(c.InterviewRpc)),
		QuestionRpc:   questionclient.NewQuestion(zrpc.MustNewClient(c.QuestionRpc)),
		WsConnections: conn.NewManager(),
	}
}
