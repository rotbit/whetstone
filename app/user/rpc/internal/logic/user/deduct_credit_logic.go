package userlogic

import (
	"context"

	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeductCreditLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeductCreditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeductCreditLogic {
	return &DeductCreditLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 面试开始前扣减次数；session_id 作幂等键，重复调用不重复扣
func (l *DeductCreditLogic) DeductCredit(in *pb.DeductCreditReq) (*pb.DeductCreditResp, error) {
	// todo: add your logic here and delete this line

	return &pb.DeductCreditResp{}, nil
}
