package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
)

// Tipos de cuenta: los consumidores gestionan sus suscripciones y las
// empresas administran el catálogo de servicios.
const (
	TipoConsumidor = "consumidor"
	TipoEmpresa    = "empresa"
)

// ErrCredenciales se devuelve cuando el email o la contraseña no coinciden.
// Es el mismo mensaje en ambos casos para no revelar qué emails existen.
var ErrCredenciales = fmt.Errorf("email o contraseña incorrectos")

func hashearPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func validarCorreo(correo string) error {
	patron := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(patron)
	if !regex.MatchString(correo) {
		return fmt.Errorf("el formato del correo electrónico no es válido")
	}
	return nil
}

func validarCamposUsuario(nombre, email, password string) error {
	if strings.TrimSpace(nombre) == "" ||
		strings.TrimSpace(email) == "" ||
		strings.TrimSpace(password) == "" {
		return fmt.Errorf("los campos nombre, email y password son obligatorios")
	}
	return validarCorreo(email)
}

// ocultarPassword evita devolver la contraseña al cliente
func ocultarPassword(u sqlc.Usuario) sqlc.Usuario {
	u.Password = ""
	return u
}

type ServicioUsuario struct {
	consultas *sqlc.Queries
}

func NuevoServicioUsuario(q *sqlc.Queries) *ServicioUsuario {
	return &ServicioUsuario{consultas: q}
}

// Crear aplica las validaciones, verifica que el correo no esté registrado
// y guarda la contraseña hasheada con bcrypt
func (s *ServicioUsuario) Crear(ctx context.Context, entrada sqlc.CreateUsuarioParams) (sqlc.Usuario, error) {
	if err := validarCamposUsuario(entrada.Nombre, entrada.Email, entrada.Password); err != nil {
		return sqlc.Usuario{}, err
	}

	if entrada.Tipo == "" {
		entrada.Tipo = TipoConsumidor
	}
	if entrada.Tipo != TipoConsumidor && entrada.Tipo != TipoEmpresa {
		return sqlc.Usuario{}, fmt.Errorf("el tipo de cuenta '%s' no es válido", entrada.Tipo)
	}

	_, err := s.consultas.GetUsuarioByEmail(ctx, entrada.Email)
	if err == nil {
		return sqlc.Usuario{}, fmt.Errorf("el correo electrónico ya está registrado")
	} else if err != sql.ErrNoRows {
		return sqlc.Usuario{}, err
	}

	hash, err := hashearPassword(entrada.Password)
	if err != nil {
		return sqlc.Usuario{}, err
	}
	entrada.Password = hash

	nuevo, err := s.consultas.CreateUsuario(ctx, entrada)
	if err != nil {
		return sqlc.Usuario{}, err
	}

	return ocultarPassword(nuevo), nil
}

// Autenticar verifica el email y la contraseña y devuelve el usuario
func (s *ServicioUsuario) Autenticar(ctx context.Context, email, password string) (sqlc.Usuario, error) {
	usr, err := s.consultas.GetUsuarioByEmail(ctx, strings.TrimSpace(email))
	if err == sql.ErrNoRows {
		return sqlc.Usuario{}, ErrCredenciales
	} else if err != nil {
		return sqlc.Usuario{}, err
	}

	if bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(password)) != nil {
		return sqlc.Usuario{}, ErrCredenciales
	}
	return ocultarPassword(usr), nil
}

// ObtenerPorID busca un usuario por su identificador único
func (s *ServicioUsuario) ObtenerPorID(ctx context.Context, id int32) (sqlc.Usuario, error) {
	usr, err := s.consultas.GetUsuario(ctx, id)
	if err == sql.ErrNoRows {
		return sqlc.Usuario{}, fmt.Errorf("usuario no encontrado")
	} else if err != nil {
		return sqlc.Usuario{}, err
	}
	return ocultarPassword(usr), nil
}

// Listar recupera la lista completa de usuarios
func (s *ServicioUsuario) Listar(ctx context.Context) ([]sqlc.Usuario, error) {
	lista, err := s.consultas.ListUsuarios(ctx)
	if err != nil {
		return nil, err
	}

	resultado := make([]sqlc.Usuario, len(lista))
	for i, u := range lista {
		resultado[i] = ocultarPassword(u)
	}
	return resultado, nil
}

// Actualizar modifica los datos del usuario cuidando la existencia y la unicidad de email
func (s *ServicioUsuario) Actualizar(ctx context.Context, id int32, entrada sqlc.UpdateUsuarioParams) (sqlc.Usuario, error) {
	if err := validarCamposUsuario(entrada.Nombre, entrada.Email, entrada.Password); err != nil {
		return sqlc.Usuario{}, err
	}

	usuarioExistente, err := s.consultas.GetUsuario(ctx, id)
	if err == sql.ErrNoRows {
		return sqlc.Usuario{}, fmt.Errorf("usuario no encontrado")
	} else if err != nil {
		return sqlc.Usuario{}, err
	}

	if entrada.Email != usuarioExistente.Email {
		existente, err := s.consultas.GetUsuarioByEmail(ctx, entrada.Email)
		if err == nil && existente.IDUsuario != id {
			return sqlc.Usuario{}, fmt.Errorf("el correo electrónico ya está en uso por otro usuario")
		} else if err != nil && err != sql.ErrNoRows {
			return sqlc.Usuario{}, err
		}
	}

	hash, err := hashearPassword(entrada.Password)
	if err != nil {
		return sqlc.Usuario{}, err
	}

	// el id siempre viene de la URL, no del body
	entrada.IDUsuario = id
	entrada.Password = hash
	if err := s.consultas.UpdateUsuario(ctx, entrada); err != nil {
		return sqlc.Usuario{}, err
	}

	return sqlc.Usuario{
		IDUsuario: id,
		Nombre:    entrada.Nombre,
		Email:     entrada.Email,
		Tipo:      usuarioExistente.Tipo,
		CreatedAt: usuarioExistente.CreatedAt,
	}, nil
}

// Eliminar quita un usuario previa comprobación de su existencia
func (s *ServicioUsuario) Eliminar(ctx context.Context, id int32) error {
	_, err := s.consultas.GetUsuario(ctx, id)
	if err == sql.ErrNoRows {
		return fmt.Errorf("usuario no encontrado")
	} else if err != nil {
		return err
	}

	return s.consultas.DeleteUsuario(ctx, id)
}

type ManejadorUsuario struct {
	servicio *ServicioUsuario
}

func NuevoManejadorUsuario(s *ServicioUsuario) *ManejadorUsuario {
	return &ManejadorUsuario{servicio: s}
}

// ManejarUsuarios atiende las peticiones a /usuarios (GET y POST)
func (h *ManejadorUsuario) ManejarUsuarios(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.obtenerTodos(w, r)
	case http.MethodPost:
		h.crear(w, r)
	default:
		http.Error(w, "Método HTTP no permitido", http.StatusMethodNotAllowed)
	}
}

// ManejarUsuarioPorID atiende las peticiones a /usuarios/{id} (GET, PUT y DELETE)
func (h *ManejadorUsuario) ManejarUsuarioPorID(w http.ResponseWriter, r *http.Request) {
	id, ok := h.extraerID(w, r)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.obtenerPorID(w, r, id)
	case http.MethodPut:
		if h.esLaPropiaCuenta(w, r, id) {
			h.actualizar(w, r, id)
		}
	case http.MethodDelete:
		if h.esLaPropiaCuenta(w, r, id) {
			h.eliminar(w, r, id)
		}
	default:
		http.Error(w, "Método HTTP no permitido", http.StatusMethodNotAllowed)
	}
}

// esLaPropiaCuenta verifica que el usuario de la sesión sea el de la URL
func (h *ManejadorUsuario) esLaPropiaCuenta(w http.ResponseWriter, r *http.Request, id int32) bool {
	usr, ok := UsuarioDeContexto(r.Context())
	if !ok || usr.IDUsuario != id {
		http.Error(w, "Solo podés modificar o eliminar tu propia cuenta", http.StatusForbidden)
		return false
	}
	return true
}

func (h *ManejadorUsuario) obtenerTodos(w http.ResponseWriter, r *http.Request) {
	usuarios, err := h.servicio.Listar(r.Context())
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusOK, usuarios)
}

func (h *ManejadorUsuario) crear(w http.ResponseWriter, r *http.Request) {
	var entrada sqlc.CreateUsuarioParams
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		http.Error(w, "El formato JSON enviado no es válido", http.StatusBadRequest)
		return
	}

	creado, err := h.servicio.Crear(r.Context(), entrada)
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusCreated, creado)
}

func (h *ManejadorUsuario) obtenerPorID(w http.ResponseWriter, r *http.Request, id int32) {
	usuario, err := h.servicio.ObtenerPorID(r.Context(), id)
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusOK, usuario)
}

func (h *ManejadorUsuario) actualizar(w http.ResponseWriter, r *http.Request, id int32) {
	var entrada sqlc.UpdateUsuarioParams
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		http.Error(w, "El formato JSON enviado no es válido", http.StatusBadRequest)
		return
	}

	actualizado, err := h.servicio.Actualizar(r.Context(), id, entrada)
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusOK, actualizado)
}

func (h *ManejadorUsuario) eliminar(w http.ResponseWriter, r *http.Request, id int32) {
	if err := h.servicio.Eliminar(r.Context(), id); err != nil {
		h.responderError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ManejadorUsuario) extraerID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	partes := strings.Split(r.URL.Path, "/")
	if len(partes) < 3 || partes[2] == "" {
		http.Error(w, "ID de usuario no especificado", http.StatusBadRequest)
		return 0, false
	}

	idNum, err := strconv.Atoi(partes[2])
	if err != nil {
		http.Error(w, "El ID de usuario debe ser numérico", http.StatusBadRequest)
		return 0, false
	}

	return int32(idNum), true
}

func (h *ManejadorUsuario) responderError(w http.ResponseWriter, err error) {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "no encontrado"):
		http.Error(w, msg, http.StatusNotFound)
	case strings.Contains(msg, "obligatorios") || strings.Contains(msg, "inválido") || strings.Contains(msg, "válido"):
		http.Error(w, msg, http.StatusBadRequest)
	case strings.Contains(msg, "registrado") || strings.Contains(msg, "uso"):
		http.Error(w, msg, http.StatusConflict)
	default:
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

func (h *ManejadorUsuario) responderJSON(w http.ResponseWriter, estado int, datos any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(datos)
}