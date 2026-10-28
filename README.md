# Requisitos previos

- [Go](https://go.dev/doc/install) 1.25 
- [Docker](https://docs.docker.com/get-docker/) 
- [sqlc](https://docs.sqlc.dev/en/latest/overview/install.html)
- [Hurl](https://hurl.dev/docs/installation.html)

# Ejecución

## Clonar repositorio
```bash
git clone https://github.com/nicolaskrohnbaldini14-rgb/Tp1-Krohn-Trapani-Cicopiedi.git
```
## Posicionarse en carpeta del proyecto
```bash
cd Tp1-Krohn-Trapani-Cicopiedi
```
## Cambiar a rama requerida
```bash
git checkout tp4
```
## Levantar la aplicación (TP4)
```bash
make run-app
```

`make run-app` realiza todas las tareas necesarias:

1. **Tareas previas:** borra contenedores y volúmenes anteriores, levanta PostgreSQL
   en Docker y espera a que el healthcheck indique que la base está lista, ejecuta
   `sqlc generate` y compila el ejecutable en `tmp/`.
2. **Ejecución:** deja corriendo el servidor en <http://localhost:8080>, que sirve
   tanto la API REST como el frontend.
3. **Tareas posteriores:** al detenerlo con `Ctrl+C` borra los contenedores y volúmenes.

La base arranca vacía. Para probarla: registrar una empresa desde el link
*"¿Sos una empresa? Registrá tu empresa"* del pie de página y crear algunos
servicios, cerrar sesión, registrarse como consumidor y crear suscripciones.

## Ejecutar pruebas (tests de sqlc + pruebas Hurl de la API)
```bash
make
```

# Frontend (TP4)

La interfaz está en [`static/`](static/) y la sirve el mismo servidor Go
(`http.FileServer` registrado en `/`), por lo que el navegador y la API
comparten origen: no hace falta configurar CORS y la cookie de sesión viaja
sola en cada `fetch`.

La aplicación está separada en páginas según quién la usa:

| Página | Quién | Contenido |
|---|---|---|
| `index.html` (`/`) | Todos | Presentación, catálogo de servicios **de solo lectura** y links a **Iniciar sesión** y **Registrarse**. |
| `login.html` | Todos | Inicio de sesión. |
| `registro.html` | Todos | Registro de consumidores. Con `?tipo=empresa` registra empresas (link al pie de página). |
| `cuenta.html` | Consumidores | *"Hola, {nombre}"*, resumen del gasto y dos opciones: ver las suscripciones o crear una nueva. |
| `suscripciones.html` | Consumidores | Lista de sus suscripciones con el gasto mensual activo y botón para eliminar. |
| `nueva-suscripcion.html` | Consumidores | Formulario de alta de suscripción. |
| `empresa.html` | Empresas | Alta y baja de servicios del catálogo. |

Todas las páginas comparten [`static/app.js`](static/app.js) y
[`static/styles.css`](static/styles.css). Cada una se identifica con
`<body data-pagina="...">` y al cargar `app.js` ejecuta la función de esa página.

Funcionamiento de `app.js`:

- **Sesión:** al cargar cualquier página pide `GET /sesion`. Según la
  respuesta arma la barra de navegación (*Iniciar sesión / Registrarse* o
  *Hola, {nombre} / Cerrar sesión*). Las páginas privadas redirigen al login
  si no hay sesión, o a la sección que corresponde si la cuenta es de otro tipo.
- **Carga de datos:** cada página hace su `fetch` al cargar
  (`GET /servicios`, `GET /suscripciones/usuario/{id}`) y genera
  dinámicamente un `<li>` por entidad, con un botón **Eliminar** donde corresponde.
- **Alta:** los formularios tienen un listener de `submit` que hace
  `preventDefault()`, arma el JSON con los datos de los inputs y envía un
  `POST`. Al crear un servicio se vuelve a pedir la lista completa; al crear
  una suscripción se pasa a la lista de suscripciones, que se carga de nuevo.
- **Borrado con delegación de eventos:** como los botones se crean
  dinámicamente, hay un único listener de `click` sobre cada lista que detecta
  el botón presionado, toma el ID desde `data-id` y hace el `DELETE`. Tras un
  borrado exitoso se refresca la lista.
- **Errores:** los mensajes de la API (email repetido, servicio con
  suscripciones, fechas inválidas, etc.) se muestran en un aviso arriba de la
  página, y el formulario conserva lo que se había escrito.
- **Seguridad:** los datos se insertan con `textContent`, nunca con
  `innerHTML`, para no inyectar HTML ingresado por los usuarios. Además, los
  permisos los controla la API: ocultar un botón en la página no alcanza.

Los estilos usan **Flexbox** (barra, formularios y listas) y **CSS Grid**
(catálogo, opciones de la cuenta y panel de empresas), con la paleta de la
primera entrega: azul acero `#567c8d`, azul petróleo `#002f44`/`#004b6e` y
crema `#ece6d4`.



# Documentación de persistencia

El modelo de datos está compuesto por 4 tablas relacionales, definidas en
[`db/schema/schema.sql`](db/schema/schema.sql):

| Tabla | Descripción |
|---|---|
| `usuario` | Cuenta de la persona o empresa que usa la app (nombre, email, contraseña hasheada con bcrypt y `tipo`: `consumidor` o `empresa`). |
| `servicio` | Catálogo de servicios disponibles (nombre + categoría), normalizado para poder agrupar el gasto por categoría. |
| `suscripcion` | Relaciona un `usuario` con un `servicio`: monto, fecha de inicio, fecha de vencimiento, credenciales de acceso al servicio (usuario/contraseña de esa cuenta) y estado (activa/pausada/cancelada). |
| `pago` | Historial real de cobros de una suscripción. Permite calcular el gasto efectivo mes a mes, no solo el proyectado. |

**Relaciones:**

<img width="1024" height="559" alt="ModeloRelacionalWEB" src="https://github.com/user-attachments/assets/a5284446-98ce-4dc8-b4a0-a3bec07d5b3e" />

- `suscripcion.id_usuario` → `usuario.id_usuario` (`ON DELETE CASCADE`)
- `suscripcion.id_servicio` → `servicio.id_servicio`
- `pago.id_suscripcion` → `suscripcion.id_suscripcion` (`ON DELETE CASCADE`)

El acceso a datos se genera con [sqlc](https://sqlc.dev/) a partir de las
queries en [`db/queries/queries.sql`](db/queries/queries.sql), que cubren las
operaciones CRUD (Create, Get, List, Update, Delete) para cada una de las 4
entidades. El código generado vive en `db/sqlc/` (`models.go`, `queries.sql.go`,
`db.go`) y **no debe editarse a mano** — se regenera automáticamente cada vez
que se corre `sqlc generate` (parte de `make run-app` y `make test-api`).

Los tests de integración (`main_test.go`) validan el CRUD contra una base
Postgres real levantada en Docker. Hay un test independiente por tabla
(`TestUsuario_CRUD`, `TestServicio_CRUD`, `TestSubscripciones_CRUD` y
`TestPago_CRUD`), y cada uno arranca con las tablas vacías y crea los datos
que necesita con funciones auxiliares compartidas (por ejemplo, el test de
pagos crea primero un usuario, un servicio y una suscripción). Se puede correr
uno solo con `go test -v -run TestPago_CRUD .`

# API REST

La lógica de negocio y los handlers HTTP de cada entidad están en `logic/`, y
las rutas se registran en `main.go`. Todas las respuestas exitosas con cuerpo
son JSON. Los errores devuelven un mensaje de texto con el código HTTP
correspondiente.

### Sesión y permisos

| Método | Ruta | Respuestas |
|---|---|---|
| `POST` | `/sesion` | `200` usuario + cookie de sesión · `401` email o contraseña incorrectos |
| `GET` | `/sesion` | `200` usuario de la sesión · `401` sin sesión |
| `DELETE` | `/sesion` | `204` cierra la sesión |

```json
{ "email": "nico@example.com", "password": "secreto" }
```

- La sesión es un token aleatorio guardado en memoria del servidor y enviado
  en una cookie `HttpOnly` y `SameSite=Lax` llamada `sesion` (dura 24 h). Si
  se reinicia el servidor hay que volver a ingresar.
- Las contraseñas se guardan **hasheadas con bcrypt** (`golang.org/x/crypto/bcrypt`).
- `logic/sesion.go` tiene el middleware `Proteger`, que en `main.go` se aplica
  solo a los métodos que lo necesitan:

| Recurso | Público | Requiere sesión |
|---|---|---|
| Usuarios | `GET`, `POST /usuarios` (registro) | `PUT`/`DELETE`: solo la propia cuenta (`403` si es otra) |
| Servicios | `GET` (catálogo) | `POST`/`PUT`/`DELETE`: solo **empresas** (`403` para consumidores) |
| Suscripciones | — | Solo **consumidores**. Se crean a nombre del usuario de la sesión y cada uno ve y modifica solo las suyas (`404` si son de otro) |
| Pagos | — | Solo **consumidores**, sobre sus propias suscripciones (`404` si son de otro) |

Sin sesión, las rutas protegidas responden `401`.

### Usuarios

| Método | Ruta | Respuestas |
|---|---|---|
| `POST` | `/usuarios` | `201` creado · `400` datos inválidos · `409` email ya registrado |
| `GET` | `/usuarios` | `200` lista de usuarios |
| `GET` | `/usuarios/{id}` | `200` · `400` ID no numérico · `404` no encontrado |
| `PUT` | `/usuarios/{id}` | `200` · `400` · `404` · `409` email en uso |
| `DELETE` | `/usuarios/{id}` | `204` · `404` |

```json
{ "nombre": "Nico", "email": "nico@example.com", "password": "secreto", "tipo": "consumidor" }
```

`tipo` es `consumidor` (por defecto) o `empresa`.

### Servicios

| Método | Ruta | Respuestas |
|---|---|---|
| `POST` | `/servicios` | `201` · `400` datos o categoría inválidos · `409` nombre repetido |
| `GET` | `/servicios` | `200` lista de servicios |
| `GET` | `/servicios/{id}` | `200` · `400` · `404` |
| `PUT` | `/servicios/{id}` | `200` · `400` · `404` · `409` nombre repetido |
| `DELETE` | `/servicios/{id}` | `204` · `404` · `409` tiene suscripciones asociadas |

```json
{ "nombre": "Netflix", "categoria": "Streaming" }
```

### Suscripciones

| Método | Ruta | Respuestas |
|---|---|---|
| `POST` | `/suscripciones` | `201` · `400` datos inválidos · `403` a nombre de otro usuario · `404` servicio inexistente · `409` ya tiene una activa |
| `GET` | `/suscripciones/{id}` | `200` · `400` · `404` |
| `GET` | `/suscripciones/usuario/{id_usuario}` | `200` suscripciones del usuario · `403` si es otro usuario |
| `PUT` | `/suscripciones/{id}` | `200` · `400` · `404` · `409` |
| `DELETE` | `/suscripciones/{id}` | `204` · `404` |

`id_usuario` es opcional: si no se envía se usa el de la sesión.

```json
{
  "id_usuario": 1,
  "id_servicio": 1,
  "monto": "2500.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "nico@example.com",
  "password_cuenta": "passcuenta123",
  "estado": "Activa"
}
```

### Pagos

| Método | Ruta | Respuestas |
|---|---|---|
| `POST` | `/pagos` | `201` · `400` datos inválidos o fecha fuera de vigencia · `404` suscripción inexistente |
| `GET` | `/pagos/{id}` | `200` · `400` · `404` |
| `GET` | `/pagos/suscripcion/{id_suscripcion}` | `200` historial de pagos · `404` suscripción inexistente |
| `DELETE` | `/pagos/{id}` | `204` · `404` |
| `PUT` / `PATCH` | `/pagos/{id}` | `405` los pagos son un historial y no se modifican |

```json
{ "id_suscripcion": 1, "monto": "2500.00", "fecha_pago": "2026-06-15T00:00:00Z" }
```

# Reglas de negocio

- **Usuarios:** nombre, email y password son obligatorios. El email tiene que tener un formato válido y no puede repetirse. La password se guarda hasheada y nunca se devuelve en las respuestas.
- **Acceso:** solo las empresas crean, modifican y borran servicios; solo los consumidores crean suscripciones y pagos, y cada uno ve y modifica únicamente los suyos.
- **Servicios:** el nombre es único. La categoría debe ser una de `Streaming`, `Música`, `Software`, `Gimnasio`, `Salud`, `Financiero`, `Turismo` u `Otros`, y no se puede cambiar después de creado. No se puede borrar un servicio que tenga suscripciones.
- **Suscripciones:**
  - El usuario y el servicio tienen que existir.
  - El monto tiene que ser mayor a cero y la fecha de inicio anterior a la de vencimiento.
  - El estado es `Activa` (por defecto), `Pausada` o `Cancelada`.
  - Un usuario no puede tener dos suscripciones activas al mismo servicio.
  - La `password_cuenta` nunca se devuelve en las respuestas.
- **Pagos:** la suscripción tiene que existir, el monto tiene que ser mayor a cero y la fecha de pago no puede ser posterior al vencimiento de la suscripción.

# Pruebas de los endpoints

Las pruebas están en `hurl/`. Hurl guarda las cookies entre pedidos del mismo
archivo, así que cada archivo inicia sesión con la cuenta que necesita antes
de las acciones protegidas.

| Archivo | Qué prueba |
|---|---|
| `frontend.hurl` | Que se sirvan las páginas, que el inicio tenga los links a login y registro y que no tenga el alta de servicios. |
| `usuario.hurl` | Registro, login correcto e incorrecto, `GET /sesion`, CRUD de la propia cuenta, `403` sobre cuentas ajenas y cierre de sesión. |
| `servicio.hurl` | CRUD de servicios como empresa, `401` sin sesión y `403` para consumidores. |
| `suscripcion.hurl` | CRUD de suscripciones como consumidor, `403` para empresas, `404` sobre suscripciones ajenas y `401` sin sesión. |
| `pago.hurl` | Pagos sobre suscripciones propias, `404` sobre pagos ajenos y `401` sin sesión. |

# Estructura del proyecto

```
.
├── main.go               # Conexión a la base, armado de capas, rutas HTTP y frontend
├── main_test.go          # Tests de integración CRUD de cada tabla (sqlc)
├── static/               # Frontend del TP4 (servido en /)
│   ├── index.html        # Inicio público
│   ├── login.html
│   ├── registro.html
│   ├── cuenta.html       # "Hola, {nombre}" (consumidores)
│   ├── suscripciones.html
│   ├── nueva-suscripcion.html
│   ├── empresa.html      # Panel de servicios (empresas)
│   ├── app.js            # Lógica de todas las páginas
│   └── styles.css
├── logic/                # Lógica de negocio y handlers HTTP de cada entidad
│   ├── sesion.go         # Sesiones, login/logout y middleware de permisos
│   ├── usuario.go
│   ├── servicio.go
│   ├── suscripcion.go
│   └── pago.go
├── db/
│   ├── schema/schema.sql # Tablas y claves foráneas
│   ├── queries/queries.sql
│   └── sqlc/             # Código generado por sqlc (no se versiona)
├── hurl/                 # Pruebas de los endpoints
│   ├── frontend.hurl
│   ├── pago.hurl
│   ├── servicio.hurl
│   ├── suscripcion.hurl
│   └── usuario.hurl
├── docker-compose.yml    # PostgreSQL con healthcheck
├── sqlc.yaml
└── makefile
```
