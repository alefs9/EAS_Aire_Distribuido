import test from 'node:test';import assert from 'node:assert/strict';
import {parseProfile} from '../web/assets/environment/model.js';
import {selectProfiles} from '../web/assets/environment/filters.js';
import {loadProfiles} from '../web/assets/environment/service.js';
const payload=(id='a')=>({job_id:id,dataset_id:'5866449404c20a23e52b65f94871a365c7979b1fcf0abd7ed91554d70ea609a6:24_horas',parametros:{k:2,max_iter:100},resultado:{centroides:[[10,5],[30,15]],conteos:[10000,8297],iteraciones:4,convergio:true}});
test('filtra grupo y contaminante sin cambiar valores ni denominador',()=>{
 const p=parseProfile(payload(),'a'),before=JSON.stringify(p);const [filtered]=selectProfiles([p],{period:'24_horas',pollutant:'pm25',group:'2'});
 assert.deepEqual(filtered.groups,[{group:2,pm10:30,pm25:15,count:8297}]);assert.deepEqual(filtered.pollutants,['pm25']);assert.equal(filtered.dataset.total,18297);assert.equal(JSON.stringify(p),before);
 assert.equal(selectProfiles([p],{period:'24_horas',group:'3'})[0].groups.length,0);
});
test('rechaza mezcla de periodos y filtros invalidos',()=>{
 const p=parseProfile(payload(),'a');assert.throws(()=>selectProfiles([p],{period:'1_hora'}));assert.throws(()=>selectProfiles([p],{period:'24_horas',pollutant:'ozono'}));
});
test('rechaza identidad unidades conteos y centroides no verificables',()=>{
 for(const change of [p=>p.job_id='b',p=>p.dataset_id='otro:24_horas',p=>p.resultado.conteos[0]=1,p=>p.resultado.centroides[0][0]=NaN,p=>p.resultado.centroides.pop()]){const p=payload();change(p);assert.throws(()=>parseProfile(p,'a'));}
});
test('consulta dos trabajos sin duplicados y conserva sus grupos independientes',async()=>{
 const result=await loadProfiles(['a','b'],{request:async path=>Response.json(payload(path.split('/')[2]))});assert.deepEqual(result.map(p=>p.id),['a','b']);await assert.rejects(loadProfiles(['a','a']));await assert.rejects(loadProfiles(['../a']));
});
test('no ofrece una comparacion parcial cuando falla una consulta',async()=>{
 for(const status of [404,409,422,503])await assert.rejects(loadProfiles(['a','b'],{request:async path=>path.includes('/b/')?Response.json({},{status}):Response.json(payload())}));
});
test('aborta consultas canceladas y cuerpos que exceden el plazo',async()=>{
 const c=new AbortController();c.abort();await assert.rejects(loadProfiles(['a'],{signal:c.signal,request:()=>{throw Error('No debe consultar');}}));
 await assert.rejects(loadProfiles(['a'],{timeoutMs:1,request:async()=>({ok:true,json:async()=>{await new Promise(r=>setTimeout(r,20));return payload();}})}));
});
