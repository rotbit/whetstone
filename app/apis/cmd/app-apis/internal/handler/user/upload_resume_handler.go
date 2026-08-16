package user

import (
	"errors"
	"net/http"

	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/logic/user"
	"github.com/rotbit/whetstone/app/apis/cmd/app-apis/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// maxResumeFileSize 是业务允许的 PDF 文件上限，不包含 multipart 边界和字段头。
	maxResumeFileSize = int64(10 << 20)
	// maxResumeRequestSize 为 multipart 元数据预留 1 MiB，防止大请求在解析前耗尽内存或磁盘。
	maxResumeRequestSize = int64(11 << 20)
	// multipartMemory 表示最多在内存中保留 1 MiB 文件数据，超出部分由标准库写入临时文件。
	multipartMemory = int64(1 << 20)
)

// UploadResumeHandler 解析 multipart/form-data 请求，并把 file 字段交给上传业务逻辑。
// 该接口无法使用 goctl 的普通 JSON 参数解析，因此必须在 handler 中直接读取 r.MultipartForm。
func UploadResumeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 在解析 multipart 前限制整个请求体，避免攻击者通过额外字段绕过单文件大小检查。
		r.Body = http.MaxBytesReader(w, r.Body, maxResumeRequestSize)
		parseErr := r.ParseMultipartForm(multipartMemory)
		if r.MultipartForm != nil {
			// ParseMultipartForm 可能创建临时文件，必须在请求结束时主动清理。
			defer r.MultipartForm.RemoveAll()
		}
		if parseErr != nil {
			writeMultipartError(r, w, parseErr)
			return
		}

		if r.MultipartForm == nil {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "缺少 PDF 文件"))
			return
		}
		// 接口约定文件字段名固定为 file；同名多文件只处理第一份。
		fileHeaders := r.MultipartForm.File["file"]
		if len(fileHeaders) == 0 {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "缺少 PDF 文件"))
			return
		}
		header := fileHeaders[0]
		// 先根据 FileHeader 快速拒绝超大文件，Logic 层还会做同样的最终边界校验。
		if header.Size > maxResumeFileSize {
			http.Error(w, "PDF 文件不能超过 10 MiB", http.StatusRequestEntityTooLarge)
			return
		}
		file, err := header.Open()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "无法读取 PDF 文件"))
			return
		}
		defer file.Close()

		// multipart.File 同时实现 io.Reader 和 io.Seeker，便于 Logic 校验 PDF 头后回到文件起点上传。
		l := user.NewUploadResumeLogic(r.Context(), svcCtx)
		resp, err := l.UploadResume(file, header.Size)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

// writeMultipartError 将请求体过大映射为 413，其余 multipart 解析错误统一映射为参数错误。
func writeMultipartError(r *http.Request, w http.ResponseWriter, err error) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		http.Error(w, "PDF 文件不能超过 10 MiB", http.StatusRequestEntityTooLarge)
		return
	}
	httpx.ErrorCtx(r.Context(), w, status.Error(codes.InvalidArgument, "multipart/form-data 格式不正确"))
}
