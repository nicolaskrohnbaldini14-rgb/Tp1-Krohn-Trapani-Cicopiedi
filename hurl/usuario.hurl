# ======================================================
# Pruebas Definitivas del Módulo de Usuario (Hurl)
# ======================================================

# ------------------------------------------------------
# 1. CREACIÓN DE USUARIOS BASE (POST HTTP 201)
# ------------------------------------------------------

# 1.1 Crear Usuario 1
POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "Nicolas Krohn Test 1",
  "email": "user1+{{newUuid}}@example.com",
  "password": "Password123"
}

HTTP 201
[Captures]
user1_id: jsonpath "$.id_usuario"
user1_email: jsonpath "$.email"
[Asserts]
jsonpath "$.id_usuario" exists
jsonpath "$.nombre" == "Nicolas Krohn Test 1"
jsonpath "$.created_at" exists
jsonpath "$.password" not exists

# 1.2 Crear Usuario 2 (auxiliar para pruebas de duplicados)
POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "Nicolas Krohn Test 2",
  "email": "user2+{{newUuid}}@example.com",
  "password": "Password123"
}

HTTP 201
[Captures]
user2_id: jsonpath "$.id_usuario"
user2_email: jsonpath "$.email"
[Asserts]
jsonpath "$.id_usuario" exists
jsonpath "$.password" not exists

# ------------------------------------------------------
# 2. VALIDACIONES DE CREACIÓN (HTTP 400 y 409)
# ------------------------------------------------------

# 2.1 Intentar crear usuario con email duplicado (409 Conflict)
POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "Intento Duplicado",
  "email": "{{user1_email}}",
  "password": "Password123"
}

HTTP 409
[Asserts]
body contains "registrado"

# 2.2 Nombre vacío (400 Bad Request)
POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "",
  "email": "nom_vacio+{{newUuid}}@example.com",
  "password": "Password123"
}

HTTP 400
[Asserts]
body contains "obligatorios"

# 2.3 Email vacío (400 Bad Request)
POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "Sin Email",
  "email": "",
  "password": "Password123"
}

HTTP 400
[Asserts]
body contains "obligatorios"

# 2.4 Password vacío (400 Bad Request)
POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "Sin Password",
  "email": "pass_vacio+{{newUuid}}@example.com",
  "password": ""
}

HTTP 400
[Asserts]
body contains "obligatorios"

# 2.5 Formato de email inválido (400 Bad Request)
POST http://localhost:8080/usuarios
Content-Type: application/json
{
  "nombre": "Email Invalido",
  "email": "correo_sin_arroba",
  "password": "Password123"
}

HTTP 400
[Asserts]
body contains "formato"

# 2.6 JSON malformado enviado en texto plano (400 Bad Request)
POST http://localhost:8080/usuarios
Content-Type: application/json
`{ "nombre": "Roto", "email": 

}`

HTTP 400
[Asserts]
body contains "JSON"

# ------------------------------------------------------
# 3. LECTURA Y CONSULTAS (GET HTTP 200 / 400 / 404)
# ------------------------------------------------------

# 3.1 Listar todos los usuarios
GET http://localhost:8080/usuarios

HTTP 200
[Asserts]
jsonpath "$" count >= 2
jsonpath "$[*].id_usuario" includes {{user1_id}}
jsonpath "$[*].id_usuario" includes {{user2_id}}
jsonpath "$[*].password" not exists

# 3.2 Obtener Usuario 1 por ID (GET profundo)
GET http://localhost:8080/usuarios/{{user1_id}}

HTTP 200
[Asserts]
jsonpath "$.id_usuario" == {{user1_id}}
jsonpath "$.email" == "{{user1_email}}"
jsonpath "$.password" not exists

# 3.3 Obtener usuario inexistente (404 Not Found)
GET http://localhost:8080/usuarios/999999999

HTTP 404
[Asserts]
body contains "no encontrado"

# 3.4 Obtener usuario con ID no numérico (400 Bad Request)
GET http://localhost:8080/usuarios/abc

HTTP 400
[Asserts]
body contains "numérico"

# ------------------------------------------------------
# 4. ACTUALIZACIÓN (PUT HTTP 200 / 400 / 404 / 409)
# ------------------------------------------------------

# 4.1 Actualización válida de Usuario 1
PUT http://localhost:8080/usuarios/{{user1_id}}
Content-Type: application/json
{
  "nombre": "Nicolas Krohn Modificado",
  "email": "user1_nuevo+{{newUuid}}@example.com",
  "password": "NuevaPassword123"
}

HTTP 200
[Captures]
user1_email_nuevo: jsonpath "$.email"
[Asserts]
jsonpath "$.id_usuario" == {{user1_id}}
jsonpath "$.nombre" == "Nicolas Krohn Modificado"
jsonpath "$.password" not exists

# 4.2 Verificar que el PUT se persistió en la BD
GET http://localhost:8080/usuarios/{{user1_id}}

HTTP 200
[Asserts]
jsonpath "$.nombre" == "Nicolas Krohn Modificado"
jsonpath "$.email" == "{{user1_email_nuevo}}"

# 4.3 PUT manteniendo su PROPIO email actual (Debe responder 200 OK)
PUT http://localhost:8080/usuarios/{{user1_id}}
Content-Type: application/json
{
  "nombre": "Nicolas Mantiene Email",
  "email": "{{user1_email_nuevo}}",
  "password": "OtraPassword123"
}

HTTP 200
[Asserts]
jsonpath "$.id_usuario" == {{user1_id}}
jsonpath "$.email" == "{{user1_email_nuevo}}"

# 4.4 Intentar actualizar usando el email del Usuario 2 (409 Conflict)
PUT http://localhost:8080/usuarios/{{user1_id}}
Content-Type: application/json
{
  "nombre": "Nicolas Colision Email",
  "email": "{{user2_email}}",
  "password": "NuevaPassword123"
}

HTTP 409
[Asserts]
body contains "uso"

# 4.5 NUEVO: Verificar con GET que el Usuario 1 NO se modificó tras el 409
GET http://localhost:8080/usuarios/{{user1_id}}

HTTP 200
[Asserts]
jsonpath "$.nombre" == "Nicolas Mantiene Email"
jsonpath "$.email" == "{{user1_email_nuevo}}"

# 4.6 Actualizar con campos vacíos (400 Bad Request)
PUT http://localhost:8080/usuarios/{{user1_id}}
Content-Type: application/json
{
  "nombre": "",
  "email": "email_invalido",
  "password": ""
}

HTTP 400
[Asserts]
body contains "obligatorios"

# 4.7 NUEVO: PUT con formato de email inválido (400 Bad Request)
PUT http://localhost:8080/usuarios/{{user1_id}}
Content-Type: application/json
{
  "nombre": "Nicolas Email Malo",
  "email": "formato_email_invalido_sin_arroba",
  "password": "Password123"
}

HTTP 400
[Asserts]
body contains "formato"

# 4.8 NUEVO: PUT con JSON malformado (400 Bad Request)
PUT http://localhost:8080/usuarios/{{user1_id}}
Content-Type: application/json
`{ "nombre": "Malo", "email": 
}`

HTTP 400
[Asserts]
body contains "JSON"

# 4.9 Actualizar usuario inexistente (404 Not Found)
PUT http://localhost:8080/usuarios/999999999
Content-Type: application/json
{
  "nombre": "Fantasma",
  "email": "fantasma+{{newUuid}}@example.com",
  "password": "Password123"
}

HTTP 404
[Asserts]
body contains "no encontrado"

# 5.1 Método PATCH no implementado (405 Method Not Allowed)
PATCH http://localhost:8080/usuarios/{{user1_id}}

HTTP 405
[Asserts]
body contains "permitido"

# 5.2 NUEVO: DELETE con ID no numérico (400 Bad Request)
DELETE http://localhost:8080/usuarios/abc

HTTP 400
[Asserts]
body contains "numérico"

# 5.3 Eliminar usuario inexistente (404 Not Found)
DELETE http://localhost:8080/usuarios/999999999

HTTP 404
[Asserts]
body contains "no encontrado"

# 5.4 LIMPIEZA: Eliminar Usuario 1
DELETE http://localhost:8080/usuarios/{{user1_id}}

HTTP 204

# 5.5 LIMPIEZA: Eliminar Usuario 2
DELETE http://localhost:8080/usuarios/{{user2_id}}

HTTP 204

# 5.6 Confirmar que Usuario 1 fue efectivamente borrado de la BD (404 Not Found)
GET http://localhost:8080/usuarios/{{user1_id}}

HTTP 404
[Asserts]
body contains "no encontrado"