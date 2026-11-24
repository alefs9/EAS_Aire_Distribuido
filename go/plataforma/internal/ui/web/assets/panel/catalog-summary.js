import {PanelDependencyUnavailable} from './dependencies.js';
let catalogLoader;
export function configurePanelCatalogLoader(loader){if(typeof loader!=='function')throw new TypeError('La consulta debe ser una función.');catalogLoader=loader;}
export function summarizeCatalog(records){
 if(!Array.isArray(records))throw new Error('Catálogo no válido.');const ids=new Set();
 const groups=['1_hora','24_horas'].map(period=>({period,datasets:0,vectors:0}));
 for(const item of records){
  if(!item || typeof item.id!=='string' || !item.id.trim() || ids.has(item.id) || !['1_hora','24_horas'].includes(item.period) || item.vectorCount!=null && (!Number.isSafeInteger(item.vectorCount)||item.vectorCount<0))throw new Error('Metadatos de catálogo no válidos.');
  ids.add(item.id);const group=groups.find(g=>g.period===item.period);group.datasets++;
  if(item.vectorCount==null)group.vectors=null;else if(group.vectors!==null){group.vectors+=item.vectorCount;if(!Number.isSafeInteger(group.vectors))throw new Error('Conteo fuera de rango.');}
 }
 return {count:records.length,groups};
}
export async function loadCatalogSummary({signal}={}){
 signal?.throwIfAborted();if(!catalogLoader)throw new PanelDependencyUnavailable('catalog');
 const records=await catalogLoader({signal});signal?.throwIfAborted();return summarizeCatalog(records);
}
export function renderCatalogSummary(summary){
 const root=document.createElement('div');
 const total=document.createElement('p');total.className='panel-status';total.textContent=`${summary.count} conjuntos recibidos`;root.append(total);
 for(const group of summary.groups){const item=document.createElement('p');item.className='panel-note';item.textContent=`${group.period==='1_hora'?'Una hora':'24 horas'}: ${group.datasets} conjuntos · ${group.vectors==null?'conteo de vectores no informado':new Intl.NumberFormat('es-PE').format(group.vectors)+' vectores'}`;root.append(item);}
 return root;
}
