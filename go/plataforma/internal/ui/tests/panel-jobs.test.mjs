import test from 'node:test';import assert from 'node:assert/strict';
import {loadJobOverview,parseJob} from '../web/assets/panel/jobs.js';import {PanelDependencyUnavailable} from '../web/assets/panel/dependencies.js';
test('requires known IDs, distinguishes an empty session and validates returned identity',async()=>{
 await assert.rejects(loadJobOverview(),PanelDependencyUnavailable);
 assert.deepEqual(await loadJobOverview({}, {references:async()=>[],request:()=>{throw new Error('must not query');}}),{totalReferenced:0,results:[]});
 for(const state of ['queued','running','completed','failed'])assert.equal(parseJob({trabajo:{job_id:'fixture',estado:state}},'fixture').state,state);
 assert.throws(()=>parseJob({trabajo:{job_id:'other',estado:'running'}},'fixture'));assert.throws(()=>parseJob({trabajo:{job_id:'fixture',estado:'unknown'}},'fixture'));
});
test('bounds concurrency, never posts jobs, and keeps missing or failed reads distinct',async()=>{
 let active=0,max=0;const paths=[];
 const overview=await loadJobOverview({}, {references:async()=>Array.from({length:12},(_,i)=>`fixture-${i}`),request:async(path,options)=>{
  paths.push(path);assert.equal(options.method,undefined);max=Math.max(max,++active);await new Promise(resolve=>setTimeout(resolve,1));active--;
  if(path.endsWith('-0'))return new Response('',{status:404});if(path.endsWith('-1'))return new Response('',{status:503});
  return Response.json({trabajo:{job_id:path.split('/').at(-1),estado:'queued'}});
 }});
 assert.equal(paths.length,10);assert.equal(overview.totalReferenced,12);assert.ok(max<=3);assert.equal(overview.results[0].kind,'not-found');assert.equal(overview.results[1].kind,'unavailable');assert.equal(overview.results[2].state,'queued');
});
