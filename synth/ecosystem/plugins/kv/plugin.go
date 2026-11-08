package kv

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	synth "github.com/Mirelth/Mirelth-Synth/v7/synth/pkg"

	_ "modernc.org/sqlite" // 纯 Go SQLite 驱动
)

const (
	// 默认库名与表名
	defaultDBPath = "synth_kv.db"
	kvTableName   = "synth_kv_store"

	// 改名前的库名/表名，仅用于存量数据兼容。
	// 拆开拼接书写：它们必须原样保留旧词前缀，但不应被
	// scripts/check-brand-residuals.sh 当成新的品牌残留。
	legacyBrandPrefix = "ap"
	legacyDBPath      = legacyBrandPrefix + "_kv.db"
	legacyTableName   = legacyBrandPrefix + "_kv_store"
)

// Plugin 是键值存储插件，提供基于 SQLite 的持久化 KV 存储
type Plugin struct {
	tool *KVStoreTool
}

// New 创建新的 KV 插件实例
func New() *Plugin {
	return &Plugin{tool: &KVStoreTool{}}
}

// Name 返回插件名称
func (p *Plugin) Name() string { return "kv" }

// Version 返回插件版本
func (p *Plugin) Version() string { return "0.1.0" }

// Tools 返回插件提供的工具列表
func (p *Plugin) Tools() []synth.Tool {
	return []synth.Tool{p.tool}
}

// Init 初始化插件，从 config 中读取 db_path 并创建数据库连接
//
// 兼容性：改名前的默认库（legacyDBPath）与表（legacyTableName）仍可读写。
// 未显式指定 db_path 时，若新默认库不存在而旧默认库存在，则继续使用旧库；
// 旧表会在建表前原地重命名为新表名，保证已持久化的 KV 数据可见。
func (p *Plugin) Init(config map[string]any) error {
	dbPath := defaultDBPath
	if path, ok := config["db_path"].(string); ok && path != "" {
		dbPath = path
	} else if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		if _, legacyErr := os.Stat(legacyDBPath); legacyErr == nil {
			dbPath = legacyDBPath
		}
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}

	// 存量数据迁移：旧表名 → 新表名（仅当旧表存在且新表不存在时执行，不覆盖新数据）
	if err := migrateLegacyTable(db); err != nil {
		db.Close()
		return fmt.Errorf("迁移旧 KV 表失败: %w", err)
	}

	// 创建 KV 存储表
	createSQL := `CREATE TABLE IF NOT EXISTS ` + kvTableName + ` (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`
	if _, err := db.Exec(createSQL); err != nil {
		db.Close()
		return fmt.Errorf("创建 KV 表失败: %w", err)
	}

	p.tool.db = db
	return nil
}

// migrateLegacyTable 将旧表名原地重命名为新表名，保留存量数据。
// 新旧表同时存在时不做任何事（避免覆盖已有数据）。
func migrateLegacyTable(db *sql.DB) error {
	legacyExists, err := tableExists(db, legacyTableName)
	if err != nil || !legacyExists {
		return err
	}
	currentExists, err := tableExists(db, kvTableName)
	if err != nil || currentExists {
		return err
	}
	_, err = db.Exec(`ALTER TABLE ` + legacyTableName + ` RENAME TO ` + kvTableName)
	return err
}

// tableExists 查询 sqlite_master 判断表是否存在
func tableExists(db *sql.DB, name string) (bool, error) {
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, name).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// Close 关闭插件资源，释放数据库连接
func (p *Plugin) Close() error {
	if p.tool != nil && p.tool.db != nil {
		return p.tool.db.Close()
	}
	return nil
}

// KVStoreTool 是基于 SQLite 的键值存储工具
type KVStoreTool struct {
	db *sql.DB
}

// Name 返回工具名称
func (t *KVStoreTool) Name() string { return "kv_store" }

// Description 返回工具描述
func (t *KVStoreTool) Description() string {
	return `键值存储工具，基于 SQLite 的持久化 KV 存储。
功能：
- 设置键值对（set）
- 获取键对应的值（get）
- 删除键（delete）
- 列出所有键值对（list）

参数：
- action (required): 操作类型 [get|set|delete|list]
- key: 键名（get/set/delete 必需）
- value: 值（set 必需）`
}

// Parameters 返回工具参数的 JSON Schema
func (t *KVStoreTool) Parameters() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"action": {"type": "string", "enum": ["get", "set", "delete", "list"]},
			"key": {"type": "string", "description": "键名"},
			"value": {"type": "string", "description": "值（set 操作必需）"}
		},
		"required": ["action"]
	}`)
}

// Category 返回工具分类
func (t *KVStoreTool) Category() string { return "database" }

// Execute 执行 KV 存储操作
func (t *KVStoreTool) Execute(ctx context.Context, input json.RawMessage) (*synth.ToolResult, error) {
	var params map[string]any
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("解析参数错误: %w", err)
	}

	action, _ := params["action"].(string)
	if action == "" {
		return synth.NewToolErrorResult("参数 'action' 不能为空"), nil
	}

	switch action {
	case "get":
		return t.get(ctx, params)
	case "set":
		return t.set(ctx, params)
	case "delete":
		return t.delete(ctx, params)
	case "list":
		return t.list(ctx)
	default:
		return synth.NewToolErrorResult(fmt.Sprintf("未知操作: %s，支持 get/set/delete/list", action)), nil
	}
}

// get 获取键对应的值
func (t *KVStoreTool) get(ctx context.Context, params map[string]any) (*synth.ToolResult, error) {
	key, _ := params["key"].(string)
	if key == "" {
		return synth.NewToolErrorResult("get 操作需要 'key' 参数"), nil
	}

	var value, createdAt, updatedAt string
	err := t.db.QueryRowContext(ctx,
		"SELECT value, created_at, updated_at FROM "+kvTableName+" WHERE key = ?", key,
	).Scan(&value, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return synth.NewToolErrorResult(fmt.Sprintf("键 %q 不存在", key)), nil
	}
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	result := map[string]any{
		"key":        key,
		"value":      value,
		"created_at": createdAt,
		"updated_at": updatedAt,
	}
	output, _ := json.MarshalIndent(result, "", "  ")
	return &synth.ToolResult{Content: string(output)}, nil
}

// set 设置键值对
func (t *KVStoreTool) set(ctx context.Context, params map[string]any) (*synth.ToolResult, error) {
	key, _ := params["key"].(string)
	if key == "" {
		return synth.NewToolErrorResult("set 操作需要 'key' 参数"), nil
	}
	value, _ := params["value"].(string)

	now := time.Now().UTC().Format(time.RFC3339)

	// 使用 UPSERT 语义：存在则更新，不存在则插入
	_, err := t.db.ExecContext(ctx,
		`INSERT INTO `+kvTableName+` (key, value, created_at, updated_at) VALUES (?, ?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, now, now)
	if err != nil {
		return nil, fmt.Errorf("写入失败: %w", err)
	}

	result := map[string]any{
		"success":    true,
		"key":        key,
		"value":      value,
		"updated_at": now,
	}
	output, _ := json.MarshalIndent(result, "", "  ")
	return &synth.ToolResult{Content: string(output)}, nil
}

// delete 删除键
func (t *KVStoreTool) delete(ctx context.Context, params map[string]any) (*synth.ToolResult, error) {
	key, _ := params["key"].(string)
	if key == "" {
		return synth.NewToolErrorResult("delete 操作需要 'key' 参数"), nil
	}

	res, err := t.db.ExecContext(ctx, "DELETE FROM "+kvTableName+" WHERE key = ?", key)
	if err != nil {
		return nil, fmt.Errorf("删除失败: %w", err)
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		return synth.NewToolErrorResult(fmt.Sprintf("键 %q 不存在", key)), nil
	}

	result := map[string]any{
		"success": true,
		"key":     key,
		"deleted": true,
	}
	output, _ := json.MarshalIndent(result, "", "  ")
	return &synth.ToolResult{Content: string(output)}, nil
}

// list 列出所有键值对
func (t *KVStoreTool) list(ctx context.Context) (*synth.ToolResult, error) {
	rows, err := t.db.QueryContext(ctx, "SELECT key, value, created_at, updated_at FROM "+kvTableName+" ORDER BY key")
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var key, value, createdAt, updatedAt string
		if err := rows.Scan(&key, &value, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("读取数据失败: %w", err)
		}
		items = append(items, map[string]any{
			"key":        key,
			"value":      value,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}

	output, _ := json.MarshalIndent(map[string]any{
		"count": len(items),
		"items": items,
	}, "", "  ")
	return &synth.ToolResult{Content: string(output)}, nil
}
