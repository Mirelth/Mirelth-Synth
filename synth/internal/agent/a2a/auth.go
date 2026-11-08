package a2a

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

// perf-v6 round 4 Task 2：a2a auth 静态错误
var (
	ErrAuthHeaderMissing  = errors.New("missing Authorization header")
	ErrAuthBearerRequired = errors.New("invalid Authorization format, Bearer token required")
)

// Principal 已认证主体
type Principal struct {
	ID       string            `json:"id"`
	Roles    []string          `json:"roles"`
	Scopes   []string          `json:"scopes"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (p *Principal) HasScope(scope string) bool {
	for _, s := range p.Scopes {
		if s == scope || s == "*" {
			return true
		}
	}
	return false
}

func (p *Principal) HasRole(role string) bool {
	for _, r := range p.Roles {
		if r == role || r == "*" {
			return true
		}
	}
	return false
}

// Authenticator 认证器接口
type Authenticator interface {
	Authenticate(r *http.Request) (*Principal, error)
}

// ErrRemoteAuthRequired 表示非 loopback 请求未配置认证。
var ErrRemoteAuthRequired = errors.New("a2a: remote access requires an explicit authenticator")

// LoopbackOnlyAuthenticator 是 A2A Server 的默认认证器：
// loopback/进程内请求放行，非 loopback 请求拒绝。
// 生产环境对外服务必须显式 WithAuth(APIKey/Bearer)；本地开发无需额外配置。
//
// 注意：判定依据是 r.RemoteAddr。若服务挂在同机反向代理（nginx 等）后面，
// 所有外部请求的来源都会显示为 127.0.0.1，本认证器会被绕过；该部署形态
// 必须配置显式认证器，不能依赖 loopback 判定。
type LoopbackOnlyAuthenticator struct{}

func NewLoopbackOnlyAuthenticator() *LoopbackOnlyAuthenticator { return &LoopbackOnlyAuthenticator{} }

func (l *LoopbackOnlyAuthenticator) Authenticate(r *http.Request) (*Principal, error) {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// 直接构造的 Request（无 RemoteAddr）按进程内调用处理；真实网络请求
		// 一定有 RemoteAddr。
		if r.RemoteAddr == "" {
			return &Principal{ID: "loopback", Roles: []string{"admin"}, Scopes: []string{"*"}}, nil
		}
		return nil, ErrRemoteAuthRequired
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return &Principal{ID: "loopback", Roles: []string{"admin"}, Scopes: []string{"*"}}, nil
	}
	return nil, ErrRemoteAuthRequired
}

// NoopAuthenticator 跳过认证（开发/测试用）
type NoopAuthenticator struct{}

func NewNoopAuthenticator() *NoopAuthenticator { return &NoopAuthenticator{} }

func (n *NoopAuthenticator) Authenticate(_ *http.Request) (*Principal, error) {
	return &Principal{ID: "noop-user", Roles: []string{"admin"}, Scopes: []string{"*"}}, nil
}

// APIKeyAuthenticator API Key 认证
type APIKeyAuthenticator struct {
	keys   map[string]string // key → principalID
	header string            // Header 名称
}

func NewAPIKeyAuthenticator(keys map[string]string, headerName string) *APIKeyAuthenticator {
	if headerName == "" {
		headerName = "X-API-Key"
	}
	return &APIKeyAuthenticator{keys: keys, header: headerName}
}

func (a *APIKeyAuthenticator) Authenticate(r *http.Request) (*Principal, error) {
	key := r.Header.Get(a.header)
	if key == "" {
		return nil, fmt.Errorf("missing auth header: %s", a.header)
	}
	principalID, ok := a.keys[key]
	if !ok {
		return nil, errors.New("invalid API key") // perf-v6 Task G：静态文案用 errors.New
	}
	return &Principal{ID: principalID, Scopes: []string{"*"}}, nil
}

// BearerTokenValidator Token 验证函数类型
type BearerTokenValidator func(token string) (*Principal, error)

// BearerTokenAuthenticator JWT/OAuth2 Bearer Token 认证
type BearerTokenAuthenticator struct {
	validate BearerTokenValidator
}

func NewBearerTokenAuthenticator(validate BearerTokenValidator) *BearerTokenAuthenticator {
	return &BearerTokenAuthenticator{validate: validate}
}

func (b *BearerTokenAuthenticator) Authenticate(r *http.Request) (*Principal, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrAuthHeaderMissing // perf-v6 round 4 Task 2
	}
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, ErrAuthBearerRequired // perf-v6 round 4 Task 2
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	return b.validate(token)
}
