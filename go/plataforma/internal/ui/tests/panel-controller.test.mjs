import test from 'node:test';import assert from 'node:assert/strict';import {createPanelController} from '../web/assets/panel/controller.js';import {PanelDependencyUnavailable} from '../web/assets/panel/dependencies.js';
test('partial failures never hide successful sections or invent training metrics',async()=>{
 const states=[];const refresh=[];
 const c=createPanelController({sources:{health:async()=>({state:'ok'}),catalog:async()=>{throw new PanelDependencyUnavailable('catalog');},jobs:async()=>{throw new Error('503');},training:()=>{throw new Error('must not request');}},onState:s=>states.push(s),onRefresh:s=>refresh.push(s),clock:()=> '2026-11-23T12:00:00Z'});
 await c.refresh();assert.equal(states.findLast(s=>s.region==='health').kind,'data');assert.equal(states.findLast(s=>s.region==='catalog').kind,'unavailable');assert.equal(states.findLast(s=>s.region==='jobs').kind,'error');assert.equal(states.findLast(s=>s.region==='training').kind,'unavailable');assert.equal(refresh.at(-1).loading,false);
});
test('overlapping refreshes abort earlier requests and discard stale updates',async()=>{
 const finish=[];const signals=[];const states=[];const parent=new AbortController();
 const c=createPanelController({signal:parent.signal,sources:{health:({signal})=>{signals.push(signal);return new Promise(resolve=>finish.push(resolve));},catalog:async()=>({count:0}),jobs:async()=>({results:[]}),training:async()=>null},onState:s=>states.push(s),onRefresh:()=>{}});
 const old=c.refresh();const latest=c.refresh();finish[1]('current');await latest;finish[0]('stale');await old;assert.equal(signals[0].aborted,true);assert.equal(states.findLast(s=>s.region==='health').value,'current');parent.abort();assert.equal(signals[1].aborted,true);
});
