export const validJobId = id => typeof id === 'string' && /^[a-zA-Z0-9_-]{1,128}$/.test(id);
const fail = () => {throw new Error('El resultado recibido no es consistente.');};
export function parseResult(payload, id) {
 const p=payload?.parametros, r=payload?.resultado;
 if(payload?.job_id!==id || !p || !r || typeof payload.dataset_id!=='string') fail();
 if(!Number.isSafeInteger(p.k)||p.k<1||p.k>20||!Number.isSafeInteger(p.semilla)||!Number.isInteger(p.max_iter)||p.max_iter<1||!Number.isFinite(p.tolerancia)||p.tolerancia<0) fail();
 if(!Array.isArray(r.centroides)||r.centroides.length!==p.k||!Array.isArray(r.conteos)||r.conteos.length!==p.k) fail();
 if(!r.centroides.every(c=>Array.isArray(c)&&c.length===2&&c.every(Number.isFinite))||!r.conteos.every(n=>Number.isSafeInteger(n)&&n>=0)) fail();
 const total=r.conteos.reduce((a,b)=>a+b,0);
 if(!Number.isSafeInteger(total)||total<1||!Number.isInteger(r.iteraciones)||r.iteraciones<1||r.iteraciones>p.max_iter||typeof r.convergio!=='boolean'||!Number.isFinite(r.inercia)||r.inercia<0||typeof r.origen!=='string') fail();
 return {id,datasetId:payload.dataset_id,parameters:p,result:r,total,source:typeof payload.fuente==='string'?payload.fuente:'No informada',notice:typeof payload.aviso==='string'?payload.aviso:null};
}
export function finishReason(model) {
 if(model.result.convergio) return 'Convergencia alcanzada';
 return model.result.iteraciones===model.parameters.max_iter ? 'Máximo de iteraciones alcanzado sin converger' : 'Finalización sin convergencia; motivo no informado';
}
