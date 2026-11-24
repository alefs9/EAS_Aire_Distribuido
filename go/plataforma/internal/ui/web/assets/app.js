import {resolveRoute, shouldNavigate} from './routes.js';
import {emptyState, loadingState, errorState} from './states.js';
import {createViewHost} from './view-host.js';
import {bindMenu} from './menu.js';
import {homeView} from './views/home.js';

// Screen modules register render functions here; keep the shell shared.
const views = new Map([['/', homeView]]);
export function registerView(path, renderView) {views.set(path, renderView);}
const descriptions = {
 '/datos': ['Consulta de datos en preparación', 'Aquí podrás consultar y seleccionar conjuntos de mediciones. Esta función todavía no está disponible.'],
 '/ejecuciones': ['Ejecuciones en preparación', 'Aquí podrás configurar un trabajo y seguir su avance. Esta función todavía no está disponible.'],
 '/resultados': ['Resultados en preparación', 'Aquí podrás consultar los resultados de tus trabajos. Esta función todavía no está disponible.'],
 '/analisis-ambiental': ['Análisis ambiental en preparación', 'Aquí podrás explorar la interpretación ambiental de los resultados. Esta función todavía no está disponible.'],
};
function placeholderView(route) {
 const section = document.createElement('section'); section.className = 'page-section';
 const label = document.createElement('p'); label.className = 'eyebrow'; label.textContent = 'Tu espacio de análisis';
 const heading = document.createElement('h1'); heading.id = 'titulo-vista'; heading.tabIndex = -1; heading.textContent = route.title;
 const [title, description] = descriptions[route.path] || ['Esta página no existe', 'Utiliza el menú para volver a una sección disponible.'];
 const back = document.createElement('a'); back.href = '/'; back.dataset.route = ''; back.className = 'button button-secondary'; back.textContent = 'Volver al inicio';
 section.append(label, heading, emptyState(title, description), back); return section;
}
const content = document.querySelector('#contenido');
const closeMenu = bindMenu(document.querySelector('.menu-toggle'), document.querySelector('#navegacion'));
const showView = createViewHost({
 replace: node => content.replaceChildren(node),
 loading: () => loadingState(),
 failure: () => errorState('No se pudo mostrar esta sección. Intenta nuevamente.', () => render(true)),
});
function render(moveFocus = false) {
 const route = resolveRoute(location.pathname);
 document.title = `EAS Aire · ${route.title}`;

 document.querySelectorAll('nav [data-route]').forEach(link => {if(link.getAttribute('href') === route.path) link.setAttribute('aria-current', 'page'); else link.removeAttribute('aria-current');});
 closeMenu();
 showView(views.get(route.path) || placeholderView, route, () => {
  if (moveFocus) {(document.querySelector('#titulo-vista') || content).focus(); window.scrollTo(0, 0);}
 });
}
document.addEventListener('click', event => {
 const anchor = event.target.closest('a[data-route]');
 if (!shouldNavigate(event, anchor, location.origin)) return;
 event.preventDefault();
 const url = new URL(anchor.href);
 if (url.pathname !== location.pathname) history.pushState(null, '', url.pathname + url.search);
 render(true);
});
window.addEventListener('popstate', () => render(true));
render();
