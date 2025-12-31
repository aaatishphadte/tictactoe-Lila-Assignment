#!/bin/bash
docker run --rm -v "$(pwd):/workspace" -w /workspace registry.heroiclabs.com/heroiclabs/nakama-pluginbuilder:3.22.0 build --buildmode=plugin -trimpath -o ./nakama/data/modules/tictactoe.so ./modules/
