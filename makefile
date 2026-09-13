APP_NAME       = Tp2-Krohn-Trapani-Cicopiedi

.PHONY: all run docker-up generate test build docker-down clean

# Al escribir 'make' o 'make run', ejecuta todo de principio a fin y corta solo
all: run

# 1. Levanta Docker con Compose
docker-up:
	@echo "==> 1. Levantando PostgreSQL en Docker..."
	@docker compose up -d 
	@sleep 2

# 2. Genera código con sqlc
generate: docker-up
	@echo "==> 2. Generando código con sqlc..."
	@sqlc generate

# 3. Ejecuta las pruebas unitarias
test: generate
	@echo "==> 3. Ejecutando pruebas unitarias CRUD..."
	@go test -v ./...

# 4. Compila la aplicación
build: test
	@echo "==> 4. Compilando ejecutable..."
	@mkdir -p tmp
	@go build -o ./tmp/$(APP_NAME) .

# 5. Apaga y remueve los contenedores de Compose
docker-down:
	@echo "==> 5. Apagando contenedor de Docker..."
	@docker compose down

# Flujo principal: Ejecuta todo en orden y al final apaga Docker y cierra el proceso
run: build
	@$(MAKE) docker-down
	@echo "==> ¡Proceso finalizado con éxito!"

clean:
	@rm -rf tmp
