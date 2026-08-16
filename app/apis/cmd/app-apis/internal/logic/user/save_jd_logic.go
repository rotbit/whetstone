package user

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/types"
	userclient "github.com/rotbit/whetstone/app/user/rpc/client/user"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// maxJDTitleRunes 按 Unicode 字符数限制标题，与 MySQL VARCHAR(128) 的字符语义保持一致。
	maxJDTitleRunes = 128
	// maxJDContentBytes 按 UTF-8 字节数限制正文，确保写入 MySQL TEXT 的 65,535 字节上限以内。
	maxJDContentBytes = 60 * 1024
)

// SaveJdLogic 负责校验并保存当前登录用户提交的目标岗位 JD。
type SaveJdLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSaveJdLogic 创建带请求上下文日志的 JD 保存逻辑。
func NewSaveJdLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SaveJdLogic {
	return &SaveJdLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// SaveJd 在 API 层快速拒绝无效请求，再把规范化后的内容交给 user-rpc 持久化。
// RPC 层仍会重复校验关键边界，避免内部调用绕过 API 校验。
func (l *SaveJdLogic) SaveJd(req *types.SaveJdReq) (resp *types.SaveJdResp, err error) {
	userID, err := userIDFromContext(l.ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "JD 内容不能为空")
	}

	// 入库前去除首尾空白，避免只包含空格的标题或正文占用数据库记录。
	title := strings.TrimSpace(req.Title)
	if title == "" || utf8.RuneCountInString(title) > maxJDTitleRunes {
		return nil, status.Error(codes.InvalidArgument, "JD 标题长度必须为 1 到 128 个字符")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" || len(content) > maxJDContentBytes {
		return nil, status.Error(codes.InvalidArgument, "JD 内容不能为空且不能超过 60 KiB")
	}

	saved, err := l.svcCtx.UserRpc.SaveJd(l.ctx, &userclient.SaveJdReq{
		UserId:  userID,
		Title:   title,
		Content: content,
	})
	if err != nil {
		return nil, err
	}
	return &types.SaveJdResp{JdId: saved.JdId}, nil
}
