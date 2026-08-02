package interviewlogic

import (
	"context"

	"github.com/rotbit/whetstone/app/interview/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/interview/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type EndSessionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEndSessionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EndSessionLogic {
	return &EndSessionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提前结束（触发报告任务入队）
func (l *EndSessionLogic) EndSession(in *pb.EndSessionReq) (*pb.EndSessionResp, error) {
	// todo: add your logic here and delete this line

	return &pb.EndSessionResp{}, nil
}
