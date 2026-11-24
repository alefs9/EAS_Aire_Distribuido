import test from 'node:test';
import assert from 'node:assert/strict';
import {createSelectionStore} from '../web/assets/datasets/selection.js';
import {executionChoice} from '../web/assets/datasets/handoff.js';
test('handoff exposes identity and period only after verification, never job parameters',()=>{
 const store=createSelectionStore({storage:null});assert.equal(executionChoice(store),null);
 store.reconcile([{id:'fixture-hour',period:'1_hora',vectorCount:12}]);store.choose('fixture-hour');
 const choice=executionChoice(store);assert.deepEqual(choice,{datasetId:'fixture-hour',period:'1_hora'});assert.ok(Object.isFrozen(choice));
 store.invalidate();assert.equal(executionChoice(store),null);
});
