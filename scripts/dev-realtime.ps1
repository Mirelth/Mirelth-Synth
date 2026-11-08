# ============================================================
# Synth Realtime Voice - One-click launcher (Windows)
#
# Runs a real voice session with the `synth realtime voice` CLI.
# Path 1 (keyless): local faster-whisper + Piper endpoints via
#                   SYNTH_ASR_URL / SYNTH_TTS_URL (or --asr-url/--tts-url).
# Path 2 (key):     SYNTH_LLM_API_KEY (or SYNTH_OPENAI_API_KEY) with
#                   OpenAI endpoints (default URLs when unset).
# Falls back to mock with a clear message when neither is configured.
#
# Usage:
#   powershell -ExecutionPolicy Bypass -File scripts\dev-realtime.ps1
# ============================================================

param(
    [string]$AsrUrl = $env:SYNTH_ASR_URL,
    [string]$TtsUrl = $env:SYNTH_TTS_URL,
    [string]$Voice  = $env:SYNTH_TTS_VOICE,
    [switch]$Mock
)

$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
$AgentDir = Join-Path $Root 'synth'

Write-Host "==> Synth realtime voice launcher" -ForegroundColor Cyan

if (-not (Get-Command 'go' -ErrorAction SilentlyContinue)) {
    Write-Error "Missing command: go (install Go 1.26+ first)"
}

# 装配真实/模拟 ASR 与 TTS
$asr = 'mock'
$tts = 'mock'
$args = @('run', './cmd/synth', 'realtime', 'voice')

if (-not $Mock) {
    $key = $env:SYNTH_LLM_API_KEY
    if (-not $key) { $key = $env:SYNTH_OPENAI_API_KEY }

    if ($AsrUrl -or $TtsUrl -or $key) {
        if (-not $AsrUrl -and $key) { $AsrUrl = 'https://api.openai.com/v1/audio/transcriptions' }
        if (-not $TtsUrl -and $key) { $TtsUrl = 'https://api.openai.com/v1/audio/speech' }
        if (-not $AsrUrl -or -not $TtsUrl) {
            Write-Host "!! Real mode needs both ASR and TTS endpoints:" -ForegroundColor Yellow
            Write-Host "   local  : set SYNTH_ASR_URL (faster-whisper) and SYNTH_TTS_URL (Piper)"
            Write-Host "   cloud  : set SYNTH_LLM_API_KEY (defaults to OpenAI endpoints)"
            Write-Host "   falling back to mock..." -ForegroundColor Yellow
        } else {
            $asr = 'openai'
            $tts = 'openai'
            $args += @('--asr=openai', '--tts=openai', "--asr-url=$AsrUrl", "--tts-url=$TtsUrl")
            if ($Voice) { $args += "--tts-voice=$Voice" }
        }
    } else {
        Write-Host "No SYNTH_ASR_URL/SYNTH_TTS_URL/SYNTH_LLM_API_KEY set - using mock (real mode hints below)." -ForegroundColor Yellow
    }
}

Write-Host "    ASR: $asr | TTS: $tts" -ForegroundColor DarkGray
if ($asr -eq 'mock') {
    Write-Host ""
    Write-Host "Real mode examples:" -ForegroundColor Cyan
    Write-Host "  powershell -File scripts\dev-realtime.ps1 -AsrUrl http://127.0.0.1:9000/v1/audio/transcriptions -TtsUrl http://127.0.0.1:5002/v1/audio/speech"
    Write-Host "  `$env:SYNTH_LLM_API_KEY=sk-... ; scripts\dev-realtime.ps1"
    Write-Host ""
}

Write-Host "==> Running: synth realtime voice ($asr/$tts)..." -ForegroundColor Green
Push-Location $AgentDir
try {
    & go @args
} finally {
    Pop-Location
}
