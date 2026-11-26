import {el} from './render.js';
export function createShell(){
 const root=el('section',undefined,'page-section'),h=el('h1','Exploración y comparación ambiental');h.id='titulo-vista';h.tabIndex=-1;root.append(h,el('p','Consulta uno o dos trabajos completados del mismo periodo para explorar sus perfiles de concentración.'));
 const form=el('form',undefined,'environment-form'),controls={};
 for(const [key,label,options] of [['a','Trabajo A'],['b','Trabajo B (opcional)'],['period','Periodo',[['1_hora','Una hora'],['24_horas','24 horas']]],['pollutant','Contaminante',[['both','PM10 y PM2.5'],['pm10','PM10'],['pm25','PM2.5']]],['group','Grupo',[['all','Todos los grupos'],...Array.from({length:20},(_,i)=>[String(i+1),`Grupo ${i+1}`])]]]){
  const field=el('div'),labelNode=el('label',label),input=el(options?'select':'input');input.id=`environment-${key}`;labelNode.htmlFor=input.id;
  if(options)for(const [value,text] of options){const o=el('option',text);o.value=value;input.append(o);}else{input.maxLength=128;input.pattern='[a-zA-Z0-9_-]+';input.required=key==='a';input.autocomplete='off';}
  field.append(labelNode,input);form.append(field);controls[key]=input;
 }
 const button=el('button','Consultar perfiles','button');button.type='submit';form.append(button);const region=el('div');region.setAttribute('aria-live','polite');root.append(form,region);return {root,form,region,controls};
}
