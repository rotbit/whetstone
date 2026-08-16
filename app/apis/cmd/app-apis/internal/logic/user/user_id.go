package user

import (
	"context"
	"encoding/json"
	"math"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// userIDFromContext 从 go-zero JWT 中间件写入的 uid claim 中提取正整数用户 ID。
// go-zero 当前使用 json.Number，但兼容 float64、int64 和 string，便于测试及不同 JWT 解析配置复用。
func userIDFromContext(ctx context.Context) (int64, error) {
	value := ctx.Value("uid")
	var (
		userID int64
		err    error
	)

	// 所有分支最终都归一化为 int64；浮点数必须是无小数部分且不溢出的正整数。
	switch typedValue := value.(type) {
	case json.Number:
		userID, err = typedValue.Int64()
	case float64:
		if typedValue <= 0 || typedValue > math.MaxInt64 || typedValue != math.Trunc(typedValue) {
			return 0, status.Error(codes.Unauthenticated, "用户身份无效")
		}
		userID = int64(typedValue)
	case int64:
		userID = typedValue
	case string:
		userID, err = strconv.ParseInt(typedValue, 10, 64)
	default:
		return 0, status.Error(codes.Unauthenticated, "用户身份无效")
	}

	if err != nil || userID <= 0 {
		return 0, status.Error(codes.Unauthenticated, "用户身份无效")
	}
	return userID, nil
}
