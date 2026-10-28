package views

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"

	"tp2-krohn-trapani-cicopiedi/db/sqlc"
)

func renderizar(t *testing.T, c templ.Component) string {
	t.Helper()
	var buf bytes.Buffer
	if err := c.Render(context.Background(), &buf); err != nil {
		t.Fatalf("Error al renderizar el componente: %v", err)
	}
	return buf.String()
}

func TestIndexPage_SinDatosMuestraMensajesVacios(t *testing.T) {
	html := renderizar(t, IndexPage("Suscripciones+", DatosInicio{}))

	esperados := []string{
		"<!doctype html>",
		"<title>Suscripciones+</title>",
		"pico.min.css",
		"Todavía no hay usuarios cargados.",
		"Todavía no hay servicios cargados.",
		"Todavía no hay suscripciones registradas.",
		"Todavía no hay pagos registrados.",
		`action="/usuarios"`,
		`action="/servicios"`,
	}
	for _, e := range esperados {
		if !strings.Contains(html, e) {
			t.Errorf("La página debería contener %q", e)
		}
	}

	// sin usuarios ni servicios no se ofrece el formulario de suscripciones
	if strings.Contains(html, `action="/suscripciones"`) {
		t.Error("No debería mostrarse el formulario de suscripciones sin usuarios ni servicios")
	}
}

func TestIndexPage_ConDatos(t *testing.T) {
	datos := DatosInicio{
		Usuarios:  []sqlc.Usuario{{IDUsuario: 7, Nombre: "Nico", Email: "nico@example.com"}},
		Servicios: []sqlc.Servicio{{IDServicio: 3, Nombre: "Netflix", Categoria: "Streaming"}},
		Suscripciones: []sqlc.ListSuscripcionesRow{{
			IDSuscripcion:     5,
			Monto:             "2500.00",
			FechaInicio:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			FechaVencimiento:  time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
			UsuarioCuenta:     "cuenta",
			Estado:            "Activa",
			UsuarioNombre:     "Nico",
			ServicioNombre:    "Netflix",
			ServicioCategoria: "Streaming",
		}},
		Error: "algo salió mal",
	}

	html := renderizar(t, IndexPage("Suscripciones+", datos))

	esperados := []string{
		"nico@example.com",
		`action="/usuarios/7/eliminar"`,
		`action="/servicios/3/eliminar"`,
		`action="/suscripciones/5/eliminar"`,
		`action="/suscripciones"`,
		`<option value="7">Nico</option>`,
		"$ 2.500,00",
		"01/01/2026 al 31/12/2026",
		"algo salió mal",
	}
	for _, e := range esperados {
		if !strings.Contains(html, e) {
			t.Errorf("La página debería contener %q", e)
		}
	}
}

func TestUsuarioList_EscapaHTML(t *testing.T) {
	html := renderizar(t, UsuarioList([]sqlc.Usuario{{IDUsuario: 1, Nombre: "<script>alert(1)</script>"}}))

	if strings.Contains(html, "<script>") {
		t.Error("templ debería escapar el contenido ingresado por el usuario")
	}
}

func TestFormatearNumero(t *testing.T) {
	casos := map[float64]string{
		0:          "0,00",
		999.5:      "999,50",
		1500:       "1.500,00",
		1234567.89: "1.234.567,89",
		-2500:      "-2.500,00",
	}
	for valor, esperado := range casos {
		if obtenido := formatearNumero(valor); obtenido != esperado {
			t.Errorf("formatearNumero(%v) = %q, se esperaba %q", valor, obtenido, esperado)
		}
	}
}
