import test from 'node:test';import assert from 'node:assert/strict';import {rememberPanelResult,getPanelResultJobId,clearPanelResult} from '../web/assets/panel/handoff.js';
test('result navigation only carries an ID from a verified completed job',()=>{
 clearPanelResult();assert.equal(getPanelResultJobId(),null);
 for(const job of [{id:'fixture',kind:'verified',state:'running'},{id:'fixture',kind:'unavailable',state:'completed'},{id:'../escape',kind:'verified',state:'completed'}])assert.equal(rememberPanelResult(job),false);
 assert.equal(rememberPanelResult({id:'fixture-completed',kind:'verified',state:'completed'}),true);assert.equal(getPanelResultJobId(),'fixture-completed');clearPanelResult();assert.equal(getPanelResultJobId(),null);
});
