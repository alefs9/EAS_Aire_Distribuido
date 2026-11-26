import {finishReason} from './model.js';
import {provenance} from './provenance.js';
export function element(tag, text, className) {
 const node=document.createElement(tag); if(text!==undefined)node.textContent=text;
 if(className)node.className=className; return node;
}
export function renderDetail(model) {
 const data=provenance(model), root=element('section',undefined,'result-detail');
 root.append(element('h2','Resultado del trabajo'),element('p',model.id,'result-id'));
 const facts=element('dl',undefined,'result-facts');
 const p=model.parameters,r=model.result;
 for(const [label,value] of Object.entries({'Periodo':data.period,'Observaciones':model.total,'Grupos (k)':p.k,'Semilla':p.semilla,'Tolerancia':p.tolerancia,'Máximo de iteraciones':p.max_iter,'Iteraciones realizadas':r.iteraciones,'Finalización':finishReason(model),'Inercia':r.inercia,'Origen del cálculo':r.origen,'Fuente de consulta':model.source})) {
  const pair=element('div'); pair.append(element('dt',label),element('dd',String(value)));facts.append(pair);
 }
 root.append(facts,element('h3','Centroides y tamaño de los grupos'),element('p',`${data.scale}. ${data.unit}.`));
 const wrap=element('div',undefined,'table-wrap'),table=element('table',undefined,'data-table');
 table.append(element('caption','Valores devueltos por la API; cada grupo conserva su orden original.'));
 const head=element('thead'),row=element('tr');
 for(const title of ['Grupo','PM10','PM2.5','Observaciones']){const th=element('th',title);th.scope='col';row.append(th);}head.append(row);table.append(head);
 const body=element('tbody');r.centroides.forEach((c,i)=>{const tr=element('tr');for(const v of [i+1,...c,r.conteos[i]])tr.append(element('td',String(v)));body.append(tr);});table.append(body);wrap.append(table);root.append(wrap);
 root.append(element('h3','Procedencia y alcance'),element('p',data.source),element('p',data.coverage),element('p',`Conjunto: ${model.datasetId}`,'result-id'),element('p','Los grupos describen similitud entre mediciones. No son categorías sanitarias ni cantidades de personas.'));
 if(model.notice)root.append(element('p',model.notice,'message'));
 return {root,data};
}
