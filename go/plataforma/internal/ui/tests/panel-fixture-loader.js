// Only panel_browser_fixture.go serves this script. Production never imports it.
import {configurePanelCatalogLoader} from '/assets/panel/catalog-summary.js';
import {configurePanelJobIdsProvider} from '/assets/panel/jobs.js';
const mode=new URLSearchParams(location.search).get('fixture');
if(mode==='full'){
 configurePanelCatalogLoader(async()=>[{id:'fixture-hour',period:'1_hora',vectorCount:12},{id:'fixture-day',period:'24_horas',vectorCount:2}]);
 configurePanelJobIdsProvider(async()=>['fixture-queued','fixture-running','fixture-completed','fixture-failed']);
}
if(mode==='empty'){configurePanelCatalogLoader(async()=>[]);configurePanelJobIdsProvider(async()=>[]);}
if(mode==='error'){configurePanelCatalogLoader(async()=>{throw new Error('Synthetic failure');});configurePanelJobIdsProvider(async()=>{throw new Error('Synthetic failure');});}
