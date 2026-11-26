import {createShell} from '../environment/shell.js';
import {loadProfiles} from '../environment/service.js';
import {selectProfiles} from '../environment/filters.js';
import {renderProfiles} from '../environment/render.js';
import {emptyState,errorState,loadingState} from '../states.js';
export function environmentView({signal}={}){
 const {root,form,region,controls:c}=createShell();let profiles=[],active,version=0;
 const clear=()=>{version++;active?.abort();profiles=[];region.setAttribute('aria-busy','false');region.replaceChildren(emptyState('Consulta los perfiles','Escribe los trabajos del periodo seleccionado y pulsa Consultar perfiles.'));};
 const render=()=>{if(!profiles.length)return;try{region.replaceChildren(renderProfiles(selectProfiles(profiles,{period:c.period.value,pollutant:c.pollutant.value,group:c.group.value})));}catch(error){region.replaceChildren(errorState(error.message));}};
 for(const control of [c.a,c.b,c.period])control.addEventListener('input',clear);
 for(const control of [c.pollutant,c.group])control.addEventListener('change',render);
 signal?.addEventListener('abort',()=>{version++;active?.abort();},{once:true});
 form.addEventListener('submit',async event=>{
  event.preventDefault();if(signal?.aborted)return;active?.abort();active=new AbortController();const current=++version;profiles=[];
  region.setAttribute('aria-busy','true');region.replaceChildren(loadingState('Consultando los trabajos seleccionados…'));
  try{const data=await loadProfiles([c.a.value.trim(),c.b.value.trim()].filter(Boolean),{signal:active.signal});if(current!==version||signal?.aborted)return;profiles=data;render();}
  catch(error){if(current===version&&!signal?.aborted)region.replaceChildren(errorState(error.message));}
  finally{if(current===version)region.setAttribute('aria-busy','false');}
 });clear();return root;
}
