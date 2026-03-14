$files = Get-ChildItem -Path "internal", "cmd" -Recurse -Filter "*.go"
foreach ($file in $files) {
    if ($file.FullName -like "*logger.go") { continue }
    $text = [IO.File]::ReadAllText($file.FullName)
    if ($text -match 'logrus\.Fields') {
        $text = $text -replace 'logrus\.Fields', 'logger.Fields'
        $text = $text -replace '(?m)^\s*"github\.com/sirupsen/logrus"\r?\n', ''
        [IO.File]::WriteAllText($file.FullName, $text)
        Write-Host "Updated $($file.FullName)"
    }
}
