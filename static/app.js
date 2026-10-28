// La página la sirve el mismo servidor Go que expone la API,
// por eso las rutas son relativas (mismo origen, sin problemas de CORS).

const listaUsuarios = document.getElementById("lista-usuarios");
const listaServicios = document.getElementById("lista-servicios");
const listaSuscripciones = document.getElementById("lista-suscripciones");
const filtroUsuario = document.getElementById("filtro-usuario");
const selectServicios = document.getElementById("select-servicios");
const totalMensual = document.getElementById("total-mensual");
const cajaMensaje = document.getElementById("mensaje");

// ---------------------------------------------------------------------------
// Utilidades
// ---------------------------------------------------------------------------

// pedir hace el fetch y lanza un Error con el mensaje que devuelve la API
async function pedir(url, opciones = {}) {
  const respuesta = await fetch(url, opciones);
  if (!respuesta.ok) {
    const texto = (await respuesta.text()).trim();
    throw new Error(texto || `Error ${respuesta.status}`);
  }
  if (respuesta.status === 204) {
    return null;
  }
  return respuesta.json();
}

function enviarJSON(url, metodo, datos) {
  return pedir(url, {
    method: metodo,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(datos),
  });
}

let temporizadorMensaje;
function mostrarMensaje(texto, tipo = "ok") {
  cajaMensaje.textContent = texto;
  cajaMensaje.className = `mensaje ${tipo}`;
  cajaMensaje.hidden = false;
  clearTimeout(temporizadorMensaje);
  temporizadorMensaje = setTimeout(() => (cajaMensaje.hidden = true), 4000);
}

// crearItem arma un <li> con un título, un detalle y el botón de eliminar.
// Se usa textContent (y no innerHTML) para no inyectar HTML con datos del usuario.
function crearItem(titulo, detalle, id, etiqueta) {
  const li = document.createElement("li");
  li.className = "item";

  const info = document.createElement("div");
  info.className = "item-info";

  const strong = document.createElement("strong");
  strong.textContent = titulo;
  if (etiqueta) {
    const span = document.createElement("span");
    span.className = "etiqueta";
    span.textContent = etiqueta;
    strong.append(" ", span);
  }

  const small = document.createElement("span");
  small.className = "item-detalle";
  small.textContent = detalle;

  info.append(strong, small);

  const boton = document.createElement("button");
  boton.type = "button";
  boton.className = "btn-eliminar";
  boton.dataset.id = id;
  boton.textContent = "Eliminar";

  li.append(info, boton);
  return li;
}

function mostrarVacio(contenedor, texto) {
  const li = document.createElement("li");
  li.className = "vacio";
  li.textContent = texto;
  contenedor.replaceChildren(li);
}

function llenarSelect(select, opciones, textoVacio) {
  const valorAnterior = select.value;
  select.replaceChildren();
  if (opciones.length === 0) {
    select.append(new Option(textoVacio, ""));
    return;
  }
  for (const { valor, texto } of opciones) {
    select.append(new Option(texto, valor));
  }
  // si la opción elegida antes sigue existiendo, se mantiene seleccionada
  if (opciones.some((o) => String(o.valor) === valorAnterior)) {
    select.value = valorAnterior;
  }
}

// formatearFecha pasa "2026-01-01T00:00:00Z" a "01/01/2026" sin corrimientos de zona horaria
function formatearFecha(fechaISO) {
  const [anio, mes, dia] = fechaISO.slice(0, 10).split("-");
  return `${dia}/${mes}/${anio}`;
}

const formatoPesos = new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS" });

// ---------------------------------------------------------------------------
// Usuarios
// ---------------------------------------------------------------------------

async function cargarUsuarios() {
  const usuarios = await pedir("/usuarios");

  if (usuarios.length === 0) {
    mostrarVacio(listaUsuarios, "Todavía no hay usuarios.");
  } else {
    listaUsuarios.replaceChildren(
      ...usuarios.map((u) => crearItem(u.nombre, u.email, u.id_usuario))
    );
  }

  const opciones = usuarios.map((u) => ({ valor: u.id_usuario, texto: u.nombre }));
  document.querySelectorAll(".select-usuarios").forEach((select) => {
    llenarSelect(select, opciones, "Primero creá un usuario");
  });
}

document.getElementById("form-usuario").addEventListener("submit", async (evento) => {
  evento.preventDefault();
  const form = evento.target;
  const datos = Object.fromEntries(new FormData(form));

  try {
    await enviarJSON("/usuarios", "POST", datos);
    form.reset();
    mostrarMensaje("Usuario creado.");
    await cargarUsuarios();
    await cargarSuscripciones();
  } catch (error) {
    mostrarMensaje(error.message, "error");
  }
});

// Delegación de eventos: un solo listener en la lista atiende todos los botones
listaUsuarios.addEventListener("click", async (evento) => {
  const boton = evento.target.closest(".btn-eliminar");
  if (!boton) return;

  try {
    await pedir(`/usuarios/${boton.dataset.id}`, { method: "DELETE" });
    mostrarMensaje("Usuario eliminado (junto con sus suscripciones).");
    await cargarUsuarios();
    await cargarSuscripciones();
  } catch (error) {
    mostrarMensaje(error.message, "error");
  }
});

// ---------------------------------------------------------------------------
// Servicios
// ---------------------------------------------------------------------------

async function cargarServicios() {
  const servicios = await pedir("/servicios");

  if (servicios.length === 0) {
    mostrarVacio(listaServicios, "Todavía no hay servicios.");
  } else {
    listaServicios.replaceChildren(
      ...servicios.map((s) => crearItem(s.nombre, "", s.id_servicio, s.categoria))
    );
  }

  llenarSelect(
    selectServicios,
    servicios.map((s) => ({ valor: s.id_servicio, texto: `${s.nombre} (${s.categoria})` })),
    "Primero creá un servicio"
  );
}

document.getElementById("form-servicio").addEventListener("submit", async (evento) => {
  evento.preventDefault();
  const form = evento.target;
  const datos = Object.fromEntries(new FormData(form));

  try {
    await enviarJSON("/servicios", "POST", datos);
    form.reset();
    mostrarMensaje("Servicio creado.");
    await cargarServicios();
  } catch (error) {
    mostrarMensaje(error.message, "error");
  }
});

listaServicios.addEventListener("click", async (evento) => {
  const boton = evento.target.closest(".btn-eliminar");
  if (!boton) return;

  try {
    await pedir(`/servicios/${boton.dataset.id}`, { method: "DELETE" });
    mostrarMensaje("Servicio eliminado.");
    await cargarServicios();
  } catch (error) {
    mostrarMensaje(error.message, "error");
  }
});

// ---------------------------------------------------------------------------
// Suscripciones
// ---------------------------------------------------------------------------

async function cargarSuscripciones() {
  const idUsuario = filtroUsuario.value;
  if (!idUsuario) {
    mostrarVacio(listaSuscripciones, "Elegí un usuario para ver sus suscripciones.");
    totalMensual.textContent = "";
    return;
  }

  const suscripciones = await pedir(`/suscripciones/usuario/${idUsuario}`);

  if (suscripciones.length === 0) {
    mostrarVacio(listaSuscripciones, "Este usuario no tiene suscripciones.");
  } else {
    listaSuscripciones.replaceChildren(
      ...suscripciones.map((s) =>
        crearItem(
          `${s.servicio_nombre} — ${formatoPesos.format(Number(s.monto))}`,
          `${s.servicio_categoria} · cuenta ${s.usuario_cuenta} · ` +
            `del ${formatearFecha(s.fecha_inicio)} al ${formatearFecha(s.fecha_vencimiento)}`,
          s.id_suscripcion,
          s.estado
        )
      )
    );
  }

  const total = suscripciones
    .filter((s) => s.estado === "Activa")
    .reduce((suma, s) => suma + Number(s.monto), 0);
  totalMensual.textContent = `Gasto mensual activo: ${formatoPesos.format(total)}`;
}

document.getElementById("form-suscripcion").addEventListener("submit", async (evento) => {
  evento.preventDefault();
  const form = evento.target;
  const datos = Object.fromEntries(new FormData(form));

  // la API espera ids numéricos y fechas completas en formato RFC 3339
  const cuerpo = {
    ...datos,
    id_usuario: Number(datos.id_usuario),
    id_servicio: Number(datos.id_servicio),
    fecha_inicio: `${datos.fecha_inicio}T00:00:00Z`,
    fecha_vencimiento: `${datos.fecha_vencimiento}T00:00:00Z`,
  };

  try {
    await enviarJSON("/suscripciones", "POST", cuerpo);
    form.reset();
    mostrarMensaje("Suscripción creada.");
    // se muestra la lista del usuario al que se le creó la suscripción
    filtroUsuario.value = datos.id_usuario;
    await cargarSuscripciones();
  } catch (error) {
    mostrarMensaje(error.message, "error");
  }
});

listaSuscripciones.addEventListener("click", async (evento) => {
  const boton = evento.target.closest(".btn-eliminar");
  if (!boton) return;

  try {
    await pedir(`/suscripciones/${boton.dataset.id}`, { method: "DELETE" });
    mostrarMensaje("Suscripción eliminada.");
    await cargarSuscripciones();
  } catch (error) {
    mostrarMensaje(error.message, "error");
  }
});

filtroUsuario.addEventListener("change", () => {
  cargarSuscripciones().catch((error) => mostrarMensaje(error.message, "error"));
});

// ---------------------------------------------------------------------------
// Carga inicial
// ---------------------------------------------------------------------------

document.addEventListener("DOMContentLoaded", async () => {
  try {
    await Promise.all([cargarUsuarios(), cargarServicios()]);
    await cargarSuscripciones();
  } catch (error) {
    mostrarMensaje(`No se pudo conectar con la API: ${error.message}`, "error");
  }
});
