@echo off
echo Final attempt: Building with exact platform settings...

docker run --rm -v "%cd%:/workspace" -w /workspace golang:1.21-alpine ^
  sh -c "apk add --no-cache git gcc musl-dev && GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -buildmode=plugin -trimpath -ldflags='-w -s' -o ./nakama/data/modules/tictactoe.so ./modules/"

if %ERRORLEVEL% EQU 0 (
    echo Build successful!
) else (
    echo Final attempt failed!
    exit /b 1
)
