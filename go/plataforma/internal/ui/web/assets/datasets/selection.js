import {validateCatalog,canSelect} from './model.js';
import {createSelectionPersistence,sessionStorageIfAvailable} from './persistence.js';
export function createSelectionStore({storage=sessionStorageIfAvailable()}={}){
 const persistence=createSelectionPersistence(storage);
 let catalog=[],candidate=persistence.read(),verified=false;
 return {
  invalidate(){verified=false;},
  reconcile(records){
   catalog=validateCatalog(records);verified=true;
   if(candidate && !catalog.some(d=>d.id===candidate.id && d.period===candidate.period && d.sha256===candidate.sha256 && canSelect(d))){candidate=null;persistence.write(null);}
  },
  choose(id){
   if(!verified)return false;
   const dataset=catalog.find(d=>d.id===id && canSelect(d));if(!dataset)return false;
   candidate={id:dataset.id,period:dataset.period,sha256:dataset.sha256};persistence.write(candidate);return true;
  },
  clear(){candidate=null;persistence.write(null);},
  current(){return verified && candidate ? catalog.find(d=>d.id===candidate.id) ?? null : null;},
 };
}
export const datasetSelection=createSelectionStore();
