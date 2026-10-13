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

// EntradaSuscripcion representa el JSON recibido en solicitudes de creación y actualización
type EntradaSuscripcion struct {
	IDUsuario        int32     `json:"id_usuario"`
	IDServicio       int32     `json:"id_servicio"`
	Monto            float64   `json:"monto"`
	FechaVencimiento time.Time `json:"fecha_vencimiento"`
	FechaInicio      time.Time `json:"fecha_inicio"`
	UsuarioCuenta    string    `json:"usuario_cuenta"`
	PasswordCuenta   string    `json:"password_cuenta"`
	Estado           string    `json:"estado"`
}

// RespuestaSuscripcion define la salida JSON enviada al cliente
type RespuestaSuscripcion struct {
	IDSuscripcion    int32     `json:"id_suscripcion"`
	IDUsuario        int32     `json:"id_usuario"`
	IDServicio       int32     `json:"id_servicio"`
	Monto            string    `json:"monto"`
	FechaVencimiento time.Time `json:"fecha_vencimiento"`
	FechaInicio      time.Time `json:"fecha_inicio"`
	UsuarioCuenta    string    `json:"usuario_cuenta"`
	Estado           string    `json:"estado"`
	CreatedAt        time.Time `json:"created_at"`
}

// Estados permitidos según las reglas de negocio
var estadosValidos = map[string]bool{
	"Activa":    true,
	"Pausada":   true,
	"Cancelada": true,
}

func validarCamposSuscripcion(datos EntradaSuscripcion) error {
	if datos.IDUsuario <= 0 || datos.IDServicio <= 0 {
		return fmt.Errorf("los identificadores id_usuario e id_servicio deben ser válidos")
	}
	if datos.Monto <= 0 {
		return fmt.Errorf("el monto debe ser mayor a cero")
	}
	if datos.FechaInicio.IsZero() || datos.FechaVencimiento.IsZero() {
		return fmt.Errorf("las fechas de inicio y vencimiento son obligatorias")
	}
	if !datos.FechaInicio.Before(datos.FechaVencimiento) {
		return fmt.Errorf("la fecha de inicio debe ser anterior a la fecha de vencimiento")
	}
	if strings.TrimSpace(datos.UsuarioCuenta) == "" || strings.TrimSpace(datos.PasswordCuenta) == "" {
		return fmt.Errorf("las credenciales de la cuenta (usuario_cuenta y password_cuenta) son obligatorias")
	}

	est := strings.TrimSpace(datos.Estado)
	if est != "" && !estadosValidos[est] {
		return fmt.Errorf("el estado '%s' no es válido (permitidos: Activa, Pausada, Cancelada)", est)
	}

	return nil
}

func mapearSuscripcionARespuesta(sub sqlc.Suscripcion) RespuestaSuscripcion {
	
	return RespuestaSuscripcion{
		IDSuscripcion:    sub.IDSuscripcion,
		IDUsuario:        sub.IDUsuario,
		IDServicio:       sub.IDServicio,
		Monto:            sub.Monto,
		FechaVencimiento: sub.FechaVencimiento,
		FechaInicio:      sub.FechaInicio,
		UsuarioCuenta:    sub.UsuarioCuenta,
		Estado:           estadoTexto,
		CreatedAt:        sub.CreatedAt,
	}
}

type ServicioSuscripcion struct {
	consultas *sqlc.Queries
}

func NuevoServicioSuscripcion(q *sqlc.Queries) *ServicioSuscripcion {
	return &ServicioSuscripcion{consultas: q}
}

// Crear valida las FKs, fechas, monto y evita suscripciones duplicadas
func (s *ServicioSuscripcion) Crear(ctx context.Context, entrada EntradaSuscripcion) (RespuestaSuscripcion, error) {
	if err := validarCamposSuscripcion(entrada); err != nil {
		return RespuestaSuscripcion{}, err
	}

	// 1. Verificar que el usuario exista
	_, err := s.consultas.GetUsuario(ctx, entrada.IDUsuario)
	if err == sql.ErrNoRows {
		return RespuestaSuscripcion{}, fmt.Errorf("el usuario especificado no existe")
	} else if err != nil {
		return RespuestaSuscripcion{}, err
	}

	// 2. Verificar que el servicio exista
	_, err = s.consultas.GetServicio(ctx, entrada.IDServicio)
	if err == sql.ErrNoRows {
		return RespuestaSuscripcion{}, fmt.Errorf("el servicio especificado no existe")
	} else if err != nil {
		return RespuestaSuscripcion{}, err
	}

	// 3. Regla de negocio: Verificar si ya existe una suscripción activa para este usuario y servicio
	activa, err := s.consultas.GetSuscripcionActiva(ctx, sqlc.GetSuscripcionActivaParams{
		IDUsuario:  entrada.IDUsuario,
		IDServicio: entrada.IDServicio,
	})
	if err == nil && activa.IDSuscripcion > 0 {
		return RespuestaSuscripcion{}, fmt.Errorf("el usuario ya posee una suscripción activa para este servicio")
	} else if err != nil && err != sql.ErrNoRows {
		return RespuestaSuscripcion{}, err
	}


	montoTexto := fmt.Sprintf("%.2f", entrada.Monto)

	nueva, err := s.consultas.CreateSuscripcion(ctx, sqlc.CreateSuscripcionParams{
		IDUsuario:        entrada.IDUsuario,
		IDServicio:       entrada.IDServicio,
		Monto:            montoTexto,
		FechaVencimiento: entrada.FechaVencimiento,
		FechaInicio:      entrada.FechaInicio,
		UsuarioCuenta:    entrada.UsuarioCuenta,
		PasswordCuenta:   entrada.PasswordCuenta,
		Estado:           entrada.Estado,
	})
	if err != nil {
		return RespuestaSuscripcion{}, err
	}

	return mapearSuscripcionARespuesta(nueva), nil
}

// ObtenerPorID busca una suscripción por id_suscripcion
func (s *ServicioSuscripcion) ObtenerPorID(ctx context.Context, id int32) (RespuestaSuscripcion, error) {
	sub, err := s.consultas.GetSuscripcion(ctx, id)
	if err == sql.ErrNoRows {
		return RespuestaSuscripcion{}, fmt.Errorf("suscripción no encontrada")
	} else if err != nil {
		return RespuestaSuscripcion{}, err
	}
	return mapearSuscripcionARespuesta(sub), nil
}

// ListarPorUsuario trae todas las suscripciones registradas para un usuario
func (s *ServicioSuscripcion) ListarPorUsuario(ctx context.Context, idUsuario int32) ([]RespuestaSuscripcion, error) {
	lista, err := s.consultas.ListSuscripcionesUsuario(ctx, idUsuario)
	if err != nil {
		return nil, err
	}

	resultado := make([]RespuestaSuscripcion, len(lista))
	for i, sub := range lista {
		resultado[i] = mapearSuscripcionARespuesta(sub)
	}
	return resultado, nil
}

// Actualizar modifica el monto, vencimiento, credenciales o estado de una suscripción
func (s *ServicioSuscripcion) Actualizar(ctx context.Context, id int32, entrada EntradaSuscripcion) (RespuestaSuscripcion, error) {
	// 1. Validaciones básicas de entrada
	if entrada.Monto <= 0 {
		return RespuestaSuscripcion{}, fmt.Errorf("el monto debe ser mayor a cero")
	}
	if entrada.FechaVencimiento.IsZero() {
		return RespuestaSuscripcion{}, fmt.Errorf("la fecha de vencimiento es obligatoria")
	}
	if strings.TrimSpace(entrada.UsuarioCuenta) == "" || strings.TrimSpace(entrada.PasswordCuenta) == "" {
		return RespuestaSuscripcion{}, fmt.Errorf("las credenciales de la cuenta son obligatorias")
	}

	// 2. Validar que el estado sea uno de los permitidos
	est := strings.TrimSpace(entrada.Estado)
	estadosValidos := map[string]bool{"Activa": true, "Pausada": true, "Cancelada": true}
	if !estadosValidos[est] {
		return RespuestaSuscripcion{}, fmt.Errorf("el estado '%s' no es válido (permitidos: Activa, Pausada, Cancelada)", est)
	}

	// 3. Buscar la suscripción existente en la base de datos
	subExistente, err := s.consultas.GetSubscripcion(ctx, id)
	if err == sql.ErrNoRows {
		return RespuestaSuscripcion{}, fmt.Errorf("la suscripción especificada no existe")
	} else if err != nil {
		return RespuestaSuscripcion{}, err
	}

	// 4. Validar coherencia de fechas (la fecha de inicio debe ser anterior al vencimiento)
	if !subExistente.FechaInicio.Before(entrada.FechaVencimiento) {
		return RespuestaSuscripcion{}, fmt.Errorf("la fecha de inicio debe ser anterior a la fecha de vencimiento")
	}

	// 5. Regla Miro: Si se intenta reactivar (pasar a 'Activa'), verificar que no posea OTRA suscripción activa para este servicio
	if est == "Activa" && subExistente.Estado != "Activa" {
		existente, err := s.consultas.GetSuscripcionActivaByUsuarioYServicio(ctx, sqlc.GetSuscripcionActivaByUsuarioYServicioParams{
			IDUsuario:  subExistente.IDUsuario,
			IDServicio: subExistente.IDServicio,
		})
		if err == nil && existente.IDSuscripcion != id {
			return RespuestaSuscripcion{}, fmt.Errorf("el usuario ya posee una suscripción activa para este servicio")
		}
	}

	// 6. Formatear monto
	montoTexto := fmt.Sprintf("%.2f", entrada.Monto)

	// 7. Actualizar en la base de datos vía sqlc
	err = s.consultas.UpdateSuscripcion(ctx, sqlc.UpdateSuscripcionParams{
		IDSuscripcion:    id,
		Monto:            montoTexto,
		FechaVencimiento: entrada.FechaVencimiento,
		Estado:           est,
		UsuarioCuenta:    entrada.UsuarioCuenta,
		PasswordCuenta:   entrada.PasswordCuenta,
	})
	if err != nil {
		return RespuestaSuscripcion{}, err
	}

	// 8. Construir DTO de respuesta
	return RespuestaSuscripcion{
		IDSuscripcion:    id,
		IDUsuario:        subExistente.IDUsuario,
		IDServicio:       subExistente.IDServicio,
		Monto:            montoTexto,
		FechaInicio:      subExistente.FechaInicio,
		FechaVencimiento: entrada.FechaVencimiento,
		UsuarioCuenta:    entrada.UsuarioCuenta,
		Estado:           est,
		CreatedAt:        subExistente.CreatedAt,
	}, nil
}
// Eliminar quita la suscripción previa comprobación
func (s *ServicioSuscripcion) Eliminar(ctx context.Context, id int32) error {
	_, err := s.consultas.GetSuscripcion(ctx, id)
	if err == sql.ErrNoRows {
		return fmt.Errorf("suscripción no encontrada")
	} else if err != nil {
		return err
	}

	return s.consultas.DeleteSuscripcion(ctx, id)
}

type ManejadorSuscripcion struct {
	servicio *ServicioSuscripcion
}

func NuevoManejadorSuscripcion(s *ServicioSuscripcion) *ManejadorSuscripcion {
	return &ManejadorSuscripcion{servicio: s}
}

// ManejarSuscripciones atiende POST a /suscripciones
func (h *ManejadorSuscripcion) ManejarSuscripciones(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.crear(w, r)
	default:
		http.Error(w, "Método HTTP no permitido", http.StatusMethodNotAllowed)
	}
}

// ManejarSuscripcionPorID atiende GET, PUT y DELETE en /suscripciones/{id}
func (h *ManejadorSuscripcion) ManejarSuscripcionPorID(w http.ResponseWriter, r *http.Request) {
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

// ManejarSuscripcionesPorUsuario atiende GET en /suscripciones/usuario/{id_usuario}
func (h *ManejadorSuscripcion) ManejarSuscripcionesPorUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Método HTTP no permitido", http.StatusMethodNotAllowed)
		return
	}

	idUsuario, ok := h.extraerID(w, r)
	if !ok {
		return
	}

	h.obtenerPorUsuario(w, r, idUsuario)
}

func (h *ManejadorSuscripcion) crear(w http.ResponseWriter, r *http.Request) {
	var entrada EntradaSuscripcion
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		http.Error(w, "El formato JSON enviado no es válido", http.StatusBadRequest)
		return
	}

	creada, err := h.servicio.Crear(r.Context(), entrada)
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusCreated, creada)
}

func (h *ManejadorSuscripcion) obtenerPorID(w http.ResponseWriter, r *http.Request, id int32) {
	sub, err := h.servicio.ObtenerPorID(r.Context(), id)
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusOK, sub)
}

func (h *ManejadorSuscripcion) actualizar(w http.ResponseWriter, r *http.Request, id int32) {
	var entrada EntradaSuscripcion
	if err := json.NewDecoder(r.Body).Decode(&entrada); err != nil {
		http.Error(w, "El formato JSON enviado no es válido", http.StatusBadRequest)
		return
	}

	actualizada, err := h.servicio.Actualizar(r.Context(), id, entrada)
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusOK, actualizada)
}

func (h *ManejadorSuscripcion) eliminar(w http.ResponseWriter, r *http.Request, id int32) {
	if err := h.servicio.Eliminar(r.Context(), id); err != nil {
		h.responderError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ManejadorSuscripcion) obtenerPorUsuario(w http.ResponseWriter, r *http.Request, idUsuario int32) {
	lista, err := h.servicio.ListarPorUsuario(r.Context(), idUsuario)
	if err != nil {
		h.responderError(w, err)
		return
	}
	h.responderJSON(w, http.StatusOK, lista)
}

func (h *ManejadorSuscripcion) extraerID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	partes := strings.Split(r.URL.Path, "/")
	if len(partes) < 3 || partes == "" {
		http.Error(w, "ID no especificado", http.StatusBadRequest)
		return 0, false
	}

	idNum, err := strconv.Atoi(partes[len(partes)-1])
	if err != nil {
		http.Error(w, "El ID debe ser un valor numérico", http.StatusBadRequest)
		return 0, false
	}

	return int32(idNum), true
}

func (h *ManejadorSuscripcion) responderError(w http.ResponseWriter, err error) {
	msg := err.Error()

	switch {
	case strings.Contains(msg, "no encontrada") || strings.Contains(msg, "no existe"):
		http.Error(w, msg, http.StatusNotFound)
	case strings.Contains(msg, "obligatorias") || strings.Contains(msg, "mayor a cero") || strings.Contains(msg, "anterior") || strings.Contains(msg, "válido"):
		http.Error(w, msg, http.StatusBadRequest)
	case strings.Contains(msg, "activa"):
		http.Error(w, msg, http.StatusConflict)
	default:
		http.Error(w, "Error interno del servidor", http.StatusInternalServerError)
	}
}

func (h *ManejadorSuscripcion) responderJSON(w http.ResponseWriter, estado int, datos any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(datos)
}