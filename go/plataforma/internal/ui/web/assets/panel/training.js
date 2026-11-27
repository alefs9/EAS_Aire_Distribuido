import {formatDecimal} from '../presentation.js';
import {executionName} from '../execution-names.js';
import {apiRequest} from '../api.js';
export function parseTraining(payload,id){
 const seconds=payload?.resultado?.segundos_ml;
 if(payload?.job_id!==id || !Number.isFinite(seconds) || seconds<0)throw new Error('Tiempo de entrenamiento no válido.');
 return Object.freeze({id,seconds});
}
export async function loadTrainingSummary(overview,{signal}={},request=apiRequest){
 const job=overview.results.find(item=>item.kind==='verified' && item.state==='completed');
 if(!job)return null;
 const response=await request(`/jobs/${job.id}/results`,{signal});
 if(!response.ok)throw new Error('No hay un resultado verificado para mostrar tiempos.');
 return parseTraining(await response.json(),job.id);
}
export function renderTraining(summary){
 const root=document.createElement('div');const value=document.createElement('p');value.className='panel-status';value.textContent=`${formatDecimal(summary.seconds)} s de entrenamiento`;root.append(value);
 const job=document.createElement('p');job.className='panel-note panel-job-id';job.textContent=`${executionName(summary.id)}`;root.append(job);
 const note=document.createElement('p');note.className='panel-note';note.textContent='Tiempo de entrenamiento de esta ejecución. No es el tiempo de respuesta HTTP ni un promedio de todos los trabajos.';root.append(note);return root;
}
