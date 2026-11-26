import {identifyDataset} from './datasets.js';
export const validId=id=>typeof id==='string'&&/^[a-zA-Z0-9_-]{1,128}$/.test(id);
export function parseProfile(body,id) {
 const r=body?.resultado,p=body?.parametros;
 if(body?.job_id!==id||!r||!p||!Number.isSafeInteger(p.k)||p.k<1||p.k>20||!Array.isArray(r.centroides)||!Array.isArray(r.conteos)||r.centroides.length!==p.k||r.conteos.length!==p.k)throw new Error('El perfil recibido no es válido.');
 if(!r.centroides.every(c=>Array.isArray(c)&&c.length===2&&c.every(v=>Number.isFinite(v)&&v>=0))||!r.conteos.every(n=>Number.isSafeInteger(n)&&n>=0))throw new Error('Las concentraciones o los conteos no son válidos.');
 if(!Number.isInteger(r.iteraciones)||r.iteraciones<1||!Number.isInteger(p.max_iter)||r.iteraciones>p.max_iter||typeof r.convergio!=='boolean')throw new Error('La finalización del trabajo no es válida.');
 const dataset=identifyDataset(body.dataset_id),total=r.conteos.reduce((a,b)=>a+b,0);
 if(total!==dataset.total)throw new Error('Los conteos no corresponden al conjunto identificado.');
 return {id,datasetId:body.dataset_id,dataset,parameters:p,converged:r.convergio,iterations:r.iteraciones,groups:r.centroides.map((c,i)=>({group:i+1,pm10:c[0],pm25:c[1],count:r.conteos[i]}))};
}
