import {createDatasetsShell} from '../datasets/shell.js';
import {emptyState} from '../states.js';
export function datasetsView() {
 const shell=createDatasetsShell();
 shell.content.append(emptyState('Catálogo todavía no disponible', 'La consulta de conjuntos aún no está conectada al servicio. No se ha verificado qué datos están disponibles para una ejecución.'));
 return shell.root;
}
