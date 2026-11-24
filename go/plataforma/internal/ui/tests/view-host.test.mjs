import test from 'node:test';
import assert from 'node:assert/strict';
import {createViewHost} from '../web/assets/view-host.js';
test('late view cannot replace the current route and receives cancellation', async () => {
 const shown=[]; let finish, oldSignal;
 const show=createViewHost({replace:x=>shown.push(x), loading:()=> 'loading', failure:()=> 'error'});
 const pending=show(({signal})=>{oldSignal=signal;return new Promise(resolve=>finish=resolve);}, {path:'/primera'});
 assert.deepEqual(shown,['loading']);
 show(()=> 'current', {path:'/segunda'});
 assert.equal(oldSignal.aborted,true);
 finish('stale'); await pending;
 assert.deepEqual(shown,['loading','current']);
});
test('sync and async failures preserve the host and allow retry', async () => {
 const shown=[];
 const show=createViewHost({replace:x=>shown.push(x), loading:()=> 'loading', failure:()=> 'error'});
 show(()=> {throw new Error('failure');},{});
 await show(()=> Promise.reject(new Error('failure')),{});
 show(()=> 'recovered',{});
 assert.deepEqual(shown,['error','loading','error','recovered']);
});
