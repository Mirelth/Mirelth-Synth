# Installation Guide

## TypeScript SDK

```bash
npm install @mirelth/synth
```

**要求：**
- Node.js >= 18
- 可选依赖 `react >= 18`（使用 React Server Components 时）
- 可选依赖 `better-sqlite3`（使用 SQLite 持久化时）

## Python SDK

```bash
pip install mirelth-synth
```

**要求：**
- Python >= 3.8

## Go Framework

```bash
go get github.com/Mirelth/Mirelth-Synth
```

**要求：**
- Go >= 1.26

## 从源码构建

```bash
git clone https://github.com/Mirelth/Mirelth-Synth.git
cd synth
make build
```

## 验证安装

```bash
# TypeScript
cd sdk/typescript && npm test

# Python
cd sdk/python && pip install -e . && pytest

# Go
go test ./...
```
