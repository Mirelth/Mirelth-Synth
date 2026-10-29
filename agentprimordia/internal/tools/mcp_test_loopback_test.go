package tools

import (
	"io"
	"net/http"
	"net/http/httptest"
)

// newLoopbackRequest 构造 loopback 来源请求。MCP 默认无 APIKey 时只接受
// loopback；远程部署必须显式配置 APIKey。
func newLoopbackRequest(method, target string, body io.Reader) *http.Request {
	req := httptest.NewRequest(method, target, body)
	req.RemoteAddr = "127.0.0.1:12345"
	return req
}
