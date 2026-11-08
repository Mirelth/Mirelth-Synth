#!/usr/bin/env python3
"""gen-brand-assets.py — 生成 Synth 品牌图片物料（海报 / 幻灯片 / 封面帧）

视觉规范取自 docs/视频脚本.md §二「创意概念」：
  深空蓝黑底 + 电光青主色 + 紫色点缀 + 发光粒子 + 线框几何体。
所有上屏数字只取 docs/视频脚本.md §一「数字白名单」中经代码核实的口径。

实现：把每个画面渲染成一份自包含 HTML，再用 Edge headless 截图，最后用 PIL
降采样到目标尺寸（2x 渲染 → LANCZOS 降采样，兼顾字重与体积）。

用法：
    ~/.local/bin/python3.12 scripts/gen-brand-assets.py            # 全部重生成
    python3 scripts/gen-brand-assets.py --only poster
    python3 scripts/gen-brand-assets.py --list
    python3 scripts/gen-brand-assets.py --check    # 只渲染不覆盖

依赖：Microsoft Edge（headless 截图）、Pillow。无第三方渲染服务。
用系统 python3（带 Pillow）运行：python3 scripts/gen-brand-assets.py
"""
from __future__ import annotations

import argparse
import os
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

EDGE_CANDIDATES = [
    "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
    "/Applications/Chromium.app/Contents/MacOS/Chromium",
]

REPO = Path(__file__).resolve().parent.parent

# ── 设计令牌 ────────────────────────────────────────────────
FONT = '"Helvetica Neue", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif'
MONO = '"SF Mono", "JetBrains Mono", Menlo, monospace'
CYAN = "#22d3ee"
CYAN_DIM = "#0e7490"
VIOLET = "#a78bfa"
TEXT = "#e9f2ff"
MUTED = "#8fa3c4"
BG = "#050a13"

BASE_CSS = f"""
* {{ margin:0; padding:0; box-sizing:border-box; }}
html, body {{ width:100%; height:100%; }}
body {{
  font-family:{FONT}; color:{TEXT}; background:{BG};
  overflow:hidden; -webkit-font-smoothing:antialiased;
  text-rendering:geometricPrecision;
}}
.stage {{ position:relative; width:100%; height:100%; overflow:hidden;
  background:
    radial-gradient(90% 70% at 12% 8%, rgba(34,211,238,.20), transparent 60%),
    radial-gradient(80% 65% at 92% 96%, rgba(167,139,250,.20), transparent 62%),
    linear-gradient(158deg,#04080f 0%, #081326 48%, #060d1a 100%);
}}
.grid {{ position:absolute; inset:0; opacity:.35;
  background-image:
    linear-gradient(rgba(34,211,238,.07) 1px, transparent 1px),
    linear-gradient(90deg, rgba(34,211,238,.07) 1px, transparent 1px);
  background-size:72px 72px;
  mask-image:radial-gradient(circle at 50% 45%, #000 35%, transparent 85%);
}}
.orb {{ position:absolute; border-radius:50%; filter:blur(2px); }}
.dots {{ position:absolute; inset:0; opacity:.55; }}
.pad {{ position:absolute; inset:0; display:flex; flex-direction:column; }}
.eyebrow {{ font-size:22px; letter-spacing:.42em; text-transform:uppercase; color:{CYAN};
  font-weight:600; }}
.brandbar {{ display:flex; align-items:center; gap:16px; }}
.mark {{ width:14px; height:14px; transform:rotate(45deg);
  background:linear-gradient(135deg,{CYAN},{VIOLET});
  box-shadow:0 0 18px rgba(34,211,238,.8); }}
.footer {{ display:flex; justify-content:space-between; align-items:flex-end;
  font-size:20px; color:{MUTED}; }}
.footer .url {{ font-family:{MONO}; color:{CYAN}; }}
h1 {{ font-size:132px; line-height:.98; letter-spacing:-.02em; font-weight:800; }}
h2 {{ font-size:66px; line-height:1.14; font-weight:800; letter-spacing:-.01em; }}
.grad {{ background:linear-gradient(120deg,{TEXT} 10%,{CYAN} 55%,{VIOLET} 92%);
  -webkit-background-clip:text; background-clip:text; color:transparent; }}
.lead {{ font-size:30px; color:{MUTED}; line-height:1.55; }}
.mono {{ font-family:{MONO}; }}
.card {{ border:1px solid rgba(34,211,238,.22); border-radius:22px;
  background:linear-gradient(160deg, rgba(20,40,66,.72), rgba(9,18,32,.72));
  box-shadow:inset 0 1px 0 rgba(255,255,255,.05); }}
.pill {{ display:inline-flex; align-items:center; gap:10px; padding:10px 20px;
  border-radius:999px; border:1px solid rgba(34,211,238,.35);
  background:rgba(34,211,238,.08); color:{CYAN}; font-size:22px; font-weight:600; }}
.num {{ font-family:{MONO}; font-size:104px; font-weight:700; letter-spacing:-.03em;
  background:linear-gradient(120deg,{CYAN},{VIOLET}); -webkit-background-clip:text;
  background-clip:text; color:transparent; }}
.kicker {{ font-size:26px; font-weight:700; color:{TEXT}; }}
.sub {{ font-size:22px; color:{MUTED}; line-height:1.5; }}
.rule {{ height:1px; background:linear-gradient(90deg,{CYAN},transparent); }}
"""


def dots(n: int = 46, w: int = 1080, h: int = 1920, seed: int = 7) -> str:
    """确定性粒子散布（LCG，无随机依赖，保证可复现）"""
    out = []
    x = seed * 2654435761 % 2147483647
    for _ in range(n):
        x = (x * 48271) % 2147483647
        px = x % w
        x = (x * 48271) % 2147483647
        py = x % h
        x = (x * 48271) % 2147483647
        r = 1 + (x % 3)
        o = 0.15 + (x % 60) / 200
        out.append(
            f'<div class="orb" style="left:{px}px;top:{py}px;width:{r*3}px;height:{r*3}px;'
            f'background:{CYAN};opacity:{o:.2f}"></div>'
        )
    return "".join(out)


def page(w: int, h: int, body: str, extra_css: str = "", seed: int = 7) -> str:
    return f"""<!doctype html><html lang="zh-CN"><head><meta charset="utf-8">
<style>{BASE_CSS}{extra_css}
.stage {{ width:{w}px; height:{h}px; }}
</style></head><body><div class="stage">
<div class="grid"></div>{dots(46, w, h, seed)}
{body}
</div></body></html>"""


# ── 画面定义 ────────────────────────────────────────────────

POSTER_CSS = """
.pad { padding:96px 84px; }
.head { display:flex; justify-content:space-between; align-items:center; }
.head .org { font-size:22px; letter-spacing:.36em; color:#8fa3c4; text-transform:uppercase; }
.body { flex:1; display:flex; flex-direction:column; justify-content:center; gap:44px; }
.foot { display:flex; justify-content:space-between; font-size:20px; color:#6c82a4; }
.tag { font-size:26px; color:#8fa3c4; letter-spacing:.2em; }
"""


def poster_head(idx: int) -> str:
    return f"""<div class="pad">
<div class="head"><div class="brandbar"><div class="mark"></div>
<span class="eyebrow">Synth</span></div>
<span class="org">Mirelth · {idx:02d}</span></div>"""


def poster_foot(note: str = "开源 AI Agent 开发框架") -> str:
    return f"""<div class="foot"><span>{note}</span>
<span class="url">github.com/Mirelth/Mirelth-Synth</span></div></div>"""


POSTERS = []

# 01 封面
POSTERS.append(("docs/posters/poster-01.png", 1080, 1920, poster_head(1) + f"""
<div class="body" style="text-align:center;align-items:center">
  <div class="pill">生产级 · 并发原生 · Go + TypeScript</div>
  <h1 class="grad" style="font-size:168px;letter-spacing:.04em">SYNTH</h1>
  <div style="font-size:52px;font-weight:700;letter-spacing:.08em">万物之源 · 智能之始</div>
  <div class="mono" style="font-size:26px;color:{MUTED};letter-spacing:.24em">FROM SYNTH, INTELLIGENCE</div>
  <div class="rule" style="width:220px;margin:8px auto"></div>
  <div class="lead" style="text-align:center">通用 AI Agent 开发框架<br>ReActLoop · AgentPool · Tool System · Memory · LLM</div>
  <div class="pill">v7.9.0</div>
</div>""" + poster_foot("Mirelth 团队作品"), 11))

# 02 定位
POSTERS.append(("docs/posters/poster-02.png", 1080, 1920, poster_head(2) + f"""
<div class="body">
  <div class="eyebrow">Positioning</div>
  <h2>通用 AI Agent<br><span class="grad">开发框架</span></h2>
  <div class="lead">不做应用，只做让应用跑得住的底座。</div>
  <div class="card" style="padding:44px 46px;display:flex;flex-direction:column;gap:30px">
    <div><div class="kicker">轻量</div><div class="sub">标准库优先，无 Web 框架 / ORM / CLI 框架依赖</div></div>
    <div><div class="kicker">并发原生</div><div class="sub">AgentPool + 并发安全的共享状态，压测在 CI 常驻</div></div>
    <div><div class="kicker">生产验证</div><div class="sub">断点续跑、熔断降级、审计与配额内建</div></div>
  </div>
</div>""" + poster_foot(), 12))

# 03 引擎
POSTERS.append(("docs/posters/poster-03.png", 1080, 1920, poster_head(3) + f"""
<div class="body">
  <div class="eyebrow">Engine</div>
  <h2><span class="grad">ReAct</span> 循环<br>每一步都可恢复</h2>
  <div class="card" style="padding:40px 42px;display:flex;flex-direction:column;gap:24px">
    <div class="kicker">35 个生命周期钩子</div>
    <div class="sub">从输入护栏到工具执行、记忆写入，每段都可插桩</div>
    <div class="rule"></div>
    <div class="kicker">每轮 checkpoint</div>
    <div class="sub">进程中断后从断点续跑，不重放已完成的工具调用</div>
    <div class="rule"></div>
    <div class="kicker">工具调用与记忆注入内建</div>
    <div class="sub">工具描述走结构化 Schema，记忆按相关性注入 Prompt</div>
  </div>
</div>""" + poster_foot(), 13))

# 04 编排
POSTERS.append(("docs/posters/poster-04.png", 1080, 1920, poster_head(4) + f"""
<div class="body">
  <div class="eyebrow">Orchestration</div>
  <h2><span class="grad">9 种</span>编排模式</h2>
  <div class="lead">单 Agent 之外，多 Agent 的协作形态都已内建。</div>
  <div style="display:grid;grid-template-columns:1fr 1fr;gap:18px">
    {"".join(f'<div class="card" style="padding:26px 28px;font-size:28px;font-weight:700{" ;grid-column:span 2" if m == "A2A" else ""}">{m}</div>' for m in ["Pipeline","Handoff","Parallel","DAG","GroupChat","Swarm","MapReduce","Debate","A2A"])}
  </div>
</div>""" + poster_foot(), 14))

# 05 能力
POSTERS.append(("docs/posters/poster-05.png", 1080, 1920, poster_head(5) + f"""
<div class="body">
  <div class="eyebrow">Capabilities</div>
  <h2>记忆 · 模型 · 工具<br><span class="grad">三层可替换</span></h2>
  <div class="card" style="padding:40px 42px;display:flex;flex-direction:column;gap:26px">
    <div><div class="kicker">记忆</div><div class="sub">SQLite FTS5 + HNSW 向量 + RAG，RRF 混合检索</div></div>
    <div class="rule"></div>
    <div><div class="kicker">模型</div><div class="sub">13 个 Provider：OpenAI / Anthropic / Azure / Gemini / GLM / Qwen / DeepSeek / Mistral / Cohere / Ollama 及多模态变体</div></div>
    <div class="rule"></div>
    <div><div class="kicker">工具</div><div class="sub">MCP Client + Server、内置工具集、插件与技能市场</div></div>
  </div>
</div>""" + poster_foot(), 15))

# 06 安全
POSTERS.append(("docs/posters/poster-06.png", 1080, 1920, poster_head(6) + f"""
<div class="body">
  <div class="eyebrow">Security</div>
  <h2><span class="grad">默认收紧</span><br>而不是默认放开</h2>
  <div style="display:flex;flex-wrap:wrap;gap:16px">
    {"".join(f'<div class="card" style="padding:22px 26px;font-size:24px">{s}</div>' for s in ["ACL 权限","WASM 沙箱（wazero）","输入/输出护栏","PII Trie 脱敏","路径遍历与 symlink 防护","AES-GCM 密钥","多租户配额"])}
  </div>
  <div class="lead">Shell 白名单 + 元字符拒绝；网络出口默认关闭；代码执行显式 opt-in。</div>
</div>""" + poster_foot(), 16))

# 07 数据卡
POSTERS.append(("docs/posters/poster-07.png", 1080, 1920, poster_head(7) + f"""
<div class="body">
  <div class="eyebrow">Benchmarks</div>
  <h2>官方基准<br><span class="sub" style="font-size:24px">README §性能，均可在仓库复现</span></h2>
  <div class="card" style="padding:38px 42px;display:flex;flex-direction:column;gap:26px">
    <div><div class="num">20,000</div><div class="sub">并发读（另 7,500 并发写）· 0 错误</div></div>
    <div class="rule"></div>
    <div><div class="num">88,182</div><div class="sub">30 分钟稳态请求总数 · 0 错误</div></div>
    <div class="rule"></div>
    <div><div class="num">104µs</div><div class="sub">9 个端点 P95 上界</div></div>
    <div class="rule"></div>
    <div><div class="num">98%</div><div class="sub">SQLite WAL 写入 P95 降幅（6.9ms → 108µs）</div></div>
  </div>
  <div class="sub" style="font-size:19px">仅引用仓库内可复现数字；不做跨框架速度对比。</div>
</div>""" + poster_foot(), 17))

# 08 生态
POSTERS.append(("docs/posters/poster-08.png", 1080, 1920, poster_head(8) + f"""
<div class="body">
  <div class="eyebrow">Ecosystem</div>
  <h2>从 IDE 到集群<br><span class="grad">一条链路</span></h2>
  <div class="card" style="padding:38px 42px;display:flex;flex-direction:column;gap:24px">
    <div class="kicker">CLI · 22 个子命令</div>
    <div class="mono sub" style="font-size:21px">init run debug loop test config mcp plugin cluster marketplace autonomy skill profile a2a realtime create-edge-agent doctor migrate completion …</div>
    <div class="rule"></div>
    <div class="kicker">部署与观测</div>
    <div class="sub">K8s Operator（AgentDeployment CRD + HPA）· Prometheus / OTel / pprof / Grafana</div>
    <div class="rule"></div>
    <div class="kicker">边缘与 IDE</div>
    <div class="sub">Edge 网关（Cloudflare Workers / Vercel Edge，WASI）· VS Code 扩展 · 浏览器扩展</div>
  </div>
</div>""" + poster_foot(), 18))

# 09 自举
POSTERS.append(("docs/posters/poster-09.png", 1080, 1920, poster_head(9) + f"""
<div class="body">
  <div class="eyebrow">Bootstrapping</div>
  <h2>用 Synth<br><span class="grad">开发 Synth</span></h2>
  <div class="card" style="padding:44px 46px;display:flex;align-items:center;gap:34px">
    <div style="text-align:center;flex:1"><div class="num" style="font-size:82px">0.33</div><div class="sub">自举成功率<br>（起点）</div></div>
    <div style="font-size:56px;color:{CYAN}">→</div>
    <div style="text-align:center;flex:1"><div class="num" style="font-size:82px">1.00</div><div class="sub">自举成功率<br>（当前）</div></div>
  </div>
  <div class="lead">自举闭环 + 季度对照回归门：框架改动由框架自己跑验收，退化即拦。</div>
</div>""" + poster_foot(), 19))

# 10 CTA
POSTERS.append(("docs/posters/poster-10.png", 1080, 1920, poster_head(10) + f"""
<div class="body">
  <div class="eyebrow">Get Started</div>
  <h2>从一行命令<br><span class="grad">到生产级 Agent</span></h2>
  <div class="card" style="padding:34px 38px;display:flex;flex-direction:column;gap:22px">
    <div class="mono" style="font-size:24px;color:{CYAN}">go get github.com/Mirelth/Mirelth-Synth/v7</div>
    <div class="rule"></div>
    <div class="mono" style="font-size:24px;color:{CYAN}">synth init myagent --template quickstart</div>
    <div class="rule"></div>
    <div class="mono" style="font-size:24px;color:{CYAN}">npm install @mirelth/synth</div>
  </div>
  <div class="lead">文档：README · docs/快速开始.md · docs/API参考.md<br>迁移自旧品牌：docs/迁移到-Synth.md</div>
</div>""" + poster_foot("Mirelth 团队 · Apache-2.0"), 20))


# ── 幻灯片（16:9）────────────────────────────────────────────

SLIDE_CSS = """
.pad { padding:64px 76px; }
.head { display:flex; justify-content:space-between; align-items:center; margin-bottom:40px; }
.head .org { font-size:19px; letter-spacing:.34em; color:#8fa3c4; text-transform:uppercase; }
h2.title { font-size:54px; }
.layer { display:flex; align-items:center; gap:34px; padding:26px 34px; margin-bottom:18px;
  border-radius:18px; border:1px solid rgba(34,211,238,.28);
  background:linear-gradient(90deg, rgba(18,38,62,.85), rgba(9,18,32,.55)); }
.layer .name { width:190px; font-size:30px; font-weight:800; }
.layer .items { flex:1; font-size:24px; color:#c6d6ee; line-height:1.45; }
"""


def slide(w: int, h: int, title: str, body: str, seed: int = 31, idx: str = "SLIDE") -> str:
    return page(w, h, f"""<div class="pad">
<div class="head"><div class="brandbar"><div class="mark"></div>
<span class="eyebrow">Synth</span></div><span class="org">Mirelth · {idx}</span></div>
<div style="flex:1;display:flex;flex-direction:column;justify-content:center;gap:38px">
<h2 class="title">{title}</h2>
<div>{body}</div>
</div>
</div>""", SLIDE_CSS, seed)


def layers_block(explain: bool = False) -> str:
    rows = [
        ("应用层", "Your Application / Synth CLI / Synth Studio / 任意 AI Agent 产品", CYAN),
        ("引擎层", "ReActLoop · AgentPool · DAG · Pipeline · GroupChat · Handoff · A2A · MapReduce", VIOLET),
        ("能力层", "Tools + MCP + Plugin · Memory + RAG + Vector · 13 Provider · Prompt", "#34d399"),
        ("治理层", "Governance · Audit · Security · Guardrail · Resilience · Health · Config", "#fbbf24"),
        ("基础设施", "Metrics · OTel · Events · Persist · Concurrency · Logger · Protocol", CYAN),
    ]
    out = []
    for name, items, color in rows:
        out.append(
            f'<div class="layer"><div class="name" style="color:{color}">{name}</div>'
            f'<div class="items">{items}</div></div>'
        )
    note = (
        '<div class="sub" style="margin-top:26px">依赖方向：严格单向（上层 → 下层，禁止反向引用）</div>'
        if explain
        else ""
    )
    return "".join(out) + note


SLIDES = [
    ("docs/images/ppt/architecture.png", 1792, 1024,
     slide(1792, 1024, "Synth 分层架构", layers_block(True), 31, "01"), "01"),
    ("docs/images/ppt/architecture_v2.png", 1792, 1024,
     slide(1792, 1024, "Synth 分层架构 · 关键机制",
           layers_block(False) + """<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:18px;margin-top:26px">
<div class="card" style="padding:22px 26px"><div class="kicker">引擎</div><div class="sub">35 生命周期钩子 · 每轮 checkpoint</div></div>
<div class="card" style="padding:22px 26px"><div class="kicker">能力</div><div class="sub">FTS5 + HNSW + RAG（RRF 融合）</div></div>
<div class="card" style="padding:22px 26px"><div class="kicker">治理</div><div class="sub">ACL · WASM 沙箱 · 审计 · 配额</div></div>
</div>""", 32, "02"), "02"),
    ("docs/images/ppt/architecture_hd.png", 2560, 1440,
     slide(2560, 1440, "Synth 分层架构", layers_block(True), 33, "03"), "03"),

    ("docs/images/ppt/react_loop.png", 1792, 1024,
     slide(1792, 1024, "ReAct 循环 · 每一步可插桩可恢复", """<div style="display:flex;gap:22px;align-items:stretch">
""" + "".join(
         f'<div class="card" style="flex:1;padding:28px 24px"><div class="num" style="font-size:40px">{i}</div>'
         f'<div class="kicker" style="margin-top:10px">{t}</div><div class="sub" style="margin-top:8px">{d}</div></div>'
         for i, (t, d) in enumerate([
             ("输入", "护栏 + 注入检测"),
             ("思考", "LLM 规划下一步"),
             ("行动", "工具调用（Scope 校验）"),
             ("观测", "结果注入上下文"),
             ("记忆", "写回 + 检索增强"),
         ], 1)
     ) + """</div>
<div class="sub" style="margin-top:26px">循环边界由 35 个生命周期钩子切分；每轮落一次 checkpoint，中断后从断点续跑。</div>""", 34, "04"), "04"),

    ("docs/images/ppt/rag_pipeline.png", 1792, 1024,
     slide(1792, 1024, "记忆与检索 · 混合检索流水线", """<div style="display:flex;gap:18px;align-items:center">
""" + "".join(
         f'<div class="card" style="flex:1;padding:24px 22px;text-align:center"><div class="kicker">{t}</div>'
         f'<div class="sub" style="margin-top:8px">{d}</div></div>' + ("<div style='font-size:30px;color:#22d3ee'>→</div>" if i < 4 else "")
         for i, (t, d) in enumerate([
             ("文档 / 对话", "原始内容"),
             ("切分 + 嵌入", "chunk → vector"),
             ("双路检索", "SQLite FTS5 + HNSW"),
             ("RRF 融合", "去重 + 排序"),
             ("注入 Prompt", "Top-K 上下文"),
         ])
     ) + """</div>
<div class="sub" style="margin-top:28px">记忆层同时服务「会话上下文」与「长期知识」：写入走 WAL，检索走 RRF 混合排序。</div>""", 35, "05"), "05"),

    ("docs/images/ppt/security_layers.png", 1792, 1024,
     slide(1792, 1024, "安全分层 · 默认收紧", """<div style="display:flex;flex-direction:column;gap:16px">
""" + "".join(
         f'<div class="layer"><div class="name" style="color:{c}">{n}</div><div class="items">{d}</div></div>'
         for n, d, c in [
             ("输入层", "注入检测 · 主题过滤 · PII Trie 脱敏", CYAN),
             ("权限层", "ACL · Scope 校验 · 敏感参数脱敏日志", VIOLET),
             ("执行层", "WASM 沙箱（wazero）· Shell 白名单 + 元字符拒绝 · 路径遍历与 symlink 防护", "#34d399"),
             ("输出层", "护栏 · 结果校验 · 截断", "#fbbf24"),
             ("治理层", "审计 · 配额 · 熔断降级 · 密钥 AES-GCM", CYAN),
         ]
     ) + """</div>
<div class="sub" style="margin-top:22px">宿主进程运行期零写入、零编译、零加载 agent 生成的代码（INV-0），代码只在 WASM 沙箱内执行。</div>""", 36, "06"), "06"),

    ("docs/images/ppt/ecosystem.png", 1792, 1024,
     slide(1792, 1024, "生态 · 从 IDE 到集群", """<div style="display:grid;grid-template-columns:repeat(3,1fr);gap:20px">
""" + "".join(
         f'<div class="card" style="padding:26px 28px"><div class="kicker">{t}</div><div class="sub" style="margin-top:10px">{d}</div></div>'
         for t, d in [
             ("SDK", "Go 模块 + TypeScript SDK，双线对等"),
             ("CLI", "22 个子命令（init / run / cluster / marketplace …）"),
             ("插件与市场", "MCP Client + Server · 插件 ECDSA 验签"),
             ("K8s Operator", "AgentDeployment CRD + HPA"),
             ("IDE 与浏览器", "VS Code 扩展 · 浏览器扩展"),
             ("边缘网关", "Cloudflare Workers / Vercel Edge（WASI）"),
         ]
     ) + """</div>""", 37, "07"), "07"),

    ("docs/images/ppt/code_stats.png", 1792, 1024,
     slide(1792, 1024, "代码规模 · 仓库实测",
           f"""<div style="display:grid;grid-template-columns:repeat(4,1fr);gap:22px">
<div class="card" style="padding:32px 28px"><div class="num" style="font-size:62px">153,868</div><div class="sub">Go 行（非测试/非生成）· 723 文件</div></div>
<div class="card" style="padding:32px 28px"><div class="num" style="font-size:62px">59,778</div><div class="sub">TypeScript src 行 · 238 文件</div></div>
<div class="card" style="padding:32px 28px"><div class="num" style="font-size:62px">28</div><div class="sub">internal 顶层包</div></div>
<div class="card" style="padding:32px 28px"><div class="num" style="font-size:62px">22</div><div class="sub">CLI 子命令</div></div>
</div>
<div class="sub" style="margin-top:28px">统计口径：`git ls-files` 范围，排除 *_test.go 与 *.pb.go；采集于 v7.9.0。</div>""", 38, "08"), "08"),

    ("docs/images/ppt/timeline.png", 1792, 1024,
     slide(1792, 1024, "版本弧线 · 从框架到可验证平台", """<div style="display:flex;align-items:flex-end;gap:16px;height:330px">
""" + "".join(
         f'<div style="flex:1;display:flex;flex-direction:column;justify-content:flex-end;gap:12px">'
         f'<div style="height:{hh}px;border-radius:14px 14px 4px 4px;background:linear-gradient(180deg,rgba(34,211,238,.85),rgba(167,139,250,.35))"></div>'
         f'<div class="sub" style="text-align:center">{v}</div><div class="mono sub" style="text-align:center;font-size:17px">{k}</div></div>'
         for v, k, hh in [
             ("v0.1", "框架骨架", 90),
             ("v2.0", "双 SDK", 140),
             ("v3.x", "A2A / 集群", 180),
             ("v5.0", "并发与性能", 230),
             ("v6.0", "安全边界", 268),
             ("v7.x", "验收弧线", 310),
             ("v7.8", "当前", 330),
         ]
     ) + """</div>""", 39, "09"), "09"),

    ("docs/images/ppt/comparison_radar.png", 1792, 1024,
     slide(1792, 1024, "能力全景 · 七大域", """<div style="display:grid;grid-template-columns:repeat(4,1fr);gap:18px">
""" + "".join(
         f'<div class="card" style="padding:24px 24px"><div class="kicker">{t}</div><div class="sub" style="margin-top:8px">{d}</div></div>'
         for t, d in [
             ("引擎", "ReActLoop · 35 钩子 · checkpoint"),
             ("编排", "9 种多 Agent 模式"),
             ("记忆", "FTS5 + HNSW + RAG（RRF）"),
             ("工具", "MCP · 插件 · 市场"),
             ("模型", "13 个 Provider"),
             ("安全", "ACL · WASM · 护栏 · PII"),
             ("可观测", "Prometheus · OTel · pprof"),
             ("生态", "CLI · Operator · IDE · Edge"),
         ]
     ) + """</div>
<div class="sub" style="margin-top:26px">只列已实现能力，不做跨框架速度对比。</div>""", 40, "10"), "10"),
]

# 无文字背景
SLIDES.append(("docs/images/ppt/cover_bg.png", 1792, 1024,
               page(1792, 1024, "", "", 41), "11"))

# 视频预览帧（16:9 1920×1080）
FRAMES = [
    ("docs/ppt_preview_frame1.png", 1920, 1080, page(1920, 1080, f"""<div class="pad" style="padding:120px 130px">
<div class="brandbar"><div class="mark"></div><span class="eyebrow">Mirelth · Synth</span></div>
<div style="flex:1;display:flex;flex-direction:column;justify-content:center;gap:30px">
<h1 class="grad" style="font-size:150px;letter-spacing:.04em">SYNTH</h1>
<div style="font-size:56px;font-weight:700">万物之源 · 智能之始</div>
<div class="lead" style="font-size:30px;letter-spacing:.18em">FROM SYNTH, INTELLIGENCE</div>
</div>
<div class="footer"><span>生产级 AI Agent 开发框架</span><span class="url">github.com/Mirelth/Mirelth-Synth</span></div>
</div>""", "", 51)),
    ("docs/ppt_preview_frame2.png", 1920, 1080, page(1920, 1080, f"""<div class="pad" style="padding:100px 120px">
<div class="head" style="display:flex;justify-content:space-between"><div class="brandbar"><div class="mark"></div><span class="eyebrow">Synth</span></div><span class="org" style="font-size:19px;letter-spacing:.34em;color:{MUTED}">MIRELTH · 02</span></div>
<div style="flex:1;display:flex;flex-direction:column;justify-content:center;gap:34px">
<h2 style="font-size:64px">引擎 · 记忆 · 编排<br><span class="grad">一条链路打通</span></h2>
<div style="display:flex;gap:22px">
""" + "".join(
        f'<div class="card" style="flex:1;padding:30px 28px"><div class="num" style="font-size:56px">{n}</div>'
        f'<div class="sub" style="margin-top:8px">{d}</div></div>'
        for n, d in [("35", "生命周期钩子"), ("13", "模型 Provider"), ("9", "编排模式"), ("22", "CLI 子命令")]
    ) + f"""</div>
<div class="sub" style="font-size:24px">断点续跑 · 熔断降级 · WASM 沙箱 · 审计与配额，默认收紧。</div>
</div>
<div class="footer"><span>所有数字可在仓库复现</span><span class="url">docs/视频脚本.md</span></div>
</div>""", "", 52)),
    ("docs/ppt_preview_frame3.png", 1920, 1080, page(1920, 1080, f"""<div class="pad" style="padding:110px 130px">
<div class="brandbar"><div class="mark"></div><span class="eyebrow">Get Started</span></div>
<div style="flex:1;display:flex;flex-direction:column;justify-content:center;gap:34px">
<h2 style="font-size:72px">从一行命令<br><span class="grad">到生产级 Agent</span></h2>
<div class="card" style="padding:34px 38px;display:flex;flex-direction:column;gap:20px">
<div class="mono" style="font-size:28px;color:{CYAN}">go get github.com/Mirelth/Mirelth-Synth/v7</div>
<div class="rule"></div>
<div class="mono" style="font-size:28px;color:{CYAN}">synth init myagent --template quickstart</div>
</div>
</div>
<div class="footer"><span>Apache-2.0 · Mirelth</span><span class="url">github.com/Mirelth/Mirelth-Synth</span></div>
</div>""", "", 53)),
]


# ── 渲染 ────────────────────────────────────────────────────

def find_edge() -> str:
    for p in EDGE_CANDIDATES:
        if Path(p).exists():
            return p
    raise SystemExit("::error::未找到 Edge/Chrome，无法渲染（headless 截图需要浏览器）")


def render(html: str, out: Path, w: int, h: int, edge: str, scale: int = 2,
           colors: int = 256) -> None:
    from PIL import Image

    with tempfile.TemporaryDirectory() as td:
        tmp = Path(td)
        page_path = tmp / "page.html"
        shot = tmp / "shot.png"
        page_path.write_text(html, encoding="utf-8")
        cmd = [
            edge,
            "--headless=new",
            "--disable-gpu",
            "--hide-scrollbars",
            "--no-first-run",
            "--no-default-browser-check",
            f"--force-device-scale-factor={scale}",
            f"--window-size={w},{h}",
            f"--screenshot={shot}",
            page_path.as_uri(),
        ]
        proc = subprocess.run(cmd, capture_output=True, text=True)
        if not shot.exists():
            raise SystemExit(
                f"::error::渲染失败 {out.name}\n引擎输出：{proc.stdout[-400:]}{proc.stderr[-400:]}"
            )
        img = Image.open(shot).convert("RGB")
        if img.size != (w, h):
            img = img.resize((w, h), Image.LANCZOS)
        # 调色板量化：渐变与字重几乎无感，体积约减半（--colors 0 关闭）
        if colors:
            img = img.convert("P", palette=Image.ADAPTIVE, colors=colors)
        out.parent.mkdir(parents=True, exist_ok=True)
        img.save(out, "PNG", optimize=True)


def main() -> int:
    ap = argparse.ArgumentParser(description="生成 Synth 品牌图片物料")
    ap.add_argument("--only", default="", help="只渲染路径包含该子串的资产")
    ap.add_argument("--list", action="store_true", help="列出全部资产")
    ap.add_argument("--colors", type=int, default=256,
                    help="调色板颜色数（默认 256，0 = 全彩）")
    args = ap.parse_args()

    assets = (
        [(rel, w, h, page(w, h, body, POSTER_CSS, seed)) for rel, w, h, body, seed in POSTERS]
        + [(rel, w, h, html) for rel, w, h, html, _ in SLIDES]
        + [(rel, w, h, html) for rel, w, h, html in FRAMES]
    )

    if args.list:
        for rel, w, h, _ in assets:
            print(f"{w:>5}x{h:<5} {rel}")
        print(f"\n共 {len(assets)} 个资产")
        return 0

    edge = find_edge()
    done = 0
    for rel, w, h, html in assets:
        if args.only and args.only not in rel:
            continue
        out = REPO / rel
        render(html, out, w, h, edge, colors=args.colors)
        size_kb = os.path.getsize(out) // 1024
        print(f"  ✓ {rel}  ({w}x{h}, {size_kb} KB)")
        done += 1
    print(f"\n完成：{done} 个资产")
    return 0


if __name__ == "__main__":
    sys.exit(main())
