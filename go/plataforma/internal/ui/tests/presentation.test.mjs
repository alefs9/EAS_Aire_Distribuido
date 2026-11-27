import test from 'node:test';import assert from 'node:assert/strict';
import {formatDecimal,formatTolerance,calculationMethod} from '../web/assets/presentation.js';
import {createExecutionNames} from '../web/assets/execution-names.js';
test('presentation rounds concentrations without rounding the data or a small tolerance to zero',()=>{
 const concentration=15.529418124857349;assert.equal(formatDecimal(concentration),'15.53');assert.equal(concentration,15.529418124857349);assert.equal(formatDecimal(5),'5.00');assert.equal(formatDecimal(6283274.552233539),'6,283,274.55');assert.equal(formatTolerance(0.0001),'1.00e-4');assert.equal(formatTolerance(0),'0.00');
 assert.equal(calculationMethod('clúster PC4 cluster_job=abc nodos=nodo-1,nodo-2'),'K-Means distribuido');
});
test('execution names remain distinct and stable across navigation, reload and reference order',()=>{
 const values=new Map();const storage={getItem:k=>values.get(k),setItem:(k,v)=>values.set(k,v)};
 const name=createExecutionNames(storage);assert.equal(name('pc4-first'),'Ejecución 1');assert.equal(name('pc4-second'),'Ejecución 2');assert.equal(name('pc4-first'),'Ejecución 1');
 const reload=createExecutionNames(storage);assert.equal(reload('pc4-second'),'Ejecución 2');assert.equal(reload('pc4-first'),'Ejecución 1');assert.equal(reload('pc4-third'),'Ejecución 3');
});
test('execution names tolerate unavailable session storage and retain at most 50 entries',()=>{
 const saved=new Map();const store={getItem:k=>saved.get(k),setItem:(k,v)=>saved.set(k,v)};const name=createExecutionNames(store);for(let i=0;i<60;i++)name(`job-${i}`);assert.equal(JSON.parse([...saved.values()][0]).entries.length,50);
 const privateName=createExecutionNames({getItem(){throw Error('private')},setItem(){throw Error('quota')}});assert.equal(privateName('job-a'),'Ejecución 1');assert.equal(privateName('job-a'),'Ejecución 1');
});
