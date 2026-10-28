package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
	"tp2-krohn-trapani-cicopiedi/logic"
	"tp2-krohn-trapani-cicopiedi/views"
)

// Web reúne los handlers que renderizan HTML con templ. Reemplazan a los
// handlers JSON del TP3: la vista se genera en el servidor.
type Web struct {
	usuarios      *logic.ServicioUsuario
	servicios     *logic.ServicioServicio
	suscripciones *logic.ServicioSuscripcion
	sesiones      *Sesiones
}

func NuevoWeb(q *sqlc.Queries, sesiones *Sesiones) *Web {
	return &Web{
		usuarios:      logic.NuevoServicioUsuario(q),
		servicios:     logic.NuevoServicioServicio(q),
		suscripciones: logic.NuevoServicioSuscripcion(q),
		sesiones:      sesiones,
	}
}

// manejadorConUsuario es un handler que recibe al usuario que inició sesión
type manejadorConUsuario func(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario)

// Registrar asocia cada ruta con su handler
func (h *Web) Registrar(mux *http.ServeMux) {
	// Páginas públicas
	mux.HandleFunc("GET /{$}", h.home)
	mux.HandleFunc("GET /login", h.formLogin)
	mux.HandleFunc("POST /login", h.login)
	mux.HandleFunc("GET /registro", h.formRegistro(false))
	mux.HandleFunc("POST /registro", h.registrar(false))
	mux.HandleFunc("GET /empresas/registro", h.formRegistro(true))
	mux.HandleFunc("POST /empresas/registro", h.registrar(true))
	mux.HandleFunc("POST /logout", h.logout)

	// Consumidores: sus propias suscripciones
	mux.HandleFunc("GET /cuenta", h.soloTipo(logic.TipoConsumidor, h.cuenta))
	mux.HandleFunc("GET /cuenta/suscripciones", h.soloTipo(logic.TipoConsumidor, h.misSuscripciones))
	mux.HandleFunc("GET /cuenta/suscripciones/nueva", h.soloTipo(logic.TipoConsumidor, h.nuevaSuscripcion))
	mux.HandleFunc("POST /cuenta/suscripciones", h.soloTipo(logic.TipoConsumidor, h.crearSuscripcion))
	mux.HandleFunc("POST /cuenta/suscripciones/{id}/eliminar", h.soloTipo(logic.TipoConsumidor, h.eliminarSuscripcion))

	// Empresas: catálogo de servicios
	mux.HandleFunc("GET /empresa", h.soloTipo(logic.TipoEmpresa, h.panelEmpresa))
	mux.HandleFunc("POST /empresa/servicios", h.soloTipo(logic.TipoEmpresa, h.crearServicio))
	mux.HandleFunc("POST /empresa/servicios/{id}/eliminar", h.soloTipo(logic.TipoEmpresa, h.eliminarServicio))

	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
}

// ---------------------------------------------------------------------------
// Sesión y control de acceso
// ---------------------------------------------------------------------------

// usuarioActual devuelve el usuario de la sesión, o nil si no ingresó
func (h *Web) usuarioActual(r *http.Request) *sqlc.Usuario {
	id, ok := h.sesiones.Usuario(r)
	if !ok {
		return nil
	}
	usr, err := h.usuarios.ObtenerPorID(r.Context(), id)
	if err != nil {
		return nil
	}
	return &usr
}

func vistaSesion(usr *sqlc.Usuario) *views.Sesion {
	if usr == nil {
		return nil
	}
	return &views.Sesion{Nombre: usr.Nombre, Tipo: usr.Tipo}
}

// destino es la página a la que va cada tipo de cuenta después de ingresar
func destino(usr sqlc.Usuario) string {
	if usr.Tipo == logic.TipoEmpresa {
		return "/empresa"
	}
	return "/cuenta"
}

// soloTipo exige una sesión iniciada con el tipo de cuenta indicado.
// Sin sesión redirige al login; con otro tipo de cuenta responde 403.
func (h *Web) soloTipo(tipo string, siguiente manejadorConUsuario) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usr := h.usuarioActual(r)
		if usr == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		if usr.Tipo != tipo {
			mensaje := "Esta sección es solo para empresas."
			if tipo == logic.TipoConsumidor {
				mensaje = "Esta sección es solo para consumidores."
			}
			render(w, r, http.StatusForbidden, views.ErrorPage(vistaSesion(usr), "Acceso no permitido", mensaje))
			return
		}
		siguiente(w, r, *usr)
	}
}

// ---------------------------------------------------------------------------
// Páginas públicas
// ---------------------------------------------------------------------------

func (h *Web) home(w http.ResponseWriter, r *http.Request) {
	servicios, err := h.servicios.Listar(r.Context())
	if err != nil {
		errorInterno(w, err)
		return
	}
	render(w, r, http.StatusOK, views.HomePage(vistaSesion(h.usuarioActual(r)), servicios))
}

func (h *Web) formLogin(w http.ResponseWriter, r *http.Request) {
	if usr := h.usuarioActual(r); usr != nil {
		http.Redirect(w, r, destino(*usr), http.StatusSeeOther)
		return
	}
	render(w, r, http.StatusOK, views.LoginPage("", ""))
}

func (h *Web) login(w http.ResponseWriter, r *http.Request) {
	if !parsearFormulario(w, r) {
		return
	}
	email := strings.TrimSpace(r.FormValue("email"))

	usr, err := h.usuarios.Autenticar(r.Context(), email, r.FormValue("password"))
	if err != nil {
		estado, mensaje := describirError(err)
		render(w, r, estado, views.LoginPage(email, mensaje))
		return
	}

	h.iniciarSesionYRedirigir(w, r, usr)
}

func (h *Web) formRegistro(empresa bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if usr := h.usuarioActual(r); usr != nil {
			http.Redirect(w, r, destino(*usr), http.StatusSeeOther)
			return
		}
		render(w, r, http.StatusOK, views.RegistroPage(views.DatosRegistro{Empresa: empresa}))
	}
}

// registrar crea la cuenta y deja la sesión iniciada
func (h *Web) registrar(empresa bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !parsearFormulario(w, r) {
			return
		}

		tipo := logic.TipoConsumidor
		if empresa {
			tipo = logic.TipoEmpresa
		}
		datos := views.DatosRegistro{
			Empresa: empresa,
			Nombre:  strings.TrimSpace(r.FormValue("nombre")),
			Email:   strings.TrimSpace(r.FormValue("email")),
		}

		usr, err := h.usuarios.Crear(r.Context(), sqlc.CreateUsuarioParams{
			Nombre:   datos.Nombre,
			Email:    datos.Email,
			Password: r.FormValue("password"),
			Tipo:     tipo,
		})
		if err != nil {
			var estado int
			estado, datos.Error = describirError(err)
			render(w, r, estado, views.RegistroPage(datos))
			return
		}

		h.iniciarSesionYRedirigir(w, r, usr)
	}
}

func (h *Web) iniciarSesionYRedirigir(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario) {
	if err := h.sesiones.Iniciar(w, usr.IDUsuario); err != nil {
		errorInterno(w, err)
		return
	}
	http.Redirect(w, r, destino(usr), http.StatusSeeOther)
}

func (h *Web) logout(w http.ResponseWriter, r *http.Request) {
	h.sesiones.Cerrar(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ---------------------------------------------------------------------------
// Consumidores
// ---------------------------------------------------------------------------

func (h *Web) cuenta(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario) {
	suscripciones, err := h.suscripciones.ListarPorUsuario(r.Context(), usr.IDUsuario)
	if err != nil {
		errorInterno(w, err)
		return
	}
	render(w, r, http.StatusOK, views.CuentaPage(vistaSesion(&usr), suscripciones))
}

func (h *Web) misSuscripciones(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario) {
	h.renderMisSuscripciones(w, r, usr, http.StatusOK, "")
}

func (h *Web) renderMisSuscripciones(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario, estado int, mensajeError string) {
	suscripciones, err := h.suscripciones.ListarPorUsuario(r.Context(), usr.IDUsuario)
	if err != nil {
		errorInterno(w, err)
		return
	}
	render(w, r, estado, views.MisSuscripcionesPage(vistaSesion(&usr), suscripciones, mensajeError))
}

func (h *Web) nuevaSuscripcion(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario) {
	h.renderNuevaSuscripcion(w, r, usr, http.StatusOK, views.FormSuscripcion{Estado: "Activa"}, "")
}

func (h *Web) renderNuevaSuscripcion(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario, estado int, form views.FormSuscripcion, mensajeError string) {
	servicios, err := h.servicios.Listar(r.Context())
	if err != nil {
		errorInterno(w, err)
		return
	}
	render(w, r, estado, views.NuevaSuscripcionPage(vistaSesion(&usr), servicios, form, mensajeError))
}

// crearSuscripcion guarda la suscripción a nombre del usuario de la sesión
// y aplica Post/Redirect/Get hacia su lista
func (h *Web) crearSuscripcion(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario) {
	if !parsearFormulario(w, r) {
		return
	}

	form := views.FormSuscripcion{
		IDServicio:       leerEntero(r.FormValue("id_servicio")),
		Monto:            strings.TrimSpace(r.FormValue("monto")),
		FechaInicio:      r.FormValue("fecha_inicio"),
		FechaVencimiento: r.FormValue("fecha_vencimiento"),
		UsuarioCuenta:    strings.TrimSpace(r.FormValue("usuario_cuenta")),
		Estado:           r.FormValue("estado"),
	}

	// los valores inválidos quedan en cero y los rechaza la capa de lógica
	_, err := h.suscripciones.Crear(r.Context(), sqlc.CreateSubscripcionParams{
		IDUsuario:        usr.IDUsuario,
		IDServicio:       form.IDServicio,
		Monto:            form.Monto,
		FechaInicio:      leerFecha(form.FechaInicio),
		FechaVencimiento: leerFecha(form.FechaVencimiento),
		UsuarioCuenta:    form.UsuarioCuenta,
		PasswordCuenta:   r.FormValue("password_cuenta"),
		Estado:           form.Estado,
	})
	if err != nil {
		estado, mensaje := describirError(err)
		h.renderNuevaSuscripcion(w, r, usr, estado, form, mensaje)
		return
	}

	http.Redirect(w, r, "/cuenta/suscripciones", http.StatusSeeOther)
}

func (h *Web) eliminarSuscripcion(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario) {
	id, ok := leerID(r)
	if !ok {
		h.renderMisSuscripciones(w, r, usr, http.StatusBadRequest, "El ID debe ser un valor numérico")
		return
	}

	if err := h.suscripciones.EliminarDeUsuario(r.Context(), id, usr.IDUsuario); err != nil {
		estado, mensaje := describirError(err)
		h.renderMisSuscripciones(w, r, usr, estado, mensaje)
		return
	}

	http.Redirect(w, r, "/cuenta/suscripciones", http.StatusSeeOther)
}

// ---------------------------------------------------------------------------
// Empresas
// ---------------------------------------------------------------------------

func (h *Web) panelEmpresa(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario) {
	h.renderPanelEmpresa(w, r, usr, http.StatusOK, "", "")
}

func (h *Web) renderPanelEmpresa(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario, estado int, nombre, mensajeError string) {
	servicios, err := h.servicios.Listar(r.Context())
	if err != nil {
		errorInterno(w, err)
		return
	}
	render(w, r, estado, views.EmpresaPage(vistaSesion(&usr), servicios, nombre, mensajeError))
}

func (h *Web) crearServicio(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario) {
	if !parsearFormulario(w, r) {
		return
	}

	nombre := strings.TrimSpace(r.FormValue("nombre"))
	_, err := h.servicios.Crear(r.Context(), sqlc.CreateServicioParams{
		Nombre:    nombre,
		Categoria: strings.TrimSpace(r.FormValue("categoria")),
	})
	if err != nil {
		estado, mensaje := describirError(err)
		h.renderPanelEmpresa(w, r, usr, estado, nombre, mensaje)
		return
	}

	http.Redirect(w, r, "/empresa", http.StatusSeeOther)
}

func (h *Web) eliminarServicio(w http.ResponseWriter, r *http.Request, usr sqlc.Usuario) {
	id, ok := leerID(r)
	if !ok {
		h.renderPanelEmpresa(w, r, usr, http.StatusBadRequest, "", "El ID debe ser un valor numérico")
		return
	}

	if err := h.servicios.Eliminar(r.Context(), id); err != nil {
		estado, mensaje := describirError(err)
		h.renderPanelEmpresa(w, r, usr, estado, "", mensaje)
		return
	}

	http.Redirect(w, r, "/empresa", http.StatusSeeOther)
}

// ---------------------------------------------------------------------------
// Auxiliares
// ---------------------------------------------------------------------------

func render(w http.ResponseWriter, r *http.Request, estado int, componente templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(estado)
	if err := componente.Render(r.Context(), w); err != nil {
		log.Printf("Error al renderizar la página: %v", err)
	}
}

func parsearFormulario(w http.ResponseWriter, r *http.Request) bool {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "No se pudieron leer los datos del formulario", http.StatusBadRequest)
		return false
	}
	return true
}

func errorInterno(w http.ResponseWriter, err error) {
	log.Printf("Error interno: %v", err)
	http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
}

// describirError traduce los errores de la capa de lógica a un código HTTP y
// al mensaje que se le muestra al usuario. Los errores inesperados no se
// muestran tal cual para no filtrar detalles internos.
func describirError(err error) (int, string) {
	if errors.Is(err, logic.ErrCredenciales) {
		return http.StatusUnauthorized, err.Error()
	}

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
		return http.StatusNotFound, msg
	case contiene("ya está registrado", "en uso", "ya existe", "ya posee", "tiene suscripciones"):
		return http.StatusConflict, msg
	case contiene("obligatori", "válid", "mayor a cero", "anterior", "posterior", "no se permite"):
		return http.StatusBadRequest, msg
	default:
		log.Printf("Error interno: %v", err)
		return http.StatusInternalServerError, "Error interno del servidor"
	}
}

func leerID(r *http.Request) (int32, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, false
	}
	return int32(id), true
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
