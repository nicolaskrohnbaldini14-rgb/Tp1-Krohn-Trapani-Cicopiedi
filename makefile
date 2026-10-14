APP_BIN := bin/tp3-api
HOST    := http://localhost:8080
HURL_FILES := hurl/usuario.hurl hurl/servicio.hurl hurl/suscripcion.hurl hurl/pago.hurl

.PHONY: test-api docker-up generate build docker-down clean

# Flujo completo: tareas previas, prueba de endpoints y tareas posteriores
test-api: docker-up build
	@echo "==> Levantando la aplicación..."
	@./$(APP_BIN) & APP_PID=$$! ; \
	curl -s -o /dev/null --retry 10 --retry-connrefused --retry-delay 1 $(HOST)/usuarios ; \
	echo "==> Probando los endpoints con Hurl..." ; \
	hurl --test --jobs 1 --variable host=$(HOST) $(HURL_FILES) ; \
	STATUS=$$? ; \
	kill $$APP_PID ; \
	$(MAKE) --no-print-directory docker-down ; \
	exit $$STATUS

# Borra contenedores y volúmenes previos, levanta PostgreSQL y espera el healthcheck
docker-up:
	@echo "==> Levantando PostgreSQL..."
	@docker compose down -v
	@docker compose up -d --wait

# Genera el código Go con sqlc
generate:
	@echo "==> Generando código con sqlc..."
	@sqlc generate

# Compila la aplicación
build: generate
	@echo "==> Compilando..."
	@go build -o $(APP_BIN) .

# Borra contenedores y volúmenes
docker-down:
	@echo "==> Borrando contenedores y volúmenes..."
	@docker compose down -v

# Borra además el binario y el código generado
clean: docker-down
	@rm -rf bin db/sqlc