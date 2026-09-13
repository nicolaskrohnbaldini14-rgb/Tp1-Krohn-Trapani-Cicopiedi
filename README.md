
# Requisitos previos

- [Go](https://go.dev/) 1.22+
- [Docker](https://www.docker.com/) con Docker Compose
- [sqlc](https://docs.sqlc.dev/) — se instala con:
  sudo snap install sqlc

  o, si ya tenés Go instalado:
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
 

# Ejecucion
## Clonar el repositorio 
git clone https://github.com/nicolaskrohnbaldini14-rgb/Tp1-Krohn-Trapani-Cicopiedi.git
## Posicionarse en la carpeta del proyecto 
cd Tp1-Krohn-Trapani-Cicopiedi
## Cambiar a la rama requerida 
git checkout tp2 
## Ejecutar las pruebas
make run


Este comando encadena automáticamente todas las tareas necesarias:

1. `sqlc generate` — regenera el código Go de acceso a datos a partir de
   `db/schema/schema.sql` y `db/queries/queries.sql`.
2. `go build` — compila el proyecto y deja el binario en `tmp/`.
3. `docker compose down -v` — borra cualquier contenedor y volumen de una
   corrida anterior, para arrancar limpio.
4. `docker compose up -d` — levanta un contenedor de Postgres nuevo, que
   aplica automáticamente `schema.sql` al iniciar.
5. Espera a que Postgres esté listo para aceptar conexiones.
6. `go test -v ./...` — corre los tests de integración (CRUD completo de las
   4 tablas) contra esa base real.
7. `docker compose down -v` — vuelve a borrar el contenedor y sus volúmenes,
   así no queda nada corriendo de fondo.




## Documentación de persistencia

El modelo de datos está compuesto por 4 tablas relacionales, definidas en
[`db/schema/schema.sql`](db/schema/schema.sql):

| Tabla | Descripción |
|---|---|
| `usuario` | Cuenta de la persona que usa la app (nombre, apellido, DNI, email, contraseña). |
| `servicio` | Catálogo de servicios disponibles (nombre + categoría), normalizado para poder agrupar el gasto por categoría. |
| `suscripcion` | Relaciona un `usuario` con un `servicio`: monto, fecha de inicio, fecha de vencimiento, credenciales de acceso al servicio (usuario/contraseña de esa cuenta) y estado (activa/pausada). |
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
que se corre `sqlc generate` (parte de `make run`).

Los tests de integración (`main_test.go`) validan el CRUD completo contra una
base Postgres real levantada en Docker, incluyendo las relaciones entre
tablas (crean un usuario y un servicio, encadenan sus IDs para crear una
suscripción, y de ahí un pago).
