package accessLogger

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
)

/**
   @author：biguanqun
   @since： 2023/10/14
   @desc：
**/

type MiddlewareBuilder struct {
	allowReqBody  *atomic.Bool
	allowRespBody bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
}

func NewmiddlewareBuilder(fn func(ctx context.Context, al *AccessLog)) *MiddlewareBuilder {
	return &MiddlewareBuilder{
		loggerFunc:   fn,
		allowReqBody: atomic.NewBool(false),
	}
}

type AccessLog struct {
	Method   string // http 请求的方法
	URL      string // 整个请求的 url
	Duration string
	ReqBody  string
	RespBody string
	Status   int
}

func (b *MiddlewareBuilder) AllowReqBody(ok bool) *MiddlewareBuilder {
	b.allowReqBody.Store(ok)
	return b
}

func (b *MiddlewareBuilder) AllowRespBody() *MiddlewareBuilder {
	b.allowRespBody = true
	return b
}

func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		url := ctx.Request.URL.String()
		al := &AccessLog{
			Method: ctx.Request.Method,
			URL:    url,
		}
		if b.allowReqBody.Load() && ctx.Request.Body != nil {
			body, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			al.ReqBody = string(body) // 这其实是一个很消耗 CPU 和 内存的操作,因为会引起复制
		}

		if b.allowRespBody {
			ctx.Writer = respWriter{
				al:             al,
				ResponseWriter: ctx.Writer,
			}
		}

		defer func() {
			al.Duration = time.Now().Sub(start).String()
			// al.Duration = time.Since(start)  同上一样效果
			b.loggerFunc(ctx, al)
		}()

		// 执行到业务逻辑
		ctx.Next()
	}
}

type respWriter struct {
	al *AccessLog
	gin.ResponseWriter
}

func (r respWriter) Write(data []byte) (int, error) {
	r.al.RespBody = string(data)
	return r.ResponseWriter.Write(data)
}

func (r respWriter) WriteHeader(statusCode int) {
	r.al.Status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r respWriter) WriteString(data string) (int, error) {
	r.al.RespBody = data
	return r.ResponseWriter.WriteString(data)
}
