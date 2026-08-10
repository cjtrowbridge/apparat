param(
    [Alias("h")]
    [switch]$Help
)

if ($Help) {
    Write-Output "Usage: powershell -ExecutionPolicy Bypass -File scripts/run_build_pipeline.ps1"
    Write-Output "Checks local Apparat build prerequisites and runs scripts/build.py once."
    exit 0
}

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$framework = Join-Path $repoRoot "agentic-pipelines\AGENTS.md"
if (-not (Test-Path -LiteralPath $framework)) {
    Write-Error "Missing agentic-pipelines submodule. Run: git submodule update --init --recursive agents agentic-pipelines"
    exit 2
}
if (-not (Get-Command python -ErrorAction SilentlyContinue)) {
    Write-Error "Python is not available on PATH."
    exit 2
}

& python (Join-Path $repoRoot "scripts\build.py")
exit $LASTEXITCODE
