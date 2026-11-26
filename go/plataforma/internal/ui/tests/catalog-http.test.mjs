import test from 'node:test';
import assert from 'node:assert/strict';
import {createCatalogLoader} from '../web/assets/datasets/catalog-source.js';

const response=(status,body)=>({ok:status>=200&&status<300,status,json:async()=>body});
test('mapea catálogo real HTTP a modelo y propaga signal',async()=>{
  let seen;
  const load=createCatalogLoader(async(path,options)=>{
    assert.equal(path,'/datasets'); seen=options.signal;
    return response(200,{datasets:[{id:'a'.repeat(64)+':1_hora',period:'1_hora',vector_count:445347,measurement_count:890694,features:['PM10','PM2.5'],unit:'µg/m³',coverage:'cobertura',sha256:'a'.repeat(64)}]});
  });
  const c=new AbortController(); const records=await load({signal:c.signal});
  assert.equal(seen,c.signal);
  assert.equal(records[0].vectorCount,445347);
  assert.equal(records[0].period,'1_hora');
  assert.equal(records[0].measurementCount, 890694);
});

test('solo respuesta 200 con lista vacía confirma catálogo vacío',async()=>{
  const records=await createCatalogLoader(async()=>response(200,{datasets:[]}))();
  assert.deepEqual(records,[]);
});

test('HTTP error y cuerpo malformado propagan error, no vacío',async()=>{
  await assert.rejects(createCatalogLoader(async()=>response(503,{error:'down'}))(),/503/);
  await assert.rejects(createCatalogLoader(async()=>response(200,{wrong:[]}))(),/no válida/);
});

test('cancelación corta consulta y reintento vuelve a llamar al servicio',async()=>{
  const c=new AbortController(); let calls=0;
  const load=createCatalogLoader(async(_path,{signal})=>{
    calls++;
    if(calls===1){c.abort();signal.throwIfAborted();}
    return response(200,{datasets:[]});
  });
  await assert.rejects(load({signal:c.signal}),{name:'AbortError'});
  let retries=0;
  const retry=createCatalogLoader(async()=>++retries===1?response(503,{error:'down'}):response(200,{datasets:[]}));
  await assert.rejects(retry(),/503/);
  assert.deepEqual(await retry(),[]); assert.equal(retries,2);
});


