# Synth Playground — 参考实现镜像（内存 demo 服务）
# 构建（必须在仓库根执行）：
#   docker build -f sdk/typescript/playground/Dockerfile -t synth-playground .
# 运行：
#   docker run -d -p 8080:8080 -e SYNTH_LLM_API_KEY=sk-xxx synth-playground
#
# 修复说明：旧 CMD 指向 tsup 未产出的 dist 子目录入口（tsup 只以
# src/index.ts 为入口），产物中并不存在该文件，容器必然启动失败。现改为零依赖的
# server.cjs（CommonJS 显式扩展名，避免 sdk/typescript/package.json 的
# "type": "module" 把 .js 当 ESM），不依赖 npm 构建，也不引入第三方运行时依赖。
#
# 该镜像是本地/演示用途的参考实现，不调用真实 LLM；生产用途请替换引擎、
# 配置鉴权（SYNTH_API_TOKEN）与持久化。

FROM node:20-alpine AS runtime

LABEL maintainer="Synth Team"
LABEL description="Synth Playground - In-memory reference server (demo)"

WORKDIR /app

RUN addgroup -S synthgroup && adduser -S synthuser -G synthgroup

COPY sdk/typescript/playground/server.cjs ./server.cjs

ENV NODE_ENV=production \
    SYNTH_PLAYGROUND_PORT=8080 \
    SYNTH_LLM_PROVIDER=openai \
    SYNTH_LLM_MODEL=gpt-4 \
    SYNTH_MAX_AGENTS=50 \
    SYNTH_LOG_LEVEL=info \
    SYNTH_PLAYGROUND_MOCK_LLM=true

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:${SYNTH_PLAYGROUND_PORT}/healthz || exit 1

USER synthuser

CMD ["node", "server.cjs"]
