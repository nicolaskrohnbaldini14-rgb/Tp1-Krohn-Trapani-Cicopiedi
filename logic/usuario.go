package logic

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
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
