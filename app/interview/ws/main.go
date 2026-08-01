// ws-gateway：实时面试长连接入口（骨架）
//
// 职责（docs/技术方案.md §6.1）：
//   - WebSocket 连接生命周期：心跳、断线重连、会话恢复
//   - M1 文字模式：转发候选人回答 → interview-rpc → 流式下发面试官回应
//   - M2 语音模式：上行音频 → ASR 流式 → 引擎 → LLM 流式 → TTS 流式下行，支持打断
//
// TODO(M1):
//  1. 引入 github.com/coder/websocket 完成协议升级
//  2. 接入 conn.Manager 管理连接与心跳
//  3. 校验 JWT（复用 user-api 的 AccessSecret）
//  4. 通过 zrpc 调用 interview-rpc（SubmitAnswer），LLM 流式回复经 pipeline 下发
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/yourname/whetstone/app/interview/ws/internal/conn"
)

var addr = flag.String("addr", ":8890", "ws-gateway listen address")

func main() {
	flag.Parse()

	mgr := conn.NewManager()
	_ = mgr // TODO: 注入 handler

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	log.Printf("ws-gateway listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: websocket.Accept(w, r, ...) 完成升级，
	// 随后进入读写循环：客户端消息 → pipeline → 流式回写。
	http.Error(w, "TODO: websocket upgrade, see docs/技术方案.md §6.1", http.StatusNotImplemented)
}
