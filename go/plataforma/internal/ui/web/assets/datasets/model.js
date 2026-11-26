export const periodLabels = Object.freeze({
  '1_hora': 'Una hora',
  '24_horas': '24 horas',
});

const optionalText = value =>
  value == null
    ? null
    : typeof value === 'string' && value.trim()
      ? value
      : (() => { throw new Error('Metadatos de texto no válidos.'); })();

export function validateCatalog(records) {
  if (!Array.isArray(records)) {
    throw new Error('El catálogo recibido no es válido.');
  }

  const ids = new Set();

  return Object.freeze(records.map(record => {
    if (
      !record ||
      typeof record.id !== 'string' ||
      !record.id.trim() ||
      ids.has(record.id) ||
      !Object.hasOwn(periodLabels, record.period)
    ) {
      throw new Error('Identificador o periodo de conjunto no válido.');
    }

    ids.add(record.id);

    if (
      record.vectorCount != null &&
      (!Number.isSafeInteger(record.vectorCount) || record.vectorCount < 0)
    ) {
      throw new Error('Conteo de vectores no válido.');
    }

    if (
      record.measurementCount != null &&
      (!Number.isSafeInteger(record.measurementCount) || record.measurementCount < 0)
    ) {
      throw new Error('Conteo de mediciones no válido.');
    }

    if (
      record.sha256 != null &&
      (typeof record.sha256 !== 'string' || !/^[a-f0-9]{64}$/i.test(record.sha256))
    ) {
      throw new Error('Huella de conjunto no válida.');
    }

    if (
      record.features != null &&
      (!Array.isArray(record.features) ||
        record.features.length !== 2 ||
        record.features[0] !== 'PM10' ||
        record.features[1] !== 'PM2.5')
    ) {
      throw new Error('Componentes del conjunto no válidos.');
    }

    return Object.freeze({
      id: record.id,
      period: record.period,
      vectorCount: record.vectorCount ?? null,
      measurementCount: record.measurementCount ?? null,
      source: optionalText(record.source),
      coverage: optionalText(record.coverage),
      unit: optionalText(record.unit),
      sha256: record.sha256 ?? null,
      features: record.features == null ? null : Object.freeze([...record.features]),
    });
  }));
}

export const canSelect = dataset =>
  Boolean(dataset && Object.hasOwn(periodLabels, dataset.period) && dataset.vectorCount !== 0);