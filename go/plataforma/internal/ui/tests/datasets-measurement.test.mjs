import test from 'node:test';
import assert from 'node:assert/strict';
import {datasetDetails} from '../web/assets/datasets/details.js';

test('muestra vectores y mediciones como conteos distintos', () => {
  const details = datasetDetails({
    period: '1_hora',
    vectorCount: 445347,
    measurementCount: 890694,
  });

  const vectors = details.find(([label]) => label === 'Vectores');
  const measurements = details.find(([label]) => label === 'Mediciones');

  assert.equal(vectors[1], new Intl.NumberFormat('es-PE').format(445347));
  assert.equal(measurements[1], new Intl.NumberFormat('es-PE').format(890694));
  assert.notEqual(vectors[1], measurements[1]);
});