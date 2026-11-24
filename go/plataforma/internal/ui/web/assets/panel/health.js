import {apiRequest} from '../api.js';
const states=new Set(['ok','degradado','no disponible']);
export function parseHealth(payload,status){
 if(!payload || !states.has(payload.estado) || !payload.dependencias || Array.isArray(payload.dependencias) || typeof payload.dependencias!=='object' || !Object.values(payload.dependencias).every(value=>typeof value==='string'))throw new Error('Respuesta de disponibilidad no válida.');
 if(!(status===200 && ['ok','degradado'].includes(payload.estado) || status===503 && payload.estado==='no disponible'))throw new Error('Estado HTTP no coherente con la disponibilidad.');
 return Object.freeze({state:payload.estado,dependencies:Object.freeze(Object.entries(payload.dependencias).map(([name,value])=>Object.freeze({name,available:value==='ok'})))});
}
export async function loadHealth({signal}={},request=apiRequest){
 const response=await request('/health',{signal});
 if(response.status!==200 && response.status!==503)throw new Error('No se pudo consultar la disponibilidad del servicio.');
 return parseHealth(await response.json(),response.status);
}
