# envia installer for Windows - https://github.com/Saimonsanbr/envia
# Uso:
#   iwr -useb https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.ps1 | iex
#   # ou baixar, inspecionar e depois executar (recomendado):
#   iwr https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.ps1 -OutFile install.ps1
#   # inspecione o arquivo, depois:
#   powershell -ExecutionPolicy Bypass -File install.ps1
#   # com versão específica:
#   powershell -ExecutionPolicy Bypass -File install.ps1 -Version v0.2.3
#   # ou
#   .\install.ps1 -Version latest

param(
    [string]$Version = "latest",
    [string]$InstallDir = ""
)

$ErrorActionPreference = "Stop"
$Repo = "Saimonsanbr/envia"
$Binary = "envia"
$BoreVersion = "v0.6.0"

Write-Host @"
========================================================================
 envia - instalador Windows (sem admin)
 Repo: https://github.com/$Repo
 Este script é 100% auditável e open source. Leia antes de executar!
 Não confie em scripts de qualquer pessoa - verifique a URL e o código.
========================================================================
"@ -ForegroundColor Cyan

function Info($msg) { Write-Host "==> $msg" -ForegroundColor Green }
function Warn($msg) { Write-Host "⚠ $msg" -ForegroundColor Yellow }
function Err($msg) { Write-Host "✗ $msg" -ForegroundColor Red }

# Verifica ExecutionPolicy e avisa
$policy = Get-ExecutionPolicy -Scope CurrentUser
if ($policy -eq "Restricted" -or $policy -eq "AllSigned") {
    Warn "Sua ExecutionPolicy atual ($policy) bloqueia scripts remotos."
    Write-Host "  Para permitir apenas este script (sem admin, só para seu usuário):"
    Write-Host "    Set-ExecutionPolicy -Scope CurrentUser -ExecutionPolicy RemoteSigned -Force" -ForegroundColor Gray
    Write-Host "  Ou execute com bypass (recomendado para instalar):"
    Write-Host "    powershell -ExecutionPolicy Bypass -File install.ps1" -ForegroundColor Gray
    Write-Host "  Ou via iwr (já usa Bypass implicitamente):"
    Write-Host "    iwr -useb https://raw.githubusercontent.com/Saimonsanbr/envia/main/install.ps1 | iex" -ForegroundColor Gray
    Write-Host ""
    Write-Host "  Segurança: este script só baixa binários das releases oficiais do envia e bore (MIT),"
    Write-Host "  não executa nada além de mover para $InstallDir. Todo código é auditável em:"
    Write-Host "  https://github.com/Saimonsanbr/envia/blob/main/install.ps1" -ForegroundColor Gray
    Write-Host ""
}

# Escolhe diretório sem admin: %LOCALAPPDATA%\Programs\envia (preferido) ou %USERPROFILE%\.local\bin
if (-not $InstallDir -or $InstallDir -eq "") {
    $localApp = $env:LOCALAPPDATA
    if ($localApp -and (Test-Path $localApp)) {
        $InstallDir = Join-Path $localApp "Programs\envia"
    } else {
        $InstallDir = Join-Path $env:USERPROFILE ".local\bin"
    }
}
Write-Host "Diretório de instalação: $InstallDir" -ForegroundColor Gray
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Detecta arch
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "amd64" }
# bore e envia só têm amd64 para Windows hoje
$platform = "windows-$arch"
$ext = ".exe"

# URLs
if ($Version -eq "latest") {
    $enviaUrl = "https://github.com/$Repo/releases/latest/download/$Binary-$platform$ext"
} else {
    $tag = $Version
    if (-not $tag.StartsWith("v")) { $tag = "v$tag" }
    $enviaUrl = "https://github.com/$Repo/releases/download/$tag/$Binary-$platform$ext"
}
$boreAsset = "bore-v0.6.0-x86_64-pc-windows-msvc.zip"
$boreUrl = "https://github.com/ekzhang/bore/releases/download/$BoreVersion/$boreAsset"

$tmp = Join-Path $env:TEMP "envia-install-$(Get-Random)"
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
trap { Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue } 

try {
    Info "Baixando $Binary $Version para $platform..."
    Write-Host "     $enviaUrl" -ForegroundColor Gray
    $dstEnvia = Join-Path $tmp "$Binary$ext"
    try {
        Invoke-WebRequest -Uri $enviaUrl -OutFile $dstEnvia -UseBasicParsing
    } catch {
        Err "Falha ao baixar $enviaUrl"
        Write-Host "  Verifique se a release $Version existe em https://github.com/$Repo/releases" -ForegroundColor Gray
        Write-Host "  Alternativa: baixe manualmente o .zip da release e descompacte" -ForegroundColor Gray
        exit 1
    }

    # Baixa bore bundle se não existir no PATH e não existir no destino
    $boreDst = Join-Path $InstallDir "bore.exe"
    $needBore = $true
    if (Get-Command bore -ErrorAction SilentlyContinue) {
        Info "bore já no PATH: $((Get-Command bore).Source) - pulando bundle"
        $needBore = $false
    } elseif (Test-Path $boreDst) {
        Info "bore já em $boreDst - pulando"
        $needBore = $false
    }

    if ($needBore) {
        Info "Baixando bore bundle $BoreVersion para $platform..."
        Write-Host "     $boreUrl" -ForegroundColor Gray
        $boreZip = Join-Path $tmp "bore.zip"
        try {
            Invoke-WebRequest -Uri $boreUrl -OutFile $boreZip -UseBasicParsing
            # Extrai bore.exe
            Add-Type -AssemblyName System.IO.Compression.FileSystem
            $extractDir = Join-Path $tmp "bore-extract"
            New-Item -ItemType Directory -Force -Path $extractDir | Out-Null
            Expand-Archive -Path $boreZip -DestinationPath $extractDir -Force
            $boreExe = Get-ChildItem -Path $extractDir -Filter "bore.exe" -Recurse | Select-Object -First 1
            if ($boreExe) {
                Copy-Item $boreExe.FullName $boreDst -Force
                Info "bore bundle instalado em $boreDst"
            } else {
                Warn "bore.exe não encontrado no zip"
            }
        } catch {
            Warn "Falha ao baixar bore bundle, continue com bore do PATH se existir: $_"
        }
    }

    # Move envia
    $finalEnvia = Join-Path $InstallDir "$Binary$ext"
    Move-Item -Path $dstEnvia -Destination $finalEnvia -Force
    Info "Instalado em $finalEnvia"

    # Adiciona ao PATH do usuário se não estiver
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallDir*") {
        Warn "$InstallDir não está no PATH"
        Write-Host "  Adicionando ao PATH do usuário..." -ForegroundColor Gray
        $newPath = "$userPath;$InstallDir".Trim(';')
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path += ";$InstallDir"
        Info "PATH atualizado. Feche e reabra o terminal, ou use:"
        Write-Host "    `$env:Path += `";$InstallDir`"" -ForegroundColor Gray
    } else {
        Info "$InstallDir já está no PATH"
    }

    # Testa versão
    try {
        $ver = & $finalEnvia --version 2>&1 | Select-Object -First 1
        Info "$ver"
    } catch {}

    # Mostra onde está bore
    if (Test-Path $boreDst) {
        try { Info "bore bundle: $((& $boreDst --version 2>&1 | Select-Object -First 1))" } catch {}
    } elseif (Get-Command bore -ErrorAction SilentlyContinue) {
        Info "bore: $((bore --version 2>&1 | Select-Object -First 1))"
    }

    Write-Host ""
    Info "Pronto! Teste:"
    Write-Host "    envia --help" -ForegroundColor Gray
    Write-Host "    envia --provider bore testes/documento-teste.txt" -ForegroundColor Gray
    Write-Host "    envia  # modo interativo" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Dica: se o Windows Defender reclamar, é falso positivo do bore (MIT, sem instalador)." -ForegroundColor Gray
}
finally {
    # trap já remove $tmp
}
