import {apiRequest} from '../api.js';
import {validateCatalog} from './model.js';

export class CatalogNotConfiguredError extends Error {
  constructor() {
    super('La consulta de conjuntos todavía no está conectada al servicio.');
    this.name = 'CatalogNotConfiguredError';
  }
}

function textIfPresent(value, field) {
  if (value == null) return undefined;
  if (typeof value !== 'string' || value.trim() === '') {
    throw new Error(`Metadato ${field} no válido.`);
  }
  return value;
}

export async function fetchCatalog({signal} = {}) {
  const response = await apiRequest('/datasets', {signal, timeoutMs: 10000});
  if (!response.ok) {
    throw new Error(`No se pudo consultar el catálogo (HTTP ${response.status}).`);
  }
  let body;
  try {
    body = await response.json();
  } catch {
    throw new Error('La respuesta del catálogo no contiene JSON válido.');
  }
  if (!body || typeof body !== 'object' || !Array.isArray(body.datasets)) {
    throw new Error('La respuesta del catálogo no tiene el formato esperado.');
  }
  const mapped = body.datasets.map(item => {
    if (!item || typeof item !== 'object') throw new Error('Conjunto no válido.');
    if (typeof item.id !== 'string' || !item.id.trim()) throw new Error('ID de conjunto no válido.');
    if (!['1_hora', '24_horas'].includes(item.period)) throw new Error('Periodo no válido.');
    const model = {id: item.id, period: item.period};
    if (item.vector_count != null) model.vectorCount = item.vector_count;
    if (item.measurement_count != null) model.measurementCount = item.measurement_count;
    if (item.features != null) model.features = item.features;
    for (const key of ['unit', 'source', 'coverage']) {
      const value = textIfPresent(item[key], key);
      if (value !== undefined) model[key] = value;
    }
    if (item.sha256 != null) model.sha256 = item.sha256;
    return model;
  });
  return mapped;
}

export function createCatalogSource(nextLoader) {
  if (nextLoader != null && typeof nextLoader !== 'function') {
    throw new TypeError('loader debe ser una función.');
  }
  return async ({signal} = {}) => {
    signal?.throwIfAborted();
    if (!nextLoader) throw new CatalogNotConfiguredError();
    const records=await nextLoader({signal});
    signal?.throwIfAborted();
    return validateCatalog(records);
  };
}

let source = createCatalogSource();
export function configureCatalogLoader(nextLoader) {
  if (typeof nextLoader !== 'function') throw new TypeError('loader debe ser una función.');
  source = createCatalogSource(nextLoader);
}
export function loadCatalog(options) {
  return source(options);
}