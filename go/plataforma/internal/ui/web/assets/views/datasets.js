import {createDatasetsShell} from '../datasets/shell.js';
import {renderCatalog} from '../datasets/catalog.js';
import {loadCatalog} from '../datasets/catalog-source.js';
import {createCatalogController} from '../datasets/controller.js';
import {emptyState,loadingState,errorState} from '../states.js';
export function createDatasetsView({load=loadCatalog}={}){
 return function datasetsView({signal}={}){
  const shell=createDatasetsShell();
  const controller=createCatalogController({load,signal,onState(state){
   shell.content.setAttribute('aria-busy',String(state.kind==='loading'));
   if(state.kind==='loading')shell.content.replaceChildren(loadingState('Consultando los conjuntos disponibles.'));
   else if(state.kind==='unavailable')shell.content.replaceChildren(emptyState('Catálogo todavía no disponible','La consulta de conjuntos aún no está conectada al servicio. No se ha verificado qué datos están disponibles para una ejecución.'));
   else if(state.kind==='empty')shell.content.replaceChildren(emptyState('No hay conjuntos disponibles','El servicio no devolvió conjuntos. Puedes volver a consultar más tarde.'));
   else if(state.kind==='error')shell.content.replaceChildren(errorState('No se pudo consultar el catálogo. Tu selección no se utilizará sin verificarla.',()=>controller.refresh()));
   else shell.content.replaceChildren(renderCatalog(state.datasets,{onSelect:()=>{}}).element);
  }});
  void controller.refresh();
  return shell.root;
 };
}
export const datasetsView=createDatasetsView();
