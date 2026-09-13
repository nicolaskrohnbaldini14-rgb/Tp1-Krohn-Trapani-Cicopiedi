--tabla usuario
CREATE TABLE usuario (
    id_usuario SERIAL PRIMARY KEY,
    nombre VARCHAR(255) NOT NULL,
    apellido VARCHAR(255) NOT NULL,
    dni VARCHAR(20) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
--tabla de suscipciones 
CREATE TABLE suscripcion (
    id_suscripcion SERIAL PRIMARY KEY,
    id_usuario INT NOT NULL,
    id_servicio INT NOT NULL,
    monto DECIMAL(10, 2) NOT NULL,
    fecha_vencimiento DATE NOT NULL,
    fecha_inicio DATE NOT NULL,
    usuario_cuenta  VARCHAR(255) NOT NULL,
    password_cuenta VARCHAR(255) NOT NULL,
    estado VARCHAR(20) DEFAULT 'Activa',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- tabla servicio 
CREATE TABLE servicio (
    id_servicio SERIAL PRIMARY KEY,
    nombre VARCHAR(100) UNIQUE NOT NULL,
    categoria VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- tabla de pagos
CREATE TABLE pago (
    id_pago SERIAL PRIMARY KEY,
    id_suscripcion INT NOT NULL,
    monto DECIMAL(10, 2) NOT NULL,
    fecha_pago DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


-- las alteraciones con las pk y fk 

ALTER TABLE suscripcion
    ADD CONSTRAINT fk_sub_usuario
    FOREIGN KEY (id_usuario) REFERENCES usuario(id_usuario) ON DELETE CASCADE;

ALTER TABLE suscripcion
    ADD CONSTRAINT fk_sub_servicio
    FOREIGN KEY (id_servicio) REFERENCES servicio(id_servicio);


ALTER TABLE pago
    ADD CONSTRAINT fk_pago_sub 
    FOREIGN KEY (id_suscripcion) REFERENCES suscripcion(id_suscripcion) ON DELETE CASCADE;
