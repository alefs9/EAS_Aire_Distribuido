import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {parseResult} from '../web/assets/results/model.js';
import {provenance} from '../web/assets/results/provenance.js';
test('las respuestas reales conservan conteos parametros y periodos',()=>{
 for(const [period,total,iterations,converged] of [['1_hora',445347,100,false],['24_horas',18297,44,true]]){
  const payload=JSON.parse(readFileSync(new URL(`../evidencias/resultados/${period}.json`,import.meta.url),'utf8').replace(/^\uFEFF/,''));
  const model=parseResult(payload,payload.job_id);
  assert.equal(model.total,total);assert.equal(model.result.iteraciones,iterations);assert.equal(model.result.convergio,converged);
  assert.deepEqual(model.parameters,{k:4,semilla:42,tolerancia:.0001,max_iter:100});assert.equal(provenance(model).unit,'µg/m³');
 }
});
