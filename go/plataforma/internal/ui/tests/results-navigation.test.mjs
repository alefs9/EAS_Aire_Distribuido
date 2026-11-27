import test from 'node:test';import assert from 'node:assert/strict';
import {initialResultJob,resultLocation} from '../web/assets/results/navigation.js';
test('URL identity overrides panel handoff; absent URL keeps manual and panel access',()=>{
 assert.deepEqual(initialResultJob('?job_id=url-job','panel-job'),{id:'url-job',error:null});
 assert.equal(initialResultJob('','panel-job').id,'panel-job');assert.equal(initialResultJob('').id,'');
});
test('invalid and ambiguous URL values never fall back to an older panel result',()=>{
 for(const query of ['?job_id=','?job_id=../escape','?job_id=a&job_id=b','?job_id=%20a'])assert.ok(initialResultJob(query,'older-job').error);
});
test('manual lookup keeps a reloadable identity and other query values',()=>{
 assert.equal(resultLocation('http://localhost/resultados?filter=x&job_id=old','new-job'),'/resultados?filter=x&job_id=new-job');assert.throws(()=>resultLocation('http://localhost/resultados','../other'));
});
