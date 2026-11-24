import {createPanelShell} from '../panel/shell.js';
import {loadTrainingSummary,renderTraining} from '../panel/training.js';
import {loadJobOverview} from '../panel/jobs.js';
import {renderJobs} from '../panel/jobs-card.js';
import {loadHealth} from '../panel/health.js';
import {loadCatalogSummary,renderCatalogSummary} from '../panel/catalog-summary.js';
import {PanelDependencyUnavailable} from '../panel/dependencies.js';
import {renderHealth} from '../panel/health-card.js';
import {emptyState,loadingState,errorState} from '../states.js';
export function panelView({signal}={}){
 const shell=createPanelShell();
 shell.health.append(loadingState('Comprobando la disponibilidad del servicio.'));
 void loadHealth({signal}).then(health=>{if(!signal?.aborted)shell.health.replaceChildren(renderHealth(health));}).catch(()=>{if(!signal?.aborted)shell.health.replaceChildren(errorState('No se pudo verificar la disponibilidad del servicio.'));});
 shell.datasets.append(loadingState('Consultando los conjuntos disponibles.'));
 void loadCatalogSummary({signal}).then(summary=>{if(!signal?.aborted)shell.datasets.replaceChildren(summary.count?renderCatalogSummary(summary):emptyState('Sin conjuntos disponibles','La consulta verificada no devolvió conjuntos.'));}).catch(error=>{if(!signal?.aborted)shell.datasets.replaceChildren(error instanceof PanelDependencyUnavailable?emptyState('Catálogo sin conectar','Todavía no se dispone de una consulta de conjuntos.'):errorState('No se pudieron verificar los conjuntos.'));});
 shell.jobs.append(loadingState('Consultando trabajos de esta sesión.'));
 void loadJobOverview({signal}).then(overview=>{if(!signal?.aborted){shell.jobs.replaceChildren(overview.totalReferenced?renderJobs(overview):emptyState('Sin trabajos para consultar','La sesión no proporcionó identificadores de trabajos.'));void loadTrainingSummary(overview,{signal}).then(summary=>{if(!signal?.aborted)shell.metrics.replaceChildren(summary?renderTraining(summary):emptyState('Sin tiempos verificados','No hay un trabajo completado y consultado para mostrar su tiempo.'));}).catch(()=>{if(!signal?.aborted)shell.metrics.replaceChildren(errorState('No se pudo verificar el tiempo de entrenamiento.'));});}}).catch(error=>{if(!signal?.aborted)shell.jobs.replaceChildren(error instanceof PanelDependencyUnavailable?emptyState('Trabajos sin conectar','La consulta de trabajos de esta sesión todavía no está conectada.'):errorState('No se pudieron verificar los trabajos de la sesión.'));});
 shell.metrics.append(emptyState('Sin tiempos verificados','No hay resultados consultados para mostrar tiempos.'));
 return shell.root;
}
