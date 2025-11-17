Param(
    [string]$InstallDir = "$env:ProgramFiles\\ComposePack"
)

$target = Join-Path $InstallDir "composepack.exe"
if (Test-Path $target) {
    Remove-Item $target -Force
    Write-Output "Removed $target"
} else {
    Write-Output "composepack not found in $InstallDir"
}
