package userlogic

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/rotbit/whetstone/app/user/model"
	"github.com/rotbit/whetstone/app/user/rpc/internal/svc"
	"github.com/rotbit/whetstone/app/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	maxJDTitleRunes   = 128
	maxJDContentBytes = 60 * 1024
)

type SaveJdLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSaveJdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveJdLogic {
	return &SaveJdLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 保存一份目标岗位 JD。
func (l *SaveJdLogic) SaveJd(in *pb.SaveJdReq) (*pb.SaveJdResp, error) {
	if in.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "用户 ID 不正确")
	}

	title := strings.TrimSpace(in.GetTitle())
	if title == "" || utf8.RuneCountInString(title) > maxJDTitleRunes {
		return nil, status.Error(codes.InvalidArgument, "JD 标题长度必须为 1 到 128 个字符")
	}

	content := strings.TrimSpace(in.GetContent())
	if content == "" || len(content) > maxJDContentBytes {
		return nil, status.Error(codes.InvalidArgument, "JD 内容不能为空且不能超过 60 KiB")
	}

	result, err := l.svcCtx.JDsModel.Insert(l.ctx, &model.JD{
		UserID:  uint64(in.GetUserId()),
		Title:   title,
		Content: content,
	})
	if err != nil {
		l.Errorf("insert JD failed: %v", err)
		return nil, status.Error(codes.Internal, "保存 JD 失败")
	}

	jdID, err := result.LastInsertId()
	if err != nil {
		l.Errorf("get inserted JD id failed: %v", err)
		return nil, status.Error(codes.Internal, "保存 JD 失败")
	}
	return &pb.SaveJdResp{JdId: jdID}, nil
}
