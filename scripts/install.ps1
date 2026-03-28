$ErrorActionPreference = "Stop"

$repo = if ($env:GOCLIP_REPO) { $env:GOCLIP_REPO } else { "Ramyprojs/go-clip" }
$version = if ($env:GOCLIP_VERSION) { $env:GOCLIP_VERSION } else { "latest" }
$installDir = if ($env:GOCLIP_INSTALL_DIR) {
    $env:GOCLIP_INSTALL_DIR
} else {
    Join-Path $HOME "AppData\Local\Programs\goclip\bin"
}

function Get-GoclipArch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64" { return "amd64" }
        "Arm64" { return "arm64" }
        default { throw "Unsupported Windows architecture: $arch" }
    }
}

$arch = Get-GoclipArch
$archive = "goclip_windows_${arch}.zip"
if ($version -eq "latest") {
    $url = "https://github.com/$repo/releases/latest/download/$archive"
} else {
    $url = "https://github.com/$repo/releases/download/$version/$archive"
}

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("goclip-install-" + [System.Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
    $archivePath = Join-Path $tempDir $archive
    Write-Host "Downloading $url"
    try {
        Invoke-WebRequest -Uri $url -OutFile $archivePath
    } catch {
        throw "Unable to download a goclip release. Publish a tagged GitHub release before using this installer."
    }

    Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null

    $binaryPath = Join-Path $installDir "goclip.exe"
    Copy-Item -Path (Join-Path $tempDir "goclip.exe") -Destination $binaryPath -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $segments = @()
    if (-not [string]::IsNullOrWhiteSpace($userPath)) {
        $segments = $userPath.Split(";", [System.StringSplitOptions]::RemoveEmptyEntries)
    }

    $normalizedInstallDir = $installDir.TrimEnd("\")
    $pathUpdated = $false
    if (-not ($segments | Where-Object { $_.TrimEnd("\") -eq $normalizedInstallDir })) {
        $newUserPath = if ([string]::IsNullOrWhiteSpace($userPath)) {
            $installDir
        } else {
            "$userPath;$installDir"
        }

        [Environment]::SetEnvironmentVariable("Path", $newUserPath, "User")
        $env:Path = "$env:Path;$installDir"
        $pathUpdated = $true
    }

    Write-Host "Installed goclip to $binaryPath"
    if ($pathUpdated) {
        Write-Host "Updated your user PATH. Restart PowerShell if the command is not available yet."
    }

    & $binaryPath version
} finally {
    Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}
