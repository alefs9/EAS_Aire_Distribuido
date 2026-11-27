let selectedResultJobId=null;
// Valentina can read this ID on the existing results route; it is not a result.
export function rememberPanelResult(job){
 if(job?.kind!=='verified' || job.state!=='completed' || typeof job.id!=='string' || !/^[a-zA-Z0-9_-]{1,128}$/.test(job.id))return false;
 selectedResultJobId=job.id;return true;
}
export const getPanelResultJobId=()=>selectedResultJobId;
export function clearPanelResult(){selectedResultJobId=null;}
export function resultLink(job){
 const link=document.createElement('a');link.href=`/resultados?job_id=${encodeURIComponent(job.id)}`;link.dataset.route='';link.className='text-link';link.textContent='Consultar resultados';link.setAttribute('aria-label',`Consultar resultados del trabajo ${job.id}`);
 link.addEventListener('click',()=>rememberPanelResult(job));return link;
}
