# Resumen de la limpieza

La fuente contiene **11,407,394 registros**. Se seleccionaron **1,580,630 mediciones de PM10 y PM2.5** y se conservaron **1,062,897** después de las exclusiones, la normalización y el control de claves.

El conjunto conserva **598,502 mediciones de PM10** y **464,395 de PM2.5** entre **2018-01-01 y 2024-12-31**, sin filtro de antigüedad. La proporción conservada es **67.25 %** del subconjunto PM.

Se retiraron **0 duplicados** en la clave documentada y **0 mediciones con clave ambigua**. Las exclusiones de PM suman **517,733 registros** y se conservan con su motivo en la auditoría. El balance satisface entrada PM = exclusiones PM + mediciones finales.

Persisten **109,288 mediciones con datum pendiente de revisión** y **1,025,663 sin final interpretable**. No se ha inventado información para completarlas. Los extremos y los ceros se conservan. Los promedios se describen por contaminante y periodo.

El manifiesto `Limpieza_PM/ultima_ejecucion.json` identifica los archivos que utilizará la preparación de vectores. Cada fila de estos archivos sigue siendo una medición individual.
