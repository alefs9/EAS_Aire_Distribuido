import {formatDecimal} from '../presentation.js';
import {element} from './detail.js';
function svgNode(tag,attrs,text) {const n=document.createElementNS('http://www.w3.org/2000/svg',tag);for(const [k,v] of Object.entries(attrs))n.setAttribute(k,String(v));if(text!==undefined)n.textContent=text;return n;}
export function renderCharts(model,data) {
 const root=element('section',undefined,'result-charts');root.append(element('h3','Distribución de los centroides'));
 const svg=svgNode('svg',{viewBox:'0 0 640 340',role:'img','aria-label':`Centroides PM10 y PM2.5. ${data.unit}. Valores con dos decimales en la tabla.`});
 const values=model.result.centroides;
 const xmin=Math.min(0,...values.map(c=>c[0])),xmax=Math.max(1,...values.map(c=>c[0]));
 const ymin=Math.min(0,...values.map(c=>c[1])),ymax=Math.max(1,...values.map(c=>c[1]));
 const x=v=>75+(v-xmin)/(xmax-xmin)*500,y=v=>275-(v-ymin)/(ymax-ymin)*220;
 svg.append(svgNode('path',{d:'M75 45 V275 H575',fill:'none',stroke:'currentColor'}));
 for(let i=0;i<=4;i++){const xv=xmin+(xmax-xmin)*i/4,yv=ymin+(ymax-ymin)*i/4;svg.append(svgNode('text',{x:x(xv),y:298,'text-anchor':'middle'},formatDecimal(xv)),svgNode('text',{x:65,y:y(yv)+4,'text-anchor':'end'},formatDecimal(yv)));}
 svg.append(svgNode('text',{x:325,y:330,'text-anchor':'middle'},`PM10 (${data.unit})`),svgNode('text',{x:75,y:22},`PM2.5 (${data.unit})`));
 values.forEach((c,i)=>{svg.append(svgNode('circle',{cx:x(c[0]),cy:y(c[1]),r:7,fill:'#176855'}),svgNode('text',{x:x(c[0])+10,y:y(c[1])-8},`G${i+1}`));});root.append(svg);
 root.append(element('h3','Observaciones por grupo'));
 const list=element('ul',undefined,'result-bars'),max=Math.max(...model.result.conteos);
 model.result.conteos.forEach((count,i)=>{const item=element('li');item.append(element('span',`Grupo ${i+1}: ${count}`));const bar=element('span',undefined,'result-bar');bar.style.width=`${count/max*100}%`;bar.setAttribute('aria-hidden','true');item.append(bar);list.append(item);});root.append(list);return root;
}
