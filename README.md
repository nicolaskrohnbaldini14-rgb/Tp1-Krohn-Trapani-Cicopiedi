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
git checkout tp3
```
## Ejecutar pruebas
```bash
make
```

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

El acceso a datos se genera con [sqlc](https://sqlc.dev/) a partir de las
queries en [`db/queries/queries.sql`](db/queries/queries.sql), que cubren las
operaciones CRUD (Create, Get, List, Update, Delete) para cada una de las 4
entidades. El código generado vive en `db/sqlc/` (`models.go`, `queries.sql.go`,
`db.go`) y **no debe editarse a mano** — se regenera automáticamente cada vez
que se corre `sqlc generate` (parte de `make test-api`).

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

### Usuarios

| Método | Ruta | Respuestas |
|---|---|---|
| `POST` | `/usuarios` | `201` creado · `400` datos inválidos · `409` email ya registrado |
| `GET` | `/usuarios` | `200` lista de usuarios |
| `GET` | `/usuarios/{id}` | `200` · `400` ID no numérico · `404` no encontrado |
| `PUT` | `/usuarios/{id}` | `200` · `400` · `404` · `409` email en uso |
| `DELETE` | `/usuarios/{id}` | `204` · `404` |

```json
{ "nombre": "Nico", "email": "nico@example.com", "password": "secreto" }
```

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
| `POST` | `/suscripciones` | `201` · `400` datos inválidos · `404` usuario o servicio inexistente · `409` ya tiene una activa |
| `GET` | `/suscripciones/{id}` | `200` · `400` · `404` |
| `GET` | `/suscripciones/usuario/{id_usuario}` | `200` suscripciones del usuario |
| `PUT` | `/suscripciones/{id}` | `200` · `400` · `404` · `409` |
| `DELETE` | `/suscripciones/{id}` | `204` · `404` |

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

- **Usuarios:** nombre, email y password son obligatorios. El email tiene que tener un formato válido y no puede repetirse. La password nunca se devuelve en las respuestas.
- **Servicios:** el nombre es único. La categoría debe ser una de `Streaming`, `Música`, `Software`, `Gimnasio`, `Salud`, `Financiero`, `Turismo` u `Otros`, y no se puede cambiar después de creado. No se puede borrar un servicio que tenga suscripciones.
- **Suscripciones:**
  - El usuario y el servicio tienen que existir.
  - El monto tiene que ser mayor a cero y la fecha de inicio anterior a la de vencimiento.
  - El estado es `Activa` (por defecto), `Pausada` o `Cancelada`.
  - Un usuario no puede tener dos suscripciones activas al mismo servicio.
  - La `password_cuenta` nunca se devuelve en las respuestas.
- **Pagos:** la suscripción tiene que existir, el monto tiene que ser mayor a cero y la fecha de pago no puede ser posterior al vencimiento de la suscripción.

# Pruebas de los endpoints

Las pruebas están en `hurl/`, un archivo por entidad (`usuario.hurl`,
`servicio.hurl`, `suscripcion.hurl` y `pago.hurl`). Cada archivo crea los datos
que necesita, prueba los casos exitosos y los de error de cada endpoint.

# Estructura del proyecto

```
.
├── main.go               # Conexión a la base, armado de capas y rutas HTTP
├── main_test.go          # Tests de integración CRUD de cada tabla (sqlc)
├── logic/                # Lógica de negocio y handlers HTTP de cada entidad
│   ├── usuario.go
│   ├── servicio.go
│   ├── suscripcion.go
│   ├── pago.go
├── db/
│   ├── schema/schema.sql # Tablas y claves foráneas
│   ├── queries/queries.sql
│   └── sqlc/             # Código generado por sqlc (no se versiona)
├── hurl/                 # Pruebas de los endpoints
│   └── pago.hurl
│   └── servicio.hurl
│   └── suscripcion.hurl
│   └── usuario.hurl              
├── docker-compose.yml    # PostgreSQL con healthcheck
├── sqlc.yaml
└── makefile
```
