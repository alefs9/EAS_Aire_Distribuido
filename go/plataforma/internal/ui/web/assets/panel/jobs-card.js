import {jobLabels} from './jobs.js';
import {resultLink} from './handoff.js';
export function summarizeJobs(overview){
 const counts={queued:0,running:0,completed:0,failed:0};let verified=0;
 for(const job of overview.results)if(job.kind==='verified' && Object.hasOwn(counts,job.state)){counts[job.state]++;verified++;}
 return {counts,verified,queried:overview.results.length,total:overview.totalReferenced};
}
export function renderJobs(overview){
 const root=document.createElement('div');const summary=summarizeJobs(overview);
 const note=document.createElement('p');note.className='panel-note';note.textContent=`${summary.queried} de ${summary.total} trabajos de la sesión consultados; ${summary.verified} estados verificados. No es un listado global del sistema.`;root.append(note);
 if(summary.verified){const counts=document.createElement('ul');counts.className='panel-job-counts';for(const [state,count] of Object.entries(summary.counts)){const item=document.createElement('li');item.textContent=`${jobLabels[state]}: ${count}`;counts.append(item);}root.append(counts);}
 const wrap=document.createElement('div');wrap.className='table-wrap';wrap.tabIndex=0;wrap.setAttribute('role','region');wrap.setAttribute('aria-label','Tabla de estados de trabajos');
 const table=document.createElement('table');table.className='data-table';
 const caption=document.createElement('caption');caption.textContent='Estados de los trabajos consultados';table.append(caption);
 const head=document.createElement('thead');const heading=document.createElement('tr');for(const label of ['Trabajo','Estado','Actualización informada','Resultados']){const th=document.createElement('th');th.scope='col';th.textContent=label;heading.append(th);}head.append(heading);table.append(head);
 const body=document.createElement('tbody');
 for(const job of overview.results){const row=document.createElement('tr');const id=document.createElement('th');id.scope='row';id.textContent=job.id;id.className='panel-job-id';
  const state=document.createElement('td');state.textContent=job.kind==='verified'?jobLabels[job.state]:job.kind==='not-found'?'No encontrado':'Sin verificar';
  const updated=document.createElement('td');updated.textContent=job.updatedAt?new Date(job.updatedAt).toLocaleString('es-PE'):'No informada';const access=document.createElement('td');if(job.kind==='verified' && job.state==='completed')access.append(resultLink(job));else access.textContent='Sin resultado verificado';row.append(id,state,updated,access);body.append(row);
 }
 table.append(body);wrap.append(table);root.append(wrap);return root;
}
