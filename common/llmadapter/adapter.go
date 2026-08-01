// Package llmadapter 统一封装 LLM / ASR / TTS 多厂商调用 —— whetstone 的 mini AI 网关。
//
// 设计原则（docs/技术方案.md §8）：
//   - 业务层只依赖本包接口，换厂商零改动
//   - 一切皆流式：channel 输出 + context 取消传播
//   - 未来可独立成库开源（这是第二条开源产出线）
package llmadapter

import "context"

// Message 对话消息。
type Message struct {
	Role    string // system | user | assistant
	Content string
}

// ChatChunk LLM 流式输出片段。
type ChatChunk struct {
	Content string
	Done    bool
	Err     error
}

// LLM 大模型统一接口。
type LLM interface {
	// ChatStream 流式对话。实现方必须监听 ctx.Done() 并及时中断上游请求。
	ChatStream(ctx context.Context, model string, msgs []Message) (<-chan ChatChunk, error)
}

// ---- M2 语音接口（占位） ----

// AsrChunk 语音识别流式结果。
type AsrChunk struct {
	Text    string
	IsFinal bool // VAD 判定句子结束
	Err     error
}

// ASR 流式语音识别。
type ASR interface {
	Transcribe(ctx context.Context, audio <-chan []byte) (<-chan AsrChunk, error)
}

// TTS 流式语音合成（按句送入，边合成边输出音频帧）。
type TTS interface {
	Synthesize(ctx context.Context, sentences <-chan string) (<-chan []byte, error)
}

// TODO(M1): 新建 deepseek/ 子包实现 LLM 接口：
//   POST https://api.deepseek.com/chat/completions  stream=true
//   解析 SSE：data: {...} 逐行读取，[DONE] 结束；
//   注意 http.Client 超时体系与连接复用（面试考点 §12）。
