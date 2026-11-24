import test from 'node:test';
import assert from 'node:assert/strict';
import {validateCatalog,canSelect} from '../web/assets/datasets/model.js';
// Synthetic UI fixtures, never served as a real catalog.
const fixture={id:'fixture-hour',period:'1_hora',vectorCount:12};
test('keeps periods distinct and leaves missing metadata unknown',()=>{
 const records=validateCatalog([fixture,{id:'fixture-day',period:'24_horas'}]);
 assert.equal(records[0].period,'1_hora');assert.equal(records[1].period,'24_horas');
 for(const field of ['source','coverage','unit','sha256','features'])assert.equal(records[0][field],null);
 assert.equal(records[1].vectorCount,null);assert.ok(Object.isFrozen(records[0]));
});
test('rejects invalid catalog, duplicate IDs, unknown periods and inaccurate counts',()=>{
 for(const value of [null,{},[fixture,fixture],[{...fixture,period:'diario'}],[{...fixture,vectorCount:-1}],[{...fixture,vectorCount:1.5}],[{...fixture,sha256:'bad'}],[{...fixture,features:['PM2.5','PM10']}]])assert.throws(()=>validateCatalog(value));
});
test('empty catalogs are valid but zero-vector datasets cannot be selected',()=>{
 assert.deepEqual(validateCatalog([]),[]);assert.equal(canSelect(validateCatalog([{...fixture,vectorCount:0}])[0]),false);assert.equal(canSelect(validateCatalog([fixture])[0]),true);
});
