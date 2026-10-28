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

// destino es la página principal de cada tipo de cuenta
function destino(usuario) {
  return usuario.tipo === "empresa" ? "empresa.html" : "cuenta.html";
}

// exigirTipo redirige al login si no hay sesión, o a su sección si la cuenta
// es de otro tipo. La API igual rechaza las acciones no permitidas (401/403).
function exigirTipo(sesion, tipo) {
  if (!sesion) {
    location.href = "login.html";
    return false;
  }
  if (sesion.tipo !== tipo) {
    location.href = destino(sesion);
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

  const items = [];
  if (sesion.tipo === "consumidor") {
    items.push(enlace("suscripciones.html", "Mis suscripciones"));
    items.push(enlace("nueva-suscripcion.html", "Nueva suscripción"));
  } else {
    items.push(enlace("empresa.html", "Mis servicios"));
  }

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

  // con sesión iniciada ya no hace falta invitar a registrar una empresa
  document.getElementById("aviso-empresas")?.remove();
}

// cargarMenuServicios completa el desplegable de la barra con el catálogo
// (solo lectura: los servicios los crean y borran las empresas en su panel)
async function cargarMenuServicios() {
  const menu = document.getElementById("menu-servicios");
  try {
    const servicios = await pedir("/servicios");
    if (servicios.length === 0) {
      mostrarVacio(menu, "Todavía no hay servicios publicados.");
      return;
    }
    menu.replaceChildren(
      ...servicios.map((s) => {
        const li = document.createElement("li");
        const nombre = document.createElement("span");
        nombre.textContent = s.nombre;
        const categoria = document.createElement("span");
        categoria.className = "etiqueta";
        categoria.textContent = s.categoria;
        li.append(nombre, categoria);
        return li;
      })
    );
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
      .replaceChildren(enlace(destino(sesion), "Ir a mi cuenta", "boton"));
  }
}

async function paginaLogin(sesion) {
  if (sesion) {
    location.href = destino(sesion);
    return;
  }

  document.getElementById("form-login").addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const datos = Object.fromEntries(new FormData(evento.target));
    try {
      const usuario = await enviarJSON("/sesion", "POST", datos);
      location.href = destino(usuario);
    } catch (error) {
      mostrarError(error);
    }
  });
}

async function paginaRegistro(sesion) {
  if (sesion) {
    location.href = destino(sesion);
    return;
  }

  // registro.html?tipo=empresa registra empresas; si no, consumidores
  const esEmpresa = new URLSearchParams(location.search).get("tipo") === "empresa";
  if (esEmpresa) {
    document.title = "Registrar empresa · Suscripciones+";
    document.getElementById("titulo-registro").textContent = "Registrar empresa";
    document.getElementById("etiqueta-nombre").textContent = "Nombre de la empresa";
    document.getElementById("ayuda-empresa").hidden = false;
  }

  document.getElementById("form-registro").addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const datos = Object.fromEntries(new FormData(evento.target));
    try {
      await enviarJSON("/usuarios", "POST", { ...datos, tipo: esEmpresa ? "empresa" : "consumidor" });
      // después de registrarse queda con la sesión iniciada
      const usuario = await enviarJSON("/sesion", "POST", { email: datos.email, password: datos.password });
      location.href = destino(usuario);
    } catch (error) {
      mostrarError(error);
    }
  });
}

// ---------------------------------------------------------------------------
// Consumidores
// ---------------------------------------------------------------------------

async function paginaCuenta(sesion) {
  if (!exigirTipo(sesion, "consumidor")) return;

  document.getElementById("saludo").textContent = `Hola, ${sesion.nombre}`;

  const suscripciones = await pedir(`/suscripciones/usuario/${sesion.id_usuario}`);
  const activas = suscripciones.filter((s) => s.estado === "Activa").length;
  const texto = activas === 1 ? "1 suscripción activa" : `${activas} suscripciones activas`;
  document.getElementById("resumen").textContent =
    `Tenés ${texto} · ${formatoPesos.format(totalActivo(suscripciones))} por mes`;
}

async function paginaSuscripciones(sesion) {
  if (!exigirTipo(sesion, "consumidor")) return;

  const lista = document.getElementById("lista-suscripciones");
  const total = document.getElementById("total-mensual");

  async function cargar() {
    const suscripciones = await pedir(`/suscripciones/usuario/${sesion.id_usuario}`);
    if (suscripciones.length === 0) {
      mostrarVacio(lista, "Todavía no registraste ninguna suscripción.");
      total.textContent = "";
      return;
    }
    lista.replaceChildren(
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
    total.textContent = `Gasto mensual activo: ${formatoPesos.format(totalActivo(suscripciones))}`;
  }

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

async function paginaNuevaSuscripcion(sesion) {
  if (!exigirTipo(sesion, "consumidor")) return;

  const select = document.getElementById("select-servicios");
  const form = document.getElementById("form-suscripcion");

  const servicios = await pedir("/servicios");
  if (servicios.length === 0) {
    select.append(new Option("Todavía no hay servicios disponibles", ""));
    form.querySelector("button").disabled = true;
  }
  for (const s of servicios) {
    select.append(new Option(`${s.nombre} (${s.categoria})`, s.id_servicio));
  }

  form.addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const datos = Object.fromEntries(new FormData(form));

    // el usuario lo toma el servidor de la sesión; la API espera el id
    // numérico y las fechas completas en formato RFC 3339
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
// Empresas
// ---------------------------------------------------------------------------

async function paginaEmpresa(sesion) {
  if (!exigirTipo(sesion, "empresa")) return;

  document.getElementById("saludo").textContent = `Hola, ${sesion.nombre}`;
  const lista = document.getElementById("lista-servicios");

  async function cargar() {
    const servicios = await pedir("/servicios");
    if (servicios.length === 0) {
      mostrarVacio(lista, "Todavía no hay servicios cargados.");
      return;
    }
    lista.replaceChildren(...servicios.map((s) => crearItem(s.nombre, "", s.id_servicio, s.categoria)));
  }

  document.getElementById("form-servicio").addEventListener("submit", async (evento) => {
    evento.preventDefault();
    const form = evento.target;
    const datos = Object.fromEntries(new FormData(form));
    try {
      await enviarJSON("/servicios", "POST", datos);
      form.reset();
      mostrarMensaje("Servicio creado.");
      // después de crear se vuelve a pedir la lista completa
      await cargar();
    } catch (error) {
      mostrarError(error);
    }
  });

  // Delegación de eventos para los botones creados dinámicamente
  lista.addEventListener("click", async (evento) => {
    const boton = evento.target.closest(".btn-eliminar");
    if (!boton) return;
    try {
      await pedir(`/servicios/${boton.dataset.id}`, { method: "DELETE" });
      mostrarMensaje("Servicio eliminado.");
      await cargar();
    } catch (error) {
      mostrarError(error);
    }
  });

  await cargar();
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
  empresa: paginaEmpresa,
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
