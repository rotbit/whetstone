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
	// 1. 从上下文中取出键为 "uid" 的值。
	//    这个值由 JWT 中间件在验证 Token 后注入，key 名与 user-rpc 签发 claim 时的字段名一致。
	v := ctx.Value("uid")

	// 2. 如果值为 nil，说明上下文中没有用户信息（可能是未登录或中间件未生效）。
	if v == nil {
		return 0, ErrNoUserID
	}

	// 3. JWT 是 JSON，数字 claim 经 jwt.MapClaims 解析回来固定为 float64，
	//    直接用 v.(int64) 必然断言失败。这里用 type switch 兼容常见数值类型。
	var id int64
	switch val := v.(type) {
	case int64:
		id = val
	case float64:
		id = int64(val)
	case int:
		id = int64(val)
	default:
		return 0, ErrNoUserID
	}

	// 4. ID <= 0 视为无效，视为未登录。
	if id <= 0 {
		return 0, ErrNoUserID
	}

	// 5. 验证通过，返回用户 ID。
	return id, nil
}
