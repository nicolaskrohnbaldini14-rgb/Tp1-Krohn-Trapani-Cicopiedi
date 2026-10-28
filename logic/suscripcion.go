package logic

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
)

// Estados permitidos según las reglas de negocio
var Estados = []string{"Activa", "Pausada", "Cancelada"}

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

// EliminarDeUsuario borra la suscripción solo si pertenece al usuario indicado.
// Si es de otro usuario responde como si no existiera.
func (s *ServicioSuscripcion) EliminarDeUsuario(ctx context.Context, id, idUsuario int32) error {
	sub, err := s.consultas.GetSubscripcion(ctx, id)
	if err == sql.ErrNoRows || (err == nil && sub.IDUsuario != idUsuario) {
		return fmt.Errorf("suscripción no encontrada")
	} else if err != nil {
		return err
	}

	return s.consultas.DeleteSubscripcion(ctx, id)
}
