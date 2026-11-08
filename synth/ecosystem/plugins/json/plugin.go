package jsonplugin

import (
	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"
)

// Plugin 是 JSON/CSV 数据处理插件，封装 synth.JSONTool 和 synth.CSVTool
type Plugin struct {
	jsonTool *synth.JSONTool
	csvTool  *synth.CSVTool
}

// New 创建新的 JSON 插件实例
func New() *Plugin {
	return &Plugin{
		jsonTool: synth.NewJSONTool(),
		csvTool:  synth.NewCSVTool(),
	}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return "json" }

// Version 返回插件版本
func (p *Plugin) Version() string { return "0.1.0" }

// Tools 返回插件提供的工具列表（JSON 和 CSV 处理器）
func (p *Plugin) Tools() []synth.Tool {
	return []synth.Tool{p.jsonTool, p.csvTool}
}

// Init 初始化插件（JSON/CSV 工具无需额外配置）
func (p *Plugin) Init(config map[string]any) error { return nil }

// Close 关闭插件资源
func (p *Plugin) Close() error { return nil }
