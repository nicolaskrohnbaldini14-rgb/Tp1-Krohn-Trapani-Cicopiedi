package main

import (
	"context"
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

	// 4. Catálogo de servicios precargado (no duplica si ya existe)
	if err := servicioServicio.CargarCatalogoInicial(context.Background()); err != nil {
		log.Fatalf("Error al precargar el catálogo de servicios: %v", err)
	}
	log.Println("Catálogo de servicios precargado")

	// las suscripciones activas que ya vencieron se renuevan por el mismo plazo
	if renovadas, err := servicioSuscripcion.RenovarVencidas(context.Background()); err != nil {
		log.Fatalf("Error al renovar suscripciones vencidas: %v", err)
	} else if renovadas > 0 {
		log.Printf("Suscripciones renovadas: %d", renovadas)
	}

	// 5. Handlers
	sesiones := logic.NuevasSesiones()
	manejadorSesion := logic.NuevoManejadorSesion(sesiones, servicioUsuario)
	manejadorUsuario := logic.NuevoManejadorUsuario(servicioUsuario)
	manejadorServicio := logic.NuevoManejadorServicio(servicioServicio)
	manejadorSuscripcion := logic.NuevoManejadorSuscripcion(servicioSuscripcion)
	manejadorPago := logic.NuevoManejadorPago(servicioPago)

	// Proteger exige sesión iniciada solo para los métodos indicados
	proteger := manejadorSesion.Proteger
	get, post, put, del := http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete

	// 6. Rutas
	http.HandleFunc("/sesion", manejadorSesion.ManejarSesion) // GET (quién soy), POST (login), DELETE (logout)

	http.HandleFunc("/usuarios", manejadorUsuario.ManejarUsuarios)                                    // GET, POST (registro público)
	http.HandleFunc("/usuarios/", proteger([]string{put, del}, manejadorUsuario.ManejarUsuarioPorID)) // GET, PUT, DELETE (solo la propia cuenta)

	http.HandleFunc("/servicios", manejadorServicio.ManejarServicios)      // GET (catálogo precargado, solo lectura)
	http.HandleFunc("/servicios/", manejadorServicio.ManejarServicioPorID) // GET

	http.HandleFunc("/suscripciones", proteger([]string{post}, manejadorSuscripcion.ManejarSuscripciones))                    // POST
	http.HandleFunc("/suscripciones/", proteger([]string{get, post, put, del}, manejadorSuscripcion.ManejarSuscripcionPorID)) // GET, PUT, DELETE y POST .../password (propias)
	http.HandleFunc("/suscripciones/usuario/", proteger([]string{get}, manejadorSuscripcion.ManejarSuscripcionesPorUsuario))  // GET (propias)

	http.HandleFunc("/pagos", proteger([]string{post}, manejadorPago.ManejarPagos))                           // POST
	http.HandleFunc("/pagos/", proteger([]string{get, del}, manejadorPago.ManejarPagoPorID))                  // GET, DELETE (propios)
	http.HandleFunc("/pagos/suscripcion/", proteger([]string{get}, manejadorPago.ManejarPagosPorSuscripcion)) // GET (propios)

	// 7. Frontend (TP4): HTML, CSS y JS estáticos que consumen la API.
	// Las rutas de la API son más específicas, así que tienen prioridad sobre "/".
	// "no-cache" hace que el navegador pregunte siempre si hay una versión nueva
	// (si no cambió, el servidor responde 304), así no queda usando un
	// app.js o styles.css viejo después de un cambio.
	estaticos := http.FileServer(http.Dir("./static"))
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		estaticos.ServeHTTP(w, r)
	}))

	// 8. Servidor HTTP
	log.Println("Servidor escuchando en http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Error al iniciar el servidor HTTP: %v", err)
	}
}
