# Medición de latencia de la API

Responsable: Nicole Bances (N3).

## Qué se mide

La latencia es el tiempo de ida y vuelta de una solicitud HTTP medido por el cliente. **No es el tiempo de entrenamiento:** `POST /jobs` responde 202 al registrar el trabajo, y el cálculo en el clúster se informa por separado en `resultado.segundos_ml`.

| Escenario | Solicitud | Condición |
|---|---|---|
| `salud` | `GET /health` | Incluye ping a MongoDB y Redis |
| `aceptacion` | `POST /jobs` | Registro en MongoDB y encolado; cada trabajo se ejecuta después en el clúster |
| `resultado_sin_cache` | `GET /jobs/{id}/results` | Primera lectura de 30 trabajos distintos (`fuente: MongoDB`) |
| `resultado_con_cache` | `GET /jobs/{id}/results` | 30 lecturas del mismo trabajo (`fuente: Redis`) |
| `estado` | `GET /jobs/{id}` | 30 trabajos distintos |
| `no_encontrado` | `GET /jobs/{id}` | Identificador inexistente (404 esperado) |

Para cada escenario se informan n, tamaño de respuesta, la primera solicitud por separado, mediana, p95 (rango más cercano), máximo, errores y la fuente de los datos. Un escenario cumple si el máximo es menor a 100 ms y no hay errores. Los tiempos individuales quedan en `mediciones.csv`.

## Cómo repetir la campaña

Con MongoDB, Redis, la API y los dos nodos en marcha (ver `docs/pc4/api.md`), desde la raíz del repositorio:

```powershell
.\scripts\api\medir-latencia.ps1 -N 30 -Periodo 24_horas
```

Los resultados se guardan en `results/api/latencia/<fecha>/`: `RESUMEN.md`, `mediciones.csv` y `entorno.txt`. Los 30 trabajos de `POST /jobs` se ejecutan de verdad en el clúster; el script espera a que terminen antes de medir las lecturas.

Además, la carpeta "3. Latencia" de la colección Postman se ejecuta con el Collection Runner (30 iteraciones) y se exporta su resultado como evidencia para el informe.

## Resultados

Ver la carpeta de cada campaña en `results/api/latencia/`. En el informe se reportan todas las solicitudes, incluidas las fallidas.
````

```powershell
git add docs/pc4/latencia.md results/api/latencia
git commit -m "docs(api): registrar la campaña de latencia y su análisis"
git push -u origin test/api-latencia