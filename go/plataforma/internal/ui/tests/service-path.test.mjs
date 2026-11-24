import test from 'node:test';
import assert from 'node:assert/strict';
import {apiRequest, getAPIConfig} from '../web/assets/api.js';
test('reject malformed or escaping service paths before making requests', async () => {
 const original=globalThis.fetch;let requests=0;globalThis.fetch=()=>{requests++;throw new Error('unexpected network');};
 try {
  for(const path of [null,42,'/../health','/%2e%2e/health','//example.org','/jobs/abc%2f..','/jobs/abc%5c..','/jobs/bad id','/health?token=x','/jobs/%zz']) await assert.rejects(apiRequest(path), /Ruta/);
  assert.equal(requests,0);
 } finally {globalThis.fetch=original;}
});
test('reject invalid configuration and allow a clean retry without extra fields', async () => {
 const original=globalThis.fetch;
 const configs=[null,{}, {apiBaseURL:'https://example.org'}, {apiBaseURL:'/api/..'}, {apiBaseURL:'/api',extra:'ignored'}];
 globalThis.fetch=async()=>Response.json(configs.shift());
 try {
  for(let i=0;i<4;i++) await assert.rejects(getAPIConfig(), /Configuración no válida/);
  assert.deepEqual(await getAPIConfig(),{apiBaseURL:'/api'});
 } finally {globalThis.fetch=original;}
});
