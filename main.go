package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib" // driver de PostgreSQL

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
	"tp2-krohn-trapani-cicopiedi/handlers"
	"tp2-krohn-trapani-cicopiedi/logic"
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

	// 3. Catálogo de servicios precargado (no duplica si ya existe)
	ctx := context.Background()
	if err := logic.NuevoServicioServicio(queries).CargarCatalogoInicial(ctx); err != nil {
		log.Fatalf("Error al precargar el catálogo de servicios: %v", err)
	}
	log.Println("Catálogo de servicios precargado")

	// las suscripciones activas que ya vencieron se renuevan por el mismo plazo
	if renovadas, err := logic.NuevoServicioSuscripcion(queries).RenovarVencidas(ctx); err != nil {
		log.Fatalf("Error al renovar suscripciones vencidas: %v", err)
	} else if renovadas > 0 {
		log.Printf("Suscripciones renovadas: %d", renovadas)
	}

	// 4. Handlers que renderizan las vistas templ (la lógica de negocio sigue en logic/)
	mux := http.NewServeMux()
	handlers.NuevoWeb(queries, handlers.NuevasSesiones()).Registrar(mux)

	// 5. Servidor HTTP
	log.Println("Servidor escuchando en http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}
