// Verified CSV fingerprints from data/vectors; never infer units for an unknown dataset.
const known={
 '2996aeea838e11ebb0613034ceabd7a51c0a7df9ba1a83f8657f908bfb267204:1_hora':{period:'Una hora',count:445347,file:'features_1_hora.csv'},
 '5866449404c20a23e52b65f94871a365c7979b1fcf0abd7ed91554d70ea609a6:24_horas':{period:'24 horas',count:18297,file:'features_24_horas.csv'},
};
// Los mismos CSV con CRLF al descargarlos con core.autocrlf en Windows.
known['cd6e82224deebf069ee6e247e12f2096835b2af747c8c2d27f5fed48609162d0:1_hora']=known['2996aeea838e11ebb0613034ceabd7a51c0a7df9ba1a83f8657f908bfb267204:1_hora'];
known['e30a4f4c1955507d3f9f3d419ed64f65b30b44d6a2fc87c44b70478aa75c1161:24_horas']=known['5866449404c20a23e52b65f94871a365c7979b1fcf0abd7ed91554d70ea609a6:24_horas'];
export function provenance(model) {
 const data=known[model.datasetId];
 if(data && model.total!==data.count) throw new Error('Los conteos no coinciden con el conjunto identificado.');
 return data ? {...data,unit:'µg/m³',scale:'Concentraciones originales, sin normalizar',source:'OEFA · Evaluaciones ambientales de seguimiento (EAS)',coverage:'Pares completos PM10–PM2.5 del periodo seleccionado. Estos vectores no incluyen fechas ni ubicaciones; no representan población expuesta.'} : {period:'No verificado',unit:'Unidad no informada',scale:'Escala no verificada: no interpretar como concentración ni como dato normalizado',source:'Procedencia no verificada',coverage:'La API no entrega metadatos suficientes para determinar unidades, fechas o ubicaciones de este conjunto.'};
}
