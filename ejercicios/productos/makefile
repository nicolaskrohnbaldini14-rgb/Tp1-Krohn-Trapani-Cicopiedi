SQLC := $(shell command -v sqlc 2>/dev/null || echo "go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1")

.PHONY: run generate clean

# Genera el código de sqlc y templ y levanta el servidor en http://localhost:8080
run: generate
	go run .

generate:
	$(SQLC) generate
	go tool templ generate

clean:
	rm -f products.db
