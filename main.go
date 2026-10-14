package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq" // Driver de PostgreSQL (o "github.com/jackc/pgx/v5/stdlib")

	// Cambiá la ruta de importación por el módulo de tu proyecto
	"tp2-krohn-trapani-cicopiedi/db/sqlc"
	"tp2-krohn-trapani-cicopiedi/logic"
)

func main() {
	// 1. Cadena de conexión a PostgreSQL (Docker)
	connStr := "postgres://postgres:postgres@localhost:5432/mi_tp2_db?sslmode=disable"

	// 2. Abrir el pool de conexiones a la Base de Datos
	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error al preparar el pool de conexiones: %v", err)
	}
	defer dbConn.Close()

	// 3. Verificar que PostgreSQL responda
	if err := dbConn.Ping(); err != nil {
		log.Fatalf("No se pudo conectar a PostgreSQL: %v", err)
	}
	log.Println("Conexión exitosa a PostgreSQL")

	// 4. Instanciar la capa de datos (sqlc)
	queries := sqlc.New(dbConn)

	// 5. Instanciar la Capa de Lógica / Servicios
	servicioUsuario := logic.NuevoServicioUsuario(queries)
	servicioServicio := logic.NuevoServicioServicio(queries)
	servicioSuscripcion := logic.NuevoServicioSuscripcion(queries)
	servicioPago := logic.NuevoServicioPago(queries)

	// 6. Instanciar la Capa de Presentación / Manejadores (Handlers)
	manejadorUsuario := logic.NuevoManejadorUsuario(servicioUsuario)
	manejadorServicio := logic.NuevoManejadorServicio(servicioServicio)
	manejadorSuscripcion := logic.NuevoManejadorSuscripcion(servicioSuscripcion)
	manejadorPago := logic.NuevoManejadorPago(servicioPago)

	// 7. Registrar las Rutas HTTP para cada Entidad

	// --- Rutas de Usuarios ---
	http.HandleFunc("/usuarios", manejadorUsuario.ManejarUsuarios)       // GET, POST
	http.HandleFunc("/usuarios/", manejadorUsuario.ManejarUsuarioPorID) // GET, PUT, DELETE /usuarios/{id}

	// --- Rutas de Servicios ---
	http.HandleFunc("/servicios", manejadorServicio.ManejarServicios)       // GET, POST
	http.HandleFunc("/servicios/", manejadorServicio.ManejarServicioPorID) // GET, PUT, DELETE /servicios/{id}

	// --- Rutas de Suscripciones ---
	http.HandleFunc("/suscripciones", manejadorSuscripcion.ManejarSuscripciones)            // POST
	http.HandleFunc("/suscripciones/", manejadorSuscripcion.ManejarSuscripcionPorID)        // GET, PUT, DELETE /suscripciones/{id}
	http.HandleFunc("/suscripciones/usuario/", manejadorSuscripcion.ManejarListarPorUsuario) // GET /suscripciones/usuario/{id_usuario}

	// --- Rutas de Pagos ---
	http.HandleFunc("/pagos", manejadorPago.ManejarPagos)                    // POST
	http.HandleFunc("/pagos/", manejadorPago.ManejarPagoPorID)               // GET, DELETE, PUT/PATCH (405) /pagos/{id}
	http.HandleFunc("/pagos/suscripcion/", manejadorPago.ManejarListarPorSuscripcion) // GET /pagos/suscripcion/{id_suscripcion}

	// 8. Iniciar el Servidor HTTP en el puerto 8080
	log.Println("Servidor escuchando en http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}