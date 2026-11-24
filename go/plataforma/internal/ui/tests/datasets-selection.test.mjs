import test from 'node:test';
import assert from 'node:assert/strict';
import {createSelectionStore} from '../web/assets/datasets/selection.js';
const fixtures=[{id:'fixture-hour',period:'1_hora',vectorCount:12},{id:'fixture-day',period:'24_horas',vectorCount:2}];
test('selection is exclusive, requires a verified catalog and distinguishes periods',()=>{
 const store=createSelectionStore();assert.equal(store.choose('fixture-hour'),false);store.reconcile(fixtures);
 assert.equal(store.choose('fixture-hour'),true);assert.equal(store.current().period,'1_hora');
 assert.equal(store.choose('fixture-day'),true);assert.equal(store.current().period,'24_horas');
 assert.equal(store.choose('not-present'),false);assert.equal(store.current().id,'fixture-day');
});
test('invalidation hides stale choices and changed or removed data clears selection',()=>{
 const store=createSelectionStore();store.reconcile(fixtures);store.choose('fixture-hour');store.invalidate();assert.equal(store.current(),null);
 store.reconcile(fixtures);assert.equal(store.current().id,'fixture-hour');
 store.reconcile([{...fixtures[0],sha256:'a'.repeat(64)}]);assert.equal(store.current(),null);
 store.choose('fixture-hour');store.reconcile([]);assert.equal(store.current(),null);
});
test('zero vectors cannot be selected and explicit clearing removes the selection',()=>{
 const store=createSelectionStore();store.reconcile([{...fixtures[0],vectorCount:0},fixtures[1]]);
 assert.equal(store.choose('fixture-hour'),false);store.choose('fixture-day');store.clear();assert.equal(store.current(),null);
});
