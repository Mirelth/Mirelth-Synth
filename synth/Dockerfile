# ============================================================
# Stage 1: Go builder
# ============================================================
FROM golang:1.26.10-alpine AS go-builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
RUN go build -ldflags="-s -w" -o /bin/synth-basic      ./synth/ecosystem/examples/basic/ \
 && go build -ldflags="-s -w" -o /bin/synth-with-tools  ./synth/ecosystem/examples/with-tools/ \
 && go build -ldflags="-s -w" -o /bin/synth-multi-agent ./synth/ecosystem/examples/multi-agent/

# ============================================================
# Stage 2: Node builder
# ============================================================
FROM node:20-alpine AS node-builder

WORKDIR /build

COPY sdk/typescript/package.json sdk/typescript/tsconfig.json ./
COPY sdk/typescript/src/ ./src/

RUN npm install
RUN npm run build

# ============================================================
# Stage 3: Runtime
# ============================================================
FROM node:20-alpine

RUN apk add --no-cache bash

COPY --from=go-builder /bin/synth-basic      /usr/local/bin/synth-basic
COPY --from=go-builder /bin/synth-with-tools /usr/local/bin/synth-with-tools
COPY --from=go-builder /bin/synth-multi-agent /usr/local/bin/synth-multi-agent

COPY --from=node-builder /build/dist        /opt/synth-sdk/dist
COPY --from=node-builder /build/src         /opt/synth-sdk/src
COPY --from=node-builder /build/package.json /opt/synth-sdk/package.json
COPY --from=node-builder /build/node_modules /opt/synth-sdk/node_modules

COPY synth/demo/ /opt/synth-demo/
RUN chmod +x /opt/synth-demo/run-go.sh /opt/synth-demo/run-ts.sh

# Security hardening
RUN addgroup -S synth && adduser -S synth -G synth
USER synth

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD pgrep -f "synth-basic" > /dev/null || exit 1

WORKDIR /opt/synth-demo

CMD ["bash", "/opt/synth-demo/run-go.sh"]
