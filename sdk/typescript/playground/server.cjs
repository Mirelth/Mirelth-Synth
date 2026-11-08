#!/usr/bin/env node
'use strict';

// Synth Playground — 参考实现（内存 demo，非生产 LLM 后端）
//
// 这个服务只使用 Node 内置模块，不依赖 tsup 构建产物。它实现
// sdk/python/synth/client.py 所约定的 REST/SSE 契约，使 Playground 镜像
// 能真正启动、通过 /healthz 健康检查，并让 SDK 可以连通。
//
// 能力边界（与框架「不夸大能力」的约定一致）：
//   - 所有回复为确定性 demo 回显，不调用任何 LLM；
//   - 即使设置了 SYNTH_LLM_API_KEY 也不会发起外部请求；
//   - 数据全部在内存中，进程重启即丢失；
//   - 生产部署请自行替换为真实引擎并配置鉴权/持久化。
//
// 可选鉴权：设置 SYNTH_API_TOKEN 后，除 /healthz 外的所有请求
// 需带 `Authorization: Bearer <token>`；未设置时保持本地开发可直连。

const http = require('node:http');
const crypto = require('node:crypto');

const PORT = Number(process.env.SYNTH_PLAYGROUND_PORT || 8080);
const MAX_AGENTS = Number(process.env.SYNTH_MAX_AGENTS || 50);
const API_TOKEN = process.env.SYNTH_API_TOKEN || '';
const MOCK = process.env.SYNTH_PLAYGROUND_MOCK_LLM !== 'false';
const MAX_BODY = 1024 * 1024;

const agents = new Map();
const sessions = new Map();
const tools = new Map();

function sendJson(res, status, body) {
  const payload = JSON.stringify(body);
  res.writeHead(status, {
    'Content-Type': 'application/json; charset=utf-8',
    'Content-Length': Buffer.byteLength(payload),
    'X-Content-Type-Options': 'nosniff',
  });
  res.end(payload);
}

function readJson(req) {
  return new Promise((resolve, reject) => {
    let size = 0;
    const chunks = [];
    req.on('data', (chunk) => {
      size += chunk.length;
      if (size > MAX_BODY) {
        reject(new Error('request body too large'));
        req.destroy();
        return;
      }
      chunks.push(chunk);
    });
    req.on('end', () => {
      const raw = Buffer.concat(chunks).toString('utf8');
      if (!raw) return resolve({});
      try {
        resolve(JSON.parse(raw));
      } catch (err) {
        reject(new Error('invalid JSON body'));
      }
    });
    req.on('error', reject);
  });
}

function authorized(req) {
  if (!API_TOKEN) return true;
  return req.headers.authorization === `Bearer ${API_TOKEN}`;
}

function agentIdFrom(pathname, suffix) {
  const match = pathname.match(/^\/api\/playground\/agent\/([^/]+)(\/[^/]+)?$/);
  if (!match) return null;
  if (suffix && match[2] !== suffix) return null;
  return decodeURIComponent(match[1]);
}

function demoReply(agent, message) {
  return `[${agent.name}] ${MOCK ? 'demo' : 'engine'} response: ${message}`;
}

const server = http.createServer(async (req, res) => {
  try {
    const url = new URL(req.url, `http://${req.headers.host || 'localhost'}`);
    const pathname = url.pathname;
    const method = req.method || 'GET';

    if (method === 'GET' && pathname === '/healthz') {
      return sendJson(res, 200, {
        status: 'ok',
        service: 'synth-playground',
        mock: MOCK,
        agents: agents.size,
      });
    }

    if (!authorized(req)) {
      return sendJson(res, 401, { error: 'unauthorized' });
    }

    if (method === 'POST' && pathname === '/api/playground/agent') {
      const body = await readJson(req);
      if (agents.size >= MAX_AGENTS) {
        return sendJson(res, 429, { error: 'agent limit reached' });
      }
      const id = crypto.randomUUID();
      agents.set(id, {
        id,
        name: String(body.name || 'agent'),
        turns: 0,
        messages: [],
      });
      return sendJson(res, 200, { id, name: agents.get(id).name });
    }

    if (method === 'GET' && pathname === '/api/playground/agents') {
      return sendJson(res, 200, { agents: [...agents.values()].map(({ id, name, turns }) => ({ id, name, turns })) });
    }

    const chatId = agentIdFrom(pathname, '/chat');
    if (method === 'POST' && chatId) {
      const agent = agents.get(chatId);
      if (!agent) return sendJson(res, 404, { error: 'agent not found' });
      const body = await readJson(req);
      const message = String(body.message || '');
      const response = demoReply(agent, message);
      agent.turns += 1;
      agent.messages.push({ role: 'user', content: message }, { role: 'assistant', content: response });
      return sendJson(res, 200, { response });
    }

    const streamId = agentIdFrom(pathname, '/stream');
    if (method === 'POST' && streamId) {
      const agent = agents.get(streamId);
      if (!agent) return sendJson(res, 404, { error: 'agent not found' });
      const body = await readJson(req);
      const message = String(body.message || '');
      const chunks = [`[${agent.name}] `, message, '\n'];
      res.writeHead(200, {
        'Content-Type': 'text/event-stream; charset=utf-8',
        'Cache-Control': 'no-cache',
        Connection: 'keep-alive',
        'X-Content-Type-Options': 'nosniff',
      });
      for (const chunk of chunks) {
        res.write(`data: ${JSON.stringify({ chunk })}\n\n`);
      }
      agent.turns += 1;
      return res.end();
    }

    const statsId = agentIdFrom(pathname, '/stats');
    if (method === 'GET' && statsId) {
      const agent = agents.get(statsId);
      if (!agent) return sendJson(res, 404, { error: 'agent not found' });
      return sendJson(res, 200, { agentId: statsId, turns: agent.turns, messages: agent.messages.length });
    }

    const deleteId = agentIdFrom(pathname, null);
    if (method === 'DELETE' && deleteId && !pathname.includes('/chat') && !pathname.includes('/stream') && !pathname.includes('/stats')) {
      if (!agents.has(deleteId)) return sendJson(res, 404, { error: 'agent not found' });
      agents.delete(deleteId);
      return sendJson(res, 200, { deleted: true });
    }

    if (method === 'POST' && pathname === '/api/tools') {
      const body = await readJson(req);
      const id = crypto.randomUUID();
      tools.set(id, { id, name: String(body.name || 'tool'), description: String(body.description || '') });
      return sendJson(res, 200, tools.get(id));
    }

    const memoryMatch = pathname.match(/^\/api\/agent\/([^/]+)\/memory\/query$/);
    if (method === 'POST' && memoryMatch) {
      return sendJson(res, 200, { results: [] });
    }

    const sessionMatch = pathname.match(/^\/api\/agent\/([^/]+)\/session$/);
    if (method === 'POST' && sessionMatch) {
      const id = crypto.randomUUID();
      sessions.set(id, { id, agentId: decodeURIComponent(sessionMatch[1]), messages: [] });
      return sendJson(res, 200, { id });
    }

    const sessionMessages = pathname.match(/^\/api\/session\/([^/]+)\/messages$/);
    if (method === 'GET' && sessionMessages) {
      const session = sessions.get(decodeURIComponent(sessionMessages[1]));
      if (!session) return sendJson(res, 404, { error: 'session not found' });
      return sendJson(res, 200, { messages: session.messages });
    }

    return sendJson(res, 404, { error: 'not found' });
  } catch (err) {
    return sendJson(res, 400, { error: err instanceof Error ? err.message : String(err) });
  }
});

server.listen(PORT, '0.0.0.0', () => {
  console.log(`Synth Playground (${MOCK ? 'demo' : 'engine'} mode) listening on :${PORT}`);
  if (API_TOKEN) console.log('SYNTH_API_TOKEN is set: Bearer auth required');
});
