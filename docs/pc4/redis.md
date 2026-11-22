# Redis y pruebas

Responsable: Valentina Montenegro (A3).

Redis guarda resultados completados durante el TTL configurado. Si la copia no existe o vence, se recupera MongoDB. Si Redis no responde, se consulta MongoDB. Los trabajos pendientes y fallidos no se presentan como resultados completados.

## Ejecutar
Desde la raíz del repositorio, con Docker Desktop en modo Linux y Node.js con `npx.cmd`:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\api\cerrar-evidencias.ps1
```

El script genera la referencia secuencial para ambos periodos, ejecuta la demostración con TTL de 3 s y comprueba las transiciones con `PC4_INTEGRATION=1`. La API usa su propio TTL y guarda en caché el objeto identificado por el `job_id` externo. Los resultados y logs están en la [guía de persistencia](../../results/persistencia/README.md).

## Colab
[Cuaderno ejecutado y salidas](https://colab.research.google.com/drive/1buYE0xcebFx30bH_K65zK4uSz4u_VuUb).

La copia en `notebooks/Persistencia_cache.ipynb` permite repetir el proceso. El código del paquete conserva la lógica del Colab; solo se cambia `package main` por `persistencia` y se separa el comando para que el equipo pueda importarlo en el módulo existente.

El Colab fue una demostración independiente con resultados de la referencia secuencial. La [integración HTTP y TCP](api.md) está implementada y dispone de un [reporte de ejecución con MongoDB y Redis reales](../../results/api/postman/docker-20261122-pc4/funcionales.json).
