#!/usr/bin/env bash
# check-brand-residuals.sh — 旧品牌标识残留门禁
#
# 目的：保证 Synth 改名成果不回潮。扫描工作区所有文本文件，
# 命中旧品牌名 / 旧组织名 / 旧 CRD group / 旧环境变量与指标前缀即失败。
#
# 用法：bash scripts/check-brand-residuals.sh
# CI：.cnb.yml 的「品牌名残留检查」stage、.github/workflows/ci.yml 的
#     version-consistency job 均调用本脚本。
#
# 说明：
#   - 模式用字符类断开（如 agent[p]rimordia），避免脚本自身被自己命中；
#   - 白名单文件按设计必须引用旧名（迁移指南、迁移脚本、资产说明、冻结评测语料），故排除；
#   - .git / .workbuddy-ai / .thinkrail / node_modules / dist 等非仓库内容不扫描；
#   - grep -I 自动跳过二进制，二进制资产的处置见 docs/品牌资产与历史说明.md。
#
# 参考标准：与 scripts/version-check.sh 一致的 REPO_ROOT 定位与 ::error:: 标注风格。

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

# 边界：允许行首、非字母字符（含 _ " 等）与转义换行/制表符前缀；
# 排除 map_ / cap_ / gap_ 这类以字母结尾的普通词。
PATTERN='agent[p]rimordia|[G]leamseekers|agent\.primordia\.dev|\.ap\.(yaml|json)|(^|[^A-Za-z]|\\[nt])AP_|(^|[^A-Za-z0-9_])ap-[a-z]|(^|[^A-Za-z0-9_])ap/[a-z]'

EXCLUDES=(
  --exclude-dir=.git
  --exclude-dir=.hg
  --exclude-dir=.workbuddy-ai
  --exclude-dir=.codebuddy
  --exclude-dir=.thinkrail
  --exclude-dir=node_modules
  --exclude-dir=dist
  --exclude-dir=coverage
  --exclude=check-brand-residuals.sh
  --exclude=migrate-to-synth.py
  --exclude='迁移到-Synth.md'
  --exclude='品牌资产与历史说明.md'
  # 冻结评测语料：sha256 台账由 internal/eval 的冻结门守住，内容一经冻结
  # 不得修改（包括品牌字样）；历史评测结果的可比性依赖它们保持原样。
  --exclude=embedding-corpus-v1.json
  --exclude=gap-tools-v1.json
  --exclude=p4-intelligence.json
  # 更名记账文档：必须写明「改了什么」，因此会引用旧名（迁移指南同理）。
  --exclude=CHANGELOG.md
  --exclude=CHANGELOG-v7.9.md
  --exclude='发布说明-v7.9.1.md'
)

echo "========================================"
echo "  旧品牌标识残留检查"
echo "========================================"

set +e
OUTPUT="$(grep -rInE -i "${EXCLUDES[@]}" "$PATTERN" . 2>/dev/null)"
STATUS=$?
set -e

if [ "$STATUS" -eq 2 ]; then
  echo "::error::grep 执行失败，请检查 grep 是否支持 --exclude-dir / --exclude"
  exit 2
fi

if [ "$STATUS" -eq 0 ]; then
  COUNT="$(printf '%s\n' "$OUTPUT" | wc -l | tr -d ' ')"
  echo "::error::检测到 ${COUNT} 处旧品牌标识残留："
  printf '%s\n' "$OUTPUT" | head -40
  echo ""
  echo "::error::请修正后重跑；白名单规则见 scripts/check-brand-residuals.sh 文件头。"
  exit 1
fi

echo "OK: 未发现旧品牌标识残留"
echo "白名单（按设计引用旧名）：迁移指南、迁移脚本、资产说明、冻结评测语料、更名记账文档、本脚本"
echo "========================================"
