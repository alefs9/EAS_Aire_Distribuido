import test from 'node:test';
import assert from 'node:assert/strict';
import {routes, resolveRoute, shouldNavigate} from '../web/assets/routes.js';
import {apiRequest, getAPIConfig} from '../web/assets/api.js';

test('all five routes resolve, trailing slash and unknown path are handled', () => {
  assert.equal(routes.length, 5);
  for (const route of routes) assert.equal(resolveRoute(route.path + (route.path === '/' ? '' : '/')).title, route.title);
  assert.equal(resolveRoute('/no-existe').title, 'Página no encontrada');
});
test('SPA navigation preserves browser modified clicks and external links', () => {
  const origin = 'http://localhost:8090';
  const anchor = (href, extra = {}) => ({href, target: '', hasAttribute: () => false, ...extra});
  const normal = {button: 0};
  assert.equal(shouldNavigate(normal, anchor(origin + '/datos'), origin), true);
  for (const modifier of ['ctrlKey', 'metaKey', 'shiftKey', 'altKey']) assert.equal(shouldNavigate({...normal, [modifier]: true}, anchor(origin + '/datos'), origin), false);
  for (const link of [anchor('https://example.org'), anchor(origin + '/#recorrido'), anchor(origin + '/datos', {target: '_blank'}), anchor(origin + '/datos', {hasAttribute: () => true})]) assert.equal(shouldNavigate(normal, link, origin), false);
  assert.equal(shouldNavigate({button: 1}, anchor(origin + '/datos'), origin), false);
  assert.equal(shouldNavigate(normal, null, origin), false);
});
test('API configuration retries errors, is shared, and preserves real HTTP errors', async () => {
  const originalFetch = globalThis.fetch;
  const calls = [];
  let fail = true;
  globalThis.fetch = async (url, options) => {
    calls.push({url, options});
    if (url === '/ui-config.json') {
      if (fail) return new Response('', {status: 503});
      return Response.json({apiBaseURL: '/api'});
    }
    return Response.json({codigo: 'no_encontrado'}, {status: 404});
  };
  try {
    await assert.rejects(getAPIConfig(), /configuración/);
    fail = false;
    const [one, two] = await Promise.all([getAPIConfig(), getAPIConfig()]);
    assert.equal(one, two);
    assert.equal(calls.filter(c => c.url === '/ui-config.json').length, 2);
    const response = await apiRequest('/jobs/ausente');
    assert.equal(response.status, 404);
    assert.deepEqual(await response.json(), {codigo: 'no_encontrado'});
    assert.equal(calls.at(-1).url, '/api/jobs/ausente');
    assert.equal(calls.at(-1).options.cache, 'no-store');
    for (const path of ['https://example.org', '//example.org', '/../health']) await assert.rejects(apiRequest(path), /Ruta/);
  } finally {globalThis.fetch = originalFetch;}
});
