# Páginas del frontend (TP4): las sirve el mismo servidor Go.

# ---------- Home: links a login y registro, catálogo de solo lectura ----------
GET {{host}}/
HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
xpath "string(//body/@data-pagina)" == "inicio"
xpath "count(//a[@href='login.html'])" >= 1
xpath "count(//a[@href='registro.html'])" >= 1
xpath "count(//a[@href='registro.html?tipo=empresa'])" == 1
xpath "count(//ul[@id='catalogo'])" == 1
# en el inicio no está el formulario de servicios
xpath "count(//form[@id='form-servicio'])" == 0
xpath "string(//script/@src)" == "app.js"

# ---------- Cada página indica cuál es para app.js ----------
GET {{host}}/login.html
HTTP 200
[Asserts]
xpath "string(//body/@data-pagina)" == "login"
xpath "count(//form[@id='form-login'])" == 1

GET {{host}}/registro.html
HTTP 200
[Asserts]
xpath "string(//body/@data-pagina)" == "registro"

GET {{host}}/cuenta.html
HTTP 200
[Asserts]
xpath "string(//body/@data-pagina)" == "cuenta"
xpath "count(//a[@href='suscripciones.html'])" >= 1
xpath "count(//a[@href='nueva-suscripcion.html'])" >= 1

GET {{host}}/suscripciones.html
HTTP 200
[Asserts]
xpath "count(//ul[@id='lista-suscripciones'])" == 1

GET {{host}}/nueva-suscripcion.html
HTTP 200
[Asserts]
xpath "count(//form[@id='form-suscripcion'])" == 1
# el usuario lo toma el servidor de la sesión
xpath "count(//select[@name='id_usuario'])" == 0

GET {{host}}/empresa.html
HTTP 200
[Asserts]
xpath "count(//form[@id='form-servicio'])" == 1
xpath "count(//ul[@id='lista-servicios'])" == 1

GET {{host}}/app.js
HTTP 200
[Asserts]
header "Content-Type" contains "javascript"

GET {{host}}/styles.css
HTTP 200
[Asserts]
header "Content-Type" contains "text/css"
