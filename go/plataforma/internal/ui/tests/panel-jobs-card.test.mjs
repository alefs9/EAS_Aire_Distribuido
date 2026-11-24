import test from 'node:test';import assert from 'node:assert/strict';import {summarizeJobs} from '../web/assets/panel/jobs-card.js';
test('counts only verified states and retains the limited scope of the query',()=>{
 const summary=summarizeJobs({totalReferenced:12,results:[{kind:'verified',state:'queued'},{kind:'verified',state:'running'},{kind:'verified',state:'completed'},{kind:'verified',state:'failed'},{kind:'not-found'},{kind:'unavailable'}]});
 assert.deepEqual(summary.counts,{queued:1,running:1,completed:1,failed:1});assert.equal(summary.verified,4);assert.equal(summary.queried,6);assert.equal(summary.total,12);
});
