package questionlogic

import (
	"context"

	"github.com/rotbit/whetstone/app/question/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/question/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PickQuestionsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPickQuestionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PickQuestionsLogic {
	return &PickQuestionsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 按岗位 + 简历/JD 标签混合检索选题（向量 + 标签过滤 + rerank）
func (l *PickQuestionsLogic) PickQuestions(in *pb.PickQuestionsReq) (*pb.PickQuestionsResp, error) {
	// todo: add your logic here and delete this line

	return &pb.PickQuestionsResp{}, nil
}
