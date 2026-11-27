import {test,expect} from '@playwright/test';
let errors;
test.beforeEach(async({page})=>{errors=[];page.on('pageerror',error=>errors.push(error.message));});
test.afterEach(()=>expect(errors).toEqual([]));

test('loads the actual entry module and renders all registered routes',async({page})=>{
 await page.goto('/');await expect(page.locator('main h1')).toContainText('Los datos del aire');
 for(const [name,title] of [['Datos','Elige las mediciones'],['Ejecuciones','Configura tu ejecución'],['Panel general','Panel general'],['Resultados','Resultados de K-Means'],['Análisis ambiental','Exploración y comparación ambiental'],['Inicio','Los datos del aire']]){
  await page.getByRole('navigation',{name:'Navegación principal'}).getByRole('link',{name,exact:true}).click();await expect(page.locator('main h1')).toContainText(title);
 }
});

test('opens a completed result directly, preserves charts and survives reload',async({page})=>{
 await page.goto('/resultados?job_id=fixture-completed');await expect(page.locator('#result-job')).toHaveValue('fixture-completed');await expect(page.locator('.result-detail > .result-id').first()).toHaveText(/Ejecución \d+/);await expect(page.locator('.result-charts svg')).toBeVisible();
 await page.reload();await expect(page.locator('.result-detail > .result-id').first()).toHaveText(/Ejecución \d+/);
});

for(const [id,title] of [['fixture-pending','Trabajo en ejecución'],['fixture-failed','El trabajo falló'],['fixture-missing','Trabajo no encontrado']])test(`shows ${id} without invented results`,async({page})=>{
 await page.goto(`/resultados?job_id=${id}`);await expect(page.getByRole('heading',{name:title,exact:true})).toBeVisible();await expect(page.locator('.result-detail')).toHaveCount(0);await expect(page.locator('.result-charts')).toHaveCount(0);
});

test('rejects ambiguous links and clears previous results when choosing another execution',async({page})=>{
 await page.addInitScript(()=>sessionStorage.setItem('eas-air:known-jobs:v1',JSON.stringify({version:1,ids:['fixture-completed','fixture-second']})));
 await page.goto('/resultados?job_id=a&job_id=b');await expect(page.getByRole('heading',{name:'No pudimos completar la consulta'})).toBeVisible();await expect(page.locator('.result-detail')).toHaveCount(0);
 await page.locator('#result-job').selectOption('fixture-completed');await page.getByRole('button',{name:'Consultar',exact:true}).click();await expect(page.locator('.result-detail')).toBeVisible();
 await page.locator('#result-job').selectOption('fixture-second');await expect(page.locator('.result-detail')).toHaveCount(0);await page.getByRole('button',{name:'Consultar',exact:true}).click();await expect(page.locator('.result-detail > .result-id').first()).toHaveText(/Ejecución \d+/);await expect(page).toHaveURL(/job_id=fixture-second/);
});

test('handles a new query on the same SPA route and supports history',async({page})=>{
 await page.goto('/resultados?job_id=fixture-completed');await expect(page.locator('.result-detail')).toBeVisible();
 // A test-only link exercises the same delegated handler as execution/panel links.
 await page.evaluate(()=>{const a=document.createElement('a');a.textContent='Enlace de prueba a otro resultado';a.href='/resultados?job_id=fixture-second';a.dataset.route='';document.body.append(a);});
 await page.getByRole('link',{name:'Enlace de prueba a otro resultado'}).click();await expect(page).toHaveURL(/job_id=fixture-second/);await expect(page.locator('.result-detail > .result-id').first()).toHaveText(/Ejecución \d+/);
 await page.goBack();await expect(page.locator('.result-detail > .result-id').first()).toHaveText(/Ejecución \d+/);
});

test('cancels pending reads on navigation and ignores their late results',async({page})=>{
 let aborted=false;page.on('requestfailed',request=>{if(request.url().includes('/jobs/fixture-slow/results'))aborted=true;});
 await page.goto('/resultados?job_id=fixture-slow');await expect(page.getByRole('heading',{name:'Cargando información'})).toBeVisible();await page.getByRole('navigation',{name:'Navegación principal'}).getByRole('link',{name:'Inicio',exact:true}).click();await expect(page.locator('main h1')).toContainText('Los datos del aire');await expect.poll(()=>aborted).toBe(true);await expect(page.locator('.result-detail')).toHaveCount(0);
});

test('created work reaches results and session references feed the real panel provider',async({page})=>{
 await page.goto('/datos');await page.getByRole('radio',{name:'Una hora',exact:true}).check();await page.getByRole('link',{name:'Continuar a ejecuciones',exact:true}).click();await page.getByLabel('Cantidad de grupos (k)',{exact:true}).fill('2');await page.getByRole('button',{name:'Iniciar ejecución',exact:true}).click();
 const link=page.getByRole('link',{name:'Ver detalle de resultados'});await expect(link).toBeVisible();await expect(link).toHaveAttribute('href','/resultados?job_id=fixture-created');await link.click();await expect(page.locator('.result-detail > .result-id').first()).toHaveText(/Ejecución \d+/);
 await page.getByRole('navigation',{name:'Navegación principal'}).getByRole('link',{name:'Panel general',exact:true}).click();await expect(page.getByRole('row').filter({has:page.locator('a[href="/resultados?job_id=fixture-created"]')})).toContainText('Completado');await expect(page.locator('[data-panel-metrics]')).toContainText('0.25 s');
 await page.reload();await expect(page.getByRole('row').filter({has:page.locator('a[href="/resultados?job_id=fixture-created"]')})).toContainText('Completado');await page.getByRole('link',{name:'Consultar resultados de Ejecución 1',exact:true}).click();await expect(page).toHaveURL(/job_id=fixture-created/);await expect(page.locator('.result-detail > .result-id').first()).toHaveText(/Ejecución \d+/);
});

test('empty sessions and failed reads remain different, without local confirmed states',async({page})=>{
 await page.goto('/panel');await expect(page.getByRole('heading',{name:'Sin trabajos para consultar'})).toBeVisible();await expect(page.getByText('Trabajos sin conectar',{exact:true})).toHaveCount(0);
 await page.evaluate(()=>sessionStorage.setItem('eas-air:known-jobs:v1',JSON.stringify({version:1,ids:['fixture-missing'],estado:'completed'})));await page.reload();await expect(page.getByRole('cell',{name:'No encontrado',exact:true})).toBeVisible();await expect(page.locator('.panel-job-counts')).toHaveCount(0);
 await page.route('**/api/datasets',route=>route.fulfill({status:503,contentType:'application/json',body:'{"codigo":"no_disponible"}'}));await page.getByRole('navigation',{name:'Navegación principal'}).getByRole('link',{name:'Datos',exact:true}).click();await expect(page.getByRole('heading',{name:'No pudimos completar la consulta'})).toBeVisible();await expect(page.getByRole('radio')).toHaveCount(0);
});

test('a failing view does not prevent returning to another route',async({page})=>{
 await page.goto('/');await expect(page.locator('main h1')).toBeVisible();await page.evaluate(async()=>{const {registerView}=await import('/assets/app.js');registerView('/resultados',()=>{throw new Error('Synthetic view failure');});});await page.getByRole('navigation',{name:'Navegación principal'}).getByRole('link',{name:'Resultados',exact:true}).click();await expect(page.getByRole('heading',{name:'No pudimos completar la consulta'})).toBeVisible();await page.getByRole('navigation',{name:'Navegación principal'}).getByRole('link',{name:'Inicio',exact:true}).click();await expect(page.locator('main h1')).toContainText('Los datos del aire');
});

test('keeps the results link hidden until a running job has actual results',async({page})=>{
 await page.route('**/api/jobs/fixture-created',route=>route.fulfill({status:200,contentType:'application/json',body:JSON.stringify({trabajo:{job_id:'fixture-created',estado:'running'}})}));
 await page.goto('/datos');await page.getByRole('radio',{name:'Una hora',exact:true}).check();await page.getByRole('link',{name:'Continuar a ejecuciones',exact:true}).click();await page.getByRole('button',{name:'Iniciar ejecución',exact:true}).click();await expect(page.getByText('Estado: En ejecución',{exact:true})).toBeVisible();await expect(page.locator('a[href="/resultados?job_id=fixture-created"]')).toBeHidden();
});


test('presents readable execution names and two decimals without exposing internal codes',async({page})=>{
 await page.goto('/resultados?job_id=fixture-completed');await expect(page.locator('.result-detail')).toBeVisible();
 await expect(page.locator('.result-detail .data-table tbody tr').first()).toContainText('10.00');
 await expect(page.locator('.result-detail')).toContainText('1.00e-4');
 const text=await page.locator('main').innerText();expect(text).not.toMatch(/fixture-completed|fixture-hour-data|cluster_job|PC4/);
 await expect(page.getByRole('combobox',{name:'Ejecución',exact:true})).toHaveValue('fixture-completed');
 await expect(page.locator('#result-job option:checked')).toHaveText('Ejecución 1');
});
