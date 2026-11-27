import {chromium,expect} from '@playwright/test';
import {mkdir,writeFile} from 'node:fs/promises';
import {resolve,join} from 'node:path';
import {execFileSync} from 'node:child_process';
const output=resolve(process.argv[2] || 'test-results/real-compose');
await mkdir(join(output,'capturas'),{recursive:true});
const browser=await chromium.launch({channel:process.env.PLAYWRIGHT_CHANNEL || 'msedge',headless:true});
const context=await browser.newContext({viewport:{width:1280,height:900}});
const page=await context.newPage();const errors=[];page.on('pageerror',e=>errors.push(e.message));
const base=process.env.TB2_UI_URL || 'http://127.0.0.1:8090';
const report={kind:'API, nodos TCP, MongoDB y Redis reales; sin fixtures',started:new Date().toISOString(),revision:execFileSync('git',['rev-parse','HEAD'],{encoding:'utf8'}).trim(),base,desktop:{width:1280,height:900},mobile:{width:390,height:844},checks:[]};
const shot=async name=>{await page.screenshot({path:join(output,'capturas',name),animations:'disabled'});};
const nav=name=>page.getByRole('navigation',{name:'Navegación principal'}).getByRole('link',{name,exact:true}).click();
try {
 await page.goto(base);await expect(page.locator('main h1')).toContainText('Los datos del aire');await shot('01-inicio-escritorio.png');
 const response=await context.request.get(`${base}/api/datasets`);expect(response.ok()).toBe(true);report.catalog=await response.json();expect(report.catalog.datasets).toHaveLength(2);
 await nav('Datos');await expect(page.getByRole('radio',{name:'Una hora',exact:true})).toBeVisible();await page.getByRole('radio',{name:'Una hora',exact:true}).check();await shot('02-catalogo-real-horario.png');
 await page.getByRole('radio',{name:'24 horas',exact:true}).check();await shot('03-seleccion-real-24-horas.png');
 await page.getByRole('link',{name:'Continuar a ejecuciones',exact:true}).click();await expect(page.getByRole('heading',{name:'Conjunto verificado',exact:true})).toBeVisible();await shot('04-seleccion-en-ejecuciones.png');
 await page.getByLabel('Cantidad de grupos (k)',{exact:true}).fill('4');
 const accepted=page.waitForResponse(r=>r.request().method()==='POST' && new URL(r.url()).pathname==='/api/jobs');
 await page.getByRole('button',{name:'Iniciar ejecución',exact:true}).click();const created=await accepted;expect(created.status()).toBe(202);report.created=await created.json();report.jobId=report.created.job_id;
 const link=page.getByRole('link',{name:'Ver detalle de resultados'});await expect(link).toBeVisible({timeout:180000});await expect(link).toHaveAttribute('href',`/resultados?job_id=${report.jobId}`);await shot('05-ejecucion-completada.png');await link.click();
 await expect(page.locator('.result-detail > .result-id').first()).toHaveText(report.jobId);await expect(page.locator('.result-charts svg')).toBeVisible();await shot('06-resultado-mismo-trabajo.png');await page.locator('.result-charts').scrollIntoViewIfNeeded();await shot('06b-graficos-resultado-real.png');await page.reload();await expect(page.locator('.result-detail > .result-id').first()).toHaveText(report.jobId);
 const result=await context.request.get(`${base}/api/jobs/${report.jobId}/results`);expect(result.status()).toBe(200);report.result=await result.json();
 await nav('Panel general');await expect(page.getByRole('row').filter({hasText:report.jobId})).toContainText('Completado');await expect(page.locator('[data-panel-metrics]')).toContainText('s');await shot('07-panel-trabajo-verificado.png');await page.getByRole('row').filter({hasText:report.jobId}).scrollIntoViewIfNeeded();await shot('07b-panel-referencias-y-tiempo.png');await page.reload();await expect(page.getByRole('row').filter({hasText:report.jobId})).toContainText('Completado');
 await nav('Análisis ambiental');await expect(page.locator('main h1')).toContainText('Exploración y comparación ambiental');await page.getByLabel('Trabajo A',{exact:true}).fill(report.jobId);await page.getByLabel('Periodo',{exact:true}).selectOption('24_horas');await page.getByRole('button',{name:'Consultar perfiles',exact:true}).click();await expect(page.locator('.environment-bars').first()).toBeVisible();await shot('08-analisis-ambiental-accesible.png');await page.locator('.environment-bars').first().scrollIntoViewIfNeeded();await shot('08b-perfiles-ambientales-reales.png');
 await page.setViewportSize(report.mobile);await page.goto(base);await expect(page.locator('main h1')).toContainText('Los datos del aire');await shot('09-inicio-movil.png');
 expect(errors).toEqual([]);report.checks=['Inicio y navegación','Catálogo real de ambos periodos','Selección conservada en ejecuciones','POST 202 y seguimiento hasta completado','Mismo job_id en Resultados, enlace directo y recarga','Panel consulta y revalida referencias de sesión','Análisis ambiental existente con trabajo real','Inicio móvil'];report.success=true;
} catch(error){report.success=false;report.error=String(error);await shot('error-recorrido-real.png').catch(()=>{});throw error;}
finally {report.pageErrors=errors;report.finished=new Date().toISOString();await writeFile(join(output,'recorrido-real.json'),JSON.stringify(report,null,2));await browser.close();}
