# Pruebas de la página y del ABM de usuarios con formularios HTML.
# Las altas y bajas exitosas responden 303 See Other hacia "/" (patrón PRG).

# ---------- 1. Página principal (200) ----------
GET {{host}}/
HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
xpath "normalize-space(//h1)" == "Suscripciones+"
xpath "string(//link[contains(@href, 'pico')]/@href)" contains "pico.min.css"
xpath "count(//form[@action='/usuarios'][@method='post'])" == 1
xpath "count(//section[@id='usuarios']//input[@name='nombre'])" == 1
xpath "count(//section[@id='usuarios']//input[@name='email'])" == 1
xpath "count(//section[@id='usuarios']//input[@name='password'])" == 1


# ---------- 2. Crear usuario: redirige a / (303) ----------
POST {{host}}/usuarios
[FormParams]
nombre: Ana Hurl
email: ana.hurl@example.com
password: secreto123
HTTP 303
[Asserts]
header "Location" == "/"


# ---------- 3. El usuario aparece en la lista (y no se muestra la contraseña) ----------
GET {{host}}/
HTTP 200
[Captures]
eliminar_ana: xpath "string(//section[@id='usuarios']//tr[td[normalize-space(.)='Ana Hurl']]//form/@action)"
[Asserts]
xpath "count(//section[@id='usuarios']//td[normalize-space(.)='ana.hurl@example.com'])" == 1
body not contains "secreto123"


# ---------- 4. Siguiendo la redirección se llega a la página actualizada ----------
POST {{host}}/usuarios
[Options]
location: true
[FormParams]
nombre: Beto Hurl
email: beto.hurl@example.com
password: secreto123
HTTP 200
[Asserts]
url == "{{host}}/"
xpath "count(//section[@id='usuarios']//td[normalize-space(.)='Beto Hurl'])" == 1


# ---------- 5. Email repetido (409) ----------
POST {{host}}/usuarios
[FormParams]
nombre: Otra Persona
email: ana.hurl@example.com
password: secreto123
HTTP 409
[Asserts]
header "Content-Type" contains "text/html"
xpath "normalize-space(//article[@role='alert'])" contains "ya está registrado"


# ---------- 6. Email con formato inválido (400) ----------
POST {{host}}/usuarios
[FormParams]
nombre: Sin Arroba
email: correo-invalido
password: secreto123
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "no es válido"


# ---------- 7. Campos obligatorios vacíos (400) ----------
POST {{host}}/usuarios
[FormParams]
nombre:
email: vacio@example.com
password: secreto123
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "obligatorios"


# ---------- 8. Eliminar usuario (303) ----------
POST {{host}}{{eliminar_ana}}
HTTP 303
[Asserts]
header "Location" == "/"


# ---------- 9. El usuario eliminado ya no aparece ----------
GET {{host}}/
HTTP 200
[Asserts]
xpath "count(//section[@id='usuarios']//td[normalize-space(.)='Ana Hurl'])" == 0


# ---------- 10. Eliminar usuario inexistente (404) ----------
POST {{host}}/usuarios/999999/eliminar
HTTP 404
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "no encontrado"


# ---------- 11. ID no numérico (400) ----------
POST {{host}}/usuarios/abc/eliminar
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "numérico"


# ---------- 12. Método no permitido: ya no hay API JSON (405) ----------
GET {{host}}/usuarios
HTTP 405
