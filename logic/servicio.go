package logic

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
)

// Categorías predefinidas según la regla de negocio (no texto libre).
// El slice mantiene el orden para mostrarlas en el formulario.
var Categorias = []string{"Streaming", "Música", "Software", "Gimnasio", "Salud", "Financiero", "Turismo", "Otros"}

var categoriasValidas = func() map[string]bool {
	validas := make(map[string]bool, len(Categorias))
	for _, c := range Categorias {
		validas[c] = true
	}
	return validas
}()

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
