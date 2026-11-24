import test from 'node:test';
import assert from 'node:assert/strict';
import {createCatalogController} from '../web/assets/datasets/controller.js';
import {CatalogNotConfiguredError} from '../web/assets/datasets/catalog-source.js';
test('reports loading, data, empty, missing integration and HTTP failure distinctly',async()=>{
 for(const [load,expected] of [[async()=>[{id:'fixture'}],'data'],[async()=>[],'empty'],[async()=>{throw new CatalogNotConfiguredError();},'unavailable'],[async()=>{throw new Error('HTTP 503');},'error']]){
  const states=[];const c=createCatalogController({load,onState:s=>states.push(s.kind)});await c.refresh();assert.deepEqual(states,['loading',expected]);
 }
});
test('ignores stale responses and cancels reads after leaving the route',async()=>{
 const states=[];const completions=[];const signals=[];const parent=new AbortController();
 const c=createCatalogController({signal:parent.signal,onState:s=>states.push(s),load:({signal})=>{signals.push(signal);return new Promise(resolve=>completions.push(resolve));}});
 const first=c.refresh();const second=c.refresh();assert.equal(signals[0].aborted,true);
 completions[1]([{id:'latest'}]);await second;completions[0]([{id:'stale'}]);await first;
 assert.deepEqual(states.at(-1).datasets,[{id:'latest'}]);
 const last=c.refresh();parent.abort();completions[2]([{id:'late'}]);await last;
 assert.equal(signals[2].aborted,true);assert.equal(states.at(-1).kind,'loading');
});
