# Registro, inicio y cierre de sesión.
# Hurl guarda las cookies entre pedidos del mismo archivo, como un navegador.

# ---------- 1. Registrarse: crea la cuenta, inicia sesión y va a /cuenta (303) ----------
POST {{host}}/registro
[FormParams]
nombre: Ana Hurl
email: ana.hurl@example.com
password: secreto123
HTTP 303
[Asserts]
header "Location" == "/cuenta?ok=bienvenida"
cookie "sesion" exists
cookie "sesion[HttpOnly]" exists


# ---------- 2. La cuenta saluda, confirma el registro y resume el mes ----------
GET {{host}}/cuenta?ok=bienvenida
HTTP 200
[Asserts]
xpath "normalize-space(//main//h1)" == "Hola, Ana Hurl"
xpath "normalize-space(//div[contains(@class,'mensaje')])" == "¡Listo! Tu cuenta se creó y ya iniciaste sesión."
xpath "normalize-space(//section[@class='bienvenida']/p)" contains "vencen 0 suscripciones · $ 0,00"
xpath "count(//main//a[@href='/cuenta/suscripciones'])" == 1
xpath "count(//main//a[@href='/cuenta/suscripciones/nueva'])" == 1
# con sesión, cada servicio del panel lleva a suscribirse
xpath "count(//a[contains(@href,'/cuenta/suscripciones/nueva?servicio=')])" >= 40
xpath "count(//form[@action='/logout'])" == 1


# ---------- 3. Con sesión iniciada, /login y /registro llevan a la cuenta ----------
GET {{host}}/login
HTTP 303
[Asserts]
header "Location" == "/cuenta"


# ---------- 4. Cerrar sesión ----------
POST {{host}}/logout
HTTP 303
[Asserts]
header "Location" == "/"

GET {{host}}/cuenta
HTTP 303
[Asserts]
header "Location" == "/login"


# ---------- 5. Login con contraseña incorrecta (401) ----------
POST {{host}}/login
[FormParams]
email: ana.hurl@example.com
password: equivocada
HTTP 401
[Asserts]
xpath "normalize-space(//div[contains(@class,'mensaje')])" == "email o contraseña incorrectos"
# el email ingresado se conserva en el formulario
xpath "string(//input[@name='email']/@value)" == "ana.hurl@example.com"


# ---------- 6. Login correcto (303 a /cuenta) ----------
POST {{host}}/login
[Options]
location: true
[FormParams]
email: ana.hurl@example.com
password: secreto123
HTTP 200
[Asserts]
url == "{{host}}/cuenta"
xpath "normalize-space(//main//h1)" == "Hola, Ana Hurl"

POST {{host}}/logout
HTTP 303


# ---------- 7. Email repetido (409) ----------
POST {{host}}/registro
[FormParams]
nombre: Otra Persona
email: ana.hurl@example.com
password: secreto123
HTTP 409
[Asserts]
xpath "normalize-space(//div[contains(@class,'mensaje')])" contains "ya está registrado"
xpath "string(//input[@name='nombre']/@value)" == "Otra Persona"


# ---------- 8. Email inválido y campos vacíos (400) ----------
POST {{host}}/registro
[FormParams]
nombre: Sin Arroba
email: correo-invalido
password: secreto123
HTTP 400
[Asserts]
xpath "normalize-space(//div[contains(@class,'mensaje')])" contains "no es válido"

POST {{host}}/registro
[FormParams]
nombre:
email: vacio@example.com
password: secreto123
HTTP 400
[Asserts]
xpath "normalize-space(//div[contains(@class,'mensaje')])" contains "obligatorios"
