// Shared route catalog. Add a view here without changing the shell.
export const routes = Object.freeze([
  {path: '/', title: 'Inicio'},
  {path: '/panel', title: 'Panel general'},
  {path: '/datos', title: 'Datos'},
  {path: '/ejecuciones', title: 'Ejecuciones'},
  {path: '/resultados', title: 'Resultados'},
  {path: '/analisis-ambiental', title: 'Análisis ambiental'},
]);
export function resolveRoute(pathname) {
  const path = pathname.replace(/\/$/, '') || '/';
  return routes.find(route => route.path === path) || {path, title: 'Página no encontrada'};
}
export function shouldNavigate(event, anchor, origin) {
  if (!anchor || event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey || anchor.target || anchor.hasAttribute('download')) return false;
  const url = new URL(anchor.href, origin);
  return url.origin === origin && !url.hash;
}
