@echo off
echo Building with Go 1.22.4 (Debian Bookworm) matching Nakama 3.22.0...

docker run --rm -v "%cd%:/workspace" -w /workspace golang:1.22.4-bookworm ^
  sh -c "go build -buildmode=plugin -trimpath -ldflags='-w -s' -o ./nakama/data/modules/tictactoe.so ./modules/"

if %ERRORLEVEL% EQU 0 (
    echo Build successful with Golden Version!
) else (
    echo Build failed!
    exit /b 1
)
