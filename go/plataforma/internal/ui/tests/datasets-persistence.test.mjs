import test from 'node:test';
import assert from 'node:assert/strict';
import {createSelectionStore} from '../web/assets/datasets/selection.js';
const fixture={id:'fixture-day',period:'24_horas',vectorCount:2,source:'not stored'};
const memory=()=>{const values=new Map();return {getItem:k=>values.get(k),setItem:(k,v)=>values.set(k,v),removeItem:k=>values.delete(k),values};};
test('restored choice stays untrusted until its exact identity is revalidated',()=>{
 const storage=memory();const first=createSelectionStore({storage});first.reconcile([fixture]);first.choose(fixture.id);
 const saved=JSON.parse([...storage.values.values()][0]);assert.deepEqual(Object.keys(saved).sort(),['id','period','sha256','version']);
 const next=createSelectionStore({storage});assert.equal(next.current(),null);next.reconcile([fixture]);assert.equal(next.current().id,fixture.id);
 next.reconcile([]);assert.equal(storage.values.size,0);
});
test('malformed storage and quota or privacy failures do not break selection',()=>{
 const storage={getItem:()=>'{broken',setItem:()=>{throw new Error('quota');},removeItem:()=>{throw new Error('privacy');}};
 const store=createSelectionStore({storage});assert.equal(store.current(),null);store.reconcile([fixture]);assert.equal(store.choose(fixture.id),true);store.clear();assert.equal(store.current(),null);
});
