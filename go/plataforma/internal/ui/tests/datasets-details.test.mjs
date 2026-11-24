import test from 'node:test';
import assert from 'node:assert/strict';
import {validateCatalog} from '../web/assets/datasets/model.js';
import {datasetDetails} from '../web/assets/datasets/details.js';
test('metadata presentation preserves reported values and identifies unknown fields',()=>{
 const dataset=validateCatalog([{id:'fixture-hour',period:'1_hora',vectorCount:12,unit:'µg/m³',source:'Fuente de prueba',coverage:'Cobertura de prueba'}])[0];
 const details=Object.fromEntries(datasetDetails(dataset));
 assert.equal(details.Periodo,'Una hora');assert.equal(details.Vectores,'12');assert.equal(details.Unidades,'µg/m³');assert.equal(details.Procedencia,'Fuente de prueba');assert.equal(details.Cobertura,'Cobertura de prueba');assert.equal(details.Componentes,'No informado por el servicio');assert.equal(details['SHA-256'],'No informado por el servicio');
});
