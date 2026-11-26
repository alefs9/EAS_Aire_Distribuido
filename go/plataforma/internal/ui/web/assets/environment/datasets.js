// Identidades verificadas del manifiesto de vectores (LF y checkout Windows CRLF).
const datasets = [
 ['1_hora',445347,'2996aeea838e11ebb0613034ceabd7a51c0a7df9ba1a83f8657f908bfb267204','cd6e82224deebf069ee6e247e12f2096835b2af747c8c2d27f5fed48609162d0'],
 ['24_horas',18297,'5866449404c20a23e52b65f94871a365c7979b1fcf0abd7ed91554d70ea609a6','e30a4f4c1955507d3f9f3d419ed64f65b30b44d6a2fc87c44b70478aa75c1161'],
];
export const periodLabels={'1_hora':'Una hora','24_horas':'24 horas'};
export function identifyDataset(id) {
 for(const [period,total,...hashes] of datasets)if(hashes.some(hash=>id===`${hash}:${period}`))return {period,total,unit:'µg/m³',scale:'Concentraciones originales, sin normalizar',source:'OEFA · Evaluaciones ambientales de seguimiento'};
 throw new Error('No se verificaron el periodo, las unidades y la escala de este conjunto. No se puede comparar.');
}
