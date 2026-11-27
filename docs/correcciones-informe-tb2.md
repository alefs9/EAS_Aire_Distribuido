# Correcciones propuestas para el informe TB2

Textos propuestos a partir de la campaña local del 27 de noviembre de 2026. No se modificó Google Docs. Las cifras siguientes corresponden a una comprobación funcional, no a una nueva campaña comparativa de rendimiento.

## Actualización de 5.1: catálogo y panel

La interfaz web permite consultar el catálogo publicado por la API mediante `GET /datasets`. En el entorno verificado se recibieron dos conjuntos: 445 347 pares completos PM10–PM2.5 de una hora y 18 297 pares de 24 horas. La selección se conserva al pasar a la configuración de una ejecución y vuelve a verificarse contra el catálogo antes de crear el trabajo. El navegador consulta metadatos y no descarga el CSV masivo.

El panel general integra el estado de las dependencias, el catálogo y los trabajos conocidos durante la sesión del navegador. Las referencias se registran después de una creación aceptada por la API; se conservan hasta 50 identificadores y se consultan hasta los 10 más recientes, con concurrencia máxima de tres. Los estados se obtienen nuevamente del servicio y no se deducen de información local. Una sesión sin referencias se distingue de una consulta fallida o de un trabajo inexistente. Esta implementación no constituye un listado global del sistema. El tiempo de entrenamiento procede de `segundos_ml` y se distingue de la latencia HTTP.

## Sustitución de la evidencia de Resultados

Retirar de la descripción del estado actual la figura «Resultados en preparación». La pantalla existente se encuentra conectada a la navegación y a `GET /jobs/{id}/results`. El identificador se transfiere mediante `job_id` en la URL, se precarga al abrir el enlace y se conserva al actualizar la página. Se mantiene la consulta manual y el acceso desde el panel. La vista muestra los gráficos y parámetros del trabajo consultado; los estados pendiente, fallido e inexistente no se sustituyen por datos simulados.

En la comprobación real se completó el trabajo `pc4-86357b4c37f392421dd8a01d00956c15`, con 18 297 vectores de 24 horas, k=4, semilla 42, tolerancia 0.0001 y máximo de 100 iteraciones. Participaron nodo-1 y nodo-2. Se alcanzó convergencia en 44 iteraciones; los conteos fueron 10 075, 992, 4 848 y 2 382, cuya suma corresponde al conjunto seleccionado. El servicio informó 0.018809535 segundos de entrenamiento. Este valor corresponde a una ejecución funcional y no estima por sí solo escalabilidad ni rendimiento general.

## Resumen: incorporación de interfaz y evaluación final

Se implementó e integró una interfaz web servida por Go para consultar los conjuntos históricos de OEFA, configurar y seguir ejecuciones de K-Means distribuido y consultar sus resultados y perfiles ambientales. La revisión de integración corrigió el registro de vistas, el traspaso del identificador entre pantallas y la consulta de referencias de trabajos en el panel. Se verificó el recorrido con la API, dos nodos TCP, MongoDB y Redis reales mediante Docker Compose. La evaluación automatizada incluyó 60 pruebas unitarias de JavaScript y 12 pruebas de navegador con respuestas sintéticas, además de pruebas Go, revisión estática y detector de carreras en Linux. Las respuestas sintéticas se utilizaron para comprobar regresiones y estados de error, sin presentarlas como ejecución científica real.

## Recomendaciones y numeración

Sustituir la recomendación de «completar Resultados» por: «Ampliar las comprobaciones de la interfaz a más combinaciones de parámetros y periodos, fortalecer la accesibilidad y evaluar las condiciones necesarias para un despliegue fuera del laboratorio. Mantener la trazabilidad entre identificadores, datos y resultados, así como la distinción entre latencia HTTP y tiempo de entrenamiento».

Corregir la referencia de recomendaciones TB2 de 7.2 a **7.3**, verificando su correspondencia con el índice y la numeración definitiva del informe.

## Capturas que deben incorporarse

Directorio: `results/evaluacion-web/2026-11-27-integracion-web-tb2-110325/validacion-entrega/capturas/`. La nota de campaña contiene comandos y revisión exacta. Todas las imágenes de este directorio utilizan la aplicación normal y servicios reales; el inicio y el menú no contienen mediciones simuladas.

| Archivo | Título propuesto | Nota propuesta |
| --- | --- | --- |
| 01-inicio-escritorio.png | Inicio y navegación del sistema | Interfaz en escritorio; propósito y acceso a las vistas implementadas. |
| 02-catalogo-real-horario.png | Consulta del catálogo de mediciones horarias | Metadatos del conjunto de una hora recibidos de la API. |
| 03-seleccion-real-24-horas.png | Selección del conjunto de 24 horas | Selección del periodo de 24 horas con sus metadatos reales. |
| 04-seleccion-en-ejecuciones.png | Conservación del conjunto seleccionado | La configuración recibe el conjunto elegido en Datos. Evidencia de integración con Ejecuciones, no atribución exclusiva a Alexandra. |
| 05-ejecucion-completada.png | Seguimiento del trabajo distribuido | La API confirmó que el trabajo terminó y habilitó su enlace a Resultados. |
| 06-resultado-mismo-trabajo.png | Consulta del resultado mediante su identificador | La vista recibe el mismo job_id desde Ejecuciones; reemplaza el estado obsoleto de preparación. |
| 06b-graficos-resultado-real.png | Centroides y distribución del trabajo completado | Gráficos existentes de Resultados con concentraciones históricas y conteos; no son pronósticos ni categorías sanitarias. |
| 07-panel-trabajo-verificado.png | Estado general y catálogo del servicio | Dependencias disponibles y ambos conjuntos consultados. |
| 07b-panel-referencias-y-tiempo.png | Estado y tiempo del trabajo conocido en la sesión | Un trabajo completado verificado por API; segundos de entrenamiento distintos de tiempo HTTP. |
| 08-analisis-ambiental-accesible.png | Acceso a la exploración ambiental | Vista ambiental existente accesible desde la navegación con un trabajo real. |
| 08b-perfiles-ambientales-reales.png | Perfiles de concentración del trabajo consultado | Barras de centroides del periodo de 24 horas; no atribuir esta pantalla como desarrollo exclusivo de Alexandra. |
| 09-inicio-movil.png | Inicio adaptable a dispositivo móvil | Vista de 390 × 844 píxeles; no constituye una auditoría integral de accesibilidad. |

No asignar numeración definitiva de figuras hasta consolidar el documento. Resultados, Ejecuciones y Análisis ambiental muestran integración del trabajo del equipo; conservar la autoría de las implementaciones existentes.

## Pendientes editoriales y de alcance

- Actualizar índice, referencias cruzadas y lista de figuras después de sustituir las evidencias.
- Completar títulos, notas y referencias incompletas de los anexos, distinguiendo campañas históricas y esta comprobación.
- Incorporar el enlace real del video cuando esté disponible. No se proporcionó ni se inventó un enlace.
- La predicción temporal permanece como diferencia entre los objetivos y el alcance implementado. Debe decidirse con autorización si se ajustan los objetivos o se planifica un modelo temporal. K-Means aplicado aquí agrupa mediciones históricas y no produce pronósticos.
- La verificación se realizó en un laboratorio local. No acredita autenticación, TLS, exposición pública, auditoría WCAG completa, despliegue en producción ni una nueva comparación experimental de rendimiento.
- Windows bloqueó ejecutables temporales de algunas pruebas Go. Se ejecutaron satisfactoriamente pruebas, vet y race dentro de Docker Linux, sin desactivar sus políticas de seguridad.
