# ======================================================
# Pruebas Definitivas del Módulo de Suscripción (Hurl)
# ======================================================

# ------------------------------------------------------
# 1. SETUP PREVIO: Crear Usuario y Servicio Auxiliares
# ------------------------------------------------------

# 1.1 Crear Usuario para la prueba
POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "Usuario Prueba Suscripcion",
  "email": "user_sub+{{newUuid}}@example.com",
  "password": "Password123"
}

HTTP 201
[Captures]
user_id: jsonpath "$.id_usuario"
[Asserts]
jsonpath "$.id_usuario" exists

# 1.2 Crear Servicio para la prueba
POST http://localhost:8080/servicios
Content-Type: application/json
{
  "nombre": "Servicio Sub Test {{newUuid}}",
  "categoria": "Streaming"
}

HTTP 201
[Captures]
service_id: jsonpath "$.id_servicio"
[Asserts]
jsonpath "$.id_servicio" exists

# ------------------------------------------------------
# 2. VALIDACIONES DE CREACIÓN (HTTP 400 / 404)
# ------------------------------------------------------

# 2.1 Error: Fecha de inicio posterior a fecha de vencimiento (400 Bad Request)
POST http://localhost:8080/suscripciones
Content-Type: application/json
{
  "id_usuario": {{user_id}},
  "id_servicio": {{service_id}},
  "monto": 1250.50,
  "fecha_inicio": "2026-11-12T00:00:00Z",
  "fecha_vencimiento": "2026-10-12T00:00:00Z",
  "usuario_cuenta": "cuenta_test@stream.com",
  "password_cuenta": "Clave123",
  "estado": "Activa"
}

HTTP 400
[Asserts]
body contains "anterior"

# 2.2 Error: Fechas de inicio y vencimiento iguales (400 Bad Request)
POST http://localhost:8080/suscripciones
Content-Type: application/json
{
  "id_usuario": {{user_id}},
  "id_servicio": {{service_id}},
  "monto": 1250.50,
  "fecha_inicio": "2026-10-12T00:00:00Z",
  "fecha_vencimiento": "2026-10-12T00:00:00Z",
  "usuario_cuenta": "cuenta_test@stream.com",
  "password_cuenta": "Clave123",
  "estado": "Activa"
}

HTTP 400
[Asserts]
body contains "anterior"

# 2.3 Error: Monto menor o igual a cero (400 Bad Request)
POST http://localhost:8080/suscripciones
Content-Type: application/json
{
  "id_usuario": {{user_id}},
  "id_servicio": {{service_id}},
  "monto": 0.00,
  "fecha_inicio": "2026-10-12T00:00:00Z",
  "fecha_vencimiento": "2026-11-12T00:00:00Z",
  "usuario_cuenta": "cuenta_test@stream.com",
  "password_cuenta": "Clave123",
  "estado": "Activa"
}

HTTP 400
[Asserts]
body contains "monto"

# 2.4 Error: Estado no permitido / inventado (400 Bad Request)
POST http://localhost:8080/suscripciones
Content-Type: application/json
{
  "id_usuario": {{user_id}},
  "id_servicio": {{service_id}},
  "monto": 1250.50,
  "fecha_inicio": "2026-10-12T00:00:00Z",
  "fecha_vencimiento": "2026-11-12T00:00:00Z",
  "usuario_cuenta": "cuenta_test@stream.com",
  "password_cuenta": "Clave123",
  "estado": "Inventada"
}

HTTP 400
[Asserts]
body contains "estado"

# 2.5 Error: Credenciales de cuenta vacías (400 Bad Request)
POST http://localhost:8080/suscripciones
Content-Type: application/json
{
  "id_usuario": {{user_id}},
  "id_servicio": {{service_id}},
  "monto": 1250.50,
  "fecha_inicio": "2026-10-12T00:00:00Z",
  "fecha_vencimiento": "2026-11-12T00:00:00Z",
  "usuario_cuenta": "",
  "password_cuenta": "",
  "estado": "Activa"
}

HTTP 400
[Asserts]
body contains "credenciales"

# 2.6 Error: Usuario inexistente (404 Not Found)
POST http://localhost:8080/suscripciones
Content-Type: application/json
{
  "id_usuario": 999999999,
  "id_servicio": {{service_id}},
  "monto": 1250.50,
  "fecha_inicio": "2026-10-12T00:00:00Z",
  "fecha_vencimiento": "2026-11-12T00:00:00Z",
  "usuario_cuenta": "cuenta_test@stream.com",
  "password_cuenta": "Clave123",
  "estado": "Activa"
}

HTTP 404
[Asserts]
body contains "el usuario especificado no existe"

# 2.7 Error: Servicio inexistente (404 Not Found)
POST http://localhost:8080/suscripciones
Content-Type: application/json
{
  "id_usuario": {{user_id}},
  "id_servicio": 999999999,
  "monto": 1250.50,
  "fecha_inicio": "2026-10-12T00:00:00Z",
  "fecha_vencimiento": "2026-11-12T00:00:00Z",
  "usuario_cuenta": "cuenta_test@stream.com",
  "password_cuenta": "Clave123",
  "estado": "Activa"
}

HTTP 404
[Asserts]
body contains "el servicio especificado no existe"

# 2.8 Error: JSON malformado enviado en texto plano (400 Bad Request)
POST http://localhost:8080/suscripciones
Content-Type: application/json
`{ "id_usuario": {{user_id}}, "monto": `

HTTP 400
[Asserts]
body contains "JSON"

# ------------------------------------------------------
# 3. CREACIÓN EXITOSA Y VERIFICACIÓN (HTTP 201)
# ------------------------------------------------------

POST http://localhost:8080/suscripciones
Content-Type: application/json
{
  "id_usuario": {{user_id}},
  "id_servicio": {{service_id}},
  "monto": 1250.50,
  "fecha_inicio": "2026-10-12T00:00:00Z",
  "fecha_vencimiento": "2026-11-12T00:00:00Z",
  "usuario_cuenta": "nicolas+{{newUuid}}@stream.com",
  "password_cuenta": "ClaveSecreta123",
  "estado": "Activa"
}

HTTP 201
[Captures]
sub_id: jsonpath "$.id_suscripcion"
sub_cuenta: jsonpath "$.usuario_cuenta"
[Asserts]
jsonpath "$.id_suscripcion" exists
jsonpath "$.id_usuario" == {{user_id}}
jsonpath "$.id_servicio" == {{service_id}}
jsonpath "$.monto" == "1250.50"
jsonpath "$.estado" == "Activa"
jsonpath "$.password_cuenta" not exists

# ------------------------------------------------------
# 4. REGLA DE NEGOCIO: DUPLICADO ACTIVO (HTTP 409)
# ------------------------------------------------------

POST http://localhost:8080/suscripciones
Content-Type: application/json
{
  "id_usuario": {{user_id}},
  "id_servicio": {{service_id}},
  "monto": 1250.50,
  "fecha_inicio": "2026-10-12T00:00:00Z",
  "fecha_vencimiento": "2026-11-12T00:00:00Z",
  "usuario_cuenta": "otra_cuenta+{{newUuid}}@stream.com",
  "password_cuenta": "Clave123",
  "estado": "Activa"
}

HTTP 409
[Asserts]
body contains "ya posee una suscripción activa"

# ------------------------------------------------------
# 5. CONSULTAS Y LECTURA (GET HTTP 200 / 400 / 404)
# ------------------------------------------------------

# 5.1 Obtener suscripción por ID (GET profundo)
GET http://localhost:8080/suscripciones/{{sub_id}}

HTTP 200
[Asserts]
jsonpath "$.id_suscripcion" == {{sub_id}}
jsonpath "$.id_usuario" == {{user_id}}
jsonpath "$.id_servicio" == {{service_id}}
jsonpath "$.monto" == "1250.50"
jsonpath "$.estado" == "Activa"
jsonpath "$.usuario_cuenta" == "{{sub_cuenta}}"
jsonpath "$.password_cuenta" not exists

# 5.2 Listar suscripciones por Usuario (Acceso al elemento [0] del Array)
GET http://localhost:8080/suscripciones/usuario/{{user_id}}

HTTP 200
[Asserts]
jsonpath "$" count == 1
jsonpath "$[0].id_suscripcion" == {{sub_id}}

# 5.3 Obtener suscripción inexistente (404 Not Found)
GET http://localhost:8080/suscripciones/999999999

HTTP 404
[Asserts]
body contains "la suscripción especificada no existe"

# 5.4 Obtener suscripción con ID no numérico (400 Bad Request)
GET http://localhost:8080/suscripciones/abc

HTTP 400
[Asserts]
body contains "numérico"

# ------------------------------------------------------
# 6. ACTUALIZACIÓN (PUT HTTP 200 / 400 / 404)
# ------------------------------------------------------

# 6.1 Actualización válida de la suscripción (Pausarla)
PUT http://localhost:8080/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": 1800.00,
  "fecha_vencimiento": "2026-12-12T00:00:00Z",
  "usuario_cuenta": "cuenta_modificada+{{newUuid}}@stream.com",
  "password_cuenta": "NuevaClave123",
  "estado": "Pausada"
}

HTTP 200
[Captures]
sub_cuenta_nueva: jsonpath "$.usuario_cuenta"
[Asserts]
jsonpath "$.id_suscripcion" == {{sub_id}}
jsonpath "$.monto" == "1800.00"
jsonpath "$.estado" == "Pausada"
jsonpath "$.password_cuenta" not exists

# 6.2 Verificar con GET que el PUT se persistió correctamente en la BD
GET http://localhost:8080/suscripciones/{{sub_id}}

HTTP 200
[Asserts]
jsonpath "$.monto" == "1800.00"
jsonpath "$.estado" == "Pausada"
jsonpath "$.usuario_cuenta" == "{{sub_cuenta_nueva}}"

# 6.3 PUT con estado no permitido / inventado (400 Bad Request)
PUT http://localhost:8080/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": 1800.00,
  "fecha_vencimiento": "2026-12-12T00:00:00Z",
  "usuario_cuenta": "{{sub_cuenta_nueva}}",
  "password_cuenta": "NuevaClave123",
  "estado": "Inventada"
}

HTTP 400
[Asserts]
body contains "estado"

# 6.4 PUT con fecha de vencimiento inválida (400 Bad Request alineado con 'anterior')
PUT http://localhost:8080/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": 1800.00,
  "fecha_vencimiento": "2026-09-01T00:00:00Z",
  "usuario_cuenta": "{{sub_cuenta_nueva}}",
  "password_cuenta": "NuevaClave123",
  "estado": "Pausada"
}

HTTP 400
[Asserts]
body contains "anterior"

# 6.5 PUT con monto inválido <= 0 (400 Bad Request)
PUT http://localhost:8080/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": -500.00,
  "fecha_vencimiento": "2026-12-12T00:00:00Z",
  "usuario_cuenta": "{{sub_cuenta_nueva}}",
  "password_cuenta": "NuevaClave123",
  "estado": "Pausada"
}

HTTP 400
[Asserts]
body contains "monto"

# 6.6 PUT con credenciales vacías (400 Bad Request)
PUT http://localhost:8080/suscripciones/{{sub_id}}
Content-Type: application/json
{
  "monto": 1800.00,
  "fecha_vencimiento": "2026-12-12T00:00:00Z",
  "usuario_cuenta": "",
  "password_cuenta": "",
  "estado": "Pausada"
}

HTTP 400
[Asserts]
body contains "credenciales"

# 6.7 PUT con JSON malformado (400 Bad Request)
PUT http://localhost:8080/suscripciones/{{sub_id}}
Content-Type: application/json
`{ "monto": 1800.00, "estado": `

HTTP 400
[Asserts]
body contains "JSON"

# 6.8 PUT sobre suscripción inexistente (404 Not Found)
PUT http://localhost:8080/suscripciones/999999999
Content-Type: application/json
{
  "monto": 1800.00,
  "fecha_vencimiento": "2026-12-12T00:00:00Z",
  "usuario_cuenta": "fantasma@stream.com",
  "password_cuenta": "Clave123",
  "estado": "Activa"
}

HTTP 404
[Asserts]
body contains "la suscripción especificada no existe"

# ------------------------------------------------------
# 7. MÉTODOS NO PERMITIDOS Y ELIMINACIÓN (HTTP 405 / 400 / 404 / 204)
# ------------------------------------------------------

# 7.1 Método PATCH no implementado (405 Method Not Allowed)
PATCH http://localhost:8080/suscripciones/{{sub_id}}

HTTP 405
[Asserts]
body contains "permitido"

# 7.2 DELETE con ID no numérico (400 Bad Request)
DELETE http://localhost:8080/suscripciones/abc

HTTP 400
[Asserts]
body contains "numérico"

# 7.3 Eliminar suscripción inexistente (404 Not Found)
DELETE http://localhost:8080/suscripciones/999999999

HTTP 404
[Asserts]
body contains "la suscripción especificada no existe"

# 7.4 LIMPIEZA ORDENADA
# A. Eliminar Suscripción
DELETE http://localhost:8080/suscripciones/{{sub_id}}

HTTP 204

# B. Confirmar que la suscripción ya no existe en la BD
GET http://localhost:8080/suscripciones/{{sub_id}}

HTTP 404
[Asserts]
body contains "la suscripción especificada no existe"

# C. Eliminar Servicio auxiliar
DELETE http://localhost:8080/servicios/{{service_id}}

HTTP 204

# D. Eliminar Usuario auxiliar
DELETE http://localhost:8080/usuarios/{{user_id}}

HTTP 204