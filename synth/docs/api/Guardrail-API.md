# Guardrail API

Guardrail 提供输入/输出护栏引擎，在 Agent 处理前后执行安全规则检查。

## 核心类型

```go
// Engine 护栏引擎，管理规则集并执行检查
type Engine struct { /* ... */ }

// Rule 规则接口
type Rule interface {
    Name() string
    Priority() int
    Check(ctx context.Context, input string, checkpoint CheckPoint) (*Result, error)
}

// CheckPoint 检查点
const (
    CheckInput  CheckPoint = "input"
    CheckOutput CheckPoint = "output"
)

// Action 检查动作
const (
    ActionPass     Action = "pass"
    ActionReject   Action = "reject"
    ActionSanitize Action = "sanitize"
    ActionFlag     Action = "flag"
)
```

## 使用方式

```go
import synth "github.com/Mirelth/Synth/v7/synth/pkg"

engine := synth.NewGuardrailEngine()
engine.AddRule(synth.NewPromptInjectionRule(synth.PromptInjectionConfig{})) // Prompt 注入检测
engine.AddRule(synth.NewPIIRule(synth.PIIRuleConfig{}))                     // PII 脱敏
engine.AddRule(synth.NewTopicConstraintRule(synth.TopicConstraintConfig{    // 主题过滤
    Topics: allowed,
}))

report, err := engine.Check(userInput, synth.CheckInput)
if err != nil {
    // 处理检查失败
}
if !report.Passed {
    // 拒绝或清理输入
}
```

## 内置规则

| 规则 | 优先级 | 说明 |
|------|--------|------|
| InjectionRule | 1000 | Prompt 注入检测 |
| PIIRule | 500 | 个人敏感信息脱敏 |
| OutputRule | 500 | 输出安全检查 |
| TopicRule | 100 | 主题约束过滤 |
| TrieRule | 100 | 基于 Trie 树的敏感词匹配 |

## 与 Agent 集成

通过 `WithInputGuard` 选项注入，用户输入进入 ReAct 循环前自动检查。`WithInputGuard` 接收 `InputGuard` 函数类型 `func(content string) (sanitized string, blocked bool, err error)`：

```go
agent, err := synth.NewAgent("my-agent", "You are a helpful assistant.", provider,
    synth.WithInputGuard(func(content string) (string, bool, error) {
        report, err := engine.Check(content, synth.CheckInput)
        if err != nil {
            return content, false, err
        }
        if !report.Passed {
            return content, true, nil // blocked=true 拒绝该输入
        }
        return content, false, nil
    }))
```
