.PHONY: build clean docker-build docker-up docker-down docker-logs

# Build the Go plugin
build:
	@echo "Building Nakama plugin..."
	go build -buildmode=plugin -o ./nakama/data/modules/tictactoe.so ./modules/

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f ./nakama/data/modules/*.so

# Build inside Docker container (for Linux compatibility)
docker-build:
	@echo "Building plugin in Docker..."
	docker run --rm -v "${PWD}:/workspace" -w /workspace golang:1.21 \
		go build -buildmode=plugin -o ./nakama/data/modules/tictactoe.so ./modules/

# Start Nakama and PostgreSQL services
docker-up:
	@echo "Starting Nakama services..."
	cd nakama && docker-compose up -d

# Stop services
docker-down:
	@echo "Stopping Nakama services..."
	cd nakama && docker-compose down

# View service logs
docker-logs:
	@echo "Viewing Nakama logs..."
	cd nakama && docker-compose logs -f nakama

# View all logs
docker-logs-all:
	@echo "Viewing all logs..."
	cd nakama && docker-compose logs -f
