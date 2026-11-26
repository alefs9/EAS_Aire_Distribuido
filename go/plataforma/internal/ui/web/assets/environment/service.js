import {apiRequest} from '../api.js';
import {parseProfile,validId} from './model.js';
export async function loadProfiles(ids,{signal,request=apiRequest,timeoutMs=15000}={}) {
 if(!Array.isArray(ids)||ids.length<1||ids.length>2||ids.some(id=>!validId(id))||new Set(ids).size!==ids.length)throw new Error('Escribe uno o dos identificadores distintos y válidos.');
 const controller=new AbortController(),cancel=()=>controller.abort(signal.reason);
 if(signal?.aborted)cancel();else signal?.addEventListener('abort',cancel,{once:true});
 const timer=setTimeout(()=>controller.abort(new DOMException('La consulta tardó demasiado.','TimeoutError')),timeoutMs);
 try {
  controller.signal.throwIfAborted();
  return await Promise.all(ids.map(async id=>{
   const response=await request(`/jobs/${id}/results`,{signal:controller.signal});
   if(response.status===404)throw new Error(`No se encontró el trabajo ${id}.`);
   if(response.status===409)throw new Error(`El trabajo ${id} aún no tiene resultados. Vuelve a consultar cuando termine.`);
   if(response.status===422)throw new Error(`El trabajo ${id} falló y no puede compararse.`);
   if(!response.ok)throw new Error('No se pudo consultar la API. Intenta nuevamente.');
   const body=await response.json();controller.signal.throwIfAborted();return parseProfile(body,id);
  }));
 } finally {clearTimeout(timer);signal?.removeEventListener('abort',cancel);controller.abort();}
}
