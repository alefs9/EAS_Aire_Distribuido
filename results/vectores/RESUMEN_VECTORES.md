# Resumen de la preparación de vectores

La entrada contiene **1,062,897 mediciones limpias**. La coincidencia por la clave documentada produjo **463,644 vectores [PM10, PM2.5]**, que utilizan **927,288 mediciones**. Las **135,609 mediciones restantes** no tienen el otro contaminante en la misma clave y se conservaron fuera de los pares, sin imputación.

Los vectores comprenden **2019-01-25 a 2024-12-31**, sin filtro de antigüedad. Se exportaron **445,347 vectores de una hora** y **18,297 de 24 horas**, en archivos separados para seleccionar el periodo del modelo.

Las concentraciones permanecen en **µg/m³**, sin escalado ni réplica. Se comprobaron **0 claves repetidas por contaminante**. **54,644 vectores** conservan un datum pendiente de revisión; esta condición se mantiene en sus metadatos.

La preparación no entrena el modelo ni demuestra capacidad predictiva. El número de vectores se informa separadamente del volumen de mediciones procesadas durante la limpieza.
