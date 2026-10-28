# ---------- Setup: una empresa publica el servicio ----------
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "Empresa Subs",
  "email": "empresasubs_{{newUuid}}@test.com",
  "password": "secreto123",
  "tipo": "empresa"
}
HTTP 201
[Captures]
empresa_id: jsonpath "$.id_usuario"
empresa_email: jsonpath "$.email"
[Asserts]
jsonpath "$.tipo" == "empresa"

# --- login como empresa ---
POST {{host}}/sesion
Content-Type: application/json
{
  "email": "{{empresa_email}}",
  "password": "secreto123"
}
HTTP 200

POST {{host}}/servicios
Content-Type: application/json
{
  "nombre": "ServicioSubs-{{newUuid}}",
  "categoria": "Streaming"
}
HTTP 201
[Captures]
servicio_id: jsonpath "$.id_servicio"

# una empresa no puede crear suscripciones (403)
POST {{host}}/suscripciones
Content-Type: application/json
{
  "id_servicio": {{servicio_id}},
  "monto": "1500.00",
  "fecha_inicio": "2026-01-01T00:00:00Z",
  "fecha_vencimiento": "2026-12-31T00:00:00Z",
  "usuario_cuenta": "cuenta@test.com",
  "password_cuenta": "clavecuenta",
  "estado": "Activa"
}
HTTP 403


# ---------- Setup: consumidor con sesión iniciada ----------
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "Usuario Suscripciones",
  "email": "subs_{{newUuid}}@test.com",
  "password": "secreto123",
  "tipo": "consumidor"
}
HTTP 201
[Captures]
usuario_id: jsonpath "$.id_usuario"
usuario_email: jsonpath "$.email"
[Asserts]
jsonpath "$.tipo" == "consumidor"

# --- login como consumidor ---
POST {{host}}/sesion
Content-Type: application/json
{
  "email": "{{usuario_email}}",
  "password": "secreto123"
}
HTTP 200


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


# ---------- 3. A nombre de otro usuario (403) ----------
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
HTTP 403
[Asserts]
body contains "a tu nombre"


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


# ---------- 24b. Otro consumidor no ve ni toca suscripciones ajenas ----------
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "Otro Consumidor",
  "email": "otro_{{newUuid}}@test.com",
  "password": "secreto123",
  "tipo": "consumidor"
}
HTTP 201
[Captures]
otro_id: jsonpath "$.id_usuario"
otro_email: jsonpath "$.email"
[Asserts]
jsonpath "$.tipo" == "consumidor"

# --- login como otro consumidor ---
POST {{host}}/sesion
Content-Type: application/json
{
  "email": "{{otro_email}}",
  "password": "secreto123"
}
HTTP 200

GET {{host}}/suscripciones/{{sub_id}}
HTTP 404

PUT {{host}}/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": "1.00",
  "fecha_vencimiento": "2027-06-30T00:00:00Z",
  "estado": "Cancelada",
  "usuario_cuenta": "x@test.com",
  "password_cuenta": "x"
}
HTTP 404

DELETE {{host}}/suscripciones/{{sub_id}}
HTTP 404

GET {{host}}/suscripciones/usuario/{{usuario_id}}
HTTP 403

# sin sesión (401)
DELETE {{host}}/sesion
HTTP 204

GET {{host}}/suscripciones/{{sub_id}}
HTTP 401

GET {{host}}/suscripciones/usuario/{{usuario_id}}
HTTP 401

# --- vuelvo al dueño de las suscripciones ---
POST {{host}}/sesion
Content-Type: application/json
{
  "email": "{{usuario_email}}",
  "password": "secreto123"
}
HTTP 200


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

# --- el otro consumidor borra su cuenta ---
POST {{host}}/sesion
Content-Type: application/json
{
  "email": "{{otro_email}}",
  "password": "secreto123"
}
HTTP 200

DELETE {{host}}/usuarios/{{otro_id}}
HTTP 204

# --- la empresa borra el servicio y su cuenta ---
POST {{host}}/sesion
Content-Type: application/json
{
  "email": "{{empresa_email}}",
  "password": "secreto123"
}
HTTP 200

DELETE {{host}}/servicios/{{servicio_id}}
HTTP 204

DELETE {{host}}/usuarios/{{empresa_id}}
HTTP 204