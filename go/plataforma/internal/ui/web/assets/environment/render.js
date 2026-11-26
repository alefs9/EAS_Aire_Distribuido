import {periodLabels} from './datasets.js';
export function el(tag,text,className){const node=document.createElement(tag);if(text!==undefined)node.textContent=String(text);if(className)node.className=className;return node;}
export function renderProfiles(profiles) {
 const root=el('section',undefined,'environment-profiles');
 root.append(el('h2','Comparación de perfiles'),el('p','Cada grupo pertenece a su propio trabajo: G1 de A no equivale necesariamente a G1 de B. Las barras comparten escala y comparan centroides, no observaciones individuales.'));
 const max=Math.max(1,...profiles.flatMap(p=>p.groups.flatMap(g=>p.pollutants.map(key=>g[key]))));
 for(const [index,p] of profiles.entries()) {
  const card=el('section',undefined,'environment-profile');card.append(el('h3',`Trabajo ${index===0?'A':'B'} · ${periodLabels[p.dataset.period]}`),el('p',p.id,'environment-id'),el('p',`${p.dataset.scale}. Unidad: ${p.dataset.unit}. Total del conjunto: ${p.dataset.total} pares completos.`),el('p',`k=${p.parameters.k} · ${p.iterations} iteraciones · ${p.converged?'Convergencia alcanzada':'Terminó sin converger'}.`));
  if(!p.groups.length){card.append(el('p','Este trabajo no contiene el grupo seleccionado.'));root.append(card);continue;}
  const chart=el('div',undefined,'environment-bars');chart.append(el('p',`Eje horizontal: concentración del centroide, de 0 a ${max} µg/m³. Eje vertical: grupo y contaminante.`));
  const table=el('table',undefined,'data-table');table.append(el('caption','Valores exactos del perfil seleccionado'));
  const head=el('thead'),tr=el('tr');for(const label of ['Grupo',...p.pollutants.map(k=>k==='pm10'?'PM10 (µg/m³)':'PM2.5 (µg/m³)'),'Observaciones']){const th=el('th',label);th.scope='col';tr.append(th);}head.append(tr);table.append(head);
  const body=el('tbody');
  for(const g of p.groups){const row=el('tr');for(const value of [g.group,...p.pollutants.map(key=>g[key]),g.count])row.append(el('td',value));body.append(row);
   for(const key of p.pollutants){const line=el('div',undefined,'environment-bar-row');line.append(el('span',`G${g.group} · ${key==='pm10'?'PM10':'PM2.5'}: ${g[key]} µg/m³`));const bar=el('span',undefined,`environment-bar ${key}`);bar.style.width=`${g[key]/max*100}%`;bar.setAttribute('aria-hidden','true');line.append(bar);chart.append(line);}
  }
  table.append(body);const wrap=el('div',undefined,'table-wrap');wrap.append(table);card.append(chart,wrap,el('p',`Conjunto: ${p.datasetId}`,'environment-id'));root.append(card);
 }
 root.append(el('h3','Cobertura e interpretación'),el('p','Fuente: OEFA · Evaluaciones ambientales de seguimiento. Se usan pares completos PM10–PM2.5. El filtro de grupo oculta perfiles; no vuelve a calcular centroides ni representa un filtro de mediciones.'),el('p','La consulta disponible no entrega fechas ni puntos de monitoreo. No se atribuyen ubicaciones ni tendencias temporales. Los grupos no son categorías sanitarias, población expuesta ni pronósticos.'));
 return root;
}
