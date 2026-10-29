# Páginas del frontend (TP4): las sirve el mismo servidor Go.

# ---------- Home: links a login y registro y menú de servicios en la barra ----------
GET {{host}}/
HTTP 200
[Asserts]
header "Content-Type" contains "text/html"
xpath "string(//body/@data-pagina)" == "inicio"
xpath "count(//a[@href='login.html'])" >= 1
xpath "count(//a[@href='registro.html'])" >= 1
# ya no hay registro de empresas
xpath "count(//a[@href='registro.html?tipo=empresa'])" == 0
# el catálogo se despliega en un panel a pantalla completa
xpath "count(//nav//div[@id='panel-servicios'])" == 1
xpath "count(//div[@id='panel-servicios']//div[@id='menu-servicios'])" == 1
xpath "count(//ul[@id='catalogo'])" == 0
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
# vencimientos por mes: links a otros meses y calendario al costado
xpath "count(//a[@id='mes-siguiente'])" == 1
xpath "count(//a[@id='mes-anterior'])" == 1
xpath "count(//aside//div[@id='calendario'])" == 1
# buscador y aviso de solo lectura para meses anteriores
xpath "count(//input[@id='buscar-suscripciones'][@type='search'])" == 1
xpath "count(//p[@id='aviso-solo-lectura'])" == 1
# el detalle permite modificar el monto, el usuario y la contraseña de la cuenta
xpath "count(//dialog//form[@id='form-editar']//input[@name='monto'])" == 1
xpath "count(//dialog//form[@id='form-editar']//input[@name='usuario_cuenta'])" == 1
xpath "count(//dialog//form[@id='form-editar']//input[@name='password_cuenta'])" == 1
# el detalle de cada suscripción se abre en un <dialog>
xpath "count(//dialog[@id='detalle'])" == 1
xpath "count(//dialog//form[@id='form-confirmar']//input[@type='password'])" == 1

GET {{host}}/nueva-suscripcion.html
HTTP 200
[Asserts]
xpath "count(//form[@id='form-suscripcion'])" == 1
# el usuario lo toma el servidor de la sesión
xpath "count(//select[@name='id_usuario'])" == 0
# se crea siempre activa y las credenciales de la cuenta son opcionales
xpath "count(//select[@name='estado'])" == 0
xpath "count(//input[@name='usuario_cuenta'][@required])" == 0
xpath "count(//input[@name='password_cuenta'][@required])" == 0

# el panel de empresas ya no existe
GET {{host}}/empresa.html
HTTP 404

GET {{host}}/app.js
HTTP 200
[Asserts]
header "Content-Type" contains "javascript"
# el navegador tiene que revalidar los archivos para no usar versiones viejas
header "Cache-Control" == "no-cache"

GET {{host}}/styles.css
HTTP 200
[Asserts]
header "Content-Type" contains "text/css"
