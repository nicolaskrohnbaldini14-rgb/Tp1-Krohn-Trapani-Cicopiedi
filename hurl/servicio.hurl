
# ---------- 1. Crear servicio (201) ----------
POST {{host}}/servicios
Content-Type: application/json
{
  "nombre": "Servicio-{{newUuid}}",
  "categoria": "Streaming"
}
HTTP 201
[Captures]
servicio_id: jsonpath "$.id_servicio"
servicio_nombre: jsonpath "$.nombre"
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$.id_servicio" > 0
jsonpath "$.categoria" == "Streaming"
jsonpath "$.created_at" exists


# ---------- 2. Nombre duplicado (409) ----------
POST {{host}}/servicios
Content-Type: application/json
{
  "nombre": "{{servicio_nombre}}",
  "categoria": "Música"
}
HTTP 409
[Asserts]
body contains "ya existe"


# ---------- 3. Nombre vacio (400) ----------
POST {{host}}/servicios
Content-Type: application/json
{
  "nombre": "",
  "categoria": "Streaming"
}
HTTP 400
[Asserts]
body contains "obligatorio"


# ---------- 4. Categoria vacia (400) ----------
# OJO: el mensaje es "la categoría es obligatoria" y responderError busca
# "obligatorio", asi que hoy devuelve 500. Ver nota al final del chat.
POST {{host}}/servicios
Content-Type: application/json
{
  "nombre": "SinCategoria-{{newUuid}}",
  "categoria": ""
}
HTTP 400


# ---------- 5. Categoria fuera de las permitidas (400) ----------
POST {{host}}/servicios
Content-Type: application/json
{
  "nombre": "Invalida-{{newUuid}}",
  "categoria": "Inventada"
}
HTTP 400
[Asserts]
body contains "no es válida"


# ---------- 6. JSON invalido (400) ----------
POST {{host}}/servicios
Content-Type: application/json
```
{ esto no es json }
```
HTTP 400


# ---------- 7. Obtener por ID (200) ----------
GET {{host}}/servicios/{{servicio_id}}
HTTP 200
[Asserts]
jsonpath "$.id_servicio" == {{servicio_id}}
jsonpath "$.nombre" == "{{servicio_nombre}}"


# ---------- 8. Obtener inexistente (404) ----------
GET {{host}}/servicios/999999
HTTP 404
[Asserts]
body contains "no encontrado"


# ---------- 9. ID no numerico (400) ----------
GET {{host}}/servicios/abc
HTTP 400


# ---------- 10. Listar (200) ----------
GET {{host}}/servicios
HTTP 200
[Asserts]
jsonpath "$" count >= 1


# ---------- 11. Actualizar nombre (200) ----------
PUT {{host}}/servicios/{{servicio_id}}
Content-Type: application/json
{
  "nombre": "Renombrado-{{newUuid}}",
  "categoria": "Streaming"
}
HTTP 200
[Captures]
servicio_nombre_nuevo: jsonpath "$.nombre"
[Asserts]
jsonpath "$.id_servicio" == {{servicio_id}}
jsonpath "$.categoria" == "Streaming"


# ---------- 12. Actualizar sin mandar categoria (200, se mantiene) ----------
PUT {{host}}/servicios/{{servicio_id}}
Content-Type: application/json
{
  "nombre": "{{servicio_nombre_nuevo}}"
}
HTTP 200
[Asserts]
jsonpath "$.categoria" == "Streaming"


# ---------- 13. Intentar cambiar la categoria (400) ----------
PUT {{host}}/servicios/{{servicio_id}}
Content-Type: application/json
{
  "nombre": "{{servicio_nombre_nuevo}}",
  "categoria": "Software"
}
HTTP 400
[Asserts]
body contains "no se permite modificar"


# ---------- 14. Actualizar con nombre vacio (400) ----------
PUT {{host}}/servicios/{{servicio_id}}
Content-Type: application/json
{
  "nombre": "",
  "categoria": "Streaming"
}
HTTP 400


# ---------- 15. Actualizar inexistente (404) ----------
PUT {{host}}/servicios/999999
Content-Type: application/json
{
  "nombre": "Fantasma-{{newUuid}}",
  "categoria": "Streaming"
}
HTTP 404


# ---------- 16. Actualizar con nombre de otro servicio (409) ----------
POST {{host}}/servicios
Content-Type: application/json
{
  "nombre": "Segundo-{{newUuid}}",
  "categoria": "Software"
}
HTTP 201
[Captures]
servicio2_id: jsonpath "$.id_servicio"
servicio2_nombre: jsonpath "$.nombre"

PUT {{host}}/servicios/{{servicio_id}}
Content-Type: application/json
{
  "nombre": "{{servicio2_nombre}}",
  "categoria": "Streaming"
}
HTTP 409
[Asserts]
body contains "ya existe"


# ---------- 17. Metodo no permitido (405) ----------
PATCH {{host}}/servicios/{{servicio_id}}
HTTP 405


# ---------- 18. No se puede eliminar con suscripciones asociadas (409) ----------
# Armo usuario + suscripcion sobre servicio2
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "Dueño Suscripcion",
  "email": "dueno_{{newUuid}}@test.com",
  "password": "secreto123"
}
HTTP 201
[Captures]
usuario_id: jsonpath "$.id_usuario"

POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio2_id}},
  "monto": "1500.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "cuenta@test.com",
  "password_cuenta": "clavecuenta",
  "estado": "Activa"
}
HTTP 201

DELETE {{host}}/servicios/{{servicio2_id}}
HTTP 409
[Asserts]
body contains "suscripciones asociadas"


# ---------- 19. Eliminar servicio sin suscripciones (204) y verificar (404) ----------
DELETE {{host}}/servicios/{{servicio_id}}
HTTP 204

GET {{host}}/servicios/{{servicio_id}}
HTTP 404

DELETE {{host}}/servicios/{{servicio_id}}
HTTP 404


# ---------- Limpieza ----------
# Al borrar el usuario se borran sus suscripciones (ON DELETE CASCADE),
# y recien ahi se puede borrar servicio2.
DELETE {{host}}/usuarios/{{usuario_id}}
HTTP 204

DELETE {{host}}/servicios/{{servicio2_id}}
HTTP 204