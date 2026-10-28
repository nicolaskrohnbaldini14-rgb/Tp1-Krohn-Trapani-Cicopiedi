APP_NAME := Tp4-Krohn-Trapani-Cicopiedi
HOST := http://localhost:8080
HURL_FILES := hurl/frontend.hurl hurl/usuario.hurl hurl/servicio.hurl hurl/suscripcion.hurl hurl/pago.hurl

.PHONY: all run run-app clean-docker docker-up generate build test test-api docker-down clean


all: test-api

clean-docker:
	@echo "==> Limpiando contenedores y volúmenes previos..."
	@-docker compose down -v 2>/dev/null || true

docker-up: clean-docker
	@echo "==> Levantando PostgreSQL en Docker..."
	@docker compose up -d --wait
	@echo "==> La base de datos está lista (healthcheck OK)"

generate: docker-up
	@echo "==> Generando código con sqlc..."
	@sqlc generate

build: generate
	@echo "==> Compilando ejecutable..."
	@mkdir -p tmp
	@go build -o ./tmp/$(APP_NAME) .


test:
	@echo "==> Ejecutando pruebas unitarias CRUD..."
	@go test -v -count=1 ./...


test-api: build
	@$(MAKE) --no-print-directory test || { $(MAKE) --no-print-directory docker-down; exit 1; }
	@echo "==> Levantando la aplicación..."
	@./tmp/$(APP_NAME) & APP_PID=$$! ; \
	curl -s -o /dev/null --retry 10 --retry-connrefused --retry-delay 1 $(HOST)/usuarios ; \
	echo "==> Probando los endpoints con Hurl..." ; \
	hurl --test --jobs 1 --variable host=$(HOST) $(HURL_FILES) ; \
	STATUS=$$? ; \
	kill $$APP_PID ; \
	$(MAKE) --no-print-directory docker-down ; \
	exit $$STATUS


docker-down:
	@echo "==> Borrando contenedores y volúmenes..."
	@docker compose down -v

run: test-api
	@echo "==> ¡Proceso finalizado con éxito!"

clean:
	@rm -rf tmp

# run-app: levanta la base, genera el código, compila y deja la app corriendo
# en http://localhost:8080. Al cortar con Ctrl+C borra contenedores y volúmenes.
run-app: build
	@echo "==> Aplicación disponible en $(HOST) (Ctrl+C para detener)"
	@trap '$(MAKE) --no-print-directory docker-down' EXIT; \
	trap 'exit 0' INT TERM; \
	./tmp/$(APP_NAME)
