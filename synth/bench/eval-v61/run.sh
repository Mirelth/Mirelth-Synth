#!/bin/bash
# v6.1 命题1 世界模型配对AB —— sensenova 网关配置
# 用法: bash run.sh [--limit N] [--rounds N] [--arms A,B]
set -euo pipefail

cd "$(dirname "$0")/../.."

go run ./bench/eval-v61 \
  --provider openai \
  --model "sensenova-6.8-flash-lite" \
  --base-url "https://token.sensenova.cn/v1" \
  --api-key "sk-IE84nyrP9MdJXleKAfAKGNiawk81sHZW" \
  "$@"
