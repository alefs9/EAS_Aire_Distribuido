const key='eas-air:selected-dataset:v1';
export function sessionStorageIfAvailable(){try{return typeof window==='undefined' ? null : window.sessionStorage;}catch{return null;}}
export function createSelectionPersistence(storage){
 return {
  read(){
   try{
    const value=JSON.parse(storage?.getItem(key) ?? 'null');
    if(!value || value.version!==1 || typeof value.id!=='string' || !value.id.trim() || !['1_hora','24_horas'].includes(value.period) || !(value.sha256===null || typeof value.sha256==='string' && /^[a-f0-9]{64}$/i.test(value.sha256)))return null;
    return {id:value.id,period:value.period,sha256:value.sha256};
   }catch{return null;}
  },
  write(value){try{if(value)storage?.setItem(key,JSON.stringify({version:1,id:value.id,period:value.period,sha256:value.sha256}));else storage?.removeItem(key);}catch{/* Selection still works in memory when storage is unavailable. */}},
 };
}
