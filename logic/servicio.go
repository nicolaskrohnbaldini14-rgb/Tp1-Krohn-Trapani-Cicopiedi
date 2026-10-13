package logic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
)

// EntradaServicio representa el JSON recibido en solicitudes POST y PUT
type EntradaServicio struct {
	Nombre    string `json:"nombre"`
	Categoria string `json:"categoria"`
}

// RespuestaServicio define el objeto retornado en las respuestas JSON
type RespuestaServicio struct {
	IDServicio int32     `json:"id_servicio"`
	Nombre     string    `json:"nombre"`
	Categoria  string    `json:"categoria"`
	CreatedAt  time.Time `json:"created_at"`
}

// Categorías predefinidas según la regla de negocio (no texto libre)
var categoriasValidas = map[string]bool{
	"Streaming": true,
	"Música":    true,
	"Software":  true,
	"Gimnasio":  true,
	"Otros":     true,
}


func validarCamposServicio(datos EntradaServicio) error {
	nom := strings.TrimSpace(datos.Nombre)
	cat := strings.TrimSpace(datos.Categoria)

	if nom == "" {
		return fmt.Errorf("el nombre del servicio es obligatorio")
	}
	if cat == "" {
		return fmt.Errorf("la categoría es obligatoria")
	}
	if !categoriasValidas[cat] {
		return fmt.Errorf("la categoría '%s' no es válida (permitidas: Streaming, Música, Software, Gimnasio, Otros)", cat)
	}

	return nil
}

func mapearServicioARespuesta(s sqlc.Servicio) RespuestaServicio {
	return RespuestaServicio{
		IDServicio: s.IDServicio,
		Nombre:     s.Nombre,
		Categoria:  s.Categoria,
		CreatedAt:  s.CreatedAt,
	}
}


type ServicioServicio struct {
	consultas *sqlc.Queries
}

func NuevoServicioServicio(q *sqlc.Queries) *ServicioServicio {
	return &ServicioServicio{consultas: q}
}

// Crear valida la categoría y comprueba que el nombre sea único
func (s *ServicioServicio) Crear(ctx context.Context, entrada EntradaServicio) (RespuestaServicio, error) {
	if err := validarCamposServicio(entrada); err != nil {
		return RespuestaServicio{}, err
	}

	// Regla Miro: El nombre debe ser único
	_, err := s.consultas.GetServicioByNombre(ctx, entrada.Nombre)
	if err == nil {
		return RespuestaServicio{}, fmt.Errorf("ya existe un servicio con el nombre '%s'", entrada.Nombre)
	} else if err != sql.ErrNoRows {
		return RespuestaServicio{}, err
	}

	nuevo, err := s.consultas.CreateServicio(ctx, sqlc.CreateServicioParams{
		Nombre:    entrada.Nombre,
		Categoria: entrada.Categoria,
	})
	if err != nil {
		return RespuestaServicio{}, err
	}

	return mapearServicioARespuesta(nuevo), nil
}

// ObtenerPorID busca un servicio por id_servicio
func (s *ServicioServicio) ObtenerPorID(ctx context.Context, id int32) (RespuestaServicio, error) {
	srv, err := s.consultas.GetServicio(ctx, id)
	if err == sql.ErrNoRows {
		return RespuestaServicio{}, fmt.Errorf("servicio no encontrado")
	} else if err != nil {
		return RespuestaServicio{}, err
	}
	return mapearServicioARespuesta(srv), nil
}

// Listar obtiene la lista de todos los servicios
func (s *ServicioServicio) Listar(ctx context.Context) ([]RespuestaServicio, error) {
	lista, err := s.consultas.ListServicios(ctx)
	if err != nil {
		return nil, err
	}

	resultado := make([]RespuestaServicio, len(lista))
	for i, item := range lista {
		resultado[i] = mapearServicioARespuesta(item)
	}
	return resultado, nil
}

// Actualizar modifica el nombre manteniendo la restricción de unicidad y no permitiendo cambiar la categoría
func (s *ServicioServicio) Actualizar(ctx context.Context, id int32, entrada EntradaServicio) (RespuestaServicio, error) {
	if strings.TrimSpace(entrada.Nombre) == "" {
		return RespuestaServicio{}, fmt.Errorf("el nombre del servicio es obligatorio")
	}

	servicioExistente, err := s.consultas.GetServicio(ctx, id)
	if err == sql.ErrNoRows {
		return RespuestaServicio{}, fmt.Errorf("servicio no encontrado")
	} else if err != nil {
		return RespuestaServicio{}, err
	}


	if entrada.Categoria != "" && entrada.Categoria != servicioExistente.Categoria {
		return RespuestaServicio{}, fmt.Errorf("no se permite modificar la categoría del servicio")
	}

	
	if entrada.Nombre != servicioExistente.Nombre {
		existente, err := s.consultas.GetServicioByNombre(ctx, entrada.Nombre)
		if err == nil && existente.IDServicio != id {
			return RespuestaServicio{}, fmt.Errorf("ya existe otro servicio registrado con el nombre '%s'", entrada.Nombre)
		} else if err != nil && err != sql.ErrNoRows {
			return RespuestaServicio{}, err
		}
	}

	err = s.consultas.UpdateServicio(ctx, sqlc.UpdateServicioParams{
		IDServicio: id,
		Nombre:     entrada.Nombre,
		Categoria:  servicioExistente.Categoria,
	})
	if err != nil {
		return RespuestaServicio{}, err
	}

	return RespuestaServicio{
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

// No se puede eliminar si tiene suscripciones asociadas, esto lo hacemos para que no nos devuelva unerror 
	//generico en la base de datos y si un error conocido de logica que es el 409
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
	var entrada EntradaServicio
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
	var entrada EntradaServicio
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