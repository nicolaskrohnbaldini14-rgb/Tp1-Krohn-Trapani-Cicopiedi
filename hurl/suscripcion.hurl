# Pruebas del ABM de suscripciones con formularios HTML.

# ---------- Setup: usuario y servicio ----------
POST {{host}}/usuarios
[FormParams]
nombre: Usuario Subs
email: usuario.subs@example.com
password: secreto123
HTTP 303

POST {{host}}/servicios
[FormParams]
nombre: Spotify Subs
categoria: Música
HTTP 303

# los ids se toman de las opciones de los <select> del formulario
GET {{host}}/
HTTP 200
[Captures]
usuario_id: xpath "string(//select[@name='id_usuario']/option[normalize-space(.)='Usuario Subs']/@value)"
servicio_id: xpath "string(//select[@name='id_servicio']/option[starts-with(normalize-space(.), 'Spotify Subs')]/@value)"
eliminar_servicio: xpath "string(//section[@id='servicios']//tr[td[normalize-space(.)='Spotify Subs']]//form/@action)"
[Asserts]
xpath "count(//form[@action='/suscripciones'])" == 1


# ---------- 1. Crear suscripción (303) ----------
POST {{host}}/suscripciones
[FormParams]
id_usuario: {{usuario_id}}
id_servicio: {{servicio_id}}
monto: 1500
fecha_inicio: 2026-01-01
fecha_vencimiento: 2026-12-31
usuario_cuenta: cuenta@test.com
password_cuenta: clavecuenta
estado: Activa
HTTP 303
[Asserts]
header "Location" == "/"


# ---------- 2. La suscripción aparece con los datos formateados ----------
GET {{host}}/
HTTP 200
[Captures]
eliminar_sub: xpath "string(//section[@id='suscripciones']//tr[td[normalize-space(.)='Usuario Subs']]//form/@action)"
[Asserts]
xpath "normalize-space(//section[@id='suscripciones']//tr[td[normalize-space(.)='Usuario Subs']]/td[3])" == "$ 1.500,00"
xpath "normalize-space(//section[@id='suscripciones']//tr[td[normalize-space(.)='Usuario Subs']]/td[4])" == "01/01/2026 al 31/12/2026"
xpath "normalize-space(//section[@id='suscripciones']//tr[td[normalize-space(.)='Usuario Subs']]/td[6])" == "Activa"
xpath "normalize-space(//section[@id='suscripciones']//p[@class='total'])" matches "^Gasto mensual activo: \\$ [0-9.]+,[0-9]{2}$"
body not contains "clavecuenta"


# ---------- 3. Segunda suscripción activa al mismo servicio (409) ----------
POST {{host}}/suscripciones
[FormParams]
id_usuario: {{usuario_id}}
id_servicio: {{servicio_id}}
monto: 1500
fecha_inicio: 2026-02-01
fecha_vencimiento: 2026-12-31
usuario_cuenta: cuenta@test.com
password_cuenta: clavecuenta
estado: Activa
HTTP 409
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "ya posee una suscripción activa"


# ---------- 4. Monto inválido (400) ----------
POST {{host}}/suscripciones
[FormParams]
id_usuario: {{usuario_id}}
id_servicio: {{servicio_id}}
monto: 0
fecha_inicio: 2026-01-01
fecha_vencimiento: 2026-12-31
usuario_cuenta: cuenta@test.com
password_cuenta: clavecuenta
estado: Pausada
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "mayor a cero"


# ---------- 5. Fecha de inicio posterior al vencimiento (400) ----------
POST {{host}}/suscripciones
[FormParams]
id_usuario: {{usuario_id}}
id_servicio: {{servicio_id}}
monto: 1500
fecha_inicio: 2026-12-31
fecha_vencimiento: 2026-01-01
usuario_cuenta: cuenta@test.com
password_cuenta: clavecuenta
estado: Pausada
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "anterior"


# ---------- 6. Fechas faltantes (400) ----------
POST {{host}}/suscripciones
[FormParams]
id_usuario: {{usuario_id}}
id_servicio: {{servicio_id}}
monto: 1500
fecha_inicio:
fecha_vencimiento:
usuario_cuenta: cuenta@test.com
password_cuenta: clavecuenta
estado: Pausada
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "obligatorias"


# ---------- 7. Usuario inexistente (404) ----------
POST {{host}}/suscripciones
[FormParams]
id_usuario: 999999
id_servicio: {{servicio_id}}
monto: 1500
fecha_inicio: 2026-01-01
fecha_vencimiento: 2026-12-31
usuario_cuenta: cuenta@test.com
password_cuenta: clavecuenta
estado: Activa
HTTP 404
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "no existe"


# ---------- 8. Estado inválido (400) ----------
POST {{host}}/suscripciones
[FormParams]
id_usuario: {{usuario_id}}
id_servicio: {{servicio_id}}
monto: 1500
fecha_inicio: 2026-01-01
fecha_vencimiento: 2026-12-31
usuario_cuenta: cuenta@test.com
password_cuenta: clavecuenta
estado: Vencida
HTTP 400


# ---------- 9. No se puede borrar un servicio con suscripciones (409) ----------
POST {{host}}{{eliminar_servicio}}
HTTP 409
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "tiene suscripciones asociadas"


# ---------- 10. Eliminar suscripción (303) ----------
POST {{host}}{{eliminar_sub}}
HTTP 303


# ---------- 11. Ya no aparece en la lista ----------
GET {{host}}/
HTTP 200
[Asserts]
xpath "count(//section[@id='suscripciones']//td[normalize-space(.)='Usuario Subs'])" == 0


# ---------- 12. Eliminar suscripción inexistente (404) ----------
POST {{host}}/suscripciones/999999/eliminar
HTTP 404
