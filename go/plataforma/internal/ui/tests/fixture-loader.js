// Synthetic fixtures are installed ONLY by browser_fixture.go, never by cmd/ui.
import {configureCatalogLoader} from '/assets/datasets/catalog-source.js';
const mode=new URLSearchParams(location.search).get('fixture');
const records=[
 {id:'fixture-hour',period:'1_hora',vectorCount:12,features:['PM10','PM2.5'],unit:'µg/m³',source:'Datos sintéticos de prueba',coverage:'No corresponde a mediciones reales',sha256:'a'.repeat(64)},
 {id:'fixture-day',period:'24_horas',vectorCount:2,features:['PM10','PM2.5'],unit:'µg/m³',source:'Datos sintéticos de prueba',coverage:null,sha256:'b'.repeat(64)},
];
if(mode==='two')configureCatalogLoader(async()=>records);
if(mode==='empty')configureCatalogLoader(async()=>[]);
if(mode==='error')configureCatalogLoader(async()=>{throw new Error('Error sintético de prueba');});
if(mode==='invalid')configureCatalogLoader(async()=>[{id:'fixture-invalid',period:'mensual'}]);
