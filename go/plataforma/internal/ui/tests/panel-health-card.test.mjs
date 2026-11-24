import test from 'node:test';import assert from 'node:assert/strict';
import {parseHealth} from '../web/assets/panel/health.js';import {healthPresentation} from '../web/assets/panel/health-card.js';
test('health labels are explicit and do not expose backend failure strings',()=>{
 const model=healthPresentation(parseHealth({estado:'degradado',dependencias:{mongodb:'ok',redis:'sensitive backend error'}},200));
 assert.equal(model.title,'Disponible con limitaciones');assert.equal(model.dependencies[1].status,'No disponible');assert.equal(JSON.stringify(model).includes('sensitive'),false);
});
