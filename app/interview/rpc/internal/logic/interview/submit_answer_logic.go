package interviewlogic

import (
	"context"

	"github.com/rotbit/whetstone/app/interview/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/interview/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SubmitAnswerLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSubmitAnswerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SubmitAnswerLogic {
	return &SubmitAnswerLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 提交候选人回答 → 引擎决策：追问 / 下一题 / 结束
func (l *SubmitAnswerLogic) SubmitAnswer(in *pb.SubmitAnswerReq) (*pb.SubmitAnswerResp, error) {
	// todo: add your logic here and delete this line

	return &pb.SubmitAnswerResp{}, nil
}
