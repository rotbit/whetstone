package interviewlogic

import (
	"context"

	"github.com/rotbit/whetstone/app/interview/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/interview/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSessionStateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetSessionStateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSessionStateLogic {
	return &GetSessionStateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// ws 断线重连时恢复会话状态
func (l *GetSessionStateLogic) GetSessionState(in *pb.GetSessionStateReq) (*pb.SessionState, error) {
	// todo: add your logic here and delete this line

	return &pb.SessionState{}, nil
}
