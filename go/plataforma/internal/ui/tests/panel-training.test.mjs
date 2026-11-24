import test from 'node:test';import assert from 'node:assert/strict';import {loadTrainingSummary,parseTraining} from '../web/assets/panel/training.js';
test('training uses segundos_ml from a matching completed result, never elapsed HTTP time',async()=>{
 const result=await loadTrainingSummary({results:[{id:'fixture-done',kind:'verified',state:'completed'}]}, {},async path=>{assert.equal(path,'/jobs/fixture-done/results');return Response.json({job_id:'fixture-done',resultado:{segundos_ml:1.25}});});assert.deepEqual(result,{id:'fixture-done',seconds:1.25});
 for(const value of [-1,null,'1',Infinity])assert.throws(()=>parseTraining({job_id:'fixture',resultado:{segundos_ml:value}},'fixture'));
 assert.throws(()=>parseTraining({job_id:'other',resultado:{segundos_ml:1}},'fixture'));
});
test('pending or failed jobs do not trigger result requests',async()=>{
 const result=await loadTrainingSummary({results:[{kind:'verified',state:'queued'},{kind:'verified',state:'failed'}]}, {},()=>{throw new Error('unexpected request');});assert.equal(result,null);
});
