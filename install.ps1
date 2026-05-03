$ErrorActionPreference = 'Stop'

$Repo = 'rolfwessels/skillsync'
$Bin = 'skillsync'
$InstallDir = if ($env:INSTALL_DIR) { $env:INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\skillsync' }

$arch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
if ($arch -ne 'AMD64') {
    throw "Only windows-amd64 is supported; detected: $arch"
}

$Archive = "$Bin-windows-amd64.zip"
$Url = "https://github.com/$Repo/releases/latest/download/$Archive"

$Tmp = New-Item -ItemType Directory -Path (Join-Path $env:TEMP "skillsync-install-$(Get-Random)")
try {
    Write-Host "Downloading $Archive..."
    Invoke-WebRequest -Uri $Url -OutFile (Join-Path $Tmp $Archive) -UseBasicParsing

    Write-Host "Extracting to $InstallDir..."
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    Expand-Archive -Path (Join-Path $Tmp $Archive) -DestinationPath $InstallDir -Force

    $exe = Join-Path $InstallDir "$Bin.exe"
    Write-Host "Installed: $exe"
    & $exe --version

    $userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
    if (-not ($userPath -split ';' -contains $InstallDir)) {
        Write-Host "Adding $InstallDir to user PATH..."
        $newPath = if ($userPath) { "$userPath;$InstallDir" } else { $InstallDir }
        [Environment]::SetEnvironmentVariable('Path', $newPath, 'User')
        $env:Path = "$env:Path;$InstallDir"
        Write-Host "Restart your terminal for PATH changes to apply to new shells."
    }
} finally {
    Remove-Item -Recurse -Force $Tmp -ErrorAction SilentlyContinue
}
