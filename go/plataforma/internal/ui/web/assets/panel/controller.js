import {PanelDependencyUnavailable} from './dependencies.js';
export function createPanelController({sources,onState,onRefresh,signal,clock=()=>new Date().toISOString()}){
 let generation=0,request,disposed=false;
 const dispose=()=>{disposed=true;generation++;request?.abort();};
 if(signal?.aborted)dispose();else signal?.addEventListener('abort',dispose,{once:true});
 return {dispose,async refresh(){
  if(disposed)return;
  request?.abort();request=new AbortController();const current=++generation;const options={signal:request.signal};
  const active=()=>!disposed && current===generation;
  const emit=(region,state)=>{if(active())onState({region,...state});};
  onRefresh({loading:true,checkedAt:null});
  for(const region of ['health','catalog','jobs','training'])emit(region,{kind:'loading'});
  const read=async(region,query)=>{
   try{const value=await query();emit(region,{kind:value===null?'empty':'data',value});return {ok:true,value};}
   catch(error){emit(region,{kind:error instanceof PanelDependencyUnavailable?'unavailable':'error'});return {ok:false};}
  };
  await Promise.all([
   read('health',()=>sources.health(options)),read('catalog',()=>sources.catalog(options)),
   (async()=>{const jobs=await read('jobs',()=>sources.jobs(options));if(!active())return;if(jobs.ok)await read('training',()=>sources.training(jobs.value,options));else emit('training',{kind:'unavailable'});})(),
  ]);
  if(active())onRefresh({loading:false,checkedAt:clock()});
 }};
}
