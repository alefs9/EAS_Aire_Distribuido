# Newman con servicios reales — 22 de noviembre de 2026

Ejecución nueva en Windows con Docker Desktop en modo Linux, API y coordinador en el mismo proceso, dos nodos TCP, MongoDB 8.3.17 y Redis 7.4.12. Newman 6.2.3 ejecutó la colección versionada desde Node.js 24.21.0. [Entorno](entorno.txt) · [Configuración y versiones de Docker](../../../persistencia/docker-20261122-pc4/)

| Carpeta | Solicitudes | Aserciones superadas | Fallidas | Máximo HTTP | Salida | Archivos |
|---|---:|---:|---:|---:|---:|---|
| Funcionales | 13 | 36 | 0 | 34 ms | 0 | [JSON](funcionales.json) · [log](funcionales.log) |
| Latencia, 30 iteraciones | 90 | 180 | 0 | 21 ms | 0 | [JSON](latencia.json) · [log](latencia.log) |
| Dependencia no disponible | 2 | 4 | 0 | 2004 ms | 0 | [JSON](sin-mongodb.json) · [log](sin-mongodb.log) |

La carpeta funcional realizó dos entrenamientos distribuidos con tolerancia 0.0001. El horario completó 100 actualizaciones sin converger y el de 24 horas convergió en 44. La primera lectura de ambos resultados devolvió `fuente: MongoDB`; la segunda de 24 horas devolvió `fuente: Redis`. El `job_id` externo se conservó entre las respuestas y `resultado.origen` identificó otro ID como `cluster_job`. Las respuestas HTTP 202 solo confirmaron la aceptación.

Al detener MongoDB, salud y creación devolvieron HTTP 503 en 524 y 2004 ms, respectivamente. Los tiempos y los códigos esperados se conservan en el JSON. Esta prueba corresponde a una interrupción deliberada y se ejecutó por separado de las otras carpetas y de la [campaña publicada del cliente Go](../../latencia/20261121-1230/RESUMEN.md).

Para reproducir, desde la raíz con Docker Desktop en modo Linux:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\api\cerrar-evidencias.ps1
```

El script ejecuta primero la carpeta funcional y exporta sus variables para la carpeta de latencia. Luego detiene únicamente su MongoDB aislado, ejecuta la carpeta de indisponibilidad y vuelve a iniciar el servicio. Los logs guardan los comandos y códigos de salida. Los reportes de esta carpeta son archivos nativos del reportero JSON de Newman.

Si aparece la exportación histórica del Collection Runner, guárdala como `results/api/postman/runner-historico/runner-original.json` con su fecha y entorno. Para generar una exportación **nueva** desde Postman Desktop, importa la colección y el entorno versionados, selecciona el entorno, ejecuta solo «1. Funcionales» con persistencia de variables habilitada y verifica que `job_id` contenga el trabajo completado. Ejecuta solo «3. Latencia» con 30 iteraciones, selecciona **Export Results** y guarda el archivo en `results/api/postman/runner-desktop-nuevo/runner.json`. Usa el entorno de servicios activos descrito en la [guía de API](../../../../docs/pc4/api.md).
