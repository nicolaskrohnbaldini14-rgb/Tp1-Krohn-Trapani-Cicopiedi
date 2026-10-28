# Pruebas del ABM de servicios con formularios HTML.

# ---------- 1. El formulario ofrece las categorías válidas ----------
GET {{host}}/
HTTP 200
[Asserts]
xpath "count(//form[@action='/servicios']//select[@name='categoria']/option)" == 8
xpath "count(//form[@action='/servicios']//option[@value='Streaming'])" == 1


# ---------- 2. Crear servicio (303) ----------
POST {{host}}/servicios
[FormParams]
nombre: Netflix Hurl
categoria: Streaming
HTTP 303
[Asserts]
header "Location" == "/"


# ---------- 3. El servicio aparece en el catálogo con su categoría ----------
GET {{host}}/
HTTP 200
[Captures]
eliminar_netflix: xpath "string(//section[@id='servicios']//tr[td[normalize-space(.)='Netflix Hurl']]//form/@action)"
[Asserts]
xpath "normalize-space(//section[@id='servicios']//tr[td[normalize-space(.)='Netflix Hurl']]/td[2])" == "Streaming"


# ---------- 4. Nombre repetido (409) ----------
POST {{host}}/servicios
[FormParams]
nombre: Netflix Hurl
categoria: Música
HTTP 409
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "ya existe"


# ---------- 5. Categoría inválida (400) ----------
POST {{host}}/servicios
[FormParams]
nombre: Servicio Raro
categoria: Inventada
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "no es válida"


# ---------- 6. Nombre vacío (400) ----------
POST {{host}}/servicios
[FormParams]
nombre:
categoria: Software
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "obligatorio"


# ---------- 7. Eliminar servicio sin suscripciones (303) ----------
POST {{host}}{{eliminar_netflix}}
HTTP 303


# ---------- 8. Ya no aparece en el catálogo ----------
GET {{host}}/
HTTP 200
[Asserts]
xpath "count(//section[@id='servicios']//td[normalize-space(.)='Netflix Hurl'])" == 0


# ---------- 9. Eliminar servicio inexistente (404) ----------
POST {{host}}/servicios/999999/eliminar
HTTP 404
