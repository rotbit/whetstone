package interviewlogic

import (
	"context"

	"github.com/rotbit/whetstone/app/interview/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/interview/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type StartSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewStartSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *StartSessionLogic {
	return &StartSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 初始化状态机，返回开场白与第一题
func (l *StartSessionLogic) StartSession(in *pb.StartSessionReq) (*pb.StartSessionResp, error) {
	// todo: add your logic here and delete this line

	return &pb.StartSessionResp{}, nil
}
