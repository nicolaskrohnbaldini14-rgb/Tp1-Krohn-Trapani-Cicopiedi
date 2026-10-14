APP_NAME := Tp2-Krohn-Trapani-Cicopiedi
HURL_DIR := hurl

.PHONY: all run clean-docker docker-up generate build test-unit test-hurl test-api docker-down clean

# Por defecto ejecuta todo el flujo completo de pruebas de la API
all: test-api

# ============================================================================
# 1. TAREAS PREVIAS
# ============================================================================

# Limpia contenedores y volúmenes de ejecuciones anteriores
clean-docker:
	@echo "==> [PRE] Limpiando contenedores y volúmenes previos..."
	@-docker compose down -v 2>/dev/null || true

# Levanta la base de datos PostgreSQL en Docker
docker-up: clean-docker
	@echo "==> [PRE] Levantando PostgreSQL en Docker..."
	@docker compose up -d || docker-compose up -d
	@echo "==> [PRE] Esperando a que la base de datos esté lista..."
	@sleep 3

# Genera el código tipado Go con sqlc
generate:
	@echo "==> [PRE] Generando código Go con sqlc..."
	@sqlc generate

# Compila el ejecutable binario de la aplicación Go
build: generate
	@echo "==> [PRE] Compilando ejecutable de la aplicación..."
	@mkdir -p tmp
	@go build -o ./tmp/$(APP_NAME) .


# ============================================================================
# 2. PRUEBAS DE ENDPOINTS CON HURL
# ============================================================================

// lo hacemos con el air a la parte del hurl?

# Inicia la aplicación en segundo plano, ejecuta los Hurl y detiene el proceso
test-hurl:
	@echo "==> [EXEC] Levantando la aplicación en segundo plano..."
	@./tmp/$(APP_NAME) & APP_PID=$$! ; \
	echo "==> [EXEC] Esperando a que el servidor HTTP escuche en el puerto 8080..." ; \
	sleep 2 ; \
	echo "==> [EXEC] Ejecutando pruebas de endpoints con Hurl..." ; \
	hurl --test $(HURL_DIR)/usuario.hurl ; \
	hurl --test $(HURL_DIR)/servicio.hurl ; \
	hurl --test $(HURL_DIR)/suscripcion.hurl ; \
	hurl --test $(HURL_DIR)/pago.hurl ; \
	STATUS=$$? ; \
	echo "==> [EXEC] Deteniendo servidor de la aplicación (PID $$APP_PID)..." ; \
	kill $$APP_PID 2>/dev/null || true ; \
	exit $$STATUS

# ============================================================================
# 3. FLUJO COMPLETO (PRE + EXEC + POST)
# ============================================================================

# Target principal que coordina todo el ciclo de vida exigido por el TP3
test-api: docker-up build test-unit
	@$(MAKE) test-hurl ; \
	STATUS=$$? ; \
	$(MAKE) docker-down ; \
	if [ $$STATUS -eq 0 ]; then \
		echo "==> ¡Todas las pruebas de la API finalización con ÉXITO!"; \
	else \
		echo "==> Las pruebas de la API FALLARON."; \
		exit 1; \
	fi

run: test-api

# ============================================================================
# 4. TAREAS POSTERIORES Y LIMPIEZA
# ============================================================================

# Elimina contenedores, volúmenes de Docker y procesos remanentes
docker-down:
	@echo "==> [POST] Deteniendo procesos y eliminando contenedores/volúmenes Docker..."
	@-pkill -f "./tmp/$(APP_NAME)" 2>/dev/null || true
	@docker compose down -v

# Limpia carpetas temporales y binarios
clean: docker-down
	@echo "==> [CLEAN] Limpiando archivos binarios temporales..."
	@rm -rf tmp