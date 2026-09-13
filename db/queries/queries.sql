
-- consultas para el usuario 

-- name: CreateUsuario :one
INSERT INTO usuario (
    nombre, apellido, dni, email, password
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetUsuario :one
SELECT * FROM usuario 
WHERE id_usuario = $1;

-- name: ListUsuarios :many
SELECT * FROM usuario 
ORDER BY apellido, nombre;

-- name: UpdateUsuario :exec
UPDATE usuario 
SET nombre = $2, 
    apellido = $3, 
    dni = $4, 
    email = $5, 
    password = $6
WHERE id_usuario = $1;

-- name: DeleteUsuario :exec
DELETE FROM usuario 
WHERE id_usuario = $1;

-- consultas del servicio

-- name: CreateServicio :one
INSERT INTO servicio (
    nombre, categoria
) VALUES (
    $1, $2
)
RETURNING *;

-- name: GetServicio :one
SELECT * FROM servicio 
WHERE id_servicio = $1;

-- name: ListServicios :many
SELECT * FROM servicio 
ORDER BY nombre;

-- name: UpdateServicio :exec
UPDATE servicio 
SET nombre = $2, 
    categoria = $3
WHERE id_servicio = $1;

-- name: DeleteServicio :exec
DELETE FROM servicio 
WHERE id_servicio = $1;

-- consulta para suscripcion 

-- name: CreateSubscripcion :one
INSERT INTO suscripcion (
    id_usuario, id_servicio, monto, fecha_vencimiento, fecha_inicio, usuario_cuenta, password_cuenta, estado
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: GetSubscripcion :one
SELECT * FROM suscripcion 
WHERE id_suscripcion = $1;

-- name: ListSuscripcionesUsuario :many
SELECT 
    sub.id_suscripcion
,
    sub.monto,
    sub.fecha_vencimiento,
    sub.fecha_inicio,
    sub.usuario_cuenta,
    sub.estado,
    ser.nombre AS servicio_nombre,
    ser.categoria AS servicio_categoria
FROM suscripcion sub
JOIN servicio ser ON sub.id_servicio = ser.id_servicio
WHERE sub.id_usuario = $1
ORDER BY ser.nombre ASC;

-- name: UpdateSubscripcion :exec
UPDATE suscripcion 
SET monto = $2, 
    fecha_vencimiento = $3, 
    estado = $4, 
    usuario_cuenta = $5, 
    password_cuenta = $6
WHERE id_suscripcion = $1;

-- name: DeleteSubscripcion :exec
DELETE FROM suscripcion 
WHERE id_suscripcion = $1;


-- consultas  de pago 

-- name: CreatePago :one
INSERT INTO pago (
    id_suscripcion
, monto, fecha_pago
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: GetPago :one
SELECT * FROM pago 
WHERE id_pago = $1;

-- name: ListPagos :many
SELECT * FROM pago 
WHERE id_suscripcion = $1 
ORDER BY fecha_pago DESC;

-- name: DeletePago :exec
DELETE FROM pago 
WHERE id_pago = $1;

-- el update de un pago no existe ya que seria un historial

