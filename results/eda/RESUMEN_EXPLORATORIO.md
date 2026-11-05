# Resumen del análisis exploratorio

La fuente contiene **11,407,394 registros y 23 variables**. Se seleccionaron **1,580,630 mediciones de PM10 y PM2.5** y se conservaron **1,137,870** después de aplicar las reglas documentadas. El conjunto preparado comprende **2018-01-01 a 2024-12-31**, sin filtro por antigüedad.

Se conservaron **645,346 mediciones de PM10** y **492,524 de PM2.5**. La proporción conservada fue **71.99 %** del subconjunto seleccionado. La eliminación de duplicados en la clave documentada retiró **0 registros**.

Los promedios, medianas, cuartiles, desviaciones y percentiles se presentan por contaminante y periodo en `23_concentraciones_por_periodo.csv`. Los registros no cuantificables y sus códigos se describen en `19_codigos_no_cuantificables_PM.csv`. Las irregularidades temporales y espaciales se cuantifican en `27_temporalidad_y_ubicacion_limpias.csv`.

Las diferencias entre años describen cobertura; los extremos IQR y P99 son señales para revisar registros y no criterios de eliminación automática. Los valores cero no demuestran ausencia de contaminación. Las coordenadas y el datum requieren contraste con la fuente antes de elaborar mapas. Las estadísticas descriptivas no acreditan capacidad predictiva ni cumplimiento de umbrales de salud.
