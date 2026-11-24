import test from 'node:test';import assert from 'node:assert/strict';
import {parseHealth,loadHealth} from '../web/assets/panel/health.js';
test('uses actual health fields and distinguishes degradation from unavailability',()=>{
 assert.equal(parseHealth({estado:'ok',dependencias:{memoria:'ok'}},200).state,'ok');
 assert.equal(parseHealth({estado:'degradado',dependencias:{mongodb:'ok',redis:'no disponible'}},200).dependencies[1].available,false);
 assert.equal(parseHealth({estado:'no disponible',dependencias:{mongodb:'no disponible'}},503).state,'no disponible');
 for(const [payload,status] of [[{},200],[{estado:'ok',dependencias:[]},200],[{estado:'ok',dependencias:{}},503],[{estado:'no disponible',dependencias:{}},200]])assert.throws(()=>parseHealth(payload,status));
});
test('only reads /health, preserves cancellation and rejects gateway and invalid JSON responses',async()=>{
 const signal=new AbortController().signal;
 const health=await loadHealth({signal},async(path,options)=>{assert.equal(path,'/health');assert.equal(options.signal,signal);return Response.json({estado:'ok',dependencias:{memoria:'ok'}});});assert.equal(health.state,'ok');
 await assert.rejects(loadHealth({},async()=>new Response('gateway',{status:502})));
 await assert.rejects(loadHealth({},async()=>new Response('not json')));
});
