import {createDatasetsShell} from '../datasets/shell.js';
import {renderCatalog} from '../datasets/catalog.js';
import {loadCatalog} from '../datasets/catalog-source.js';
import {createCatalogController} from '../datasets/controller.js';
import {datasetSelection} from '../datasets/selection.js';
import {periodLabels} from '../datasets/model.js';
import {renderSelectionActions} from '../datasets/handoff.js';
import {emptyState,loadingState,errorState} from '../states.js';
export function createDatasetsView({load=loadCatalog,selection=datasetSelection}={}){
 return function datasetsView({signal}={}){
  const shell=createDatasetsShell();
  const updateSummary=()=>{
   renderSelectionActions(shell,selection,()=>{selection.clear();shell.root.querySelectorAll('input[name=dataset]').forEach(input=>input.checked=false);updateSummary();shell.root.querySelector('input[name=dataset]:not(:disabled)')?.focus();});
   const selected=selection.current();
  shell.summary.textContent=selected ? `Conjunto seleccionado: ${periodLabels[selected.period]}. La ejecución todavía no se ha iniciado.` : 'Selecciona un conjunto verificado para continuar.';
  };
  const controller=createCatalogController({load,signal,onState(state){
   if(state.kind==='data' || state.kind==='empty')selection.reconcile(state.datasets);else selection.invalidate();
   updateSummary();
   shell.refresh.disabled=state.kind==='loading' || state.kind==='unavailable';
   shell.content.setAttribute('aria-busy',String(state.kind==='loading'));
   if(state.kind==='loading')shell.content.replaceChildren(loadingState('Consultando los conjuntos disponibles.'));
   else if(state.kind==='unavailable')shell.content.replaceChildren(emptyState('Catálogo todavía no disponible','La consulta de conjuntos aún no está conectada al servicio. No se ha verificado qué datos están disponibles para una ejecución.'));
   else if(state.kind==='empty')shell.content.replaceChildren(emptyState('No hay conjuntos disponibles','El servicio no devolvió conjuntos. Puedes volver a consultar más tarde.'));
   else if(state.kind==='error')shell.content.replaceChildren(errorState('No se pudo consultar el catálogo. Tu selección no se utilizará sin verificarla.',()=>refresh()));
   else shell.content.replaceChildren(renderCatalog(state.datasets,{selectedId:selection.current()?.id,onSelect:id=>{selection.choose(id);updateSummary();}}).element);
  }});
  const refresh=async()=>{
   await controller.refresh();if(signal?.aborted)return;
   const target=shell.content.querySelector('h2') || shell.root.querySelector('h1');target.tabIndex=-1;target.focus();
  };
  shell.refresh.addEventListener('click',refresh);
  void controller.refresh();
  return shell.root;
 };
}
export const datasetsView=createDatasetsView();
