import {loadResult} from '../results/service.js';
import {validJobId} from '../results/model.js';
import {initialResultJob,resultLocation} from '../results/navigation.js';
import {element,renderDetail} from '../results/detail.js';
import {renderCharts} from '../results/charts.js';
import {getPanelResultJobId} from '../panel/handoff.js';
import {loadingState,emptyState,errorState} from '../states.js';
export function resultsView({signal,search=globalThis.location?.search ?? ''}={}) {
 const root=element('section',undefined,'page-section'),heading=element('h1','Resultados de K-Means');heading.id='titulo-vista';heading.tabIndex=-1;
 root.append(heading,element('p','Consulta un trabajo para revisar sus grupos y parámetros.'));
 const form=element('form',undefined,'result-form'),label=element('label','Identificador del trabajo'),input=element('input'),button=element('button','Consultar','button');
 input.id='result-job';input.name='job';input.required=true;input.maxLength=128;input.pattern='[a-zA-Z0-9_-]+';input.autocomplete='off';label.htmlFor=input.id;button.type='submit';form.append(label,input,button);
 const region=element('div');region.setAttribute('aria-live','polite');root.append(form,region);
 let active,version=0;
 signal?.addEventListener('abort',()=>{version++;active?.abort();},{once:true});
 input.addEventListener('input',()=>{version++;active?.abort();region.setAttribute('aria-busy','false');region.replaceChildren(emptyState('Identificador modificado','Pulsa Consultar para obtener el resultado de este trabajo.'));});
 async function consult() {
  if(signal?.aborted)return;
  const id=input.value.trim();
  if(!validJobId(id)){region.replaceChildren(errorState('Escribe un identificador válido: letras, números, guiones o guion bajo.'));return;}
  active?.abort();active=new AbortController();const current=++version;
  if(globalThis.location && globalThis.history)history.replaceState(null,'',resultLocation(location.href,id));
  region.setAttribute('aria-busy','true');region.replaceChildren(loadingState('Consultando el resultado…'));
  try {
   const state=await loadResult(id,{signal:active.signal});
   if(current!==version||signal?.aborted)return;
   if(state.kind==='completed'){const {root:detail,data}=renderDetail(state.model);region.replaceChildren(detail,renderCharts(state.model,data));}
   else if(state.kind==='pending')region.replaceChildren(emptyState(state.state==='queued'?'Trabajo en cola':'Trabajo en ejecución','Todavía no hay resultados. Pulsa Consultar para actualizar.'));
   else if(state.kind==='failed')region.replaceChildren(emptyState('El trabajo falló',state.message));
   else region.replaceChildren(emptyState('Trabajo no encontrado','Revisa el identificador e intenta nuevamente.'));
  } catch(error) {
   if(current===version&&!signal?.aborted)region.replaceChildren(errorState(error.message||'No se pudo consultar el resultado.',consult));
  } finally {if(current===version)region.setAttribute('aria-busy','false');}
 }
 form.addEventListener('submit',event=>{event.preventDefault();void consult();});
 const initial=initialResultJob(search,getPanelResultJobId());
 input.value=initial.id;
 if(initial.error)region.append(errorState(initial.error));
 else if(input.value)void consult();
 else region.append(emptyState('Selecciona un trabajo','Usa un enlace de Ejecuciones o Panel general, o escribe su identificador.'));
 return root;
}
