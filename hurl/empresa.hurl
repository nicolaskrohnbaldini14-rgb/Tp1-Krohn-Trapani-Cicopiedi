# Panel de empresas: alta y baja de servicios del catálogo.

# ---------- 1. Registrar empresa: va directo al panel (303) ----------
POST {{host}}/empresas/registro
[FormParams]
nombre: Empresa Hurl
email: empresa.hurl@example.com
password: secreto123
HTTP 303
[Asserts]
header "Location" == "/empresa"


# ---------- 2. El panel ofrece las categorías válidas ----------
GET {{host}}/empresa
HTTP 200
[Asserts]
xpath "normalize-space(//main//h1)" == "Hola, Empresa Hurl"
xpath "count(//form[@action='/empresa/servicios']//select[@name='categoria']/option)" == 8


# ---------- 3. Una empresa no accede a la sección de consumidores (403) ----------
GET {{host}}/cuenta
HTTP 403


# ---------- 4. Crear servicio (303) ----------
POST {{host}}/empresa/servicios
[FormParams]
nombre: Netflix Hurl
categoria: Streaming
HTTP 303
[Asserts]
header "Location" == "/empresa"


# ---------- 5. Aparece en el panel y en el catálogo público ----------
GET {{host}}/empresa
HTTP 200
[Captures]
eliminar_netflix: xpath "string(//tr[td[normalize-space(.)='Netflix Hurl']]//form/@action)"
[Asserts]
xpath "normalize-space(//tr[td[normalize-space(.)='Netflix Hurl']]/td[2])" == "Streaming"

GET {{host}}/
HTTP 200
[Asserts]
xpath "count(//ul[@class='catalogo']/li[strong='Netflix Hurl'])" == 1


# ---------- 6. Nombre repetido (409) ----------
POST {{host}}/empresa/servicios
[FormParams]
nombre: Netflix Hurl
categoria: Música
HTTP 409
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "ya existe"


# ---------- 7. Categoría inválida (400) ----------
POST {{host}}/empresa/servicios
[FormParams]
nombre: Servicio Raro
categoria: Inventada
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "no es válida"


# ---------- 8. Nombre vacío (400) ----------
POST {{host}}/empresa/servicios
[FormParams]
nombre:
categoria: Software
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "obligatorio"


# ---------- 9. Eliminar servicio sin suscripciones (303) ----------
POST {{host}}{{eliminar_netflix}}
HTTP 303

GET {{host}}/empresa
HTTP 200
[Asserts]
xpath "count(//td[normalize-space(.)='Netflix Hurl'])" == 0


# ---------- 10. Eliminar servicio inexistente (404) e ID inválido (400) ----------
POST {{host}}/empresa/servicios/999999/eliminar
HTTP 404

POST {{host}}/empresa/servicios/abc/eliminar
HTTP 400

POST {{host}}/logout
HTTP 303
