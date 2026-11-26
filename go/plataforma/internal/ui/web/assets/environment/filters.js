export function selectProfiles(profiles,{period,pollutant='both',group='all'}) {
 if(!['1_hora','24_horas'].includes(period)||!['both','pm10','pm25'].includes(pollutant)||!(group==='all'||/^[1-9][0-9]*$/.test(String(group))))throw new Error('Filtros no válidos.');
 if(profiles.some(profile=>profile.dataset.period!==period))throw new Error('Los trabajos deben pertenecer al periodo seleccionado. No se mezclan una hora y 24 horas.');
 return profiles.map(profile=>({...profile,groups:profile.groups.filter(item=>group==='all'||item.group===Number(group)),pollutants:pollutant==='both'?['pm10','pm25']:[pollutant]}));
}
