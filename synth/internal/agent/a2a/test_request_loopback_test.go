package a2a

import (
	"io"
	"net/http"
	"net/http/httptest"
)

// newLoopbackRequest 构造 loopback 来源的请求。A2A 默认认证器只放行 loopback；
// 生产远程访问必须显式 WithAuth。测试用此 helper 避免依赖 httptest 默认的
// 192.0.2.1 文档地址。
func newLoopbackRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.RemoteAddr = "127.0.0.1:12345"
	return req
}
