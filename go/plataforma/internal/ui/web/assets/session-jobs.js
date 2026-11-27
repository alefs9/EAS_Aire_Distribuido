import {validJobId} from './results/model.js';
export const MAX_SESSION_JOBS=50;
const key='eas-air:known-jobs:v1';
const normalize=ids=>[...new Set(ids.filter(validJobId))].slice(0,MAX_SESSION_JOBS);
function sessionStorageIfAvailable(){try{return typeof window==='undefined'?null:window.sessionStorage;}catch{return null;}}
// Stores references only. Neither persisted status nor results are trusted or retained.
export function createJobReferences({storage=sessionStorageIfAvailable()}={}){
 let ids=[];
 try{const raw=storage?.getItem(key);if(raw && raw.length<=16384){const saved=JSON.parse(raw);if(saved?.version===1 && Array.isArray(saved.ids))ids=normalize(saved.ids);}}catch{/* Private-mode or malformed storage starts an empty session. */}
 const persist=()=>{try{storage?.setItem(key,JSON.stringify({version:1,ids}));}catch{/* References remain usable in memory. */}};
 return Object.freeze({
  add(id){if(!validJobId(id))return false;ids=normalize([id,...ids]);persist();return true;},
  list(){return Object.freeze([...ids]);},
 });
}
export const sessionJobReferences=createJobReferences();
export function rememberAcceptedJob(status,body,references=sessionJobReferences){
 if(status!==202 || !validJobId(body?.job_id))throw new Error('La API no devolvió un trabajo aceptado con identificador válido.');
 references.add(body.job_id);return body.job_id;
}
