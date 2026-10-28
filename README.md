# Suscripciones+ · TP5 — Renderizado del lado del servidor con templ

Aplicación web en Go para registrar suscripciones y gastos recurrentes
(streaming, gimnasio, software, etc.). En este TP se reemplazó el frontend
del TP4 (HTML + JavaScript que consumía una API JSON) por **vistas
renderizadas en el servidor con [templ](https://templ.guide/)**. El acceso a
datos sigue siendo con [sqlc](https://sqlc.dev/) sobre PostgreSQL en Docker.

La entrega **no contiene archivos JavaScript**: las altas y bajas se hacen con
formularios HTML clásicos y recargas de página completas.

La aplicación está separada en secciones según quién la usa:

- **Inicio (`/`)**: presentación, catálogo público de servicios y links a
  **Iniciar sesión** y **Registrarse**.
- **Consumidores**: al ingresar ven *"Hola, {nombre}"* con un resumen de su
  gasto y dos opciones: **ver sus suscripciones** y **crear una nueva**.
- **Empresas**: se registran desde *"¿Sos una empresa? Registrá tu empresa"*
  (al pie de página) y son las únicas que pueden crear y borrar servicios.

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

La base arranca vacía. Para probarla: registrar una empresa desde el link del
pie de página y crear algunos servicios, cerrar sesión, registrarse como
consumidor y crear suscripciones.

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
Navegador ──(GET de páginas, POST de formularios + cookie de sesión)──▶ handlers/ ──▶ logic/ ──▶ db/sqlc ──▶ PostgreSQL
                         ◀──────────── HTML completo ──────────── views/ (templ)
```

- **`views/`** — componentes templ (capa de presentación).
- **`handlers/`** — handlers HTTP (`web.go`) y sesiones (`sesiones.go`): leen
  el formulario, verifican la sesión y el tipo de cuenta, llaman a la lógica
  y renderizan la vista o redirigen. Reemplazan a los handlers JSON del TP3.
- **`logic/`** — reglas de negocio de cada entidad (validaciones, unicidad,
  existencia de claves foráneas, hash y verificación de contraseñas). Es la
  misma capa del TP3, sin la parte HTTP.
- **`db/`** — esquema, queries y código generado por sqlc.

## Cuentas y sesiones

- La tabla `usuario` tiene una columna `tipo`: `consumidor` (por defecto) o
  `empresa`. `/registro` crea consumidores y `/empresas/registro` crea empresas.
- Las contraseñas se guardan **hasheadas con bcrypt**
  (`golang.org/x/crypto/bcrypt`), nunca en texto plano.
- Al registrarse o iniciar sesión se genera un token aleatorio que se guarda en
  memoria del servidor y se envía en una cookie `HttpOnly` y `SameSite=Lax`
  llamada `sesion` (dura 24 h). Si se reinicia el servidor hay que volver a ingresar.
- Las páginas privadas usan `soloTipo(...)`: sin sesión redirigen a `/login`
  y con el tipo de cuenta equivocado responden `403`.
- El usuario de una suscripción se toma de la sesión, no del formulario, y un
  consumidor solo puede ver y borrar sus propias suscripciones.

## Componentes templ

Se partió del `index.html` del TP4 y se lo dividió en componentes y páginas:

| Archivo | Componentes | Descripción |
|---|---|---|
| `layout.templ` | `Layout(titulo, sesion, contenido)`, `Navegacion` | Estructura base (`<!DOCTYPE>`, `<head>`, `<body>`), link a **Pico.css** y a `static/styles.css`. La barra muestra *Iniciar sesión / Registrarse* o *Hola, {nombre} / Cerrar sesión* según la sesión. |
| `home.templ` | `HomePage` | Inicio público con el catálogo de servicios en modo lectura. |
| `login.templ` | `LoginPage` | Formulario de inicio de sesión. |
| `usuario_form.templ` | `RegistroPage`, `UsuarioForm` | Registro de consumidores y de empresas. |
| `cuenta.templ` | `CuentaPage` | *"Hola, {nombre}"*, resumen del gasto y las dos opciones del consumidor. |
| `suscripcion_list.templ` | `MisSuscripcionesPage`, `SuscripcionList` | Tabla con las suscripciones del usuario y su gasto mensual activo. |
| `suscripcion_form.templ` | `NuevaSuscripcionPage`, `SuscripcionForm` | Alta de suscripción. Si hay un error conserva lo ingresado (salvo la contraseña). |
| `servicio_form.templ` | `EmpresaPage`, `ServicioForm` | Panel de empresas con el alta de servicios. |
| `servicio_list.templ` | `ServicioList`, `CatalogoServicios` | Tabla editable (empresas) y catálogo de solo lectura (inicio). |
| `componentes.templ` | `BotonEliminar`, `MensajeVacio`, `AvisoError`, `ErrorPage` | Componentes reutilizables y formato de montos y fechas. |

Cada lista muestra un mensaje cuando está vacía. Los estilos mantienen la
paleta de la primera entrega (azul acero `#567c8d`, azul petróleo
`#002f44`/`#004b6e` y crema `#ece6d4`) aplicada sobre Pico.css.

## Rutas

| Método | Ruta | Acceso | Descripción |
|---|---|---|---|
| `GET` | `/` | Público | Inicio y catálogo (usa `ListServicios` de sqlc). |
| `GET` / `POST` | `/login` | Público | Formulario / inicio de sesión. |
| `GET` / `POST` | `/registro` | Público | Registro de consumidores. |
| `GET` / `POST` | `/empresas/registro` | Público | Registro de empresas. |
| `POST` | `/logout` | Con sesión | Cierra la sesión. |
| `GET` | `/cuenta` | Consumidor | *"Hola, {nombre}"* y opciones. |
| `GET` | `/cuenta/suscripciones` | Consumidor | Sus suscripciones (`ListSubscripcionesUsuario`). |
| `GET` | `/cuenta/suscripciones/nueva` | Consumidor | Formulario de alta. |
| `POST` | `/cuenta/suscripciones` | Consumidor | Crea la suscripción. |
| `POST` | `/cuenta/suscripciones/{id}/eliminar` | Consumidor | Borra una suscripción propia. |
| `GET` | `/empresa` | Empresa | Panel con el catálogo de servicios. |
| `POST` | `/empresa/servicios` | Empresa | Crea un servicio. |
| `POST` | `/empresa/servicios/{id}/eliminar` | Empresa | Borra un servicio sin suscripciones. |
| `GET` | `/static/styles.css` | Público | Hoja de estilos propia. |

Los borrados usan `POST .../eliminar` porque los formularios HTML no pueden
enviar `DELETE`.

### Patrón Post/Redirect/Get

Los handlers `POST` parsean el formulario (`r.ParseForm()`), convierten los
valores (ids, montos y fechas de `<input type="date">`) y llaman al método
`Create...`/`Delete...` correspondiente a través de la capa de lógica:

- **Si sale bien** responden `303 See Other` (`http.Redirect`) hacia la página
  que muestra el resultado (`/cuenta/suscripciones`, `/empresa`, etc.). El
  navegador la pide de nuevo y se ve la lista actualizada; recargar no reenvía
  el formulario.
- **Si falla** se vuelve a renderizar el formulario con el mensaje de error y
  el código HTTP correspondiente: `400` datos inválidos, `401` credenciales
  incorrectas, `404` registro inexistente, `409` conflicto (email o nombre
  repetido, suscripción activa duplicada, servicio con suscripciones).

# Documentación de persistencia

El modelo de datos está compuesto por 4 tablas relacionales, definidas en
[`db/schema/schema.sql`](db/schema/schema.sql):

| Tabla | Descripción |
|---|---|
| `usuario` | Cuenta de la persona o empresa que usa la app (nombre, email, contraseña hasheada y `tipo`: `consumidor` o `empresa`). |
| `servicio` | Catálogo de servicios disponibles (nombre + categoría), normalizado para poder agrupar el gasto por categoría. |
| `suscripcion` | Relaciona un `usuario` con un `servicio`: monto, fecha de inicio, fecha de vencimiento, credenciales de acceso al servicio (usuario/contraseña de esa cuenta) y estado (activa/pausada/cancelada). |
| `pago` | Historial real de cobros de una suscripción. Permite calcular el gasto efectivo mes a mes, no solo el proyectado. |

**Relaciones:**

<img width="1024" height="559" alt="ModeloRelacionalWEB" src="https://github.com/user-attachments/assets/a5284446-98ce-4dc8-b4a0-a3bec07d5b3e" />

- `suscripcion.id_usuario` → `usuario.id_usuario` (`ON DELETE CASCADE`)
- `suscripcion.id_servicio` → `servicio.id_servicio`
- `pago.id_suscripcion` → `suscripcion.id_suscripcion` (`ON DELETE CASCADE`)

El acceso a datos se genera con sqlc a partir de
[`db/queries/queries.sql`](db/queries/queries.sql). `CreateUsuario` ahora
recibe el `tipo` de cuenta, y el login usa `GetUsuarioByEmail` para comparar el
hash de la contraseña. La tabla `pago` y su lógica se mantienen del TP3, pero
la interfaz de este TP no los muestra.

El código generado (`db/sqlc/`) y los `*_templ.go` **no se versionan**: se
regeneran con `make generate` (incluido en `make run-app`).

# Reglas de negocio

- **Usuarios:** nombre, email y password son obligatorios. El email tiene que tener un formato válido y no puede repetirse. La password se guarda hasheada y nunca se muestra.
- **Acceso:** solo las empresas crean y borran servicios; solo los consumidores crean suscripciones, y cada uno ve y borra únicamente las suyas.
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
   - `main_test.go`: CRUD de cada tabla con sqlc contra PostgreSQL en Docker.
   - `views/views_test.go`: renderiza los componentes templ sin base de datos
     (links según la sesión, que el inicio no permita crear servicios, datos
     conservados en el formulario, formato de montos/fechas, escape de HTML).
2. **Pruebas Hurl** (`hurl/`): navegan la aplicación como lo haría el
   navegador (Hurl guarda las cookies de sesión entre pedidos) y verifican
   redirecciones, el HTML con consultas XPath y los códigos de error.

| Archivo | Qué prueba |
|---|---|
| `publico.hurl` | Inicio, login y registro; sin sesión las páginas privadas redirigen a `/login`. |
| `cuentas.hurl` | Registro, *"Hola, {nombre}"*, logout, login correcto e incorrecto, validaciones, `403` de un consumidor en el panel de empresas. |
| `empresa.hurl` | Registro de empresa y ABM de servicios. |
| `suscripcion.hurl` | Alta y baja de suscripciones, validaciones, que otro consumidor no pueda borrarlas y que no se pueda borrar un servicio con suscripciones. |

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
├── handlers/
│   ├── web.go              # Rutas, control de acceso y handlers HTML
│   └── sesiones.go         # Sesiones con cookie HttpOnly
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
