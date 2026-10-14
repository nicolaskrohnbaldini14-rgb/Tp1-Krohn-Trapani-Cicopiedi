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

// Estados permitidos según las reglas de negocio
var estadosValidos = map[string]bool{
	"Activa":    true,
	"Pausada":   true,
	"Cancelada": true,
}


// pregunta si esta bien esto o lo cambiamos en el schema
// parsearMonto convierte el monto (string en sqlc) a número y valida que sea > 0
func parsearMonto(monto string) (float64, error) {
	valor, err := strconv.ParseFloat(strings.TrimSpace(monto), 64)
	if err != nil {
		return 0, fmt.Errorf("el monto debe ser un número válido")
	}
	if valor <= 0 {
		return 0, fmt.Errorf("el monto debe ser mayor a cero")
	}
	return valor, nil
}

func validarCamposSuscripcion(datos sqlc.CreateSubscripcionParams) error {
	if datos.IDUsuario <= 0 || datos.IDServicio <= 0 {
		return fmt.Errorf("los identificadores id_usuario e id_servicio deben ser válidos")
	}
	if _, err := parsearMonto(datos.Monto); err != nil {
		return err
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

// ocultarPasswordCuenta evita devolver la contraseña de la cuenta al cliente
func ocultarPasswordCuenta(sub sqlc.Suscripcion) sqlc.Suscripcion {
	sub.PasswordCuenta = ""
	return sub
}

type ServicioSuscripcion struct {
	consultas *sqlc.Queries
}

func NuevoServicioSuscripcion(q *sqlc.Queries) *ServicioSuscripcion {
	return &ServicioSuscripcion{consultas: q}
}

// Crear valida las FKs, fechas, monto y evita suscripciones duplicadas
func (s *ServicioSuscripcion) Crear(ctx context.Context, entrada sqlc.CreateSubscripcionParams) (sqlc.Suscripcion, error) {
	// Si no viene estado, por defecto la suscripción arranca como 'Activa'
	entrada.Estado = strings.TrimSpace(entrada.Estado)
	if entrada.Estado == "" {
		entrada.Estado = "Activa"
	}

	if err := validarCamposSuscripcion(entrada); err != nil {
		return sqlc.Suscripcion{}, err
	}

	// 1. Verificar que el usuario exista
	_, err := s.consultas.GetUsuario(ctx, entrada.IDUsuario)
	if err == sql.ErrNoRows {
		return sqlc.Suscripcion{}, fmt.Errorf("el usuario especificado no existe")
	} else if err != nil {
		return sqlc.Suscripcion{}, err
	}

	// 2. Verificar que el servicio exista
	_, err = s.consultas.GetServicio(ctx, entrada.IDServicio)
	if err == sql.ErrNoRows {
		return sqlc.Suscripcion{}, fmt.Errorf("el servicio especificado no existe")
	} else if err != nil {
		return sqlc.Suscripcion{}, err
	}

	//  Verificar si ya existe una suscripción activa para este usuario y servicio
	activa, err := s.consultas.GetSuscripcionActiva(ctx, sqlc.GetSuscripcionActivaParams{
		IDUsuario:  entrada.IDUsuario,
		IDServicio: entrada.IDServicio,
	})
	if err == nil && activa.IDSuscripcion > 0 {
		return sqlc.Suscripcion{}, fmt.Errorf("el usuario ya posee una suscripción activa para este servicio")
	} else if err != nil && err != sql.ErrNoRows {
		return sqlc.Suscripcion{}, err
	}

	// el monto ya fue validado en validarCamposSuscripcion
	monto, _ := parsearMonto(entrada.Monto)
	entrada.Monto = fmt.Sprintf("%.2f", monto)

	nueva, err := s.consultas.CreateSubscripcion(ctx, entrada)
	if err != nil {
		return sqlc.Suscripcion{}, err
	}

	return ocultarPasswordCuenta(nueva), nil
}

// ObtenerPorID busca una suscripción por id_suscripcion
func (s *ServicioSuscripcion) ObtenerPorID(ctx context.Context, id int32) (sqlc.Suscripcion, error) {
	sub, err := s.consultas.GetSubscripcion(ctx, id)
	if err == sql.ErrNoRows {
		return sqlc.Suscripcion{}, fmt.Errorf("suscripción no encontrada")
	} else if err != nil {
		return sqlc.Suscripcion{}, err
	}
	return ocultarPasswordCuenta(sub), nil
}

// ListarPorUsuario trae todas las suscripciones registradas para un usuario
func (s *ServicioSuscripcion) ListarPorUsuario(ctx context.Context, idUsuario int32) ([]sqlc.ListSubscripcionesUsuarioRow, error) {
	return s.consultas.ListSubscripcionesUsuario(ctx, idUsuario)
}

// Actualizar modifica el monto, vencimiento, credenciales o estado de una suscripción
func (s *ServicioSuscripcion) Actualizar(ctx context.Context, id int32, entrada sqlc.UpdateSubscripcionParams) (sqlc.Suscripcion, error) {
	// 1. Validaciones básicas de entrada
	monto, err := parsearMonto(entrada.Monto)
	if err != nil {
		return sqlc.Suscripcion{}, err
	}
	if entrada.FechaVencimiento.IsZero() {
		return sqlc.Suscripcion{}, fmt.Errorf("la fecha de vencimiento es obligatoria")
	}
	if strings.TrimSpace(entrada.UsuarioCuenta) == "" || strings.TrimSpace(entrada.PasswordCuenta) == "" {
		return sqlc.Suscripcion{}, fmt.Errorf("las credenciales de la cuenta son obligatorias")
	}

	// 2. Validar que el estado sea uno de los permitidos
	est := strings.TrimSpace(entrada.Estado)
	if !estadosValidos[est] {
		return sqlc.Suscripcion{}, fmt.Errorf("el estado '%s' no es válido (permitidos: Activa, Pausada, Cancelada)", est)
	}

	// 3. Buscar la suscripción existente en la base de datos
	subExistente, err := s.consultas.GetSubscripcion(ctx, id)
	if err == sql.ErrNoRows {
		return sqlc.Suscripcion{}, fmt.Errorf("la suscripción especificada no existe")
	} else if err != nil {
		return sqlc.Suscripcion{}, err
	}

	// 4. Validar coherencia de fechas (la fecha de inicio debe ser anterior al vencimiento)
	if !subExistente.FechaInicio.Before(entrada.FechaVencimiento) {
		return sqlc.Suscripcion{}, fmt.Errorf("la fecha de inicio debe ser anterior a la fecha de vencimiento")
	}

	// 5. Si se intenta reactivar (pasar a 'Activa'), verificar que no posea OTRA suscripción activa para este servicio
	if est == "Activa" && subExistente.Estado != "Activa" {
		existente, err := s.consultas.GetSuscripcionActiva(ctx, sqlc.GetSuscripcionActivaParams{
			IDUsuario:  subExistente.IDUsuario,
			IDServicio: subExistente.IDServicio,
		})
		if err == nil && existente.IDSuscripcion != id {
			return sqlc.Suscripcion{}, fmt.Errorf("el usuario ya posee una suscripción activa para este servicio")
		}
	}

	// 6. Formatear monto y completar los datos que no vienen del body
	entrada.IDSuscripcion = id // el id siempre viene de la URL
	entrada.Monto = fmt.Sprintf("%.2f", monto)
	entrada.Estado = est

	// 7. Actualizar en la base de datos vía sqlc
	if err := s.consultas.UpdateSubscripcion(ctx, entrada); err != nil {
		return sqlc.Suscripcion{}, err
	}

	// 8. Construir la respuesta (sin password_cuenta)
	return sqlc.Suscripcion{
		IDSuscripcion:    id,
		IDUsuario:        subExistente.IDUsuario,
		IDServicio:       subExistente.IDServicio,
		Monto:            entrada.Monto,
		FechaInicio:      subExistente.FechaInicio,
		FechaVencimiento: entrada.FechaVencimiento,
		UsuarioCuenta:    entrada.UsuarioCuenta,
		Estado:           est,
		CreatedAt:        subExistente.CreatedAt,
	}, nil
}

// Eliminar quita la suscripción previa comprobación
func (s *ServicioSuscripcion) Eliminar(ctx context.Context, id int32) error {
	_, err := s.consultas.GetSubscripcion(ctx, id)
	if err == sql.ErrNoRows {
		return fmt.Errorf("suscripción no encontrada")
	} else if err != nil {
		return err
	}

	return s.consultas.DeleteSubscripcion(ctx, id)
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
	var entrada sqlc.CreateSubscripcionParams
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
	var entrada sqlc.UpdateSubscripcionParams
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
	if len(partes) < 3 || partes[len(partes)-1] == "" {
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
	case strings.Contains(msg, "obligatoria") || strings.Contains(msg, "mayor a cero") || strings.Contains(msg, "anterior") || strings.Contains(msg, "válido"):
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