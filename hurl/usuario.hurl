# ---------- 1. Crear usuario (201) ----------
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "Usuario Test",
  "email": "user_{{newUuid}}@test.com",
  "password": "secreto123"
}
HTTP 201
[Captures]  
usuario_id: jsonpath "$.id_usuario"
usuario_email: jsonpath "$.email"
[Asserts]
header "Content-Type" == "application/json"
jsonpath "$.id_usuario" > 0
jsonpath "$.nombre" == "Usuario Test"
jsonpath "$.password" == ""
jsonpath "$.created_at" exists


# ---------- 2. Crear con email duplicado (409) ----------
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "Otro Nombre",
  "email": "{{usuario_email}}",
  "password": "secreto123"
}
HTTP 409
[Asserts]
body contains "ya está registrado"


# ---------- 3. Crear con campos vacios (400) ----------
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "",
  "email": "",
  "password": ""
}
HTTP 400
[Asserts]
body contains "obligatorios"


# ---------- 4. Crear con email invalido (400) ----------
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "Mail Roto",
  "email": "esto-no-es-un-mail",
  "password": "secreto123"
}
HTTP 400
[Asserts]
body contains "no es válido"


# ---------- 5. Crear con JSON invalido (400) ----------
POST {{host}}/usuarios
Content-Type: application/json
```
{ esto no es json}
```
HTTP 400
[Asserts]
body contains "JSON"


# ---------- 6. Obtener por ID (200) ----------
GET {{host}}/usuarios/{{usuario_id}}
HTTP 200
[Asserts]
jsonpath "$.id_usuario" == {{usuario_id}}
jsonpath "$.email" == "{{usuario_email}}"
jsonpath "$.password" == ""


# ---------- 7. Obtener inexistente (404) ----------
GET {{host}}/usuarios/999999
HTTP 404
[Asserts]
body contains "no encontrado"


# ---------- 8. ID no numerico (400) ----------
GET {{host}}/usuarios/abc
HTTP 400
[Asserts]
body contains "numérico"


# ---------- 9. Listar (200) ----------
GET {{host}}/usuarios
HTTP 200
[Asserts]
jsonpath "$" count >= 1


# ---------- 10. Actualizar (200) ----------
PUT {{host}}/usuarios/{{usuario_id}}
Content-Type: application/json
{
  "nombre": "Usuario Editado",
  "email": "{{usuario_email}}",
  "password": "nuevaclave456"
}
HTTP 200
[Asserts]
jsonpath "$.id_usuario" == {{usuario_id}}
jsonpath "$.nombre" == "Usuario Editado"
jsonpath "$.password" == ""


# ---------- 11. Verificar que el cambio persistio ----------
GET {{host}}/usuarios/{{usuario_id}}
HTTP 200
[Asserts]
jsonpath "$.nombre" == "Usuario Editado"


# ---------- 12. Actualizar inexistente (404) ----------
PUT {{host}}/usuarios/999999
Content-Type: application/json
{
  "nombre": "Fantasma",
  "email": "fantasma_{{newUuid}}@test.com",
  "password": "x12345"
}
HTTP 404


# ---------- 13. Actualizar con email de otro usuario (409) ----------
# Primero creo un segundo usuario
POST {{host}}/usuarios
Content-Type: application/json
{
  "nombre": "Segundo Usuario",
  "email": "segundo_{{newUuid}}@test.com",
  "password": "secreto123"
}
HTTP 201
[Captures]
otro_id: jsonpath "$.id_usuario"
otro_email: jsonpath "$.email"

PUT {{host}}/usuarios/{{usuario_id}}
Content-Type: application/json
{
  "nombre": "Usuario Editado",
  "email": "{{otro_email}}",
  "password": "nuevaclave456"
}
HTTP 409
[Asserts]
body contains "en uso"


# ---------- 14. Actualizar con datos invalidos (400) ----------
PUT {{host}}/usuarios/{{usuario_id}}
Content-Type: application/json
{
  "nombre": "",
  "email": "{{usuario_email}}",
  "password": "nuevaclave456"
}
HTTP 400


# ---------- 15. Metodo no permitido (405) ----------
PATCH {{host}}/usuarios/{{usuario_id}}
HTTP 405


# ---------- 16. Eliminar (204) y verificar (404) ----------
DELETE {{host}}/usuarios/{{usuario_id}}
HTTP 204

GET {{host}}/usuarios/{{usuario_id}}
HTTP 404

DELETE {{host}}/usuarios/{{usuario_id}}
HTTP 404


# ---------- Limpieza ----------
DELETE {{host}}/usuarios/{{otro_id}}
HTTP 204