// All views share this configuration. No API request is made on the home page.
let configuration;
export async function getAPIConfig() {
  if (!configuration) {
    configuration = fetch('/ui-config.json', {cache: 'no-store'})
      .then(response => {if (!response.ok) throw new Error('No se pudo cargar la configuración.'); return response.json();})
      .then(config => {if (config.apiBaseURL !== '/api') throw new Error('Configuración no válida.'); return Object.freeze(config);})
      .catch(error => {configuration = undefined; throw error;});
  }
  return configuration;
}
// Response formats are owned by the existing API; views handle their own schema.
export async function apiRequest(path, options = {}) {
  if (!path.startsWith('/') || path.startsWith('//') || /[?#]/.test(path) || path.split('/').includes('..')) throw new Error('Ruta de servicio no válida.');
  const {apiBaseURL} = await getAPIConfig();
  return fetch(`${apiBaseURL}${path}`, {...options, cache: 'no-store'});
}
