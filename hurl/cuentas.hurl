# Registro, inicio y cierre de sesión de consumidores.
# Hurl guarda las cookies entre pedidos del mismo archivo, como un navegador.

# ---------- 1. Registrarse: crea la cuenta, inicia sesión y va a /cuenta (303) ----------
POST {{host}}/registro
[FormParams]
nombre: Ana Hurl
email: ana.hurl@example.com
password: secreto123
HTTP 303
[Asserts]
header "Location" == "/cuenta"
cookie "sesion" exists
cookie "sesion[HttpOnly]" exists


# ---------- 2. La cuenta saluda al usuario y ofrece las dos opciones ----------
GET {{host}}/cuenta
HTTP 200
[Asserts]
xpath "normalize-space(//main//h1)" == "Hola, Ana Hurl"
xpath "count(//main//a[@href='/cuenta/suscripciones'])" == 1
xpath "count(//main//a[@href='/cuenta/suscripciones/nueva'])" == 1
xpath "count(//form[@action='/logout'])" == 1


# ---------- 3. Con sesión iniciada, /login y /registro llevan a la cuenta ----------
GET {{host}}/login
HTTP 303
[Asserts]
header "Location" == "/cuenta"


# ---------- 4. Un consumidor no puede entrar al panel de empresas (403) ----------
GET {{host}}/empresa
HTTP 403
[Asserts]
xpath "normalize-space(//article/p)" contains "solo para empresas"

POST {{host}}/empresa/servicios
[FormParams]
nombre: Servicio Pirata
categoria: Otros
HTTP 403


# ---------- 5. Cerrar sesión ----------
POST {{host}}/logout
HTTP 303
[Asserts]
header "Location" == "/"

GET {{host}}/cuenta
HTTP 303
[Asserts]
header "Location" == "/login"


# ---------- 6. Login con contraseña incorrecta (401) ----------
POST {{host}}/login
[FormParams]
email: ana.hurl@example.com
password: equivocada
HTTP 401
[Asserts]
xpath "normalize-space(//article[@role='alert'])" == "email o contraseña incorrectos"
# el email ingresado se conserva en el formulario
xpath "string(//input[@name='email']/@value)" == "ana.hurl@example.com"


# ---------- 7. Login con email inexistente (401, mismo mensaje) ----------
POST {{host}}/login
[FormParams]
email: nadie@example.com
password: secreto123
HTTP 401
[Asserts]
xpath "normalize-space(//article[@role='alert'])" == "email o contraseña incorrectos"


# ---------- 8. Login correcto (303 a /cuenta) ----------
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


# ---------- 9. Email repetido (409) ----------
POST {{host}}/registro
[FormParams]
nombre: Otra Persona
email: ana.hurl@example.com
password: secreto123
HTTP 409
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "ya está registrado"
# se conservan los datos ingresados
xpath "string(//input[@name='nombre']/@value)" == "Otra Persona"


# ---------- 10. Email inválido (400) ----------
POST {{host}}/registro
[FormParams]
nombre: Sin Arroba
email: correo-invalido
password: secreto123
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "no es válido"


# ---------- 11. Campos obligatorios vacíos (400) ----------
POST {{host}}/registro
[FormParams]
nombre:
email: vacio@example.com
password: secreto123
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "obligatorios"
