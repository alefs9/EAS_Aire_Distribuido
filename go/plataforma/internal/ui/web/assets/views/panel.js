import {createPanelShell} from '../panel/shell.js';
import {emptyState} from '../states.js';
export function panelView(){
 const shell=createPanelShell();
 shell.health.append(emptyState('Estado sin consultar','La disponibilidad todavía no se ha verificado.'));
 shell.datasets.append(emptyState('Catálogo sin conectar','Todavía no se dispone de una consulta de conjuntos.'));
 shell.jobs.append(emptyState('Trabajos sin conectar','No se dispone de un listado de trabajos para consultar.'));
 shell.metrics.append(emptyState('Sin tiempos verificados','No hay resultados consultados para mostrar tiempos.'));
 return shell.root;
}
