import test from 'node:test';
import assert from 'node:assert/strict';
import {createCatalogSource,CatalogNotConfiguredError} from '../web/assets/datasets/catalog-source.js';
test('missing integration is distinct from a real empty catalog and makes no network request',async()=>{
 await assert.rejects(createCatalogSource()(),CatalogNotConfiguredError);
 assert.deepEqual(await createCatalogSource(async()=>[])(),[]);
});
test('passes cancellation, validates responses and preserves service failures',async()=>{
 const controller=new AbortController();let received;
 const source=createCatalogSource(async({signal})=>{received=signal;return [{id:'fixture-hour',period:'1_hora'}];});
 assert.equal((await source({signal:controller.signal}))[0].id,'fixture-hour');assert.equal(received,controller.signal);
 controller.abort();await assert.rejects(source({signal:controller.signal}),{name:'AbortError'});
 await assert.rejects(createCatalogSource(async()=>({}))(),/catálogo/);
 const failure=new Error('service unavailable');await assert.rejects(createCatalogSource(async()=>{throw failure;})(),error=>error===failure);
});
