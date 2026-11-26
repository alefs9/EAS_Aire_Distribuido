import {loadResult} from '../results/service.js';
import {element,renderDetail} from '../results/detail.js';
import {renderCharts} from '../results/charts.js';
import {getPanelResultJobId} from '../panel/handoff.js';
import {loadingState,emptyState,errorState} from '../states.js';
export function resultsView({signal}={}) {
 const root=element('section',undefined,'page-section'),heading=element('h1','Resultados de K-Means');heading.id='titulo-vista';heading.tabIndex=-1;
 root.append(heading,element('p','Consulta un trabajo para revisar sus grupos y parámetros.'));
 const form=element('form',undefined,'result-form'),label=element('label','Identificador del trabajo'),input=element('input'),button=element('button','Consultar','button');
 input.id='result-job';input.name='job';input.required=true;input.maxLength=128;input.pattern='[a-zA-Z0-9_-]+';input.autocomplete='off';label.htmlFor=input.id;button.type='submit';form.append(label,input,button);
 const region=element('div');region.setAttribute('aria-live','polite');root.append(form,region);
 let active,version=0;
 signal?.addEventListener('abort',()=>active?.abort(),{once:true});
 async function consult() {
  if(signal?.aborted)return;active?.abort();active=new AbortController();const current=++version;
  region.setAttribute('aria-busy','true');region.replaceChildren(loadingState('Consultando el resultado…'));
  try {
   const state=await loadResult(input.value.trim(),{signal:active.signal});
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
 input.value=getPanelResultJobId()||'';
 if(input.value)void consult();else region.append(emptyState('Selecciona un trabajo','Usa el enlace del panel general o escribe su identificador.'));
 return root;
}
