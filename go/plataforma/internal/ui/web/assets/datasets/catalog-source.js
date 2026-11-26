import {apiRequest} from '../api.js';
import {validateCatalog} from './model.js';

export class CatalogNotConfiguredError extends Error {
  constructor() {
    super('La consulta de conjuntos todavía no está conectada al servicio.');
    this.name = 'CatalogNotConfiguredError';
  }
}

function toDatasetModel(item) {
  if (
    !item ||
    typeof item !== 'object' ||
    typeof item.id !== 'string' ||
    !item.id.trim() ||
    !['1_hora', '24_horas'].includes(item.period)
  ) {
    throw new Error('Conjunto de catálogo no válido.');
  }

  const model = {id: item.id, period: item.period};
  if (item.vector_count != null) model.vectorCount = item.vector_count;
  if (item.measurement_count != null) model.measurementCount = item.measurement_count;
  if (item.features != null) model.features = item.features;

  for (const key of ['unit', 'source', 'coverage']) {
    if (item[key] != null) model[key] = item[key];
  }

  if (item.sha256 != null) model.sha256 = item.sha256;
  return model;
}

export function createCatalogLoader(request = apiRequest) {
  return async ({signal} = {}) => {
    signal?.throwIfAborted();

    const response = await request('/datasets', {signal, timeoutMs: 10000});
    if (!response.ok) {
      throw new Error(`Catálogo no disponible (HTTP ${response.status}).`);
    }

    let body;
    try {
      body = await response.json();
    } catch {
      throw new Error('La respuesta del catálogo no contiene JSON válido.');
    }

    if (!body || typeof body !== 'object' || !Array.isArray(body.datasets)) {
      throw new Error('Respuesta de catálogo no válida.');
    }

    signal?.throwIfAborted();
    return validateCatalog(body.datasets.map(toDatasetModel));
  };
}

export const fetchCatalog = createCatalogLoader();

export function createCatalogSource(nextLoader) {
  if (nextLoader != null && typeof nextLoader !== 'function') {
    throw new TypeError('loader debe ser una función.');
  }

  return async ({signal} = {}) => {
    signal?.throwIfAborted();
    if (!nextLoader) throw new CatalogNotConfiguredError();

    const records = await nextLoader({signal});
    signal?.throwIfAborted();
    return validateCatalog(records);
  };
}

let source = createCatalogSource();

export function configureCatalogLoader(nextLoader) {
  if (typeof nextLoader !== 'function') {
    throw new TypeError('loader debe ser una función.');
  }
  source = createCatalogSource(nextLoader);
}

export function loadCatalog(options) {
  return source(options);
}