# Suscripciones+ · TP5 — Renderizado del lado del servidor con templ

Aplicación web en Go para registrar suscripciones y gastos recurrentes
(streaming, gimnasio, software, etc.). En este TP se reemplazó el frontend
del TP4 (HTML + JavaScript que consumía una API JSON) por **vistas
renderizadas en el servidor con [templ](https://templ.guide/)**. El acceso a
datos sigue siendo con [sqlc](https://sqlc.dev/) sobre PostgreSQL en Docker.

La entrega **no contiene archivos JavaScript**: las altas y bajas se hacen con
formularios HTML clásicos y recargas de página completas.

# Requisitos previos

- [Go](https://go.dev/doc/install) 1.25 o superior (si la versión instalada es
  anterior, Go descarga automáticamente la indicada en `go.mod`)
- [Docker](https://docs.docker.com/get-docker/) con Docker Compose
- `make`
- [Hurl](https://hurl.dev/docs/installation.html) (solo para `make test-api`)
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html) (opcional: si no
  está instalado, el makefile lo ejecuta con `go run`)

templ **no** hace falta instalarlo: está declarado como herramienta en `go.mod`
(directiva `tool`) y se ejecuta con `go tool templ`.

# Ejecución

```bash
git clone https://github.com/nicolaskrohnbaldini14-rgb/Tp1-Krohn-Trapani-Cicopiedi.git
```
```bash
cd Tp1-Krohn-Trapani-Cicopiedi
```
```bash
git checkout tp5
```
```bash
make run-app
```

Luego abrir <http://localhost:8080>. Para detener la aplicación: `Ctrl+C`.

`make run-app` realiza todas las tareas necesarias:

1. **Tareas previas**
   - `sqlc generate`: genera el acceso a datos en `db/sqlc/`.
   - `templ generate`: compila los componentes `.templ` a código Go (`*_templ.go`).
   - Compila el ejecutable en `tmp/`.
   - Borra contenedores y volúmenes previos (`docker compose down -v`).
   - Levanta PostgreSQL (`docker compose up -d --wait`) y espera a que el
     healthcheck (`pg_isready`) indique que la base está lista. El esquema se
     crea al iniciar el contenedor (`db/schema/schema.sql` se monta en
     `docker-entrypoint-initdb.d`).
2. **Ejecución** del servidor en el puerto 8080.
3. **Tareas posteriores**: al cortar con `Ctrl+C` (o si la app termina) se
   borran los contenedores y volúmenes.

## Otros comandos

| Comando | Qué hace |
|---|---|
| `make` / `make test-api` | Genera, compila, levanta la base, corre los tests de Go y prueba la aplicación con Hurl. Al final borra contenedores y volúmenes. |
| `make generate` | Solo genera el código de sqlc y templ. |
| `make build` | Genera y compila. |
| `make docker-down` | Borra contenedores y volúmenes manualmente. |
| `make clean` | Borra el ejecutable. |

# Arquitectura

```
Navegador ──(GET / , POST de formularios)──▶ handlers/ ──▶ logic/ ──▶ db/sqlc ──▶ PostgreSQL
                     ◀──── HTML completo ──── views/ (templ)
```

- **`views/`** — componentes templ (capa de presentación).
- **`handlers/`** — handlers HTTP: leen el formulario, llaman a la lógica y
  renderizan la vista o redirigen. Reemplazan a los handlers JSON del TP3.
- **`logic/`** — reglas de negocio de cada entidad (validaciones, unicidad,
  existencia de claves foráneas, etc.). Es la misma capa del TP3, sin la parte HTTP.
- **`db/`** — esquema, queries y código generado por sqlc.

## Componentes templ

Se partió del `index.html` del TP4 y se lo dividió en componentes:

| Archivo | Componente | Descripción |
|---|---|---|
| `layout.templ` | `Layout(titulo, contenido)` | Estructura base (`<!DOCTYPE>`, `<head>`, `<body>`), link a **Pico.css** y a `static/styles.css`. Recibe un `templ.Component` como contenido. |
| `index.templ` | `IndexPage(titulo, datos)` | Página principal: usa `Layout` y renderiza los formularios y listas de cada entidad, y el mensaje de error si lo hay. |
| `usuario_list.templ` / `usuario_form.templ` | `UsuarioList([]sqlc.Usuario)` / `UsuarioForm()` | Tabla y formulario de usuarios. |
| `servicio_list.templ` / `servicio_form.templ` | `ServicioList([]sqlc.Servicio)` / `ServicioForm()` | Tabla y formulario de servicios. |
| `suscripcion_list.templ` / `suscripcion_form.templ` | `SuscripcionList(...)` / `SuscripcionForm(usuarios, servicios)` | Tabla (con gasto mensual activo) y formulario de suscripciones. |
| `pago_list.templ` / `pago_form.templ` | `PagoList(...)` / `PagoForm(suscripciones)` | Historial y formulario de pagos. |
| `componentes.templ` | `BotonEliminar`, `MensajeVacio` | Componentes reutilizables y funciones de formato (montos y fechas). |

Cada lista muestra un mensaje cuando está vacía. Los estilos mantienen la
paleta de la primera entrega (azul acero `#567c8d`, azul petróleo
`#002f44`/`#004b6e` y crema `#ece6d4`) aplicada sobre Pico.css.

## Rutas

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/` | Obtiene todos los registros con los métodos `List...` de sqlc y renderiza `IndexPage`. |
| `POST` | `/usuarios` | Crea un usuario. |
| `POST` | `/servicios` | Crea un servicio. |
| `POST` | `/suscripciones` | Crea una suscripción. |
| `POST` | `/pagos` | Registra un pago. |
| `POST` | `/{usuarios,servicios,suscripciones,pagos}/{id}/eliminar` | Elimina el registro. Se usa `POST` porque los formularios HTML no pueden enviar `DELETE`. |
| `GET` | `/static/styles.css` | Hoja de estilos propia. |

### Patrón Post/Redirect/Get

Los handlers `POST` parsean el formulario (`r.ParseForm()`), convierten los
valores (ids, montos y fechas de `<input type="date">`) y llaman al método
`Create...`/`Delete...` correspondiente a través de la capa de lógica:

- **Si sale bien** responden `303 See Other` con `Location: /`
  (`http.Redirect`). El navegador pide `/` de nuevo y se ve la lista
  actualizada; recargar la página no reenvía el formulario.
- **Si falla una regla de negocio** se vuelve a renderizar la página con el
  mensaje de error y el código HTTP correspondiente: `400` datos inválidos,
  `404` registro inexistente, `409` conflicto (email o nombre repetido,
  suscripción activa duplicada, servicio con suscripciones).

# Documentación de persistencia

El modelo de datos está compuesto por 4 tablas relacionales, definidas en
[`db/schema/schema.sql`](db/schema/schema.sql):

| Tabla | Descripción |
|---|---|
| `usuario` | Cuenta de la persona que usa la app (nombre, email, contraseña). |
| `servicio` | Catálogo de servicios disponibles (nombre + categoría), normalizado para poder agrupar el gasto por categoría. |
| `suscripcion` | Relaciona un `usuario` con un `servicio`: monto, fecha de inicio, fecha de vencimiento, credenciales de acceso al servicio (usuario/contraseña de esa cuenta) y estado (activa/pausada/cancelada). |
| `pago` | Historial real de cobros de una suscripción. Permite calcular el gasto efectivo mes a mes, no solo el proyectado. |

**Relaciones:**

<img width="1024" height="559" alt="ModeloRelacionalWEB" src="https://github.com/user-attachments/assets/a5284446-98ce-4dc8-b4a0-a3bec07d5b3e" />

- `suscripcion.id_usuario` → `usuario.id_usuario` (`ON DELETE CASCADE`)
- `suscripcion.id_servicio` → `servicio.id_servicio`
- `pago.id_suscripcion` → `suscripcion.id_suscripcion` (`ON DELETE CASCADE`)

El acceso a datos se genera con sqlc a partir de
[`db/queries/queries.sql`](db/queries/queries.sql). Para la página principal
se agregaron dos consultas con `JOIN`, así las tablas muestran nombres en
lugar de ids:

- `ListSuscripciones`: todas las suscripciones con el nombre del usuario y del servicio.
- `ListTodosLosPagos`: todos los pagos con el usuario y el servicio de la suscripción.

El código generado (`db/sqlc/`) y los `*_templ.go` **no se versionan**: se
regeneran con `make generate` (incluido en `make run-app`).

# Reglas de negocio

- **Usuarios:** nombre, email y password son obligatorios. El email tiene que tener un formato válido y no puede repetirse. La password nunca se muestra.
- **Servicios:** el nombre es único. La categoría debe ser una de `Streaming`, `Música`, `Software`, `Gimnasio`, `Salud`, `Financiero`, `Turismo` u `Otros`. No se puede borrar un servicio que tenga suscripciones.
- **Suscripciones:**
  - El usuario y el servicio tienen que existir.
  - El monto tiene que ser mayor a cero y la fecha de inicio anterior a la de vencimiento.
  - El estado es `Activa` (por defecto), `Pausada` o `Cancelada`.
  - Un usuario no puede tener dos suscripciones activas al mismo servicio.
  - La contraseña de la cuenta nunca se muestra.
- **Pagos:** la suscripción tiene que existir, el monto tiene que ser mayor a cero y la fecha de pago no puede ser posterior al vencimiento de la suscripción.

# Pruebas

`make` (o `make test-api`) corre:

1. **Tests de Go** (`go test ./...`):
   - `main_test.go`: CRUD de cada tabla con sqlc contra PostgreSQL en Docker,
     más `TestListadosPaginaPrincipal` para las consultas nuevas.
   - `views/views_test.go`: renderiza los componentes templ sin base de datos
     (mensajes de lista vacía, datos y formato de montos/fechas, escape de HTML).
2. **Pruebas Hurl** (`hurl/`): envían los formularios como lo haría el
   navegador (`[FormParams]`) y verifican el `303` hacia `/`, que los datos
   aparezcan en el HTML (consultas XPath) y los códigos y mensajes de error.

# Ejercicios del TP5

En [`ejercicios/productos/`](ejercicios/productos/) están los ejercicios 1 a 4
del práctico (productos con SQLite, sqlc y templ). Es un módulo de Go
independiente: usa `mattn/go-sqlite3`, que necesita `gcc` (cgo).

```bash
cd ejercicios/productos
```
```bash
make run
```

Genera el código con sqlc y templ y levanta el servidor en
<http://localhost:8080> (crea `products.db` si no existe). `GET /` lista los
productos y el formulario hace `POST /products`, que redirige a `/`.

# Estructura del proyecto

```
.
├── main.go                 # Conexión a la base, armado del mux y servidor HTTP
├── main_test.go            # Tests de integración de sqlc
├── handlers/web.go         # Handlers HTML (GET / y POST de formularios)
├── logic/                  # Reglas de negocio de cada entidad
├── views/                  # Componentes templ + views_test.go
├── static/styles.css       # Ajustes de estilo sobre Pico.css
├── db/
│   ├── schema/schema.sql
│   ├── queries/queries.sql
│   └── sqlc/               # Generado por sqlc (no se versiona)
├── hurl/                   # Pruebas de la aplicación
├── ejercicios/productos/   # Ejercicios 1 a 4 del TP5
├── docker-compose.yml      # PostgreSQL con healthcheck
├── sqlc.yaml
└── makefile
```
