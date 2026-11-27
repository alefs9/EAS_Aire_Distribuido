# Despliegue y evaluación web

## Datos

Los datos reales se montan desde `data/vectors/` en `/data/vectors`, como solo lectura. La carpeta debe contener `features_1_hora.csv`, `features_24_horas.csv` y `manifest.json`. Los CSV deben tener el encabezado `PM10,PM2.5`; antes de usarlos, verifica filas y SHA-256 contra el manifiesto.

La fuente de los datos originales es la [ficha de OEFA sobre vigilancia y seguimiento de la calidad del aire](https://www.datosabiertos.gob.pe/dataset/vigilancia-y-seguimiento-ambiental-en-la-calidad-del-aire-organismo-de-evaluaci%C3%B3n-y). El manifiesto local no verifica `source`; por eso ese campo no se publica en la API.

## Versiones y puertos

Imágenes declaradas: `golang:1.27-alpine`, `alpine:3.22`, `mongo:8.0` y `redis:7.4-alpine`. En cada campaña guarda también los IDs y digests de imagen que Docker usó.

| Servicio | Puerto | Acceso |
|---|---:|---|
| UI | 8090 | Host local, `127.0.0.1` |
| API HTTP | 8080 | Red interna de Compose |
| Coordinador TCP | 9000 | Red interna de Compose |
| MongoDB | 27017 | Red interna de Compose |
| Redis | 6379 | Red interna de Compose |
| Nodos ML | — | Red interna de Compose |

## Inicio y parada

Desde la raíz del repositorio:

```powershell
docker compose -f go/plataforma/compose.yml config
docker compose -f go/plataforma/compose.yml build
docker compose -f go/plataforma/compose.yml up -d
docker compose -f go/plataforma/compose.yml ps
Invoke-RestMethod http://127.0.0.1:8090/api/health
Invoke-RestMethod http://127.0.0.1:8090/api/datasets
Start-Process http://127.0.0.1:8090/datos
docker compose -f go/plataforma/compose.yml logs api nodo-1 nodo-2
docker compose -f go/plataforma/compose.yml down
```
