import test from 'node:test';import assert from 'node:assert/strict';
import {createJobReferences,rememberAcceptedJob,MAX_SESSION_JOBS} from '../web/assets/session-jobs.js';
const memory=()=>{const values=new Map();return {getItem:k=>values.get(k),setItem:(k,v)=>values.set(k,v),values};};
test('only accepted valid jobs are remembered, deduplicated and bounded',()=>{
 const store=createJobReferences({storage:null});assert.throws(()=>rememberAcceptedJob(503,{job_id:'job-x'},store));assert.throws(()=>rememberAcceptedJob(200,{job_id:'job-x'},store));assert.throws(()=>rememberAcceptedJob(202,{job_id:'../bad'},store));assert.deepEqual(store.list(),[]);
 for(let i=0;i<70;i++)rememberAcceptedJob(202,{job_id:`job-${i}`},store);store.add('job-69');assert.equal(store.list().length,MAX_SESSION_JOBS);assert.equal(store.list()[0],'job-69');assert.equal(new Set(store.list()).size,MAX_SESSION_JOBS);assert.ok(Object.isFrozen(store.list()));
});
test('reload retains IDs only; local status and malformed references are not trusted',()=>{
 const storage=memory();const store=createJobReferences({storage});store.add('job-real');const raw=JSON.parse([...storage.values.values()][0]);assert.deepEqual(raw,{version:1,ids:['job-real']});assert.deepEqual(createJobReferences({storage}).list(),['job-real']);
 const untrusted={getItem:()=>JSON.stringify({version:1,ids:['job-valid','../escape','job-valid',null],estado:'completed'})};assert.deepEqual(createJobReferences({storage:untrusted}).list(),['job-valid']);
 assert.deepEqual(createJobReferences({storage:{getItem:()=>'{bad'}}).list(),[]);
});
test('quota or privacy failures preserve in-memory references without saving results',()=>{
 const store=createJobReferences({storage:{getItem:()=>{throw new Error('private');},setItem:()=>{throw new Error('quota');}}});store.add('job-1');assert.deepEqual(store.list(),['job-1']);
});
