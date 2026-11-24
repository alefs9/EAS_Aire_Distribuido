// Owns navigation lifecycle, independent of the business screens and the DOM.
export function createViewHost({replace, loading, failure}) {
 let revision = 0;
 let controller;
 return function show(factory, route, ready = () => {}) {
  controller?.abort();
  controller = new AbortController();
  const current = ++revision;
  const context = {...route, signal: controller.signal};
  const commit = node => {if (current === revision) {replace(node); ready();}};
  const reject = error => {if (current === revision) {replace(failure(error)); ready();}};
  try {
   const result = factory(context);
   if (result && typeof result.then === 'function') {
    replace(loading());
    return Promise.resolve(result).then(commit).catch(reject);
   }
   commit(result);
  } catch (error) {reject(error);}
 };
}
