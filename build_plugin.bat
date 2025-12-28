@echo off
docker run --rm -v "%cd%:/workspace" -w /workspace heroiclabs/nakama-pluginbuilder:3.22.0 build --buildmode=plugin -trimpath -o ./nakama/data/modules/tictactoe.so ./modules/
