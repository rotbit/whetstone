package user

import (
	"context"
	"errors"
)

// ErrNoUserID 表示上下文中没有找到用户 ID，或 ID 不合法。
// 通常由 JWT 中间件未正确注入导致。
var ErrNoUserID = errors.New("用户未登录")

// userIDFromContext 从 Context 中提取经过 JWT 认证的用户 ID。
// 如果缺失或类型不正确，返回 ErrNoUserID，由上层决定如何处理（如返回 401）。
func userIDFromContext(ctx context.Context) (int64, error) {
	// 1. 从上下文中取出键为 "user_id" 的值。
	//    这个值由 JWT 中间件在验证 Token 后注入。
	v := ctx.Value("user_id")

	// 2. 如果值为 nil，说明上下文中没有用户信息（可能是未登录或中间件未生效）。
	if v == nil {
		return 0, ErrNoUserID
	}

	// 3. 类型断言：尝试将值转换为 int64。
	//    JWT 中间件存入时通常是 int64，但为了安全需做类型检查。
	id, ok := v.(int64)

	// 4. 如果断言失败（不是 int64）或 ID <= 0（无效值），视为未登录。
	if !ok || id <= 0 {
		return 0, ErrNoUserID
	}

	// 5. 验证通过，返回用户 ID。
	return id, nil
}
