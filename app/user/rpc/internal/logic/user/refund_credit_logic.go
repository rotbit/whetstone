package userlogic

import (
	"context"

	"github.com/yourname/whetstone/app/user/rpc/internal/svc"
	"github.com/yourname/whetstone/app/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefundCreditLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefundCreditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefundCreditLogic {
	return &RefundCreditLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 面试异常中断时回补次数（同样以 session_id 幂等）
func (l *RefundCreditLogic) RefundCredit(in *pb.RefundCreditReq) (*pb.RefundCreditResp, error) {
	// todo: add your logic here and delete this line

	return &pb.RefundCreditResp{}, nil
}
