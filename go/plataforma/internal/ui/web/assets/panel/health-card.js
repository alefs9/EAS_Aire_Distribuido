const labels={'ok':'Disponible','degradado':'Disponible con limitaciones','no disponible':'No disponible'};
const dependencies={mongodb:'Almacenamiento persistente',redis:'Caché de resultados',memoria:'Almacenamiento de desarrollo'};
export function healthPresentation(health){return {title:labels[health.state],dependencies:health.dependencies.map(item=>({name:dependencies[item.name] ?? 'Dependencia del servicio',status:item.available?'Disponible':'No disponible'}))};}
export function renderHealth(health){
 const root=document.createElement('div');const model=healthPresentation(health);
 const status=document.createElement('p');status.className=`panel-status panel-status-${health.state==='ok'?'ok':'warning'}`;status.textContent=model.title;root.append(status);
 const list=document.createElement('dl');list.className='panel-facts';
 for(const item of model.dependencies){const name=document.createElement('dt');name.textContent=item.name;const value=document.createElement('dd');value.textContent=item.status;list.append(name,value);}root.append(list);
 const note=document.createElement('p');note.className='panel-note';note.textContent='Esta comprobación no confirma la disponibilidad de nodos ni el éxito de un entrenamiento.';root.append(note);return root;
}
