@echo off
echo Building plugin with Nakama-compatible Go version...

REM Use the exact Nakama image as the build environment to guarantee version matching
docker run --rm -v "%cd%:/workspace" ^
  registry.heroiclabs.com/heroiclabs/nakama:3.22.0 ^
  sh -c "apk add --no-cache git && cd /workspace && go build -buildmode=plugin -trimpath -o ./nakama/data/modules/tictactoe.so ./modules/"

if %ERRORLEVEL% EQU 0 (
    echo Build successful! Plugin created with matching Go version.
) else (
    echo Build failed!
    exit /b 1
)
