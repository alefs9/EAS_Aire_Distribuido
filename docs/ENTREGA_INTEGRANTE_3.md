# Entrega de Valentina — integrante 3

Se desarrollaron localmente N1, N2 y N3 del plan, que originalmente figuraban como responsabilidad de Nicole. Se consultaron el informe de Drive, el repositorio del equipo y sus datos reales. No se modificó el informe compartido ni se publicaron ramas o pull requests.

## Archivos principales

| Entregable | Resultado |
|---|---|
| Informe | [Secciones 3.3.1–3.3.5, 3.4.1–3.4.3, 3.5.1–3.5.2 y aportes a conclusiones/recomendaciones](informe-integrante-3.md) |
| N1 | [K-Means concurrente](../go/kmeans_secuencial/concurrent.go), [pruebas](../go/kmeans_secuencial/concurrent_test.go) y [explicación](kmeans-concurrente.md) |
| N2 | [Comando común](../go/kmeans_secuencial/cli.go), [integración](../go/kmeans_secuencial/integration_test.go), [guía](pruebas-integracion.md) y [registros](../results/pruebas) |
| N3 | [Metodología](rendimiento.md), [50 mediciones crudas](../results/rendimiento/ejecucion-local/mediciones.csv), [tabla](../results/rendimiento/ejecucion-local/RESUMEN.md) y [figura](../results/rendimiento/ejecucion-local/tiempos.png) |
| GitHub | [Títulos, cuerpos y orden de los tres PR](pull-requests-integrante-3.md) |

## Verificación realizada

- Pruebas normales y `-race` en ambos módulos: PASS; incluye los dos datasets reales.
- Equivalencia con 1, 2, 4 y 8 workers: mismas asignaciones, 96 iteraciones horarias y 44 diarias; centroides e inercia dentro de las tolerancias documentadas.
- Carga completa: 1 062 897 aceptados y cero rechazados en ambos modos, con conteos reconciliados.
- Clon local limpio: pruebas de ambos módulos y `go vet ./...` aprobados; registro en `results/pruebas/clon-limpio.txt`.
- Rendimiento: cinco repeticiones y un calentamiento por configuración; 4 workers obtuvieron speedup ML de 2,694 para una hora y 1,919 para 24 horas. Se conserva también el resultado diario más lento con un worker concurrente.

## Organización Git

Las ramas locales son `feature/kmeans-concurrente`, `test/integracion-kmeans` y `perf/comparacion-kmeans`, con 10 commits propios en N1, 11 en N2 y 12 en N3 (33 en total), obtenidos al reorganizar los mismos cambios existentes. Se conserva la [guía de publicación en GitHub y Drive](GUIA_GITHUB_DRIVE.md). Están encadenadas porque N2 depende de N1 y N3 depende de N2. La rama activa final contiene todo el trabajo. `develop` permanece en la base descargada.

Para mantener tres PR hacia `develop`, integrar N1 antes de revisar el diff final de N2 e integrar N2 antes de revisar N3. Si el equipo usa squash, actualizar las ramas dependientes cuidadosamente antes de publicarlas; no hacer force-push sobre trabajo compartido sin coordinarlo. Los cuerpos propuestos están preparados para revisión.

Quedan como acciones externas la publicación de las tres ramas/PR, incorporar el texto al Google Doc y la revisión independiente del equipo. No se simularon esas acciones ni se registraron números de PR inexistentes.
