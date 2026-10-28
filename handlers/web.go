package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
	"tp2-krohn-trapani-cicopiedi/logic"
	"tp2-krohn-trapani-cicopiedi/views"
)

const tituloPagina = "Suscripciones+"

// Web reúne los handlers que renderizan HTML con templ. Reemplazan a los
// handlers JSON del TP3: la vista se genera en el servidor.
type Web struct {
	usuarios      *logic.ServicioUsuario
	servicios     *logic.ServicioServicio
	suscripciones *logic.ServicioSuscripcion
	pagos         *logic.ServicioPago
}

func NuevoWeb(q *sqlc.Queries) *Web {
	return &Web{
		usuarios:      logic.NuevoServicioUsuario(q),
		servicios:     logic.NuevoServicioServicio(q),
		suscripciones: logic.NuevoServicioSuscripcion(q),
		pagos:         logic.NuevoServicioPago(q),
	}
}

// Registrar asocia cada ruta con su handler
func (h *Web) Registrar(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.inicio)

	mux.HandleFunc("POST /usuarios", h.crearUsuario)
	mux.HandleFunc("POST /usuarios/{id}/eliminar", h.eliminarUsuario)

	mux.HandleFunc("POST /servicios", h.crearServicio)
	mux.HandleFunc("POST /servicios/{id}/eliminar", h.eliminarServicio)

	mux.HandleFunc("POST /suscripciones", h.crearSuscripcion)
	mux.HandleFunc("POST /suscripciones/{id}/eliminar", h.eliminarSuscripcion)

	mux.HandleFunc("POST /pagos", h.crearPago)
	mux.HandleFunc("POST /pagos/{id}/eliminar", h.eliminarPago)

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
}

// ---------------------------------------------------------------------------
// Página principal
// ---------------------------------------------------------------------------

func (h *Web) inicio(w http.ResponseWriter, r *http.Request) {
	h.renderInicio(w, r, http.StatusOK, "")
}

// renderInicio obtiene todos los registros con sqlc y renderiza IndexPage
func (h *Web) renderInicio(w http.ResponseWriter, r *http.Request, estado int, mensajeError string) {
	ctx := r.Context()
	datos := views.DatosInicio{Error: mensajeError}

	var err error
	if datos.Usuarios, err = h.usuarios.Listar(ctx); err != nil {
		h.errorInterno(w, err)
		return
	}
	if datos.Servicios, err = h.servicios.Listar(ctx); err != nil {
		h.errorInterno(w, err)
		return
	}
	if datos.Suscripciones, err = h.suscripciones.Listar(ctx); err != nil {
		h.errorInterno(w, err)
		return
	}
	if datos.Pagos, err = h.pagos.ListarTodos(ctx); err != nil {
		h.errorInterno(w, err)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(estado)
	if err := views.IndexPage(tituloPagina, datos).Render(ctx, w); err != nil {
		log.Printf("Error al renderizar la página: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Altas: parsean el formulario, guardan con sqlc (vía la capa de lógica) y
// redirigen a "/" con 303 See Other (patrón Post/Redirect/Get).
// ---------------------------------------------------------------------------

func (h *Web) crearUsuario(w http.ResponseWriter, r *http.Request) {
	if !h.parsearFormulario(w, r) {
		return
	}

	_, err := h.usuarios.Crear(r.Context(), sqlc.CreateUsuarioParams{
		Nombre:   strings.TrimSpace(r.FormValue("nombre")),
		Email:    strings.TrimSpace(r.FormValue("email")),
		Password: r.FormValue("password"),
	})
	h.redirigirOMostrarError(w, r, err)
}

func (h *Web) crearServicio(w http.ResponseWriter, r *http.Request) {
	if !h.parsearFormulario(w, r) {
		return
	}

	_, err := h.servicios.Crear(r.Context(), sqlc.CreateServicioParams{
		Nombre:    strings.TrimSpace(r.FormValue("nombre")),
		Categoria: strings.TrimSpace(r.FormValue("categoria")),
	})
	h.redirigirOMostrarError(w, r, err)
}

func (h *Web) crearSuscripcion(w http.ResponseWriter, r *http.Request) {
	if !h.parsearFormulario(w, r) {
		return
	}

	// los valores inválidos quedan en cero y los rechaza la capa de lógica
	_, err := h.suscripciones.Crear(r.Context(), sqlc.CreateSubscripcionParams{
		IDUsuario:        leerEntero(r.FormValue("id_usuario")),
		IDServicio:       leerEntero(r.FormValue("id_servicio")),
		Monto:            r.FormValue("monto"),
		FechaInicio:      leerFecha(r.FormValue("fecha_inicio")),
		FechaVencimiento: leerFecha(r.FormValue("fecha_vencimiento")),
		UsuarioCuenta:    strings.TrimSpace(r.FormValue("usuario_cuenta")),
		PasswordCuenta:   r.FormValue("password_cuenta"),
		Estado:           r.FormValue("estado"),
	})
	h.redirigirOMostrarError(w, r, err)
}

func (h *Web) crearPago(w http.ResponseWriter, r *http.Request) {
	if !h.parsearFormulario(w, r) {
		return
	}

	_, err := h.pagos.Crear(r.Context(), sqlc.CreatePagoParams{
		IDSuscripcion: leerEntero(r.FormValue("id_suscripcion")),
		Monto:         r.FormValue("monto"),
		FechaPago:     leerFecha(r.FormValue("fecha_pago")),
	})
	h.redirigirOMostrarError(w, r, err)
}

// ---------------------------------------------------------------------------
// Bajas: los formularios HTML no pueden enviar DELETE, por eso se usa
// POST /<entidades>/{id}/eliminar.
// ---------------------------------------------------------------------------

func (h *Web) eliminarUsuario(w http.ResponseWriter, r *http.Request) {
	h.eliminar(w, r, h.usuarios.Eliminar)
}

func (h *Web) eliminarServicio(w http.ResponseWriter, r *http.Request) {
	h.eliminar(w, r, h.servicios.Eliminar)
}

func (h *Web) eliminarSuscripcion(w http.ResponseWriter, r *http.Request) {
	h.eliminar(w, r, h.suscripciones.Eliminar)
}

func (h *Web) eliminarPago(w http.ResponseWriter, r *http.Request) {
	h.eliminar(w, r, h.pagos.Eliminar)
}

func (h *Web) eliminar(w http.ResponseWriter, r *http.Request, borrar func(ctx context.Context, id int32) error) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		h.renderInicio(w, r, http.StatusBadRequest, "El ID debe ser un valor numérico")
		return
	}
	h.redirigirOMostrarError(w, r, borrar(r.Context(), int32(id)))
}

// ---------------------------------------------------------------------------
// Auxiliares
// ---------------------------------------------------------------------------

func (h *Web) parsearFormulario(w http.ResponseWriter, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		h.renderInicio(w, r, http.StatusBadRequest, "No se pudieron leer los datos del formulario")
		return false
	}
	return true
}

// redirigirOMostrarError aplica PRG si la operación salió bien. Si falló,
// vuelve a renderizar la página con el mensaje y el código HTTP adecuado.
func (h *Web) redirigirOMostrarError(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	estado := codigoDeError(err)
	mensaje := err.Error()
	if estado == http.StatusInternalServerError {
		log.Printf("Error interno: %v", err)
		mensaje = "Error interno del servidor"
	}
	h.renderInicio(w, r, estado, mensaje)
}

func (h *Web) errorInterno(w http.ResponseWriter, err error) {
	log.Printf("Error al consultar la base de datos: %v", err)
	http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
}

// codigoDeError traduce los mensajes de la capa de lógica a códigos HTTP,
// con el mismo criterio que usaba la API del TP3.
func codigoDeError(err error) int {
	msg := err.Error()
	contiene := func(palabras ...string) bool {
		for _, p := range palabras {
			if strings.Contains(msg, p) {
				return true
			}
		}
		return false
	}

	switch {
	case contiene("no encontrado", "no encontrada", "no existe"):
		return http.StatusNotFound
	case contiene("ya está registrado", "en uso", "ya existe", "ya posee", "tiene suscripciones"):
		return http.StatusConflict
	case contiene("obligatori", "válid", "mayor a cero", "anterior", "posterior", "no se permite"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func leerEntero(valor string) int32 {
	n, err := strconv.Atoi(strings.TrimSpace(valor))
	if err != nil {
		return 0
	}
	return int32(n)
}

// leerFecha interpreta el formato de <input type="date"> (AAAA-MM-DD)
func leerFecha(valor string) time.Time {
	fecha, err := time.Parse("2006-01-02", strings.TrimSpace(valor))
	if err != nil {
		return time.Time{}
	}
	return fecha
}
