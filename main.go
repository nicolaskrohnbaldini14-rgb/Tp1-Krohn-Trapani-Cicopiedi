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
	sesiones := logic.NuevasSesiones()
	manejadorSesion := logic.NuevoManejadorSesion(sesiones, servicioUsuario)
	manejadorUsuario := logic.NuevoManejadorUsuario(servicioUsuario)
	manejadorServicio := logic.NuevoManejadorServicio(servicioServicio)
	manejadorSuscripcion := logic.NuevoManejadorSuscripcion(servicioSuscripcion)
	manejadorPago := logic.NuevoManejadorPago(servicioPago)

	// Proteger exige sesión (y tipo de cuenta) solo para los métodos indicados
	proteger := manejadorSesion.Proteger
	empresa, consumidor, cualquiera := logic.TipoEmpresa, logic.TipoConsumidor, ""
	get, post, put, del := http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete

	// 5. Rutas
	http.HandleFunc("/sesion", manejadorSesion.ManejarSesion) // GET (quién soy), POST (login), DELETE (logout)

	http.HandleFunc("/usuarios", manejadorUsuario.ManejarUsuarios)                                                // GET, POST (registro público)
	http.HandleFunc("/usuarios/", proteger(cualquiera, []string{put, del}, manejadorUsuario.ManejarUsuarioPorID)) // GET, PUT, DELETE (solo la propia cuenta)

	http.HandleFunc("/servicios", proteger(empresa, []string{post}, manejadorServicio.ManejarServicios))          // GET público, POST empresas
	http.HandleFunc("/servicios/", proteger(empresa, []string{put, del}, manejadorServicio.ManejarServicioPorID)) // GET público, PUT y DELETE empresas

	http.HandleFunc("/suscripciones", proteger(consumidor, []string{post}, manejadorSuscripcion.ManejarSuscripciones))                   // POST
	http.HandleFunc("/suscripciones/", proteger(consumidor, []string{get, put, del}, manejadorSuscripcion.ManejarSuscripcionPorID))      // GET, PUT, DELETE (propias)
	http.HandleFunc("/suscripciones/usuario/", proteger(consumidor, []string{get}, manejadorSuscripcion.ManejarSuscripcionesPorUsuario)) // GET (propias)

	http.HandleFunc("/pagos", proteger(consumidor, []string{post}, manejadorPago.ManejarPagos))                           // POST
	http.HandleFunc("/pagos/", proteger(consumidor, []string{get, del}, manejadorPago.ManejarPagoPorID))                  // GET, DELETE (propios)
	http.HandleFunc("/pagos/suscripcion/", proteger(consumidor, []string{get}, manejadorPago.ManejarPagosPorSuscripcion)) // GET (propios)

	// 6. Frontend (TP4): HTML, CSS y JS estáticos que consumen la API.
	// Las rutas de la API son más específicas, así que tienen prioridad sobre "/".
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// 7. Servidor HTTP
	log.Println("Servidor escuchando en http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}
