import {CatalogNotConfiguredError} from './catalog-source.js';
export function createCatalogController({load,onState,signal}){
 let revision=0,request,closed=false;
 const dispose=()=>{closed=true;revision++;request?.abort();};
 if(signal?.aborted)dispose();else signal?.addEventListener('abort',dispose,{once:true});
 return {dispose, async refresh(){
  if(closed)return;
  request?.abort();request=new AbortController();const current=++revision;
  onState({kind:'loading'});
  try{
   const datasets=await load({signal:request.signal});
   if(!closed && current===revision)onState({kind:datasets.length?'data':'empty',datasets});
  }catch(error){
   if(!closed && current===revision)onState({kind:error instanceof CatalogNotConfiguredError?'unavailable':'error'});
  }
 }};
}
