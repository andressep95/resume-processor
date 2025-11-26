.PHONY: run up down build logs ps clean

run:
	@echo "Iniciando el servidor..."
	@go run cmd/main.go

# Docker Compose commands
up:
	@echo "Levantando servicios con Docker Compose..."
	@docker-compose up -d

down:
	@echo "Deteniendo servicios..."
	@docker-compose down

build:
	@echo "Construyendo y levantando servicios..."
	@docker-compose up -d --build

logs:
	@docker-compose logs -f

ps:
	@docker-compose ps

clean:
	@echo "Deteniendo servicios y eliminando volúmenes..."
	@docker-compose down -v

