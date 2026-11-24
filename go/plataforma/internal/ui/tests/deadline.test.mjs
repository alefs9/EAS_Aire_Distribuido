import test from 'node:test';
import assert from 'node:assert/strict';
import {apiRequest} from '../web/assets/api.js';
test('slow requests expire and cancelled views do not start a request', async () => {
 const original = globalThis.fetch;
 let requests = 0;
 globalThis.fetch = async (url, {signal}) => {
  if(url === '/ui-config.json') return Response.json({apiBaseURL:'/api'});
  requests++;
  return new Promise((resolve,reject) => signal.addEventListener('abort',()=>reject(signal.reason),{once:true}));
 };
 try {
  await assert.rejects(apiRequest('/health',{timeoutMs:5}), {name:'TimeoutError'});
  const controller = new AbortController(); controller.abort();
  await assert.rejects(apiRequest('/health',{signal:controller.signal}), {name:'AbortError'});
  assert.equal(requests,1);
  await assert.rejects(apiRequest('/health',{timeoutMs:0}), /Plazo/);
 } finally {globalThis.fetch=original;}
});
