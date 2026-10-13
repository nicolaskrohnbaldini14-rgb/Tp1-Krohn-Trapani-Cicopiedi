APP_NAME := Tp2-Krohn-Trapani-Cicopiedi

.PHONY: all run clean-docker docker-up generate build test docker-down clean


all: run

clean-docker:
	@echo "==> Limpiando contenedores y volúmenes previos..."
	@-docker compose down -v 2>/dev/null || true

docker-up: clean-docker
	@echo "==> Levantando PostgreSQL en Docker..."
	@docker compose up -d || docker-compose up -d
	@echo "==> Esperando a que la base de datos esté lista..."
	@sleep 2

generate: docker-up
	@echo "==> Generando código con sqlc..."
	@sqlc generate

build: generate
	@echo "==> Compilando ejecutable..."
	@mkdir -p tmp
	@go build -o ./tmp/$(APP_NAME) .


test: build
	@echo "==> Ejecutando pruebas unitarias CRUD..."
	@go test -v ./...


docker-down:
	@echo "==> Borrando contenedores y volúmenes..."
	@docker compose down -v

run: test
	@$(MAKE) docker-down
	@echo "==> ¡Proceso finalizado con éxito!"

clean:
	@rm -rf tmp
