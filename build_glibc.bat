@echo off
echo Building with glibc compatibility...

docker run --rm -v "%cd%:/workspace" -w /workspace golang:1.21 ^
  sh -c "apt-get update && apt-get install -y gcc && GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=plugin -trimpath -ldflags='-w -s' -o ./nakama/data/modules/tictactoe.so ./modules/"

if %ERRORLEVEL% EQU 0 (
    echo Build successful with glibc!
) else (
    echo Build failed!
    exit /b 1
)
