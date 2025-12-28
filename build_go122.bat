@echo off
echo Building plugin with Go 1.22...

REM Try Go 1.22 which should match Nakama 3.22.0
docker run --rm -v "%cd%:/workspace" -w /workspace golang:1.22-alpine ^
  sh -c "apk add --no-cache git && go build -buildmode=plugin -trimpath -o ./nakama/data/modules/tictactoe.so ./modules/"

if %ERRORLEVEL% EQU 0 (
    echo Build successful with Go 1.22!
) else (
    echo Build failed! Trying Go 1.21...
    
    docker run --rm -v "%cd%:/workspace" -w /workspace golang:1.21-alpine ^
      sh -c "apk add --no-cache git && go build -buildmode=plugin -trimpath -o ./nakama/data/modules/tictactoe.so ./modules/"
    
    if !ERRORLEVEL! EQU 0 (
        echo Build successful with Go 1.21!
    ) else (
        echo Build failed with both versions!
        exit /b 1
    )
)
