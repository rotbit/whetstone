package svc

import (
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/config"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/ws/conn"
	interviewclient "github.com/rotbit/whetstone/app/interview/rpc/client/interview"
	questionclient "github.com/rotbit/whetstone/app/question/rpc/client/question"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"
	userpb "github.com/rotbit/whetstone/app/user/rpc/pb"

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
	zrpc.DontLogClientContentForMethod(userpb.User_Register_FullMethodName)
	return &ServiceContext{
		Config:        c,
		UserRpc:       userclient.NewUser(zrpc.MustNewClient(c.UserRpc)),
		InterviewRpc:  interviewclient.NewInterview(zrpc.MustNewClient(c.InterviewRpc)),
		QuestionRpc:   questionclient.NewQuestion(zrpc.MustNewClient(c.QuestionRpc)),
		WsConnections: conn.NewManager(),
	}
}
