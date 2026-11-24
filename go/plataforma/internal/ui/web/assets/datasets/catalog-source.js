import {validateCatalog} from './model.js';
export class CatalogNotConfiguredError extends Error {
 constructor(){super('La consulta de conjuntos todavía no está conectada al servicio.');this.name='CatalogNotConfiguredError';}
}
// An integration seam only: no endpoint, response shape or fallback data is assumed.
export function createCatalogSource(loader) {
 if (loader != null && typeof loader !== 'function') throw new TypeError('La consulta debe ser una función.');
 return async ({signal} = {}) => {
  signal?.throwIfAborted();
  if (!loader) throw new CatalogNotConfiguredError();
  const records=await loader({signal});
  signal?.throwIfAborted();
  return validateCatalog(records);
 };
}
let source=createCatalogSource();
export function configureCatalogLoader(loader){source=createCatalogSource(loader);}
export const loadCatalog = options => source(options);
