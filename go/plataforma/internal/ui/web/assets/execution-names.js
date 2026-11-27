import {validJobId} from './results/model.js';
import {sessionJobReferences} from './session-jobs.js';
export function createExecutionNames(storage) {
 const key='eas-air:execution-names:v1', names=new Map(); let next=1;
 try { const saved=JSON.parse(storage?.getItem(key) || 'null');
  if(saved?.version===1 && Number.isSafeInteger(saved.next) && saved.next>0 && Array.isArray(saved.entries) && saved.entries.length<=50) {
   const numbers=new Set();
   for(const pair of saved.entries) if(Array.isArray(pair) && validJobId(pair[0]) && Number.isSafeInteger(pair[1]) && pair[1]>0 && pair[1]<saved.next && !numbers.has(pair[1])) {names.set(pair[0],pair[1]);numbers.add(pair[1]);}
   next=saved.next;
  }
 } catch { /* Names work when session storage is unavailable. */ }
 return id=>{
  if(!validJobId(id)) return 'Ejecución';
  if(!names.has(id)) {names.set(id,next++);if(names.size>50)names.delete(names.keys().next().value);
   try {storage?.setItem(key,JSON.stringify({version:1,next,entries:[...names]}));}catch{ /* Keep the current page usable. */ }
  }
  return `Ejecución ${names.get(id)}`;
 };
}
let storage;try {storage=globalThis.sessionStorage;}catch{ /* Optional browser persistence. */ }
export const executionName=createExecutionNames(storage);
export function executionSelect(id,{selected='',optional=false}={}) {
 const select=document.createElement('select');select.id=id;select.required=!optional;
 const placeholder=document.createElement('option');placeholder.value='';placeholder.textContent=optional?'Sin segunda ejecución':'Selecciona una ejecución';select.append(placeholder);
 const choices=[...new Set([...sessionJobReferences.list(),selected].filter(validJobId))];
 for(const value of choices){const option=document.createElement('option');option.value=value;option.textContent=executionName(value);select.append(option);}
 select.value=selected;
 return select;
}
