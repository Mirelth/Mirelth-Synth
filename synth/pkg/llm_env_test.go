package synth

import (
	"strings"
	"testing"
)

// TestProviderFromEnv_NoModel 未设置 SYNTH_LLM_MODEL → 明确错误（区分未配置）。
func TestProviderFromEnv_NoModel(t *testing.T) {
	t.Setenv("SYNTH_LLM_PROVIDER", "openai")
	t.Setenv("SYNTH_LLM_MODEL", "")
	t.Setenv("SYNTH_LLM_API_KEY", "")
	_, err := ProviderFromEnv()
	if err == nil {
		t.Fatal("未设置模型应报错")
	}
	if !strings.Contains(err.Error(), "SYNTH_LLM_MODEL") {
		t.Errorf("错误应提示 SYNTH_LLM_MODEL，got %v", err)
	}
}

// TestProviderFromEnv_DefaultOpenAI 默认 provider=openai，模型名生效。
func TestProviderFromEnv_DefaultOpenAI(t *testing.T) {
	t.Setenv("SYNTH_LLM_MODEL", "gpt-4o-mini")
	t.Setenv("SYNTH_LLM_API_KEY", "sk-test")
	p, err := ProviderFromEnv()
	if err != nil {
		t.Fatalf("ProviderFromEnv: %v", err)
	}
	if p == nil {
		t.Fatal("provider 不应为 nil")
	}
	if got := p.Info().Name; got != "gpt-4o-mini" {
		t.Errorf("model = %q, want gpt-4o-mini", got)
	}
}

// TestProviderFromEnv_Ollama 本地服务无需 Key。
func TestProviderFromEnv_Ollama(t *testing.T) {
	t.Setenv("SYNTH_LLM_PROVIDER", "ollama")
	t.Setenv("SYNTH_LLM_MODEL", "qwen3:8b")
	t.Setenv("SYNTH_LLM_API_KEY", "")
	p, err := ProviderFromEnv()
	if err != nil {
		t.Fatalf("ProviderFromEnv: %v", err)
	}
	if got := p.Info().Name; got != "qwen3:8b" {
		t.Errorf("model = %q, want qwen3:8b", got)
	}
}

// TestProviderFromEnv_Unknown 未知 provider → 报错并列出支持项。
func TestProviderFromEnv_Unknown(t *testing.T) {
	t.Setenv("SYNTH_LLM_PROVIDER", "skynet")
	t.Setenv("SYNTH_LLM_MODEL", "x")
	_, err := ProviderFromEnv()
	if err == nil {
		t.Fatal("未知 provider 应报错")
	}
	if !strings.Contains(err.Error(), "skynet") {
		t.Errorf("错误应包含 provider 名，got %v", err)
	}
}
