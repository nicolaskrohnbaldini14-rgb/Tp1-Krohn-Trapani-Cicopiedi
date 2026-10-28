package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib" // driver de PostgreSQL

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
	"tp2-krohn-trapani-cicopiedi/handlers"
)

func main() {
	// 1. Conexión a PostgreSQL (Docker). El esquema lo crea el contenedor al iniciar
	//    (db/schema/schema.sql se monta en docker-entrypoint-initdb.d).
	connStr := "postgres://postgres:postgres@localhost:5432/mi_tp2_db?sslmode=disable"

	dbConn, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatalf("Error al preparar el pool de conexiones: %v", err)
	}
	defer dbConn.Close()

	if err := dbConn.Ping(); err != nil {
		log.Fatalf("No se pudo conectar a PostgreSQL: %v", err)
	}
	log.Println("Conexión exitosa a PostgreSQL")

	// 2. Capa de datos (sqlc)
	queries := sqlc.New(dbConn)

	// 3. Handlers que renderizan las vistas templ (la lógica de negocio sigue en logic/)
	mux := http.NewServeMux()
	handlers.NuevoWeb(queries).Registrar(mux)

	// 4. Servidor HTTP
	log.Println("Servidor escuchando en http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}
