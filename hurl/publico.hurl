# Páginas públicas y control de acceso sin sesión.

# ---------- 1. Home: presentación, links de acceso y catálogo de solo lectura ----------
GET {{host}}/
HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
xpath "normalize-space(//section[@class='portada']/h1)" == "Suscripciones+"
xpath "string(//link[contains(@href, 'pico')]/@href)" contains "pico.min.css"
xpath "count(//a[@href='/login'])" >= 1
xpath "count(//a[@href='/registro'])" >= 1
xpath "count(//a[@href='/empresas/registro'])" == 1
xpath "count(//section[@id='servicios-disponibles'])" == 1
# en el inicio no se pueden crear ni borrar servicios
xpath "count(//form[@action='/empresa/servicios'])" == 0
xpath "count(//form[contains(@action, '/eliminar')])" == 0


# ---------- 2. Páginas de login y registro ----------
GET {{host}}/login
HTTP 200
[Asserts]
xpath "count(//form[@action='/login']//input[@name='email'])" == 1
xpath "count(//form[@action='/login']//input[@name='password'])" == 1

GET {{host}}/registro
HTTP 200
[Asserts]
xpath "count(//form[@action='/registro']//input[@name='nombre'])" == 1

GET {{host}}/empresas/registro
HTTP 200
[Asserts]
xpath "count(//form[@action='/empresas/registro'])" == 1
xpath "normalize-space(//article/header)" == "Registrar empresa"


# ---------- 3. Sin sesión, las páginas privadas redirigen al login ----------
GET {{host}}/cuenta
HTTP 303
[Asserts]
header "Location" == "/login"

GET {{host}}/cuenta/suscripciones
HTTP 303
[Asserts]
header "Location" == "/login"

GET {{host}}/cuenta/suscripciones/nueva
HTTP 303
[Asserts]
header "Location" == "/login"

POST {{host}}/cuenta/suscripciones
[FormParams]
id_servicio: 1
HTTP 303
[Asserts]
header "Location" == "/login"

GET {{host}}/empresa
HTTP 303
[Asserts]
header "Location" == "/login"

POST {{host}}/empresa/servicios
[FormParams]
nombre: Intruso
categoria: Otros
HTTP 303
[Asserts]
header "Location" == "/login"


# ---------- 4. Rutas inexistentes y métodos no permitidos ----------
GET {{host}}/no-existe
HTTP 404

GET {{host}}/logout
HTTP 405
