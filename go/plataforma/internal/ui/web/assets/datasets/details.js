import {periodLabels} from './model.js';
const unknown='No informado por el servicio';
export function datasetDetails(dataset){
 return [
  ['Periodo',periodLabels[dataset.period]],
  ['Vectores',dataset.vectorCount == null ? unknown : new Intl.NumberFormat('es-PE').format(dataset.vectorCount)],
  ['Mediciones', dataset.measurementCount == null? unknown: new Intl.NumberFormat('es-PE').format(dataset.measurementCount)],
  ['Componentes',dataset.features ? dataset.features.join(' · ') : unknown],
  ['Unidades',dataset.unit ?? unknown],
  ['Procedencia',dataset.source ?? unknown],
  ['Cobertura',dataset.coverage ?? unknown],
  ['SHA-256',dataset.sha256 ?? unknown],
];
}
