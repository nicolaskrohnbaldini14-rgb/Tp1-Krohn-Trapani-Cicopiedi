// Todas las páginas comparten este archivo. Cada una indica cuál es con
// <body data-pagina="...">, y al cargar se ejecuta la función correspondiente.
// Las páginas las sirve el mismo servidor Go que expone la API, por eso las
// rutas son relativas (mismo origen: sin CORS y con la cookie de sesión).

const cajaMensaje = document.getElementById("mensaje");

// ---------------------------------------------------------------------------
// Utilidades
// ---------------------------------------------------------------------------

// pedir hace el fetch y lanza un Error con el mensaje que devuelve la API
async function pedir(url, opciones = {}) {
  const respuesta = await fetch(url, opciones);
  if (!respuesta.ok) {
    const texto = (await respuesta.text()).trim();
    const error = new Error(texto || `Error ${respuesta.status}`);
    error.status = respuesta.status;
    throw error;
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
  temporizadorMensaje = setTimeout(() => (cajaMensaje.hidden = true), 5000);
}

function mostrarError(error) {
  mostrarMensaje(error.message, "error");
}

// crearItem arma un <li> con un título, un detalle y (opcional) el botón de eliminar.
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
  li.append(info);

  if (id !== undefined) {
    const boton = document.createElement("button");
    boton.type = "button";
    boton.className = "btn-eliminar";
    boton.dataset.id = id;
    boton.textContent = "Eliminar";
    li.append(boton);
  }
  return li;
}

function mostrarVacio(contenedor, texto) {
  const li = document.createElement("li");
  li.className = "vacio";
  li.textContent = texto;
  contenedor.replaceChildren(li);
}

// formatearFecha pasa "2026-01-01T00:00:00Z" a "01/01/2026" sin corrimientos de zona horaria
function formatearFecha(fechaISO) {
  const [anio, mes, dia] = fechaISO.slice(0, 10).split("-");
  return `${dia}/${mes}/${anio}`;
}

const formatoPesos = new Intl.NumberFormat("es-AR", { style: "currency", currency: "ARS" });

function totalActivo(suscripciones) {
  return suscripciones
    .filter((s) => s.estado === "Activa")
    .reduce((suma, s) => suma + Number(s.monto), 0);
}

// ---------------------------------------------------------------------------
// Sesión
// ---------------------------------------------------------------------------

// obtenerSesion devuelve el usuario que inició sesión o null
async function obtenerSesion() {
  try {
    return await pedir("/sesion");
  } catch {
    return null;
  }
}

// exigirSesion redirige al login si no hay sesión iniciada.
// La API igual rechaza las acciones sin sesión (401).
function exigirSesion(sesion) {
  if (!sesion) {
    location.href = "login.html";
    return false;
  }
  return true;
}

function enlace(href, texto, clase) {
  const a = document.createElement("a");
  a.href = href;
  a.textContent = texto;
  if (clase) a.className = clase;
  return a;
}

// pintarBarra muestra los links que corresponden según haya sesión o no.
// El menú desplegable de servicios queda siempre primero.
function pintarBarra(sesion) {
  const nav = document.getElementById("navegacion");
  if (!sesion) return; // quedan los links de "Iniciar sesión" y "Registrarse"

  const menuServicios = document.getElementById("desplegable-servicios");

  const items = [
    enlace("suscripciones.html", "Mis suscripciones"),
    enlace("nueva-suscripcion.html", "Nueva suscripción"),
  ];

  const saludo = document.createElement("span");
  saludo.className = "saludo";
  saludo.textContent = `Hola, ${sesion.nombre}`;

  const salir = document.createElement("button");
  salir.type = "button";
  salir.className = "boton-salir";
  salir.textContent = "Cerrar sesión";
  salir.addEventListener("click", async () => {
    await pedir("/sesion", { method: "DELETE" });
    location.href = "/";
  });

  nav.replaceChildren(menuServicios, ...items, saludo, salir);
}

// mismo orden de categorías que el catálogo precargado (logic/servicio.go)
const ORDEN_CATEGORIAS = ["Streaming", "Música", "Software", "Deportes", "Salud", "Financiero", "Turismo", "Otros"];

// agruparPorCategoria devuelve [[categoria, servicios], ...] siguiendo
// ORDEN_CATEGORIAS; dentro de cada una, los servicios quedan por nombre
function agruparPorCategoria(servicios) {
  const grupos = new Map();
  for (const s of servicios) {
    if (!grupos.has(s.categoria)) grupos.set(s.categoria, []);
    grupos.get(s.categoria).push(s);
  }
  const posicion = (categoria) => {
    const i = ORDEN_CATEGORIAS.indexOf(categoria);
    return i === -1 ? ORDEN_CATEGORIAS.length : i;
  };
  return [...grupos.entries()].sort(([a], [b]) => posicion(a) - posicion(b));
}

// cargarMenuServicios completa el desplegable de la barra con el catálogo
// precargado, agrupado por categoría (solo lectura)
async function cargarMenuServicios() {
  const menu = document.getElementById("menu-servicios");
  try {
    const servicios = await pedir("/servicios");
    if (servicios.length === 0) {
      mostrarVacio(menu, "Todavía no hay servicios disponibles.");
      return;
    }
    const items = [];
    for (const [categoria, lista] of agruparPorCategoria(servicios)) {
      const titulo = document.createElement("li");
      titulo.className = "desplegable-categoria";
      titulo.textContent = categoria;
      items.push(titulo);
      for (const s of lista) {
        const li = document.createElement("li");
        li.textContent = s.nombre;
        items.push(li);
      }
    }
    menu.replaceChildren(...items);
  } catch {
    mostrarVacio(menu, "No se pudieron cargar los servicios.");
  }
}

// ---------------------------------------------------------------------------
// Páginas públicas
// ---------------------------------------------------------------------------

async function paginaInicio(sesion) {
  if (sesion) {
    document
      .getElementById("acciones-inicio")
      .replaceChildren(enlace("cuenta.html", "Ir a mi cuenta", "boton"));
  }
}

async function paginaLogin(sesion) {
  if (sesion) {
    location.href = "cuenta.html";
    return;
  }

  document.getElementById("form-login").addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const datos = Object.fromEntries(new FormData(evento.target));
    try {
      await enviarJSON("/sesion", "POST", datos);
      location.href = "cuenta.html";
    } catch (error) {
      mostrarError(error);
    }
  });
}

async function paginaRegistro(sesion) {
  if (sesion) {
    location.href = "cuenta.html";
    return;
  }

  document.getElementById("form-registro").addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const datos = Object.fromEntries(new FormData(evento.target));
    try {
      await enviarJSON("/usuarios", "POST", datos);
      // después de registrarse queda con la sesión iniciada
      await enviarJSON("/sesion", "POST", { email: datos.email, password: datos.password });
      location.href = "cuenta.html";
    } catch (error) {
      mostrarError(error);
    }
  });
}

// ---------------------------------------------------------------------------
// Cuenta y suscripciones
// ---------------------------------------------------------------------------

async function paginaCuenta(sesion) {
  if (!exigirSesion(sesion)) return;

  document.getElementById("saludo").textContent = `Hola, ${sesion.nombre}`;

  const suscripciones = await pedir(`/suscripciones/usuario/${sesion.id_usuario}`);
  const activas = suscripciones.filter((s) => s.estado === "Activa").length;
  const texto = activas === 1 ? "1 suscripción activa" : `${activas} suscripciones activas`;
  document.getElementById("resumen").textContent =
    `Tenés ${texto} · ${formatoPesos.format(totalActivo(suscripciones))} por mes`;
}

async function paginaSuscripciones(sesion) {
  if (!exigirSesion(sesion)) return;

  const lista = document.getElementById("lista-suscripciones");
  const total = document.getElementById("total-mensual");

  // se guardan para poder armar el PUT al cambiar el estado
  let suscripciones = [];

  async function cargar() {
    suscripciones = await pedir(`/suscripciones/usuario/${sesion.id_usuario}`);
    if (suscripciones.length === 0) {
      mostrarVacio(lista, "Todavía no registraste ninguna suscripción.");
      total.textContent = "";
      return;
    }
    lista.replaceChildren(...suscripciones.map(itemSuscripcion));
    total.textContent = `Gasto mensual activo: ${formatoPesos.format(totalActivo(suscripciones))}`;
  }

  // Delegación de eventos: el cambio de estado de cualquier suscripción
  lista.addEventListener("change", async (evento) => {
    const select = evento.target.closest(".select-estado");
    if (!select) return;
    const sub = suscripciones.find((s) => String(s.id_suscripcion) === select.dataset.id);
    try {
      // la contraseña no se envía: la API conserva la que estaba guardada
      await enviarJSON(`/suscripciones/${sub.id_suscripcion}`, "PUT", {
        monto: sub.monto,
        fecha_vencimiento: sub.fecha_vencimiento,
        usuario_cuenta: sub.usuario_cuenta,
        estado: select.value,
      });
      mostrarMensaje(`Suscripción a ${sub.servicio_nombre}: ${select.value.toLowerCase()}.`);
    } catch (error) {
      mostrarError(error);
    }
    await cargar();
  });

  // Delegación de eventos: un solo listener en la lista atiende todos los botones
  lista.addEventListener("click", async (evento) => {
    const boton = evento.target.closest(".btn-eliminar");
    if (!boton) return;
    try {
      await pedir(`/suscripciones/${boton.dataset.id}`, { method: "DELETE" });
      mostrarMensaje("Suscripción eliminada.");
      await cargar();
    } catch (error) {
      mostrarError(error);
    }
  });

  await cargar();
}

// itemSuscripcion arma el <li> de una suscripción con su selector de estado
// y el botón de eliminar
function itemSuscripcion(s) {
  const detalle = [s.servicio_categoria];
  if (s.usuario_cuenta) detalle.push(`cuenta ${s.usuario_cuenta}`);
  detalle.push(`del ${formatearFecha(s.fecha_inicio)} al ${formatearFecha(s.fecha_vencimiento)}`);

  const li = crearItem(
    `${s.servicio_nombre} — ${formatoPesos.format(Number(s.monto))}`,
    detalle.join(" · "),
    s.id_suscripcion
  );

  const select = document.createElement("select");
  select.className = "select-estado";
  select.dataset.id = s.id_suscripcion;
  select.setAttribute("aria-label", `Estado de ${s.servicio_nombre}`);
  for (const estado of ["Activa", "Pausada", "Cancelada"]) {
    select.append(new Option(estado, estado, false, estado === s.estado));
  }
  li.querySelector(".btn-eliminar").before(select);
  return li;
}

async function paginaNuevaSuscripcion(sesion) {
  if (!exigirSesion(sesion)) return;

  const select = document.getElementById("select-servicios");
  const form = document.getElementById("form-suscripcion");

  // las opciones se agrupan por categoría con <optgroup>
  const servicios = await pedir("/servicios");
  if (servicios.length === 0) {
    select.append(new Option("Todavía no hay servicios disponibles", ""));
    form.querySelector("button").disabled = true;
  }
  for (const [categoria, lista] of agruparPorCategoria(servicios)) {
    const grupo = document.createElement("optgroup");
    grupo.label = categoria;
    for (const s of lista) {
      grupo.append(new Option(s.nombre, s.id_servicio));
    }
    select.append(grupo);
  }

  form.addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const datos = Object.fromEntries(new FormData(form));

    // el usuario lo toma el servidor de la sesión y el estado inicial es
    // siempre Activa; la API espera el id numérico y las fechas completas en
    // formato RFC 3339. Usuario y contraseña de la cuenta son opcionales.
    const cuerpo = {
      ...datos,
      id_servicio: Number(datos.id_servicio),
      fecha_inicio: `${datos.fecha_inicio}T00:00:00Z`,
      fecha_vencimiento: `${datos.fecha_vencimiento}T00:00:00Z`,
    };

    try {
      await enviarJSON("/suscripciones", "POST", cuerpo);
      location.href = "suscripciones.html";
    } catch (error) {
      // si falla se conservan los datos del formulario para corregirlos
      mostrarError(error);
    }
  });
}

// ---------------------------------------------------------------------------
// Arranque
// ---------------------------------------------------------------------------

const paginas = {
  inicio: paginaInicio,
  login: paginaLogin,
  registro: paginaRegistro,
  cuenta: paginaCuenta,
  suscripciones: paginaSuscripciones,
  "nueva-suscripcion": paginaNuevaSuscripcion,
};

document.addEventListener("DOMContentLoaded", async () => {
  const sesion = await obtenerSesion();
  pintarBarra(sesion);
  cargarMenuServicios();
  try {
    await paginas[document.body.dataset.pagina]?.(sesion);
  } catch (error) {
    mostrarMensaje(`No se pudo conectar con la API: ${error.message}`, "error");
  }
});
