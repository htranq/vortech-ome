SHELL := /bin/bash
tag ?= local

# build images from all services
.PHONY: build-image-local
build-image-local:
	docker build -f build/Dockerfile -t stream_management:${tag} .

.PHONY: build-linux
build-linux: download-go-mod
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=readonly -o out/cmd/main cmd/main.go

test:
	echo "Test success"

.PHONY: download-go-mod
download-go-mod: go.mod
	go mod download all


# Start services
.PHONY: up
up:
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		if [ "$(build)" = "true" ]; then \
			docker compose up --build -d $(filter-out $@,$(MAKECMDGOALS)); \
		else \
			docker compose up -d $(filter-out $@,$(MAKECMDGOALS)); \
		fi; \
	else \
		if [ "$(build)" = "true" ]; then \
			docker compose up --build -d; \
		else \
			docker compose up -d; \
		fi; \
	fi

# Stop services
.PHONY: down
down:
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		docker compose stop $(filter-out $@,$(MAKECMDGOALS)); \
	else \
		docker compose down; \
	fi

# Show logs
.PHONY: logs
logs:
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		docker compose logs $(filter-out $@,$(MAKECMDGOALS)); \
	else \
		docker compose logs; \
	fi

# Follow logs
.PHONY: logsf
logsf:
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		docker compose logs -f $(filter-out $@,$(MAKECMDGOALS)); \
	else \
		docker compose logs -f; \
	fi

# Restart services
.PHONY: restart
restart:
	@if [ "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		docker compose restart $(filter-out $@,$(MAKECMDGOALS)); \
	else \
		docker compose restart; \
	fi

# Additional useful commands
.PHONY: ps
ps:
	docker ps -a
.PHONY: clean
clean:
	docker compose down -v --remove-orphans
	docker system prune -f

# Handle service names as arguments
%:
	@: 