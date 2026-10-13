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
	"time"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
)

// EntradaUsuario representa la información enviada en el JSON
type EntradaUsuario struct {
	Nombre   string `json:"nombre"`
	Correo   string `json:"email"`
	Password string `json:"password"`
}

// RespuestaUsuario es el objeto retornado al cliente (sin la contraseña)
type RespuestaUsuario struct {
	IDUsuario int32     `json:"id_usuario"`
	Nombre    string    `json:"nombre"`
	Correo    string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func validarCorreo(correo string) error {
	patron := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	regex := regexp.MustCompile(patron)
	if !regex.MatchString(correo) {
		return fmt.Errorf("el formato del correo electrónico no es válido")
	}
	return nil
}

func validarCamposUsuario(datos EntradaUsuario) error {
	if strings.TrimSpace(datos.Nombre) == "" ||
		strings.TrimSpace(datos.Correo) == "" ||
		strings.TrimSpace(datos.Password) == "" {
		return fmt.Errorf("los campos nombre, email y password son obligatorios")
	}
	return validarCorreo(datos.Correo)
}

func mapearUsuarioARespuesta(u sqlc.Usuario) RespuestaUsuario {
	return RespuestaUsuario{
		IDUsuario: u.IDUsuario,
		Nombre:    u.Nombre,
		Correo:    [se quitó el vínculo sospechoso],
		CreatedAt: u.CreatedAt,
	}
}


type ServicioUsuario struct {
	consultas *sqlc.Queries
}

func NuevoServicioUsuario(q *sqlc.Queries) *ServicioUsuario {
	return &ServicioUsuario{consultas: q}
}

// Crear aplica las validaciones y verifica que el correo no esté registrado
func (s *ServicioUsuario) Crear(ctx context.Context, entrada EntradaUsuario) (RespuestaUsuario, error) {
	if err := validarCamposUsuario(entrada); err != nil {
		return RespuestaUsuario{}, err
	}

	_, err := s.consultas.GetUsuarioByEmail(ctx, entrada.Correo)
	if err == nil {
		return RespuestaUsuario{}, fmt.Errorf("el correo electrónico ya está registrado")
	} else if err != sql.ErrNoRows {
		return RespuestaUsuario{}, err
	}

	nuevo, err := s.consultas.CreateUsuario(ctx, sqlc.CreateUsuarioParams{
		Nombre:   entrada.Nombre,
		Email:    entrada.Correo,
		Password: entrada.Password,
	})
	if err != nil {
		return RespuestaUsuario{}, err
	}

	return mapearUsuarioARespuesta(nuevo), nil
}

// ObtenerPorID busca un usuario por su identificador único
func (s *ServicioUsuario) ObtenerPorID(ctx context.Context, id int32) (RespuestaUsuario, error) {
	usr, err := s.consultas.GetUsuario(ctx, id)
	if err == sql.ErrNoRows {
		return RespuestaUsuario{}, fmt.Errorf("usuario no encontrado")
	} else if err != nil {
		return RespuestaUsuario{}, err
	}
	return mapearUsuarioARespuesta(usr), nil
}

// Listar recupera la lista completa de usuarios
func (s *ServicioUsuario) Listar(ctx context.Context) ([]RespuestaUsuario, error) {
	lista, err := s.consultas.ListUsuarios(ctx)
	if err != nil {
		return nil, err
	}

	resultado := make([]RespuestaUsuario, len(lista))
	for i, u := range lista {
		resultado[i] = mapearUsuarioARespuesta(u)
	}
	return resultado, nil
}

// Actualizar modifica los datos del usuario cuidando la existencia y la unicidad de email
func (s *ServicioUsuario) Actualizar(ctx context.Context, id int32, entrada EntradaUsuario) (RespuestaUsuario, error) {
	if err := validarCamposUsuario(entrada); err != nil {
		return RespuestaUsuario{}, err
	}

	usuarioExistente, err := s.consultas.GetUsuario(ctx, id)
	if err == sql.ErrNoRows {
		return RespuestaUsuario{}, fmt.Errorf("usuario no encontrado")
	} else if err != nil {
		return RespuestaUsuario{}, err
	}

	if entrada.Correo != [se quitó el vínculo sospechoso] {
		existente, err := s.consultas.GetUsuarioByEmail(ctx, entrada.Correo)
		if err == nil && existente.IDUsuario != id {
			return RespuestaUsuario{}, fmt.Errorf("el correo electrónico ya está en uso por otro usuario")
		} else if err != nil && err != sql.ErrNoRows {
			return RespuestaUsuario{}, err
		}
	}

	err = s.consultas.UpdateUsuario(ctx, sqlc.UpdateUsuarioParams{
		IDUsuario: id,
		Nombre:    entrada.Nombre,
		Email:     entrada.Correo,
		Password:  entrada.Password,
	})
	if err != nil {
		return RespuestaUsuario{}, err
	}

	return RespuestaUsuario{
		IDUsuario: id,
		Nombre:    entrada.Nombre,
		Correo:    entrada.Correo,
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
		h.actualizar(w, r, id)
	case http.MethodDelete:
		h.eliminar(w, r, id)
	default:
		http.Error(w, "Método HTTP no permitido", http.StatusMethodNotAllowed)
	}
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
	var entrada EntradaUsuario
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
	var entrada EntradaUsuario
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
	if len(partes) < 3 || partes == "" {
		http.Error(w, "ID de usuario no especificado", http.StatusBadRequest)
		return 0, false
	}

	idNum, err := strconv.Atoi(partes)
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