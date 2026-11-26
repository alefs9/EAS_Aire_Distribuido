import test from 'node:test';
import assert from 'node:assert/strict';
import {parseResult,finishReason} from '../web/assets/results/model.js';
import {provenance} from '../web/assets/results/provenance.js';
import {loadResult} from '../web/assets/results/service.js';
const payload=()=>({job_id:'job-1',dataset_id:'unknown',parametros:{k:2,semilla:42,max_iter:100,tolerancia:.0001},resultado:{centroides:[[10,5],[20,8]],conteos:[3,5],iteraciones:4,inercia:12,convergio:true,origen:'tcp'}});
test('conserva valores y diferencia convergencia del limite',()=>{
 const p=payload(),m=parseResult(p,'job-1');assert.equal(m.total,8);assert.deepEqual(m.result.centroides,p.resultado.centroides);assert.match(finishReason(m),/Convergencia/);
 m.result.convergio=false;m.result.iteraciones=100;assert.match(finishReason(m),/Máximo/);m.result.iteraciones=3;assert.match(finishReason(m),/no informado/);
});
test('rechaza datos incompatibles y otro identificador',()=>{
 assert.throws(()=>parseResult(payload(),'otro'));
 for(const mutate of [p=>p.resultado.conteos.pop(),p=>p.resultado.conteos[0]=-1,p=>p.resultado.centroides[0][0]=NaN,p=>p.resultado.iteraciones=101,p=>p.parametros.k=0]){const p=payload();mutate(p);assert.throws(()=>parseResult(p,'job-1'));}
});
test('no atribuye unidades a un conjunto desconocido y verifica cobertura conocida',()=>{
 const m=parseResult(payload(),'job-1');assert.match(provenance(m).unit,/no informada/);
 m.datasetId='5866449404c20a23e52b65f94871a365c7979b1fcf0abd7ed91554d70ea609a6:24_horas';assert.throws(()=>provenance(m));m.total=18297;assert.equal(provenance(m).unit,'µg/m³');
});
test('consulta resultados completos y distingue pendiente fallo y ausencia',async()=>{
 for(const [status,body,kind] of [[200,payload(),'completed'],[409,{job_id:'job-1',estado:'running'},'pending'],[422,{job_id:'job-1',error:'Sin nodos'},'failed'],[404,{},'missing']]) {
  const result=await loadResult('job-1',{request:async path=>{assert.equal(path,'/jobs/job-1/results');return Response.json(body,{status});}});assert.equal(result.kind,kind);
 }
 await assert.rejects(loadResult('../otro'));await assert.rejects(loadResult('job-1',{request:async()=>Response.json({},{status:503})}));
});
test('cancela consultas y no acepta cuerpos tardios',async()=>{
 const controller=new AbortController();controller.abort();
 await assert.rejects(loadResult('job-1',{signal:controller.signal,request:async()=>Response.json(payload())}));
 await assert.rejects(loadResult('job-1',{timeoutMs:1,request:async()=>({status:200,json:async()=>{await new Promise(r=>setTimeout(r,20));return payload();}})}));
});
