@echo off
echo Building Nakama plugin using Docker...
docker run --rm -v "%cd%:/workspace" -w /workspace heroiclabs/nakama-pluginbuilder:3.22.0 build --buildmode=plugin -o ./nakama/data/modules/tictactoe.so ./modules/
if %ERRORLEVEL% EQU 0 (
    echo Build successful! Plugin created at: nakama\data\modules\tictactoe.so
) else (
    echo Build failed!
    exit /b 1
)
