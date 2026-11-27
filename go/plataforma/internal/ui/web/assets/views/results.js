import {executionSelect} from '../execution-names.js';
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
 const initial=initialResultJob(search,getPanelResultJobId());
 const form=element('form',undefined,'result-form'),label=element('label','Ejecución'),input=executionSelect('result-job',{selected:initial.id}),button=element('button','Consultar','button');
 input.name='job';label.htmlFor=input.id;button.type='submit';form.append(label,input,button);
 const region=element('div');region.setAttribute('aria-live','polite');root.append(form,region);
 let active,version=0;
 signal?.addEventListener('abort',()=>{version++;active?.abort();},{once:true});
 input.addEventListener('input',()=>{version++;active?.abort();region.setAttribute('aria-busy','false');region.replaceChildren(emptyState('Ejecución seleccionada','Pulsa Consultar para obtener sus resultados.'));});
 async function consult() {
  if(signal?.aborted)return;
  const id=input.value.trim();
  if(!validJobId(id)){region.replaceChildren(errorState('Selecciona una ejecución disponible.'));return;}
  active?.abort();active=new AbortController();const current=++version;
  if(globalThis.location && globalThis.history)history.replaceState(null,'',resultLocation(location.href,id));
  region.setAttribute('aria-busy','true');region.replaceChildren(loadingState('Consultando el resultado…'));
  try {
   const state=await loadResult(id,{signal:active.signal});
   if(current!==version||signal?.aborted)return;
   if(state.kind==='completed'){const {root:detail,data}=renderDetail(state.model);region.replaceChildren(detail,renderCharts(state.model,data));}
   else if(state.kind==='pending')region.replaceChildren(emptyState(state.state==='queued'?'Trabajo en cola':'Trabajo en ejecución','Todavía no hay resultados. Pulsa Consultar para actualizar.'));
   else if(state.kind==='failed')region.replaceChildren(emptyState('El trabajo falló','La ejecución no pudo completarse. Revisa su seguimiento e inténtalo de nuevo.'));
   else region.replaceChildren(emptyState('Trabajo no encontrado','Esta ejecución no está disponible. Selecciona otra o vuelve a intentarlo.'));
  } catch(error) {
   if(current===version&&!signal?.aborted)region.replaceChildren(errorState(error.message||'No se pudo consultar el resultado.',consult));
  } finally {if(current===version)region.setAttribute('aria-busy','false');}
 }
 form.addEventListener('submit',event=>{event.preventDefault();void consult();});
 input.value=initial.id;
 if(initial.error)region.append(errorState(initial.error));
 else if(input.value)void consult();
 else region.append(emptyState('Selecciona un trabajo','Selecciona una ejecución reciente o abre su resultado desde Ejecuciones o el Panel general.'));
 return root;
}
