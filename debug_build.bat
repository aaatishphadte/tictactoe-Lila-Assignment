@echo off
REM Save compile errors to a file
docker run --rm -v "%cd%:/workspace" -w /workspace golang:1.21 bash -c "go build -buildmode=plugin -o test.so ./modules/ 2>&1" > build_errors.txt
type build_errors.txt
