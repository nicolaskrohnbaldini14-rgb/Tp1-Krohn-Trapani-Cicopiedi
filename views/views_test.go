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

func verificarContiene(t *testing.T, html string, esperados ...string) {
	t.Helper()
	for _, e := range esperados {
		if !strings.Contains(html, e) {
			t.Errorf("La página debería contener %q", e)
		}
	}
}

func verificarNoContiene(t *testing.T, html string, prohibidos ...string) {
	t.Helper()
	for _, p := range prohibidos {
		if strings.Contains(html, p) {
			t.Errorf("La página no debería contener %q", p)
		}
	}
}

func TestHomePage_SinSesion(t *testing.T) {
	html := renderizar(t, HomePage(nil, nil))

	verificarContiene(t, html,
		"<!doctype html>",
		"pico.min.css",
		`href="/login"`,
		`href="/registro"`,
		`href="/empresas/registro"`,
		"Todavía no hay servicios publicados.",
	)
	// el alta de servicios es solo para empresas
	verificarNoContiene(t, html, `action="/empresa/servicios"`, "Cerrar sesión")
}

func TestHomePage_ConSesionYServicios(t *testing.T) {
	servicios := []sqlc.Servicio{{IDServicio: 3, Nombre: "Netflix", Categoria: "Streaming"}}
	html := renderizar(t, HomePage(&Sesion{Nombre: "Nico", Tipo: "consumidor"}, servicios))

	verificarContiene(t, html, "Hola, Nico", "Cerrar sesión", `href="/cuenta"`, "Netflix", "Streaming")
	verificarNoContiene(t, html, `action="/empresa/servicios"`, "/eliminar")
}

func TestCuentaPage(t *testing.T) {
	suscripciones := []sqlc.ListSubscripcionesUsuarioRow{
		{IDSuscripcion: 1, Monto: "2500.00", Estado: "Activa"},
		{IDSuscripcion: 2, Monto: "1000.00", Estado: "Pausada"},
	}
	html := renderizar(t, CuentaPage(&Sesion{Nombre: "Nico", Tipo: "consumidor"}, suscripciones))

	verificarContiene(t, html,
		"Hola, Nico",
		"1 suscripción activa",
		"$ 2.500,00 por mes",
		`href="/cuenta/suscripciones"`,
		`href="/cuenta/suscripciones/nueva"`,
	)
}

func TestMisSuscripcionesPage(t *testing.T) {
	sesion := &Sesion{Nombre: "Nico", Tipo: "consumidor"}

	vacia := renderizar(t, MisSuscripcionesPage(sesion, nil, ""))
	verificarContiene(t, vacia, "Todavía no registraste ninguna suscripción.")

	suscripciones := []sqlc.ListSubscripcionesUsuarioRow{{
		IDSuscripcion:     5,
		Monto:             "2500.00",
		FechaInicio:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		FechaVencimiento:  time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC),
		UsuarioCuenta:     "cuenta",
		Estado:            "Activa",
		ServicioNombre:    "Netflix",
		ServicioCategoria: "Streaming",
	}}
	html := renderizar(t, MisSuscripcionesPage(sesion, suscripciones, "algo salió mal"))
	verificarContiene(t, html,
		`action="/cuenta/suscripciones/5/eliminar"`,
		"$ 2.500,00",
		"01/01/2026 al 31/12/2026",
		"algo salió mal",
	)
}

func TestNuevaSuscripcionPage_ConservaLosDatos(t *testing.T) {
	servicios := []sqlc.Servicio{
		{IDServicio: 1, Nombre: "Netflix", Categoria: "Streaming"},
		{IDServicio: 2, Nombre: "Spotify", Categoria: "Música"},
	}
	form := FormSuscripcion{IDServicio: 2, Monto: "999", FechaInicio: "2026-03-01", Estado: "Pausada"}
	html := renderizar(t, NuevaSuscripcionPage(&Sesion{Nombre: "Nico", Tipo: "consumidor"}, servicios, form, ""))

	verificarContiene(t, html,
		`<option value="2" selected>`,
		`value="999"`,
		`value="2026-03-01"`,
		`<option value="Pausada" selected>`,
	)
	// el usuario sale de la sesión: no se elige en el formulario
	verificarNoContiene(t, html, `name="id_usuario"`)
}

func TestEmpresaPage(t *testing.T) {
	servicios := []sqlc.Servicio{{IDServicio: 3, Nombre: "Netflix", Categoria: "Streaming"}}
	html := renderizar(t, EmpresaPage(&Sesion{Nombre: "Netflix SA", Tipo: "empresa"}, servicios, "", ""))

	verificarContiene(t, html, `action="/empresa/servicios"`, `action="/empresa/servicios/3/eliminar"`, `href="/empresa"`)
	verificarNoContiene(t, html, `href="/cuenta/suscripciones"`)
}

func TestServicioList_EscapaHTML(t *testing.T) {
	html := renderizar(t, ServicioList([]sqlc.Servicio{{IDServicio: 1, Nombre: "<script>alert(1)</script>"}}))
	verificarNoContiene(t, html, "<script>")
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
