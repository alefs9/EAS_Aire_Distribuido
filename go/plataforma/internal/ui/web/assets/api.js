// All views share this configuration. No API request is made on the home page.
let configuration;
async function withDeadline(task, signal, timeoutMs) {
  if (!Number.isFinite(timeoutMs) || timeoutMs <= 0) throw new Error('Plazo de consulta no válido.');
  const controller = new AbortController();
  const cancel = () => controller.abort(signal.reason);
  if (signal?.aborted) cancel(); else signal?.addEventListener('abort', cancel, {once: true});
  const timer = setTimeout(() => controller.abort(new DOMException('El servicio tardó demasiado en responder.', 'TimeoutError')), timeoutMs);
  try {controller.signal.throwIfAborted(); return await task(controller.signal);}
  finally {clearTimeout(timer); signal?.removeEventListener('abort', cancel);}
}
export async function getAPIConfig() {
  if (!configuration) {
    configuration = withDeadline(async signal => {
      const response = await fetch('/ui-config.json', {cache: 'no-store', signal});
      if (!response.ok) throw new Error('No se pudo cargar la configuración.');
      const config = await response.json();
      if (!config || typeof config !== 'object' || config.apiBaseURL !== '/api') throw new Error('Configuración no válida.');
      return Object.freeze({apiBaseURL: config.apiBaseURL});
    }, undefined, 5000)
      .catch(error => {configuration = undefined; throw error;});
  }
  return configuration;
}
function validateServicePath(path) {
  let decoded;
  try {decoded = decodeURIComponent(path);} catch {throw new Error('Ruta de servicio no válida.');}
  if (typeof path !== 'string' || !path.startsWith('/') || decoded.includes('//') || /[\\\s?#\x00-\x1f]/.test(decoded) || /%2f|%5c/i.test(path) || decoded.split('/').some(part => part === '.' || part === '..')) throw new Error('Ruta de servicio no válida.');
}
// Response formats are owned by the existing API; views handle their own schema.
export async function apiRequest(path, options = {}) {
  validateServicePath(path);
  const {timeoutMs = 15000, signal, ...request} = options;
  signal?.throwIfAborted();
  const {apiBaseURL} = await getAPIConfig();
  return withDeadline(combinedSignal => fetch(`${apiBaseURL}${path}`, {...request, cache: 'no-store', signal: combinedSignal}), signal, timeoutMs);
}
