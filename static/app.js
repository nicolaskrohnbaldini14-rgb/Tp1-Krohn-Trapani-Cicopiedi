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
// Meses y vencimientos
// ---------------------------------------------------------------------------

// Los meses se manejan como texto "AAAA-MM" (por ejemplo "2026-10"), que es
// lo mismo que los primeros 7 caracteres de la fecha que devuelve la API.

function claveMes(fecha) {
  return `${fecha.getFullYear()}-${String(fecha.getMonth() + 1).padStart(2, "0")}`;
}

function mesActual() {
  return claveMes(new Date());
}

// sumarMeses("2026-12", 1) devuelve "2027-01"
function sumarMeses(clave, cantidad) {
  const [anio, mes] = clave.split("-").map(Number);
  return claveMes(new Date(anio, mes - 1 + cantidad, 1));
}

// nombreMes("2026-10") devuelve "octubre de 2026"
function nombreMes(clave) {
  const [anio, mes] = clave.split("-").map(Number);
  return new Intl.DateTimeFormat("es-AR", { month: "long", year: "numeric" }).format(new Date(anio, mes - 1, 1));
}

// mesDeLaURL lee ?mes=AAAA-MM; si no viene o es inválido, usa el mes actual
function mesDeLaURL() {
  const mes = new URLSearchParams(location.search).get("mes") ?? "";
  return /^\d{4}-(0[1-9]|1[0-2])$/.test(mes) ? mes : mesActual();
}

function venceEnMes(suscripcion, clave) {
  return suscripcion.fecha_vencimiento.slice(0, 7) === clave;
}

// pintarCalendario arma la grilla del mes (de lunes a domingo) y marca con un
// pin los días en que vence alguna suscripción. Al pasar el cursor (o enfocar
// el día) se muestra un cartel con los servicios que vencen ese día.
function pintarCalendario(contenedor, clave, suscripciones) {
  const [anio, mes] = clave.split("-").map(Number);
  const diasDelMes = new Date(anio, mes, 0).getDate();
  const columnaDelDia1 = (new Date(anio, mes - 1, 1).getDay() + 6) % 7; // lunes = 0
  const hoy = new Date();
  const esMesActual = clave === claveMes(hoy);

  const vencimientosPorDia = new Map();
  for (const s of suscripciones) {
    const dia = Number(s.fecha_vencimiento.slice(8, 10));
    if (!vencimientosPorDia.has(dia)) vencimientosPorDia.set(dia, []);
    vencimientosPorDia.get(dia).push(s);
  }

  const celdas = [];
  for (const inicial of ["L", "M", "M", "J", "V", "S", "D"]) {
    const celda = document.createElement("div");
    celda.className = "cal-semana";
    celda.textContent = inicial;
    celdas.push(celda);
  }
  for (let i = 0; i < columnaDelDia1; i++) {
    celdas.push(document.createElement("div"));
  }

  for (let dia = 1; dia <= diasDelMes; dia++) {
    const celda = document.createElement("div");
    celda.className = "cal-dia";
    if (esMesActual && dia === hoy.getDate()) celda.classList.add("cal-hoy");

    const numero = document.createElement("span");
    numero.textContent = dia;
    celda.append(numero);

    const vencen = vencimientosPorDia.get(dia);
    if (vencen) {
      celda.classList.add("cal-con-vencimiento");
      // los carteles de los días de los costados se alinean hacia adentro
      const columna = (columnaDelDia1 + dia - 1) % 7;
      if (columna <= 1) celda.classList.add("cal-izquierda");
      if (columna >= 5) celda.classList.add("cal-derecha");
      celda.tabIndex = 0;

      const pin = document.createElement("span");
      pin.className = "cal-pin";
      pin.setAttribute("aria-hidden", "true");

      const cartel = document.createElement("div");
      cartel.className = "cal-cartel";
      cartel.setAttribute("role", "tooltip");
      const titulo = document.createElement("strong");
      titulo.textContent = `Vence el ${dia}:`;
      cartel.append(titulo);
      for (const s of vencen) {
        const linea = document.createElement("span");
        const estado = s.estado === "Activa" ? "" : ` (${s.estado.toLowerCase()})`;
        linea.textContent = `${s.servicio_nombre} — ${formatoPesos.format(Number(s.monto))}${estado}`;
        cartel.append(linea);
      }

      celda.setAttribute("aria-label", `${dia}: vence ${vencen.map((s) => s.servicio_nombre).join(", ")}`);
      celda.append(pin, cartel);
    }
    celdas.push(celda);
  }

  contenedor.replaceChildren(...celdas);
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

  // el resumen, igual que la lista, tiene en cuenta solo lo que vence este mes
  const mes = mesActual();
  const todas = await pedir(`/suscripciones/usuario/${sesion.id_usuario}`);
  const delMes = todas.filter((s) => venceEnMes(s, mes));
  const activas = delMes.filter((s) => s.estado === "Activa").length;
  const texto = activas === 1 ? "vence 1 suscripción activa" : `vencen ${activas} suscripciones activas`;
  document.getElementById("resumen").textContent =
    `En ${nombreMes(mes)} ${texto} · ${formatoPesos.format(totalActivo(delMes))}`;
}

async function paginaSuscripciones(sesion) {
  if (!exigirSesion(sesion)) return;

  const lista = document.getElementById("lista-suscripciones");
  const total = document.getElementById("total-mensual");

  // se muestran solo las suscripciones que vencen en el mes elegido
  // (por defecto el actual; los links llevan a ?mes=AAAA-MM)
  const mes = mesDeLaURL();
  const anterior = sumarMeses(mes, -1);
  const siguiente = sumarMeses(mes, 1);

  document.getElementById("titulo-mes").textContent =
    mes === mesActual() ? `Vencimientos de este mes: ${nombreMes(mes)}` : `Vencimientos de ${nombreMes(mes)}`;
  document.getElementById("calendario-titulo").textContent = nombreMes(mes);

  const linkAnterior = document.getElementById("mes-anterior");
  linkAnterior.href = `suscripciones.html?mes=${anterior}`;
  linkAnterior.textContent = `← ${nombreMes(anterior)}`;
  const linkSiguiente = document.getElementById("mes-siguiente");
  linkSiguiente.href = `suscripciones.html?mes=${siguiente}`;
  linkSiguiente.textContent = `Ver los vencimientos de ${nombreMes(siguiente)} →`;

  // se guardan para poder armar el PUT al cambiar el estado
  let suscripciones = [];

  async function cargar() {
    const todas = await pedir(`/suscripciones/usuario/${sesion.id_usuario}`);
    suscripciones = todas.filter((s) => venceEnMes(s, mes));

    pintarCalendario(document.getElementById("calendario"), mes, suscripciones);

    // el gasto suma solo las suscripciones activas que vencen en el mes
    total.textContent = `Gasto mensual activo de ${nombreMes(mes)}: ${formatoPesos.format(totalActivo(suscripciones))}`;

    if (suscripciones.length === 0) {
      mostrarVacio(
        lista,
        todas.length === 0
          ? "Todavía no registraste ninguna suscripción."
          : `No tenés suscripciones que venzan en ${nombreMes(mes)}.`
      );
      return;
    }
    // ordenadas por fecha de vencimiento
    suscripciones.sort((a, b) => a.fecha_vencimiento.localeCompare(b.fecha_vencimiento));
    lista.replaceChildren(...suscripciones.map(itemSuscripcion));
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
  // el detalle se abre al tocar el item (o con Enter / espacio si tiene el foco)
  const detalle = prepararDetalle();
  function abrirDesde(evento) {
    if (evento.target.closest(".btn-eliminar, .select-estado")) return false;
    const item = evento.target.closest(".item-abrible");
    if (!item) return false;
    detalle.abrir(suscripciones.find((s) => String(s.id_suscripcion) === item.dataset.id));
    return true;
  }
  lista.addEventListener("keydown", (evento) => {
    if ((evento.key === "Enter" || evento.key === " ") && evento.target.matches(".item-abrible")) {
      evento.preventDefault();
      abrirDesde(evento);
    }
  });

  lista.addEventListener("click", async (evento) => {
    if (abrirDesde(evento)) return;
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
// En la lista solo se ve el servicio, la categoría y el monto; la cuenta, las
// fechas y la contraseña están en el detalle que se abre al tocar el item.
function itemSuscripcion(s) {
  const li = crearItem(
    `${s.servicio_nombre} — ${formatoPesos.format(Number(s.monto))}`,
    "Tocá para ver los detalles",
    s.id_suscripcion,
    s.servicio_categoria
  );
  li.classList.add("item-abrible");
  li.dataset.id = s.id_suscripcion;
  li.tabIndex = 0;
  li.setAttribute("aria-label", `Ver detalles de ${s.servicio_nombre}`);

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

// prepararDetalle configura el <dialog> con el detalle de una suscripción.
// La contraseña de la cuenta no viene en la lista: para verla se confirma la
// contraseña del usuario con POST /suscripciones/{id}/password.
function prepararDetalle() {
  const dialogo = document.getElementById("detalle");
  const campoPassword = document.getElementById("detalle-password");
  const botonMostrar = document.getElementById("detalle-mostrar");
  const form = document.getElementById("form-confirmar");
  const error = document.getElementById("detalle-error");
  let actual = null;

  function ocultarPassword() {
    campoPassword.textContent = "••••••••";
    campoPassword.classList.remove("revelada");
    botonMostrar.hidden = false;
    form.hidden = true;
    form.reset();
    error.hidden = true;
  }

  function abrir(sub) {
    actual = sub;
    document.getElementById("detalle-categoria").textContent = sub.servicio_categoria;
    document.getElementById("detalle-titulo").textContent = sub.servicio_nombre;
    document.getElementById("detalle-monto").textContent = `${formatoPesos.format(Number(sub.monto))} por mes`;
    document.getElementById("detalle-estado").textContent = sub.estado;
    document.getElementById("detalle-cuenta").textContent = sub.usuario_cuenta ?? "Sin cuenta";
    document.getElementById("detalle-inicio").textContent = formatearFecha(sub.fecha_inicio);
    document.getElementById("detalle-vencimiento").textContent = formatearFecha(sub.fecha_vencimiento);
    ocultarPassword();
    dialogo.showModal();
  }

  botonMostrar.addEventListener("click", () => {
    botonMostrar.hidden = true;
    form.hidden = false;
    form.elements.password.focus();
  });

  document.getElementById("cancelar-confirmar").addEventListener("click", ocultarPassword);

  form.addEventListener("submit", async (evento) => {
    evento.preventDefault();
    try {
      const respuesta = await enviarJSON(`/suscripciones/${actual.id_suscripcion}/password`, "POST", {
        password: form.elements.password.value,
      });
      form.hidden = true;
      form.reset();
      error.hidden = true;
      if (respuesta.password_cuenta) {
        campoPassword.textContent = respuesta.password_cuenta;
        campoPassword.classList.add("revelada");
      } else {
        campoPassword.textContent = "No guardaste contraseña para esta suscripción";
      }
    } catch (e) {
      error.textContent = e.message;
      error.hidden = false;
      form.elements.password.select();
    }
  });

  // se cierra con la ×, con Escape o tocando el fondo desenfocado
  document.getElementById("detalle-cerrar").addEventListener("click", () => dialogo.close());
  dialogo.addEventListener("click", (evento) => {
    if (evento.target === dialogo) dialogo.close();
  });
  // al cerrar no queda la contraseña a la vista
  dialogo.addEventListener("close", ocultarPassword);

  return { abrir };
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
