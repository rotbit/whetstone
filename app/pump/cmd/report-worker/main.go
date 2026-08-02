// report-worker：复盘报告异步生成（骨架）。
//
// 流程（docs/技术方案.md §6.4）：
//
//	interview-rpc 在面试结束时投递任务 → 本进程消费 →
//	逐题评分（正确性/深度/STAR/表达，JSON Schema 输出 + 校验重试）→
//	聚合雷达图与改进建议 → 落库 → 通知客户端
//
// TODO(M1):
//  1. 引入 github.com/hibiken/asynq，注册 task "report:generate"
//  2. 幂等：session_id 唯一键，重复任务直接跳过
//  3. 失败重试（指数退避）+ 死信告警
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "report-worker")

	logger.Info("report-worker skeleton is running — no asynq handlers are registered yet")
	// TODO: srv := asynq.NewServer(...); mux := asynq.NewServeMux(); ...
	<-ctx.Done()
	logger.Info("report-worker shutting down")
}
