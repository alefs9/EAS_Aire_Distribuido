import {apiRequest} from '../api.js';
import {loadCatalog} from '../datasets/catalog-source.js';
import {getSelectedDataset} from '../datasets/handoff.js';
import {datasetSelection} from '../datasets/selection.js';
import {datasetDetails} from '../datasets/details.js';
import {errorState, loadingState} from '../states.js';

const labels = {queued: 'En cola', running: 'En ejecución', completed: 'Completada', failed: 'Fallida'};

const isAbort = error => error?.name === 'AbortError';

function field(id, labelText, type, value, min, max, step='1') {
  const wrap = document.createElement('div'); wrap.className = 'form-field';
  const label = document.createElement('label'); label.htmlFor = id; label.textContent = labelText;
  const input = document.createElement('input');
  input.id=id; input.name=id; input.type=type; input.value=String(value);
  input.required=true; input.min=String(min); input.max=String(max); input.step=step;
  const help=document.createElement('span'); help.className='field-help'; help.id=`${id}-help`;
  input.setAttribute('aria-describedby', help.id);
  wrap.append(label,input,help); return {wrap,input,help};
}

function publicError(status, code) {
  if (status === 400) return 'Revisa los parámetros: la API los rechazó.';
  if (status === 404) return 'No se encontró la ejecución solicitada.';
  if (status === 429) return 'Se alcanzó el límite temporal de ejecuciones. Intenta más tarde.';
  if (status === 409) return 'La ejecución continúa pendiente.';
  if (status === 422 || code === 'trabajo_fallido') return 'La ejecución falló. Revisa el estado y vuelve a intentarlo con el equipo disponible.';
  if (status === 503 || status === 502) return 'El servicio o una dependencia no está disponible. Intenta nuevamente más tarde.';
  return 'No se pudo completar la solicitud. Intenta nuevamente.';
}

async function checkedJSON(response) {
  let body={}; try {body=await response.json();} catch { /* El cuerpo de error puede venir vacío. */ }
  if (!response.ok) throw Object.assign(new Error(publicError(response.status, body.codigo)), {status:response.status, code:body.codigo});
  return body;
}

export function createExecutionsView({catalog=loadCatalog, selection=datasetSelection}={}) {
 return function executionsView({signal}={}) {
  const root=document.createElement('section'); root.className='execution-page page-section';
  const eyebrow=document.createElement('p'); eyebrow.className='eyebrow'; eyebrow.textContent='Ejecuciones';
  const title=document.createElement('h1'); title.id='titulo-vista'; title.tabIndex=-1; title.textContent='Configura tu ejecución.';
  const intro=document.createElement('p'); intro.className='lead'; intro.textContent='Elige los parámetros de K-Means para analizar el conjunto verificado.';
  const content=document.createElement('div'); content.className='execution-content'; content.setAttribute('aria-busy','true'); content.append(loadingState('Verificando el conjunto seleccionado.'));
  root.append(eyebrow,title,intro,content);

  let active=false;
  const controller=new AbortController();
  const abort=()=>controller.abort();
  if(signal?.aborted) abort(); else signal?.addEventListener('abort',abort,{once:true});
  const cleanup=()=>{active=false;controller.abort();signal?.removeEventListener('abort',abort);};
  signal?.addEventListener('abort',cleanup,{once:true});

  function showMessage(message, retry) {
    const box=errorState(message, retry);
    content.replaceChildren(box); content.setAttribute('aria-busy','false');
  }
  function renderForm(dataset) {
    content.replaceChildren(); content.setAttribute('aria-busy','false');
    const chosen=document.createElement('article'); chosen.className='card execution-dataset';
    const h=document.createElement('h2'); h.textContent='Conjunto verificado';
    const facts=document.createElement('dl'); facts.className='dataset-details';
    for(const [name,value] of datasetDetails(dataset)) {const dt=document.createElement('dt');dt.textContent=name;const dd=document.createElement('dd');dd.textContent=value;facts.append(dt,dd);}
    chosen.append(h,facts);
    const form=document.createElement('form'); form.className='card execution-form';
    const legend=document.createElement('h2'); legend.textContent='Parámetros del análisis';
    const fields=[
      field('k','Cantidad de grupos (k)','number',4,1,20),
      field('semilla','Semilla','number',42,Number.MIN_SAFE_INTEGER,Number.MAX_SAFE_INTEGER),
      field('tolerancia','Tolerancia','number',0.0001,0,1000000,'any'),
      field('max_iter','Máximo de iteraciones','number',100,1,1000),
      field('nodos','Nodos ML','number',2,1,8),
    ];
    fields[0].help.textContent='Entero entre 1 y 20.';
    fields[1].help.textContent='Entero seguro; valor inicial 42.';
    fields[2].help.textContent='Número finito mayor o igual a 0.';
    fields[3].help.textContent='Entero entre 1 y 1000.';
    fields[4].help.textContent='Entero entre 1 y 8.';
    const submit=document.createElement('button'); submit.type='submit'; submit.className='button button-primary'; submit.textContent='Iniciar ejecución';
    const feedback=document.createElement('div'); feedback.className='execution-feedback'; feedback.setAttribute('aria-live','polite'); feedback.setAttribute('aria-atomic','true');
    form.append(legend,...fields.map(f=>f.wrap),submit,feedback);
    content.append(chosen,form);

    form.addEventListener('submit', async event=>{
      event.preventDefault();
      if(!form.reportValidity() || submit.disabled) return;
      const values=fields.map(f=>Number(f.input.value));
      const [k,semilla,tolerancia,max_iter,nodos]=values;
      if(!Number.isSafeInteger(k)||!Number.isSafeInteger(semilla)||!Number.isFinite(tolerancia)||tolerancia<0||!Number.isSafeInteger(max_iter)||!Number.isSafeInteger(nodos)) {
        feedback.replaceChildren(errorState('Revisa los valores ingresados y sus rangos.')); return;
      }
      submit.disabled=true; feedback.replaceChildren(loadingState('Verificando el catálogo antes de enviar el trabajo.'));
      try {
        const fresh=await catalog({signal:controller.signal});
        selection.reconcile(fresh);
        const verified=getSelectedDataset();
        if(!verified || verified.id!==dataset.id || verified.period!==dataset.period || verified.sha256!==dataset.sha256) {
          selection.invalidate();
          throw new Error('El conjunto cambió o ya no puede verificarse. Vuelve a Datos y selecciona uno disponible.');
        }
        const payload={periodo:verified.period,k,semilla,tolerancia,max_iter,nodos};
        const created=await checkedJSON(await apiRequest('/jobs',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify(payload),signal:controller.signal}));
        if(typeof created.job_id!=='string'||!created.job_id) throw new Error('La API no devolvió un job_id válido.');
        await followJob(created.job_id, feedback, submit, controller.signal);
      } catch(error) {
        if(!isAbort(error)) feedback.replaceChildren(errorState(error.message || 'No se pudo crear el trabajo.',()=>form.requestSubmit()));
        submit.disabled=false;
      }
    });
  }

  async function followJob(jobId, feedback, submit, signal) {
  const FOLLOW_LIMIT_MS = 10 * 60 * 1000;
  const MAX_TRANSIENT_ERRORS = 5;
  const BASE_DELAY_MS = 1500;
  const MAX_DELAY_MS = 10000;

  const box = document.createElement('section');
  box.className = 'message';
  box.setAttribute('role', 'status');

  const h = document.createElement('h2');
  h.textContent = 'Seguimiento del trabajo';

  const id = document.createElement('p');
  id.className = 'job-id';
  id.textContent = `Identificador: ${jobId}`;

  const status = document.createElement('p');
  status.textContent = 'Estado: en cola';

  const note = document.createElement('p');
  note.textContent = 'La aceptación HTTP ya terminó; el entrenamiento se ejecuta de forma asíncrona.';

  const details = document.createElement('p');

  const refresh = document.createElement('button');
  refresh.type = 'button';
  refresh.className = 'button button-secondary';
  refresh.textContent = 'Consultar ahora';

  const resultLink = document.createElement('a');
  resultLink.className = 'button button-primary';
  resultLink.textContent = 'Ver detalle de resultados';
  resultLink.href = `/resultados?job_id=${encodeURIComponent(jobId)}`;
  resultLink.dataset.route = '';
  resultLink.hidden = true;

  box.append(h, id, status, note, details, refresh, resultLink);
  feedback.replaceChildren(box);

  let deadline = Date.now() + FOLLOW_LIMIT_MS;
  let failures = 0;
  let done = false;
  let checking = false;
  let timer;

  function pause(message) {
    clearTimeout(timer);
    done = true;
    status.textContent = 'Seguimiento pausado';
    note.textContent = message;
    refresh.textContent = 'Reanudar seguimiento';
    refresh.disabled = false;
    submit.disabled = false;
  }

  function schedule(ms) {
    clearTimeout(timer);
    timer = setTimeout(() => void check(), ms);
  }

  async function check(manual = false) {
    if (checking || signal.aborted) return;

    if (manual) {
      clearTimeout(timer);
      done = false;
      failures = 0;
      deadline = Date.now() + FOLLOW_LIMIT_MS;
      refresh.textContent = 'Consultar ahora';
    }

    if (done) return;

    if (Date.now() >= deadline) {
      pause('Se alcanzó el límite de seguimiento. Conservamos el identificador para reanudar.');
      return;
    }

    checking = true;
    refresh.disabled = true;

    try {
      const state = await checkedJSON(
        await apiRequest(`/jobs/${encodeURIComponent(jobId)}`, {signal})
      );

      const job = state.trabajo;
      if (!job || !labels[job.estado]) {
        throw new Error('La API devolvió un estado no reconocido.');
      }

      if (job.estado !== 'completed') failures = 0;
      status.textContent = `Estado: ${labels[job.estado]}`;

      if (job.estado === 'failed') {
        done = true;
        note.textContent = 'El trabajo falló. No se publicaron resultados completos.';
        resultLink.hidden = true;
        submit.disabled = false;
        return;
      }

      if (job.estado === 'completed') {
        const response = await apiRequest(
          `/jobs/${encodeURIComponent(jobId)}/results`,
          {signal}
        );

        if (response.status === 409) {
          schedule(BASE_DELAY_MS);
          return;
        }

        const result = await checkedJSON(response);
        failures = 0;
        done = true;
        resultLink.hidden = false;
        details.textContent = result.resultado?.origen
          ? `Origen del clúster: ${result.resultado.origen}`
          : 'El servicio no informó el origen del clúster.';
        note.textContent = `Resultado disponible. job_id externo: ${result.job_id || jobId}.`;
        submit.disabled = false;
        return;
      }

      schedule(BASE_DELAY_MS);
    } catch (error) {
      if (isAbort(error)) return;

      if (error.status === 404) {
        done = true;
        status.textContent = 'Estado: identificador no encontrado';
        note.textContent = error.message;
        submit.disabled = false;
        return;
      }

      failures++;

      if (failures >= MAX_TRANSIENT_ERRORS) {
        pause('El servicio no respondió tras varios intentos. Conservamos el job_id para reanudar.');
        return;
      }

      const wait = Math.min(
        BASE_DELAY_MS * (2 ** (failures - 1)),
        MAX_DELAY_MS
      );

      note.textContent = 'No se pudo consultar el estado; se volverá a intentar con una espera creciente.';
      schedule(wait);
    } finally {
      checking = false;
      refresh.disabled = false;
    }
  }

  refresh.addEventListener('click', () => void check(true));
  signal.addEventListener('abort', () => {
    clearTimeout(timer);
    done = true;
  }, {once: true});

  await check();
}

  void (async()=>{
    try {
      if(!getSelectedDataset()) throw new Error('No hay un conjunto seleccionado y verificado. Regresa a Datos para seleccionar uno.');
      const before=getSelectedDataset();
      const records=await loadCatalog({signal:controller.signal});
      selection.reconcile(records);
      const current=getSelectedDataset();
      if(!current || !before || current.id!==before.id || current.period!==before.period || current.sha256!==before.sha256) {
        throw new Error('La selección cambió o no pudo verificarse. Regresa a Datos y elige un conjunto disponible.');
      }
      if(!active || controller.signal.aborted)return;
      renderForm(current);
    } catch(error) {
      if(!isAbort(error) && !controller.signal.aborted) {
        const box=errorState(error.message || 'No se pudo verificar el conjunto.',()=>location.assign('/datos'));
        const link=document.createElement('a');link.href='/datos';link.dataset.route='';link.className='button button-secondary';link.textContent='Ir a Datos';box.append(link);
        content.replaceChildren(box);content.setAttribute('aria-busy','false');
      }
    }
  })();
  active=true;
  return root;
 };
}
export const executionsView=createExecutionsView();