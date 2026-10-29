package logic

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"sync"
	"time"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
)

const (
	nombreCookie   = "sesion"
	duracionSesion = 24 * time.Hour
)

// Sesiones guarda en memoria qué usuario corresponde a cada token de sesión.
// El navegador solo recibe el token aleatorio en una cookie HttpOnly.
// Al reiniciar el servidor las sesiones se pierden y hay que volver a ingresar.
type Sesiones struct {
	mu    sync.RWMutex
	datos map[string]sesion
}

type sesion struct {
	idUsuario int32
	vence     time.Time
}

func NuevasSesiones() *Sesiones {
	return &Sesiones{datos: make(map[string]sesion)}
}

// Iniciar crea una sesión para el usuario y envía la cookie
func (s *Sesiones) Iniciar(w http.ResponseWriter, idUsuario int32) error {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return err
	}
	token := hex.EncodeToString(bytes)
	vence := time.Now().Add(duracionSesion)

	s.mu.Lock()
	s.datos[token] = sesion{idUsuario: idUsuario, vence: vence}
	s.mu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     nombreCookie,
		Value:    token,
		Path:     "/",
		Expires:  vence,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// Usuario devuelve el id del usuario de la sesión actual, si hay una válida
func (s *Sesiones) Usuario(r *http.Request) (int32, bool) {
	cookie, err := r.Cookie(nombreCookie)
	if err != nil {
		return 0, false
	}

	s.mu.RLock()
	ses, ok := s.datos[cookie.Value]
	s.mu.RUnlock()

	if !ok || time.Now().After(ses.vence) {
		return 0, false
	}
	return ses.idUsuario, true
}

// Cerrar elimina la sesión y le indica al navegador que borre la cookie
func (s *Sesiones) Cerrar(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(nombreCookie); err == nil {
		s.mu.Lock()
		delete(s.datos, cookie.Value)
		s.mu.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:     nombreCookie,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// ---------------------------------------------------------------------------
// Usuario de la sesión dentro del contexto de la petición
// ---------------------------------------------------------------------------

type claveUsuario struct{}

// UsuarioDeContexto devuelve el usuario que dejó Proteger en el contexto
func UsuarioDeContexto(ctx context.Context) (sqlc.Usuario, bool) {
	usr, ok := ctx.Value(claveUsuario{}).(sqlc.Usuario)
	return usr, ok
}

// ---------------------------------------------------------------------------
// Handler de /sesion y middleware de protección
// ---------------------------------------------------------------------------

type ManejadorSesion struct {
	sesiones *Sesiones
	usuarios *ServicioUsuario
}

func NuevoManejadorSesion(sesiones *Sesiones, usuarios *ServicioUsuario) *ManejadorSesion {
	return &ManejadorSesion{sesiones: sesiones, usuarios: usuarios}
}

type credenciales struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ManejarSesion atiende /sesion: GET (quién soy), POST (login) y DELETE (logout)
func (h *ManejadorSesion) ManejarSesion(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		usr, ok := h.usuarioActual(r)
		if !ok {
			http.Error(w, "No hay una sesión iniciada", http.StatusUnauthorized)
			return
		}
		responderJSON(w, http.StatusOK, usr)
	case http.MethodPost:
		h.login(w, r)
	case http.MethodDelete:
		h.sesiones.Cerrar(w, r)
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Método HTTP no permitido", http.StatusMethodNotAllowed)
	}
}

func (h *ManejadorSesion) login(w http.ResponseWriter, r *http.Request) {
	var entrada credenciales
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		http.Error(w, "El formato JSON enviado no es válido", http.StatusBadRequest)
		return
	}

	usr, err := h.usuarios.Autenticar(r.Context(), entrada.Email, entrada.Password)
	if errors.Is(err, ErrCredenciales) {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}

	if err := h.sesiones.Iniciar(w, usr.IDUsuario); err != nil {
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
		return
	}
	responderJSON(w, http.StatusOK, usr)
}

func (h *ManejadorSesion) usuarioActual(r *http.Request) (sqlc.Usuario, bool) {
	id, ok := h.sesiones.Usuario(r)
	if !ok {
		return sqlc.Usuario{}, false
	}
	usr, err := h.usuarios.ObtenerPorID(r.Context(), id)
	if err != nil {
		return sqlc.Usuario{}, false
	}
	return usr, true
}

// Proteger exige una sesión iniciada para los métodos indicados y responde
// 401 si no la hay. El resto de los métodos pasa directo al handler (por
// ejemplo, el GET público de un usuario).
func (h *ManejadorSesion) Proteger(metodos []string, siguiente http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !slices.Contains(metodos, r.Method) {
			siguiente(w, r)
			return
		}

		usr, ok := h.usuarioActual(r)
		if !ok {
			http.Error(w, "Tenés que iniciar sesión", http.StatusUnauthorized)
			return
		}

		siguiente(w, r.WithContext(context.WithValue(r.Context(), claveUsuario{}, usr)))
	}
}

func responderJSON(w http.ResponseWriter, estado int, datos any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(datos)
}
