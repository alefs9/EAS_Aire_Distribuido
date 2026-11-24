import {apiRequest} from '../api.js';
import {PanelDependencyUnavailable} from './dependencies.js';
export const jobLabels=Object.freeze({queued:'En cola',running:'En ejecución',completed:'Completado',failed:'Fallido'});
let referencesProvider;
// Nicole can supply IDs already known to her execution screen. This never lists all jobs.
export function configurePanelJobIdsProvider(provider){if(typeof provider!=='function')throw new TypeError('Se necesita una función de identificadores.');referencesProvider=provider;}
export function parseJob(payload,id){
 const job=payload?.trabajo;
 if(!job || job.job_id!==id || !Object.hasOwn(jobLabels,job.estado))throw new Error('Respuesta de trabajo no válida.');
 return Object.freeze({id,state:job.estado,datasetId:typeof job.dataset_id==='string'?job.dataset_id:null,updatedAt:typeof job.actualizado==='string' && Number.isFinite(Date.parse(job.actualizado))?job.actualizado:null});
}
export async function loadJobOverview({signal}={}, {request=apiRequest,references=referencesProvider}={}){
 signal?.throwIfAborted();if(!references)throw new PanelDependencyUnavailable('jobs');
 const supplied=await references({signal});
 if(!Array.isArray(supplied) || !supplied.every(id=>typeof id==='string' && /^[a-zA-Z0-9_-]{1,128}$/.test(id)))throw new Error('Identificadores de trabajo no válidos.');
 const ids=[...new Set(supplied)];const results=[];
 for(let offset=0;offset<Math.min(ids.length,10);offset+=3){
  signal?.throwIfAborted();
  results.push(...await Promise.all(ids.slice(offset,Math.min(offset+3,10)).map(async id=>{
   try{const response=await request(`/jobs/${id}`,{signal});if(response.status===404)return {id,kind:'not-found'};if(!response.ok)throw new Error('Consulta no disponible');return {kind:'verified',...parseJob(await response.json(),id)};}
   catch(error){signal?.throwIfAborted();return {id,kind:'unavailable'};}
  })));
 }
 signal?.throwIfAborted();return {totalReferenced:ids.length,results};
}
