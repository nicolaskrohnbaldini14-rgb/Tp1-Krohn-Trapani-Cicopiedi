package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	sqlc "tp2-krohn-trapani-cicopiedi/db/sqlc"
)

func TestTablasDominio_CRUD(t *testing.T) {
	// 1. Conexión a PostgreSQL en Docker
	connStr := "postgres://postgres:postgres@localhost:5432/mi_tp2_db?sslmode=disable"
	dbConn, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("Error al abrir conexión con la base de datos: %v", err)
	}
	defer dbConn.Close()

	err = dbConn.Ping()
	if err != nil {
		t.Fatalf("Base de datos inalcanzable. ¿El contenedor de Docker está activo?: %v", err)
	}

	queries := sqlc.New(dbConn)
	ctx := context.Background()

	// 2. Limpieza de las 4 tablas en orden inverso por claves foráneas
	_, err = dbConn.Exec("TRUNCATE TABLE pago, suscripcion, servicio, usuario CASCADE")
	if err != nil {
		t.Fatalf("Error al limpiar las tablas de la BD: %v", err)
	}

	// Variables para encadenar los IDs generados entre tablas
	var createdUserID int32
	var createdServiceID int32
	var createdSubID int32
	var createdPagoID int32

	// =========================================================================
	// TEST 1: TABLA USUARIO
	// =========================================================================
	t.Run("Tabla_Usuario_CRUD", func(t *testing.T) {
		t.Run("CreateUsuario", func(t *testing.T) {
			usr, err := queries.CreateUsuario(ctx, sqlc.CreateUsuarioParams{
				Nombre:   "Nicolas",
				Apellido: "Krohn",
				Dni:      "40123456",
				Email:    "nicolas@example.com",
				Password: "password123",
			})
			if err != nil {
				t.Fatalf("Error al crear usuario: %v", err)
			}
			if usr.IDUsuario == 0 {
				t.Error("Se esperaba un ID de usuario autogenerado mayor a 0")
			}
			createdUserID = usr.IDUsuario
		})

		t.Run("GetUsuario", func(t *testing.T) {
			usr, err := queries.GetUsuario(ctx, createdUserID)
			if err != nil {
				t.Fatalf("Error al obtener usuario: %v", err)
			}
			if usr.Email != "nicolas@example.com" {
				t.Errorf("Email esperado 'nicolas@example.com', obtenido '%s'", usr.Email)
			}
		})

		t.Run("UpdateUsuario", func(t *testing.T) {
			err := queries.UpdateUsuario(ctx, sqlc.UpdateUsuarioParams{
				IDUsuario: createdUserID,
				Nombre:    "Nicolas Modificado",
				Apellido:  "Krohn",
				Dni:       "40123456",
				Email:     "nicolas.updated@example.com",
				Password:  "newpassword123",
			})
			if err != nil {
				t.Fatalf("Error al actualizar usuario: %v", err)
			}
		})

		t.Run("ListUsuarios", func(t *testing.T) {
			lista, err := queries.ListUsuarios(ctx)
			if err != nil {
				t.Fatalf("Error al listar usuarios: %v", err)
			}
			if len(lista) == 0 {
				t.Error("Se esperaba al menos 1 usuario en la lista")
			}
		})

		t.Run("DeleteUsuario", func(t *testing.T) {
			usrTemp, err := queries.CreateUsuario(ctx, sqlc.CreateUsuarioParams{
				Nombre:   "Usuario",
				Apellido: "A Borrar",
				Dni:      "99999999",
				Email:    "aborrar@example.com",
				Password: "password123",
			})
			if err != nil {
				t.Fatalf("Error al crear usuario auxiliar para borrar: %v", err)
			}

			err = queries.DeleteUsuario(ctx, usrTemp.IDUsuario)
			if err != nil {
				t.Fatalf("Error al eliminar usuario: %v", err)
			}

			_, err = queries.GetUsuario(ctx, usrTemp.IDUsuario)
			if err != sql.ErrNoRows {
				t.Errorf("Se esperaba sql.ErrNoRows al buscar un usuario eliminado, pero se obtuvo: %v", err)
			}
		})
	})

	// =========================================================================
	// TEST 2: TABLA SERVICIO
	// =========================================================================
	t.Run("Tabla_Servicio_CRUD", func(t *testing.T) {
		t.Run("CreateServicio", func(t *testing.T) {
			srv, err := queries.CreateServicio(ctx, sqlc.CreateServicioParams{
				Nombre: "Netflix",
			})
			if err != nil {
				t.Fatalf("Error al crear servicio: %v", err)
			}
			if srv.IDServicio == 0 {
				t.Error("Se esperaba un ID de servicio autogenerado mayor a 0")
			}
			createdServiceID = srv.IDServicio
		})

		t.Run("GetServicio", func(t *testing.T) {
			srv, err := queries.GetServicio(ctx, createdServiceID)
			if err != nil {
				t.Fatalf("Error al obtener servicio: %v", err)
			}
			if srv.Nombre != "Netflix" {
				t.Errorf("Nombre esperado 'Netflix', obtenido '%s'", srv.Nombre)
			}
		})

		t.Run("UpdateServicio", func(t *testing.T) {
			err := queries.UpdateServicio(ctx, sqlc.UpdateServicioParams{
				IDServicio: createdServiceID,
				Nombre:     "Netflix Premium",
			})
			if err != nil {
				t.Fatalf("Error al actualizar servicio: %v", err)
			}
		})

		t.Run("ListServicios", func(t *testing.T) {
			servicios, err := queries.ListServicios(ctx)
			if err != nil {
				t.Fatalf("Error al listar servicios: %v", err)
			}
			if len(servicios) == 0 {
				t.Error("Se esperaba al menos 1 servicio en la lista")
			}
		})
	})

	// =========================================================================
	// TEST 3: TABLA SUBSCRIPCIONES
	// =========================================================================
	t.Run("Tabla_Subscripcion_CRUD", func(t *testing.T) {
		ahora := time.Now()

		t.Run("CreateSubscripcion", func(t *testing.T) {
			sub, err := queries.CreateSubscripcion(ctx, sqlc.CreateSubscripcionParams{
				IDUsuario:        createdUserID,
				IDServicio:       createdServiceID,
				Monto:            "6000.00",
				FechaVencimiento: ahora.AddDate(0, 1, 0),
				FechaInicio:      ahora,
				UsuarioCuenta:    "nicolas@example.com",
				PasswordCuenta:   "passcuenta123",
				Estado:           sql.NullString{String: "Activa", Valid: true},
			})
			if err != nil {
				t.Fatalf("Error al crear suscripción: %v", err)
			}
			if sub.IDSuscripcion == 0 {
				t.Error("Se esperaba un ID de suscripción mayor a 0")
			}
			createdSubID = sub.IDSuscripcion
		})

		t.Run("GetSubscripcion", func(t *testing.T) {
			sub, err := queries.GetSubscripcion(ctx, createdSubID)
			if err != nil {
				t.Fatalf("Error al obtener suscripción: %v", err)
			}
			if sub.UsuarioCuenta != "nicolas@example.com" {
				t.Errorf("Usuario de cuenta esperado 'nicolas@example.com', obtenido '%s'", sub.UsuarioCuenta)
			}
		})

		t.Run("UpdateSubscripcion", func(t *testing.T) {
			err := queries.UpdateSubscripcion(ctx, sqlc.UpdateSubscripcionParams{
				IDSuscripcion:  createdSubID,
				Monto:            "6500.00",
				FechaVencimiento: ahora.AddDate(0, 2, 0),
				Estado:           sql.NullString{String: "Pausada", Valid: true},
			})
			if err != nil {
				t.Fatalf("Error al actualizar suscripción: %v", err)
			}
		})

		t.Run("ListSubscripcionUsuario", func(t *testing.T) {
			subs, err := queries.ListSuscripcionesUsuario(ctx, createdUserID)
			if err != nil {
				t.Fatalf("Error al listar suscripcion del usuario: %v", err)
			}
			if len(subs) == 0 {
				t.Error("Se esperaba al menos 1 suscripción listada para el usuario")
			}
		})
	})

	// =========================================================================
	// TEST 4: TABLA PAGO
	// =========================================================================
	t.Run("Tabla_Pago_CRUD", func(t *testing.T) {
		ahora := time.Now()

		t.Run("CreatePago", func(t *testing.T) {
			pago, err := queries.CreatePago(ctx, sqlc.CreatePagoParams{
				IDSuscripcion: createdSubID,
				Monto:           "6500.00",
				FechaPago:       ahora,
			})
			if err != nil {
				t.Fatalf("Error al crear pago: %v", err)
			}
			if pago.IDPago == 0 {
				t.Error("Se esperaba un ID de pago mayor a 0")
			}
			createdPagoID = pago.IDPago
		})

		t.Run("GetPago", func(t *testing.T) {
			pago, err := queries.GetPago(ctx, createdPagoID)
			if err != nil {
				t.Fatalf("Error al obtener pago: %v", err)
			}
			if pago.Monto != "6500.00" {
				t.Errorf("Monto esperado '6500.00', obtenido '%s'", pago.Monto)
			}
		})

		t.Run("ListPagos", func(t *testing.T) {
			// Si en db/sqlc/queries.sql.go el método se llama ListPagosBySuscripcion,
			// reemplaza queries.ListPagos(ctx) por queries.ListPagosBySuscripcion(ctx, createdSubID)
			pagos, err := queries.ListPagos(ctx, createdSubID)
			if err != nil {
				t.Fatalf("Error al listar pagos: %v", err)
			}
			if len(pagos) == 0 {
				t.Error("Se esperaba al menos 1 pago listado")
			}
		})

		t.Run("DeletePago", func(t *testing.T) {
			err := queries.DeletePago(ctx, createdPagoID)
			if err != nil {
				t.Fatalf("Error al eliminar pago: %v", err)
			}

			// Limpieza final de entidades asociadas
			_ = queries.DeleteSubscripcion(ctx, createdSubID)
			_ = queries.DeleteServicio(ctx, createdServiceID)
			_ = queries.DeleteUsuario(ctx, createdUserID)
		})
	})
}