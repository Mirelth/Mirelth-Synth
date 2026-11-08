#!/usr/bin/env python3
"""Migrate a downstream project from the legacy (pre-Synth) contract to Synth.

Usage:
    python3 scripts/migrate-to-synth.py --target /path/to/project            # dry-run
    python3 scripts/migrate-to-synth.py --target /path/to/project --apply    # write

Default is dry-run. Never touches .git, node_modules, dist, vendor or binary files.
Keeps UTF-8 / GBK encodings, refuses path collisions, and does not create commits.
"""
import argparse
import os
import re
import sys
from collections import Counter

B = r"(?<![A-Za-z0-9_])"
E = r"(?![A-Za-z0-9_])"

RULES = [
    ("module/v7", "github.com/Gleamseekers/AgentPrimordia/v7", "github.com/Mirelth/Synth/v7"),
    ("module/other", "github.com/Gleamseekers/AgentPrimordia", "github.com/Mirelth/Synth"),
    ("repo-short-url", "github.com/agentprimordia/ap", "github.com/Mirelth/Synth"),
    ("crd-group", "agent.primordia.dev", "synth.dev"),
    ("upper-brand", "AGENTPRIMORDIA", "SYNTH"),
    ("title-brand", "AgentPrimordia", "Synth"),
    ("lower-brand", "agentprimordia", "synth"),
    ("brand-root", re.compile(B + "Primordial" + E), "Synth"),
    ("ghcr-owner", re.compile(B + "gleamseekers" + E), "mirelth"),
    ("org", "Gleamseekers", "Mirelth"),
    ("env-prefix", re.compile(r"(?<![A-Za-z])AP_"), "SYNTH_"),
    ("metric-prefix", re.compile(r"(?<![A-Za-z])ap_"), "synth_"),
    # 字符串字面量里的转义前缀（\nap_pool / \tAP_x）：前一个字符是 n/t/r 等字母，
    # 上面的 lookbehind 会把它们当成普通单词而漏改。
    ("env-prefix-escaped", re.compile(r"(\\[ntr\"'])" + "AP_"), r"\1SYNTH_"),
    ("metric-prefix-escaped", re.compile(r"(\\[ntr\"'])" + "ap_"), r"\1synth_"),
    # 裸词 AP / ap 不参与自动改写：它们同时是 CAP 的 Availability
    # （“CP 优先于 AP”）、下游局部变量与业务缩写，盲改会破坏语义。
    # 仅收窄到已知语境，剩下的列入人工确认清单（见 MANUAL_RULES）。
    ("config-suffix", re.compile(r"[.]ap(?=[-.](?:yaml|yml|json|toml)|-(?:autonomy|cluster|skills|guard)\b)"), ".synth"),
    ("go-import-alias", re.compile(r'(?m)^([ \t]*)ap([ \t]+")'), r"\1synth\2"),
    ("go-package", re.compile(r"(?m)^package[ \t]+ap$"), "package synth"),
    ("go-qualifier", re.compile(B + r"ap[.](?=[A-Z])"), "synth."),
    ("cli-cmd", re.compile(B + r"ap(?=[ \t]+(?:init|run|a2a|config|doctor|version|completion|migrate)\b)"), "synth"),
    ("path-ap-dir", re.compile(B + "ap/"), "synth/"),
    ("path-ap-prefix", re.compile(B + "ap(?=-)"), "synth"),
]

# 裸词命中只报告不改写：需人工判断是否属于品牌语境
MANUAL_RULES = [
    ("AP-word", re.compile(B + "AP" + E)),
    ("ap-word", re.compile(B + "ap" + E)),
]


SKIP_DIRS = {".git", ".hg", ".svn", "node_modules", "dist", "build", "vendor", "coverage", ".cache"}
SKIP_SUFFIXES = (".png", ".jpg", ".jpeg", ".gif", ".webp", ".pdf", ".wasm", ".prof", ".so", ".dll", ".exe")


def transform(text):
    hits = Counter()
    for name, src, dst in RULES:
        if isinstance(src, re.Pattern):
            text, n = src.subn(dst, text)
        else:
            n = text.count(src)
            if n:
                text = text.replace(src, dst)
        hits[name] += n
    return text, hits


def scan_manual(text):
    """统计裸词 AP / ap 命中数，仅用于人工确认清单，不改写"""
    hits = Counter()
    for name, pat in MANUAL_RULES:
        n = len(pat.findall(text))
        if n:
            hits[name] += n
    return hits


def decode(raw):
    try:
        return raw.decode("utf-8"), "utf-8"
    except UnicodeDecodeError:
        try:
            return raw.decode("gb18030"), "gb18030"
        except UnicodeDecodeError:
            return None, None


def collect(root, extra_excludes):
    out = []
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in SKIP_DIRS and d not in extra_excludes]
        for fn in filenames:
            rel = os.path.relpath(os.path.join(dirpath, fn), root)
            if fn.lower().endswith(SKIP_SUFFIXES):
                continue
            if any(part in SKIP_DIRS for part in rel.split(os.sep)):
                continue
            out.append(rel)
    return sorted(out)


def main():
    ap = argparse.ArgumentParser(description="Migrate a downstream project to the Synth contract")
    ap.add_argument("--target", required=True, help="path to the downstream project root")
    ap.add_argument("--apply", action="store_true", help="actually write changes (default: dry-run)")
    ap.add_argument("--exclude", action="append", default=[], help="extra directory name to skip (repeatable)")
    args = ap.parse_args()

    root = os.path.abspath(args.target)
    if not os.path.isdir(root):
        print(f"::error::target is not a directory: {root}", file=sys.stderr)
        return 2

    files = collect(root, set(args.exclude))
    contents, encodings = {}, {}
    binary = 0
    for rel in files:
        raw = open(os.path.join(root, rel), "rb").read()
        if b"\0" in raw[:8000]:
            binary += 1
            continue
        text, enc = decode(raw)
        if text is None:
            binary += 1
            continue
        contents[rel] = text
        encodings[rel] = enc

    total_hits = Counter()
    manual = []
    changed = []
    for rel, text in contents.items():
        new, hits = transform(text)
        total_hits.update(hits)
        if new != text:
            changed.append((rel, new))
        m = scan_manual(new)
        if m:
            manual.append((rel, m))

    renames = []
    for rel in files:
        new_rel, _ = transform(rel)
        if new_rel != rel:
            renames.append((rel, new_rel))

    targets = {}
    for old, new in renames:
        targets.setdefault(new, []).append(old)
    collisions = {k: v for k, v in targets.items() if len(v) > 1}

    print("=" * 56)
    print("  Synth downstream migration", "APPLY" if args.apply else "DRY-RUN")
    print("=" * 56)
    print(f"target:            {root}")
    print(f"files scanned:     {len(files)} (text {len(contents)}, skipped {binary})")
    print(f"content changes:   {len(changed)}")
    print(f"path renames:      {len(renames)}")
    print(f"collisions:        {len(collisions)}")
    print("\nrule hits:")
    for k, v in total_hits.most_common():
        print(f"  {k:20} {v}")
    print("\ncontent change preview:")
    for rel, _ in changed[:15]:
        print(f"  M {rel}")
    print("\nrename preview:")
    for old, new in renames[:15]:
        print(f"  R {old} -> {new}")

    # 裸词 AP / ap 命中：不自动改写，仅提示人工确认
    if manual:
        files_with = len(manual)
        total_manual = sum(sum(m.values()) for _, m in manual)
        print(f"\nmanual review (未改写, 疑似裸词 AP/ap): {total_manual} 处 / {files_with} 文件")
        for rel, m in manual[:15]:
            detail = ", ".join(f"{k}={v}" for k, v in sorted(m.items()))
            print(f"  ? {rel}  ({detail})")
        print("  上述位置需人工判断：CAP 的 AP（Availability）、局部变量、业务缩写都不应改名。")

    if collisions:
        for k, v in collisions.items():
            print(f"::error::rename collision on {k}: {v}", file=sys.stderr)
        return 1
    if not args.apply:
        print("\nDRY-RUN ONLY. Re-run with --apply to write.")
        return 0

    for rel, new in changed:
        with open(os.path.join(root, rel), "wb") as fh:
            fh.write(new.encode(encodings[rel]))

    for old, new in sorted(renames, key=lambda p: p[0].count(os.sep), reverse=True):
        dst = os.path.join(root, new)
        parent = os.path.dirname(dst)
        if parent and not os.path.isdir(parent):
            os.makedirs(parent, exist_ok=True)
        if os.path.exists(dst):
            print(f"::error::target already exists, refusing overwrite: {new}", file=sys.stderr)
            return 1
        os.rename(os.path.join(root, old), dst)

    print(f"\nDONE: rewrote {len(changed)} files, renamed {len(renames)} paths.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
