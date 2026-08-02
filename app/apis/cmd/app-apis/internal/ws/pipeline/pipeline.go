// Package pipeline 编排 app-apis 中一轮问答的流式数据通路。
//
// M1 文字模式：
//
//	候选人文本 → interview-rpc.SubmitAnswer → llmadapter.ChatStream
//	→ token 逐段回写客户端（打字机效果）
//
// M2 语音模式（技术含金量最高，docs/技术方案.md §6.1）：
//
//	上行音频帧 → ASR 流式（partial 字幕回显）→ final 文本进引擎
//	→ LLM 流式生成 → 按句切分 → TTS 流式合成 → 音频帧下行
//	打断（barge-in）：上行 VAD 检测到说话 → 停止下行播放
//	→ context.CancelFunc 级联取消 LLM/TTS 调用，防 goroutine 泄漏
package pipeline

import "context"

// Chunk 下行流式片段。
type Chunk struct {
	Type    string // text | audio | subtitle | event
	Payload []byte
	Done    bool
}

// Round 处理一轮问答，返回只读流；调用方 range 读取并回写 WebSocket。
// ctx 取消（客户端断连 / 打断）时，实现方必须立即停止上游调用并关闭 channel。
func Round(ctx context.Context, sessionId string, answer string) (<-chan Chunk, error) {
	// TODO(M1):
	//  1. zrpc 调 interview-rpc.SubmitAnswer 拿到决策与回应模板
	//  2. llmadapter.ChatStream 流式生成，token 写入 channel
	//  3. select { case <-ctx.Done(): ... } 保证取消传播
	ch := make(chan Chunk)
	close(ch)
	return ch, nil
}
