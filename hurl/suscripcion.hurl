# Suscripciones de consumidores: cada uno ve y borra solo las suyas.

# ---------- Setup: una empresa publica un servicio ----------
POST {{host}}/empresas/registro
[FormParams]
nombre: Spotify SA
email: spotify.hurl@example.com
password: secreto123
HTTP 303

POST {{host}}/empresa/servicios
[FormParams]
nombre: Spotify Subs
categoria: Música
HTTP 303

GET {{host}}/empresa
HTTP 200
[Captures]
eliminar_servicio: xpath "string(//tr[td[normalize-space(.)='Spotify Subs']]//form/@action)"

POST {{host}}/logout
HTTP 303


# ---------- Setup: se registra un consumidor ----------
POST {{host}}/registro
[FormParams]
nombre: Usuario Subs
email: usuario.subs@example.com
password: secreto123
HTTP 303

GET {{host}}/cuenta/suscripciones
HTTP 200
[Asserts]
xpath "normalize-space(//p[@class='vacio'])" contains "Todavía no registraste"

# el id del servicio se toma de las opciones del <select>
GET {{host}}/cuenta/suscripciones/nueva
HTTP 200
[Captures]
servicio_id: xpath "string(//select[@name='id_servicio']/option[starts-with(normalize-space(.), 'Spotify Subs')]/@value)"
[Asserts]
# el usuario sale de la sesión, no hay campo para elegirlo
xpath "count(//select[@name='id_usuario'])" == 0


# ---------- 1. Crear suscripción (303 a la lista) ----------
POST {{host}}/cuenta/suscripciones
[FormParams]
id_servicio: {{servicio_id}}
monto: 1500
fecha_inicio: 2026-01-01
fecha_vencimiento: 2026-12-31
usuario_cuenta: cuenta@test.com
password_cuenta: clavecuenta
estado: Activa
HTTP 303
[Asserts]
header "Location" == "/cuenta/suscripciones"


# ---------- 2. Aparece en la lista con los datos formateados ----------
GET {{host}}/cuenta/suscripciones
HTTP 200
[Captures]
eliminar_sub: xpath "string(//tr[td[starts-with(normalize-space(.), 'Spotify Subs')]]//form/@action)"
[Asserts]
xpath "normalize-space(//tr[td[starts-with(normalize-space(.), 'Spotify Subs')]]/td[2])" == "$ 1.500,00"
xpath "normalize-space(//tr[td[starts-with(normalize-space(.), 'Spotify Subs')]]/td[3])" == "01/01/2026 al 31/12/2026"
xpath "normalize-space(//tr[td[starts-with(normalize-space(.), 'Spotify Subs')]]/td[5])" == "Activa"
xpath "normalize-space(//p[@class='total'])" == "Gasto mensual activo: $ 1.500,00"
body not contains "clavecuenta"

GET {{host}}/cuenta
HTTP 200
[Asserts]
xpath "normalize-space(//p[@class='lema'])" contains "Tenés 1 suscripción activa"


# ---------- 3. Segunda suscripción activa al mismo servicio (409) ----------
POST {{host}}/cuenta/suscripciones
[FormParams]
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


# ---------- 4. Monto inválido (400) y se conservan los datos ----------
POST {{host}}/cuenta/suscripciones
[FormParams]
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
xpath "string(//input[@name='fecha_inicio']/@value)" == "2026-01-01"
xpath "string(//select[@name='estado']/option[@selected]/@value)" == "Pausada"


# ---------- 5. Fecha de inicio posterior al vencimiento (400) ----------
POST {{host}}/cuenta/suscripciones
[FormParams]
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
POST {{host}}/cuenta/suscripciones
[FormParams]
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


# ---------- 7. Servicio inexistente (404) ----------
POST {{host}}/cuenta/suscripciones
[FormParams]
id_servicio: 999999
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
POST {{host}}/cuenta/suscripciones
[FormParams]
id_servicio: {{servicio_id}}
monto: 1500
fecha_inicio: 2026-01-01
fecha_vencimiento: 2026-12-31
usuario_cuenta: cuenta@test.com
password_cuenta: clavecuenta
estado: Vencida
HTTP 400

POST {{host}}/logout
HTTP 303


# ---------- 9. Otro consumidor no ve ni puede borrar esa suscripción (404) ----------
POST {{host}}/registro
[FormParams]
nombre: Otro Consumidor
email: otro.consumidor@example.com
password: secreto123
HTTP 303

GET {{host}}/cuenta/suscripciones
HTTP 200
[Asserts]
xpath "count(//td[starts-with(normalize-space(.), 'Spotify Subs')])" == 0

POST {{host}}{{eliminar_sub}}
HTTP 404
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "no encontrada"

POST {{host}}/logout
HTTP 303


# ---------- 10. La empresa no puede borrar un servicio con suscripciones (409) ----------
POST {{host}}/login
[FormParams]
email: spotify.hurl@example.com
password: secreto123
HTTP 303

POST {{host}}{{eliminar_servicio}}
HTTP 409
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "tiene suscripciones asociadas"

POST {{host}}/logout
HTTP 303


# ---------- 11. El dueño elimina su suscripción (303) ----------
POST {{host}}/login
[FormParams]
email: usuario.subs@example.com
password: secreto123
HTTP 303

POST {{host}}{{eliminar_sub}}
HTTP 303
[Asserts]
header "Location" == "/cuenta/suscripciones"

GET {{host}}/cuenta/suscripciones
HTTP 200
[Asserts]
xpath "count(//td[starts-with(normalize-space(.), 'Spotify Subs')])" == 0


# ---------- 12. Eliminar suscripción inexistente (404) ----------
POST {{host}}/cuenta/suscripciones/999999/eliminar
HTTP 404

POST {{host}}/logout
HTTP 303
