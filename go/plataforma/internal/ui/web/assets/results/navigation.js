import {validJobId} from './model.js';
// An explicit URL always takes precedence, including invalid or duplicated values.
export function initialResultJob(search,panelId=null){
 const values=new URLSearchParams(search).getAll('job_id');
 if(values.length){
  if(values.length!==1 || !validJobId(values[0]))return {id:values.length===1?values[0]:'',error:'No pudimos abrir esta ejecución. Comprueba el enlace.'};
  return {id:values[0],error:null};
 }
 return {id:validJobId(panelId)?panelId:'',error:null};
}
export function resultLocation(href,id){
 if(!validJobId(id))throw new Error('Identificador no válido.');
 const url=new URL(href);url.searchParams.set('job_id',id);return url.pathname+url.search+url.hash;
}
