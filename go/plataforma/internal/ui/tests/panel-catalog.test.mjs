import test from 'node:test';import assert from 'node:assert/strict';
import {summarizeCatalog,loadCatalogSummary,configurePanelCatalogLoader} from '../web/assets/panel/catalog-summary.js';import {PanelDependencyUnavailable} from '../web/assets/panel/dependencies.js';
test('summaries keep periods separate and never turn missing counts into zero',()=>{
 const s=summarizeCatalog([{id:'fixture-a',period:'1_hora',vectorCount:12},{id:'fixture-b',period:'24_horas'}]);assert.equal(s.count,2);assert.equal(s.groups[0].vectors,12);assert.equal(s.groups[1].vectors,null);
 assert.equal(summarizeCatalog([]).count,0);assert.throws(()=>summarizeCatalog([{id:'fixture',period:'week'}]));assert.throws(()=>summarizeCatalog([{id:'same',period:'1_hora'},{id:'same',period:'24_horas'}]));
});
test('catalog integration is optional and no endpoint is guessed',async()=>{
 await assert.rejects(loadCatalogSummary(),PanelDependencyUnavailable);
 configurePanelCatalogLoader(async()=>[]);assert.equal((await loadCatalogSummary()).count,0);
});
