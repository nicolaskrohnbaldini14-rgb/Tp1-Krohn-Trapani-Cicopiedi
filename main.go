package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib" // driver de PostgreSQL

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
	"tp2-krohn-trapani-cicopiedi/logic"
)

func main() {
	// 1. Conexión a PostgreSQL (Docker)
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

	// 3. Capa de lógica
	servicioUsuario := logic.NuevoServicioUsuario(queries)
	servicioServicio := logic.NuevoServicioServicio(queries)
	servicioSuscripcion := logic.NuevoServicioSuscripcion(queries)
	servicioPago := logic.NuevoServicioPago(queries)

	// 4. Handlers
	manejadorUsuario := logic.NuevoManejadorUsuario(servicioUsuario)
	manejadorServicio := logic.NuevoManejadorServicio(servicioServicio)
	manejadorSuscripcion := logic.NuevoManejadorSuscripcion(servicioSuscripcion)
	manejadorPago := logic.NuevoManejadorPago(servicioPago)

	// 5. Rutas
	http.HandleFunc("/usuarios", manejadorUsuario.ManejarUsuarios)      // GET, POST
	http.HandleFunc("/usuarios/", manejadorUsuario.ManejarUsuarioPorID) // GET, PUT, DELETE

	http.HandleFunc("/servicios", manejadorServicio.ManejarServicios)      // GET, POST
	http.HandleFunc("/servicios/", manejadorServicio.ManejarServicioPorID) // GET, PUT, DELETE

	http.HandleFunc("/suscripciones", manejadorSuscripcion.ManejarSuscripciones)                    // POST
	http.HandleFunc("/suscripciones/", manejadorSuscripcion.ManejarSuscripcionPorID)                // GET, PUT, DELETE
	http.HandleFunc("/suscripciones/usuario/", manejadorSuscripcion.ManejarSuscripcionesPorUsuario) // GET

	http.HandleFunc("/pagos", manejadorPago.ManejarPagos)                            // POST
	http.HandleFunc("/pagos/", manejadorPago.ManejarPagoPorID)                       // GET, DELETE
	http.HandleFunc("/pagos/suscripcion/", manejadorPago.ManejarPagosPorSuscripcion) // GET

	// 6. Servidor HTTP
	log.Println("Servidor escuchando en http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}