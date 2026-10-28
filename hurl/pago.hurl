# Pruebas del registro de pagos con formularios HTML.

# ---------- Setup: usuario, servicio y suscripción ----------
POST {{host}}/usuarios
[FormParams]
nombre: Usuario Pagos
email: usuario.pagos@example.com
password: secreto123
HTTP 303

POST {{host}}/servicios
[FormParams]
nombre: Gimnasio Pagos
categoria: Gimnasio
HTTP 303

GET {{host}}/
HTTP 200
[Captures]
usuario_id: xpath "string(//select[@name='id_usuario']/option[normalize-space(.)='Usuario Pagos']/@value)"
servicio_id: xpath "string(//select[@name='id_servicio']/option[starts-with(normalize-space(.), 'Gimnasio Pagos')]/@value)"

POST {{host}}/suscripciones
[FormParams]
id_usuario: {{usuario_id}}
id_servicio: {{servicio_id}}
monto: 32000
fecha_inicio: 2026-01-01
fecha_vencimiento: 2026-06-30
usuario_cuenta: socio-123
password_cuenta: clave
estado: Activa
HTTP 303

GET {{host}}/
HTTP 200
[Captures]
sub_id: xpath "string(//select[@name='id_suscripcion']/option[normalize-space(.)='Usuario Pagos · Gimnasio Pagos']/@value)"
[Asserts]
xpath "count(//form[@action='/pagos'])" == 1


# ---------- 1. Registrar pago (303) ----------
POST {{host}}/pagos
[FormParams]
id_suscripcion: {{sub_id}}
monto: 32000
fecha_pago: 2026-03-10
HTTP 303
[Asserts]
header "Location" == "/"


# ---------- 2. El pago aparece en el historial ----------
GET {{host}}/
HTTP 200
[Captures]
eliminar_pago: xpath "string(//section[@id='pagos']//tr[td[normalize-space(.)='Usuario Pagos']]//form/@action)"
[Asserts]
xpath "normalize-space(//section[@id='pagos']//tr[td[normalize-space(.)='Usuario Pagos']]/td[1])" == "10/03/2026"
xpath "normalize-space(//section[@id='pagos']//tr[td[normalize-space(.)='Usuario Pagos']]/td[3])" == "Gimnasio Pagos"
xpath "normalize-space(//section[@id='pagos']//tr[td[normalize-space(.)='Usuario Pagos']]/td[4])" == "$ 32.000,00"


# ---------- 3. Fecha de pago posterior al vencimiento (400) ----------
POST {{host}}/pagos
[FormParams]
id_suscripcion: {{sub_id}}
monto: 32000
fecha_pago: 2026-07-15
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "posterior"


# ---------- 4. Monto inválido (400) ----------
POST {{host}}/pagos
[FormParams]
id_suscripcion: {{sub_id}}
monto: -10
fecha_pago: 2026-03-10
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "mayor a cero"


# ---------- 5. Fecha faltante (400) ----------
POST {{host}}/pagos
[FormParams]
id_suscripcion: {{sub_id}}
monto: 100
fecha_pago:
HTTP 400
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "obligatoria"


# ---------- 6. Suscripción inexistente (404) ----------
POST {{host}}/pagos
[FormParams]
id_suscripcion: 999999
monto: 100
fecha_pago: 2026-03-10
HTTP 404
[Asserts]
xpath "normalize-space(//article[@role='alert'])" contains "no existe"


# ---------- 7. Eliminar pago (303) ----------
POST {{host}}{{eliminar_pago}}
HTTP 303


# ---------- 8. Ya no aparece en el historial ----------
GET {{host}}/
HTTP 200
[Asserts]
xpath "count(//section[@id='pagos']//td[normalize-space(.)='Usuario Pagos'])" == 0


# ---------- 9. Eliminar pago inexistente (404) ----------
POST {{host}}/pagos/999999/eliminar
HTTP 404
