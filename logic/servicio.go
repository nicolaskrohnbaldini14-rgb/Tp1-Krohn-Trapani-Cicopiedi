package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
)

// Categorías predefinidas según la regla de negocio (no texto libre)
var categoriasValidas = map[string]bool{
	"Streaming": true,
	"Música":    true,
	"Software":  true,
	"Deporte":  true,
	"Salud":     true,
	"Financiero": true ,
	"Turismo":   true,
	"Otros":     true,
}

func validarCamposServicio(datos sqlc.CreateServicioParams) error {
	nom := strings.TrimSpace(datos.Nombre)
	cat := strings.TrimSpace(datos.Categoria)

	if nom == "" {
		return fmt.Errorf("el nombre del servicio es obligatorio")
	}
	if cat == "" {
    return fmt.Errorf("la categoría es obligatorio") 
}
	if !categoriasValidas[cat] {
		return fmt.Errorf("la categoría '%s' no es válida (permitidas: Streaming, Música, Software, Gimnasio,Salud,Financiero,Turismo Otros)", cat)
	}

	return nil
}

type ServicioServicio struct {
	consultas *sqlc.Queries
}

func NuevoServicioServicio(q *sqlc.Queries) *ServicioServicio {
	return &ServicioServicio{consultas: q}
}

// Crear valida la categoría y comprueba que el nombre sea único
func (s *ServicioServicio) Crear(ctx context.Context, entrada sqlc.CreateServicioParams) (sqlc.Servicio, error) {
	if err := validarCamposServicio(entrada); err != nil {
		return sqlc.Servicio{}, err
	}

	// Regla Miro: El nombre debe ser único
	_, err := s.consultas.GetServicioByNombre(ctx, entrada.Nombre)
	if err == nil {
		return sqlc.Servicio{}, fmt.Errorf("ya existe un servicio con el nombre '%s'", entrada.Nombre)
	} else if err != sql.ErrNoRows {
		return sqlc.Servicio{}, err
	}

	return s.consultas.CreateServicio(ctx, entrada)
}

// ObtenerPorID busca un servicio por id_servicio
func (s *ServicioServicio) ObtenerPorID(ctx context.Context, id int32) (sqlc.Servicio, error) {
	srv, err := s.consultas.GetServicio(ctx, id)
	if err == sql.ErrNoRows {
		return sqlc.Servicio{}, fmt.Errorf("servicio no encontrado")
	} else if err != nil {
		return sqlc.Servicio{}, err
	}
	return srv, nil
}

// Listar obtiene la lista de todos los servicios
func (s *ServicioServicio) Listar(ctx context.Context) ([]sqlc.Servicio, error) {
	return s.consultas.ListServicios(ctx)
}

// Actualizar modifica el nombre manteniendo la restricción de unicidad y no permitiendo cambiar la categoría
func (s *ServicioServicio) Actualizar(ctx context.Context, id int32, entrada sqlc.UpdateServicioParams) (sqlc.Servicio, error) {
	if strings.TrimSpace(entrada.Nombre) == "" {
		return sqlc.Servicio{}, fmt.Errorf("el nombre del servicio es obligatorio")
	}

	servicioExistente, err := s.consultas.GetServicio(ctx, id)
	if err == sql.ErrNoRows {
		return sqlc.Servicio{}, fmt.Errorf("servicio no encontrado")
	} else if err != nil {
		return sqlc.Servicio{}, err
	}

	if entrada.Categoria != "" && entrada.Categoria != servicioExistente.Categoria {
		return sqlc.Servicio{}, fmt.Errorf("no se permite modificar la categoría del servicio")
	}

	if entrada.Nombre != servicioExistente.Nombre {
		existente, err := s.consultas.GetServicioByNombre(ctx, entrada.Nombre)
		if err == nil && existente.IDServicio != id {
			return sqlc.Servicio{}, fmt.Errorf("ya existe otro servicio registrado con el nombre '%s'", entrada.Nombre)
		} else if err != nil && err != sql.ErrNoRows {
			return sqlc.Servicio{}, err
		}
	}

	// el id siempre viene de la URL y la categoría se mantiene la existente
	entrada.IDServicio = id
	entrada.Categoria = servicioExistente.Categoria

	if err := s.consultas.UpdateServicio(ctx, entrada); err != nil {
		return sqlc.Servicio{}, err
	}

	return sqlc.Servicio{
		IDServicio: id,
		Nombre:     entrada.Nombre,
		Categoria:  servicioExistente.Categoria,
		CreatedAt:  servicioExistente.CreatedAt,
	}, nil
}

// Eliminar quita un servicio verificando que no tenga suscripciones asociadas
func (s *ServicioServicio) Eliminar(ctx context.Context, id int32) error {
	_, err := s.consultas.GetServicio(ctx, id)
	if err == sql.ErrNoRows {
		return fmt.Errorf("servicio no encontrado")
	} else if err != nil {
		return err
	}

	//preguntar estoo si estabien o dejamos que lo haga la base sola
	// No se puede eliminar si tiene suscripciones asociadas, esto lo hacemos para que no nos devuelva un error
	// genérico en la base de datos y sí un error conocido de lógica que es el 409
	cantidadSubs, err := s.consultas.CountSuscripcionesByServicio(ctx, id)
	if err != nil {
		return err
	}
	if cantidadSubs > 0 {
		return fmt.Errorf("no se puede eliminar el servicio porque tiene suscripciones asociadas")
	}

	return s.consultas.DeleteServicio(ctx, id)
}

type ManejadorServicio struct {
	servicio *ServicioServicio
}

func NuevoManejadorServicio(s *ServicioServicio) *ManejadorServicio {
	return &ManejadorServicio{servicio: s}
}

// ManejarServicios atiende GET y POST en /servicios
func (h *ManejadorServicio) ManejarServicios(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.obtenerTodos(w, r)
	case http.MethodPost:
		h.crear(w, r)
	default:
		http.Error(w, "Método HTTP no permitido", http.StatusMethodNotAllowed)
	}
}

// ManejarServicioPorID atiende GET, PUT y DELETE en /servicios/{id}
func (h *ManejadorServicio) ManejarServicioPorID(w http.ResponseWriter, r *http.Request) {
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

func (h *ManejadorServicio) obtenerTodos(w http.ResponseWriter, r *http.Request) {
	servicios, err := h.servicio.Listar(r.Context())
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusOK, servicios)
}

func (h *ManejadorServicio) crear(w http.ResponseWriter, r *http.Request) {
	var entrada sqlc.CreateServicioParams
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

func (h *ManejadorServicio) obtenerPorID(w http.ResponseWriter, r *http.Request, id int32) {
	srv, err := h.servicio.ObtenerPorID(r.Context(), id)
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusOK, srv)
}

func (h *ManejadorServicio) actualizar(w http.ResponseWriter, r *http.Request, id int32) {
	var entrada sqlc.UpdateServicioParams
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

func (h *ManejadorServicio) eliminar(w http.ResponseWriter, r *http.Request, id int32) {
	if err := h.servicio.Eliminar(r.Context(), id); err != nil {
		h.responderError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ManejadorServicio) extraerID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	partes := strings.Split(r.URL.Path, "/")
	if len(partes) < 3 || partes[len(partes)-1] == "" {
		http.Error(w, "ID de servicio no especificado", http.StatusBadRequest)
		return 0, false
	}

	idNum, err := strconv.Atoi(partes[len(partes)-1])
	if err != nil {
		http.Error(w, "El ID debe ser un valor numérico", http.StatusBadRequest)
		return 0, false
	}

	return int32(idNum), true
}

func (h *ManejadorServicio) responderError(w http.ResponseWriter, err error) {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "no encontrado"):
		http.Error(w, msg, http.StatusNotFound)
	case strings.Contains(msg, "obligatorio") || strings.Contains(msg, "no es válida") || strings.Contains(msg, "no se permite modificar"):
		http.Error(w, msg, http.StatusBadRequest)
	case strings.Contains(msg, "ya existe") || strings.Contains(msg, "suscripciones asociadas"):
		http.Error(w, msg, http.StatusConflict)
	default:
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

func (h *ManejadorServicio) responderJSON(w http.ResponseWriter, estado int, datos any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(datos)
}
