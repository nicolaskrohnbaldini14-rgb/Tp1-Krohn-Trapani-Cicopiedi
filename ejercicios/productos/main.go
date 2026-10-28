package main

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"productos/db/sqlc"
	"productos/views"
)

//go:embed db/schema/schema.sql
var schema string

func main() {
	// Conexión a SQLite y creación de la tabla si no existe
	conn, err := sql.Open("sqlite3", "products.db")
	if err != nil {
		log.Fatalf("Error al abrir la base de datos: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(context.Background(), schema); err != nil {
		log.Fatalf("Error al crear el esquema: %v", err)
	}

	queries := db.New(conn)

	// GET /: lista los productos y renderiza la página
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		products, err := queries.ListProducts(r.Context())
		if err != nil {
			http.Error(w, "Error al listar productos", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := views.IndexPage("Productos", products).Render(r.Context(), w); err != nil {
			log.Printf("Error al renderizar: %v", err)
		}
	})

	// POST /products: crea el producto y redirige a / (Post/Redirect/Get)
	http.HandleFunc("POST /products", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Formulario inválido", http.StatusBadRequest)
			return
		}

		name := strings.TrimSpace(r.FormValue("name"))
		price, err := strconv.ParseFloat(r.FormValue("price"), 64)
		if name == "" || err != nil || price < 0 {
			http.Error(w, "Nombre o precio inválido", http.StatusBadRequest)
			return
		}

		if err := queries.CreateProduct(r.Context(), db.CreateProductParams{Name: name, Price: price}); err != nil {
			http.Error(w, "Error al guardar el producto", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	log.Println("Servidor escuchando en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
