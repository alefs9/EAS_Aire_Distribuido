import {createPanelShell} from '../panel/shell.js';
import {loadHealth} from '../panel/health.js';
import {renderHealth} from '../panel/health-card.js';
import {emptyState,loadingState,errorState} from '../states.js';
export function panelView({signal}={}){
 const shell=createPanelShell();
 shell.health.append(loadingState('Comprobando la disponibilidad del servicio.'));
 void loadHealth({signal}).then(health=>{if(!signal?.aborted)shell.health.replaceChildren(renderHealth(health));}).catch(()=>{if(!signal?.aborted)shell.health.replaceChildren(errorState('No se pudo verificar la disponibilidad del servicio.'));});
 shell.datasets.append(emptyState('Catálogo sin conectar','Todavía no se dispone de una consulta de conjuntos.'));
 shell.jobs.append(emptyState('Trabajos sin conectar','No se dispone de un listado de trabajos para consultar.'));
 shell.metrics.append(emptyState('Sin tiempos verificados','No hay resultados consultados para mostrar tiempos.'));
 return shell.root;
}
