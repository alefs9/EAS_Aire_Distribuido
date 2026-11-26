import {apiRequest} from '../api.js';
import {parseResult,validJobId} from './model.js';
export async function loadResult(id,{signal,request=apiRequest,timeoutMs=15000}={}) {
 if(!validJobId(id)) throw new Error('Escribe un identificador válido: letras, números, guiones o guion bajo.');
 const controller=new AbortController();const abort=()=>controller.abort(signal.reason);
 if(signal?.aborted)abort();else signal?.addEventListener('abort',abort,{once:true});
 const timer=setTimeout(()=>controller.abort(new DOMException('La consulta tardó demasiado.','TimeoutError')),timeoutMs);
 try {
  const response=await request(`/jobs/${id}/results`,{signal:controller.signal});
  if(response.status===404)return {kind:'missing'};
  if(![200,409,422].includes(response.status))throw new Error('No se pudo consultar la API. Intenta nuevamente.');
  const body=await response.json();controller.signal.throwIfAborted();
  if(response.status===409 && body.job_id===id && ['queued','running'].includes(body.estado))return {kind:'pending',state:body.estado};
  if(response.status===422 && body.job_id===id)return {kind:'failed',message:typeof body.error==='string'?body.error:'El trabajo terminó con un error.'};
  if(response.status!==200)throw new Error('La respuesta del servicio no es válida.');
  return {kind:'completed',model:parseResult(body,id)};
 } finally {clearTimeout(timer);signal?.removeEventListener('abort',abort);}
}
