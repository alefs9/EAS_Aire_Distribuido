import {periodLabels,canSelect} from './model.js';
import {datasetDetails} from './details.js';
export function renderCatalog(datasets,{onSelect,selectedId=null}){
 const fieldset=document.createElement('fieldset');fieldset.className='dataset-options';
 const legend=document.createElement('legend');legend.textContent='Selecciona un conjunto para continuar';fieldset.append(legend);
 const radios=[];
 for(const [period,title] of Object.entries(periodLabels)){
  const group=document.createElement('section');group.className='dataset-period';
  const heading=document.createElement('h2');heading.textContent=title;group.append(heading);
  const records=datasets.filter(dataset=>dataset.period===period);
  if(!records.length){const note=document.createElement('p');note.textContent='No se recibieron conjuntos de este periodo.';group.append(note);}
  for(const dataset of records){
   const index=radios.length;
   const card=document.createElement('article');card.className='card dataset-card';
   const choice=document.createElement('div');choice.className='dataset-choice';
   const input=document.createElement('input');input.type='radio';input.name='dataset';input.id=`dataset-${index}`;input.value=dataset.id;input.checked=dataset.id===selectedId;input.disabled=!canSelect(dataset);
   const label=document.createElement('label');label.htmlFor=input.id;label.textContent=title;
   input.addEventListener('change',()=>{if(input.checked)onSelect(dataset.id);});
   choice.append(input,label);card.append(choice);
   const details=document.createElement('dl');details.className='dataset-details';
   for(const [name,value] of datasetDetails(dataset)){const dt=document.createElement('dt');dt.textContent=name;const dd=document.createElement('dd');dd.textContent=value;details.append(dt,dd);}
   card.append(details);
   if(input.disabled){const note=document.createElement('p');note.textContent='Este conjunto no contiene vectores y no se puede seleccionar.';card.append(note);}
   group.append(card);radios.push(input);
  }
  fieldset.append(group);
 }
 return {element:fieldset,radios};
}
