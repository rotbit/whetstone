package questionlogic

import (
	"context"

	"github.com/yourname/whetstone/app/question/rpc/internal/svc"
	"github.com/yourname/whetstone/app/question/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateFollowupLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGenerateFollowupLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateFollowupLogic {
	return &GenerateFollowupLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 基于候选人回答生成追问（项目深挖不走题库，LLM 现场生成）
func (l *GenerateFollowupLogic) GenerateFollowup(in *pb.GenerateFollowupReq) (*pb.GenerateFollowupResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GenerateFollowupResp{}, nil
}
