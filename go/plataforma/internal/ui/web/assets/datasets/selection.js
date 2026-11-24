import {validateCatalog,canSelect} from './model.js';
export function createSelectionStore(){
 let catalog=[],candidate=null,verified=false;
 return {
  invalidate(){verified=false;},
  reconcile(records){
   catalog=validateCatalog(records);verified=true;
   if(candidate && !catalog.some(d=>d.id===candidate.id && d.period===candidate.period && d.sha256===candidate.sha256 && canSelect(d)))candidate=null;
  },
  choose(id){
   if(!verified)return false;
   const dataset=catalog.find(d=>d.id===id && canSelect(d));if(!dataset)return false;
   candidate={id:dataset.id,period:dataset.period,sha256:dataset.sha256};return true;
  },
  clear(){candidate=null;},
  current(){return verified && candidate ? catalog.find(d=>d.id===candidate.id) ?? null : null;},
 };
}
export const datasetSelection=createSelectionStore();
