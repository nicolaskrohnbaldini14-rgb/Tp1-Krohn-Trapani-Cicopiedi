# ---------- Setup: usuario y servicio ----------
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "Usuario Suscripciones",
  "email": "subs_{{newUuid}}@test.com",
  "password": "secreto123"
}
HTTP 201
[Captures]
usuario_id: jsonpath "$.id_usuario"

POST {{host}}/servicios
Content-Type: application/json
{
  "nombre": "ServicioSubs-{{newUuid}}", //{{newUuid}} preguntar es un id unico y universal para que no de error
  "categoria": "Streaming"
}
HTTP 201
[Captures]
servicio_id: jsonpath "$.id_servicio"


# ---------- 1. Crear suscripcion (201) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio_id}},
  "monto": "1500.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "cuenta@test.com",
  "password_cuenta": "clavecuenta",
  "estado": "Activa"
}
HTTP 201
[Captures]
sub_id: jsonpath "$.id_suscripcion"
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$.id_suscripcion" > 0
jsonpath "$.id_usuario" == {{usuario_id}}
jsonpath "$.id_servicio" == {{servicio_id}}
jsonpath "$.monto" == "1500.00"
jsonpath "$.estado" == "Activa"
jsonpath "$.usuario_cuenta" == "cuenta@test.com"
jsonpath "$.password_cuenta" == ""
jsonpath "$.fecha_inicio" startsWith "2026-01-01"
jsonpath "$.fecha_vencimiento" startsWith "2026-12-31"


# ---------- 2. Duplicada activa para mismo usuario y servicio (409) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio_id}},
  "monto": "900.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "otra@test.com",
  "password_cuenta": "clave",
  "estado": "Activa"
}
HTTP 409
[Asserts]
body contains "suscripción activa"


# ---------- 3. Usuario inexistente (404) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": 999999,
  "id_servicio": {{servicio_id}},
  "monto": "100.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "x@test.com",
  "password_cuenta": "clave",
  "estado": "Activa"
}
HTTP 404
[Asserts]
body contains "usuario"


# ---------- 4. Servicio inexistente (404) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": 999999,
  "monto": "100.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "x@test.com",
  "password_cuenta": "clave",
  "estado": "Activa"
}
HTTP 404
[Asserts]
body contains "servicio"


# ---------- 5. IDs invalidos (400) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": 0,
  "id_servicio": 0,
  "monto": "100.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "x@test.com",
  "password_cuenta": "clave",
  "estado": "Activa"
}
HTTP 400


# ---------- 6. Monto cero (400) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio_id}},
  "monto": "0",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "x@test.com",
  "password_cuenta": "clave",
  "estado": "Activa"
}
HTTP 400
[Asserts]
body contains "mayor a cero"


# ---------- 7. Monto no numerico (400) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio_id}},
  "monto": "abc",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "x@test.com",
  "password_cuenta": "clave",
  "estado": "Activa"
}
HTTP 400
[Asserts]
body contains "número válido"


# ---------- 8. Inicio posterior al vencimiento (400) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio_id}},
  "monto": "100.00",
  "fecha_inicio": "2027-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-01-01T00:00:00Z",
  "usuario_cuenta": "x@test.com",
  "password_cuenta": "clave",
  "estado": "Activa"
}
HTTP 400
[Asserts]
body contains "anterior"


# ---------- 9. Sin fechas (400) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio_id}},
  "monto": "100.00",
  "usuario_cuenta": "x@test.com",
  "password_cuenta": "clave",
  "estado": "Activa"
}
HTTP 400
[Asserts]
body contains "obligatorias"


# ---------- 10. Sin credenciales de cuenta (400) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio_id}},
  "monto": "100.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "",
  "password_cuenta": "",
  "estado": "Activa"
}
HTTP 400
[Asserts]
body contains "obligatorias"


# ---------- 11. Estado invalido (400) ----------
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio_id}},
  "monto": "100.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "x@test.com",
  "password_cuenta": "clave",
  "estado": "Inventada"
}
HTTP 400
[Asserts]
body contains "no es válido"


# ---------- 12. JSON invalido (400) ----------
POST {{host}}/suscripciones
Content-Type: application/json
```
{ esto no es json}
```
HTTP 400


# ---------- 13. Obtener por ID (200) ----------
GET {{host}}/suscripciones/{{sub_id}}
HTTP 200
[Asserts]
jsonpath "$.id_suscripcion" == {{sub_id}}
jsonpath "$.estado" == "Activa"
jsonpath "$.password_cuenta" == ""


# ---------- 14. Obtener inexistente (404) ----------
GET {{host}}/suscripciones/999999
HTTP 404
[Asserts]
body contains "no encontrada"


# ---------- 15. ID no numerico (400) ----------
GET {{host}}/suscripciones/abc
HTTP 400


# ---------- 16. Actualizar monto y estado (200) ----------
PUT {{host}}/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": "1999.50",
  "fecha_vencimiento": "2027-06-30T00:00:00Z",
  "estado": "Pausada",
  "usuario_cuenta": "nuevacuenta@test.com",
  "password_cuenta": "nuevaclave"
}
HTTP 200
[Asserts]
jsonpath "$.id_suscripcion" == {{sub_id}}
jsonpath "$.monto" == "1999.50"
jsonpath "$.estado" == "Pausada"
jsonpath "$.usuario_cuenta" == "nuevacuenta@test.com"
jsonpath "$.fecha_vencimiento" startsWith "2027-06-30"
jsonpath "$.fecha_inicio" startsWith "2026-01-01"
jsonpath "$.password_cuenta" == ""


# ---------- 17. Verificar que persistio ----------
GET {{host}}/suscripciones/{{sub_id}}
HTTP 200
[Asserts]
jsonpath "$.estado" == "Pausada"
jsonpath "$.monto" == "1999.50"


# ---------- 18. Actualizar con estado invalido (400) ----------
PUT {{host}}/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": "100.00",
  "fecha_vencimiento": "2027-06-30T00:00:00Z",
  "estado": "Inventada",
  "usuario_cuenta": "a@test.com",
  "password_cuenta": "clave"
}
HTTP 400
[Asserts]
body contains "no es válido"


# ---------- 19. Actualizar con monto cero (400) ----------
PUT {{host}}/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": "0",
  "fecha_vencimiento": "2027-06-30T00:00:00Z",
  "estado": "Pausada",
  "usuario_cuenta": "a@test.com",
  "password_cuenta": "clave"
}
HTTP 400


# ---------- 20. Vencimiento anterior al inicio (400) ----------
PUT {{host}}/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": "100.00",
  "fecha_vencimiento": "2025-01-01T00:00:00Z",
  "estado": "Pausada",
  "usuario_cuenta": "a@test.com",
  "password_cuenta": "clave"
}
HTTP 400
[Asserts]
body contains "anterior"


# ---------- 21. Actualizar inexistente (404) ----------
PUT {{host}}/suscripciones/999999
Content-Type: application/json
{
  "monto": "100.00",
  "fecha_vencimiento": "2027-06-30T00:00:00Z",
  "estado": "Pausada",
  "usuario_cuenta": "a@test.com",
  "password_cuenta": "clave"
}
HTTP 404


# ---------- 22. Reactivar con otra activa existente (409) ----------
# La sub1 esta Pausada, asi que puedo crear una segunda Activa
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_usuario": {{usuario_id}},
  "id_servicio": {{servicio_id}},
  "monto": "800.00",
  "fecha_inicio": "2026-02-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "segunda@test.com",
  "password_cuenta": "clave2",
  "estado": "Activa"
}
HTTP 201
[Captures]
sub2_id: jsonpath "$.id_suscripcion"

PUT {{host}}/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": "1999.50",
  "fecha_vencimiento": "2027-06-30T00:00:00Z",
  "estado": "Activa",
  "usuario_cuenta": "nuevacuenta@test.com",
  "password_cuenta": "nuevaclave"
}
HTTP 409
[Asserts]
body contains "suscripción activa"


# ---------- 23. Cancelar la segunda y ahora si reactivar la primera (200) ----------
PUT {{host}}/suscripciones/{{sub2_id}}
Content-Type: application/json
{
  "monto": "800.00",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "estado": "Cancelada",
  "usuario_cuenta": "segunda@test.com",
  "password_cuenta": "clave2"
}
HTTP 200
[Asserts]
jsonpath "$.estado" == "Cancelada"

PUT {{host}}/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": "1999.50",
  "fecha_vencimiento": "2027-06-30T00:00:00Z",
  "estado": "Activa",
  "usuario_cuenta": "nuevacuenta@test.com",
  "password_cuenta": "nuevaclave"
}
HTTP 200
[Asserts]
jsonpath "$.estado" == "Activa"


# ---------- 24. Listar por usuario (200) ----------
GET {{host}}/suscripciones/usuario/{{usuario_id}}
HTTP 200
[Asserts]
jsonpath "$" count == 2
jsonpath "$[0].servicio_nombre" exists
jsonpath "$[0].servicio_categoria" == "Streaming"
jsonpath "$[0].password_cuenta" not exists


# ---------- 25. Metodos no permitidos (405) ----------
GET {{host}}/suscripciones
HTTP 405

PATCH {{host}}/suscripciones/{{sub_id}}
HTTP 405

POST {{host}}/suscripciones/usuario/{{usuario_id}}
HTTP 405


# ---------- 26. Eliminar (204) y verificar (404) ----------
DELETE {{host}}/suscripciones/{{sub2_id}}
HTTP 204

GET {{host}}/suscripciones/{{sub2_id}}
HTTP 404

DELETE {{host}}/suscripciones/{{sub2_id}}
HTTP 404


# ---------- Limpieza ----------
# Borrar el usuario elimina la suscripcion restante (CASCADE)
DELETE {{host}}/usuarios/{{usuario_id}}
HTTP 204

DELETE {{host}}/servicios/{{servicio_id}}
HTTP 204