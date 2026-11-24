import {clearPanelResult} from '../panel/handoff.js';
import {createPanelShell} from '../panel/shell.js';
import {createPanelController} from '../panel/controller.js';
import {loadHealth} from '../panel/health.js';
import {renderHealth} from '../panel/health-card.js';
import {loadCatalogSummary,renderCatalogSummary} from '../panel/catalog-summary.js';
import {loadJobOverview} from '../panel/jobs.js';
import {renderJobs} from '../panel/jobs-card.js';
import {loadTrainingSummary,renderTraining} from '../panel/training.js';
import {emptyState,loadingState,errorState} from '../states.js';
export function createPanelView({sources={health:loadHealth,catalog:loadCatalogSummary,jobs:loadJobOverview,training:loadTrainingSummary}}={}){
 return function panelView({signal}={}){
  const shell=createPanelShell();
  const regions={health:shell.health,catalog:shell.datasets,jobs:shell.jobs,training:shell.metrics};
  const renderers={health:renderHealth,catalog:value=>value.count?renderCatalogSummary(value):emptyState('Sin conjuntos disponibles','La consulta verificada no devolvió conjuntos.'),jobs:value=>value.totalReferenced?renderJobs(value):emptyState('Sin trabajos para consultar','La sesión no proporcionó identificadores de trabajos.'),training:renderTraining};
  const unavailable={health:['Estado no disponible','No se pudo comprobar el servicio.'],catalog:['Catálogo sin conectar','Todavía no se dispone de una consulta de conjuntos.'],jobs:['Trabajos sin conectar','La consulta de trabajos de esta sesión todavía no está conectada.'],training:['Sin tiempos verificados','No hay un resultado completado y consultado para mostrar su tiempo.']};
  const controller=createPanelController({sources,signal,onState(state){
   const region=regions[state.region];region.setAttribute('aria-busy',String(state.kind==='loading'));
   let content;
   if(state.kind==='loading')content=loadingState('Consultando información del servicio.');
   else if(state.kind==='error')content=errorState('No se pudo verificar esta información. Los valores anteriores no se muestran como actuales.',()=>refresh());
   else if(state.kind==='unavailable' || state.kind==='empty')content=emptyState(...unavailable[state.region]);
   else content=renderers[state.region](state.value);
   region.replaceChildren(content);
  },onRefresh(state){if(state.loading)clearPanelResult();shell.refresh.disabled=state.loading;shell.updated.textContent=state.loading?'Actualización en curso…':`Último intento: ${new Date(state.checkedAt).toLocaleString('es-PE')} (hora de este navegador).`;}});
  const refresh=async()=>{await controller.refresh();if(!signal?.aborted)shell.root.querySelector('h1').focus();};
  shell.refresh.addEventListener('click',refresh);void controller.refresh();return shell.root;
 };
}
export const panelView=createPanelView();
