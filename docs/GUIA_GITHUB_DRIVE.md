# Publicación de la entrega en GitHub y Drive

Responsable: Valentina, integrante 3, después del intercambio con Nicole.

## GitHub

La entrega contiene **33 commits propios: 10 en N1, 11 en N2 y 12 en N3** sobre la base de `develop`:

| Entregable | Rama | Commits propios |
|---|---|---:|
| N1: concurrencia | `feature/kmeans-concurrente` | 10 |
| N2: integración y pruebas | `test/integracion-kmeans` | 11 |
| N3: rendimiento, informe y guía de publicación | `perf/comparacion-kmeans` | 12 |

Las ramas son dependientes. La última contiene los 33 commits acumulados; no deben sumarse los conteos acumulados de las tres ramas como si fueran aportes diferentes.

1. Publicar las tres ramas del repositorio local.
2. Crear y revisar los PR hacia `develop` en orden N1, N2 y N3.
3. Usar los títulos y descripciones de [pull-requests-integrante-3.md](pull-requests-integrante-3.md).
4. Antes de integrar cada PR, comprobar que su diff no repite cambios ya integrados. Si el equipo usa squash, actualizar las ramas dependientes antes de continuar.

No es necesario subir el ZIP de respaldo dentro del repositorio. Los archivos deben conservar sus rutas `go/`, `scripts/`, `docs/` y `results/`. Las carpetas del IDE, los ejecutables y los entornos locales no forman parte de esta entrega.

## Drive

Dentro de la carpeta `Concurrente`, conservar los datos y cuadernos existentes y subir la carpeta `Integrante_3_Valentina` preparada para la entrega:

| Carpeta | Contenido |
|---|---|
| `01_Informe` | Texto de las secciones 3.3–3.5, conclusiones, recomendaciones y figura |
| `02_Pruebas` | Comandos, salidas y comprobaciones de integración y carreras |
| `03_Rendimiento` | Mediciones crudas, entorno, resumen y gráficos |
| `04_Codigo_y_Colab` | Respaldo del código y cuaderno complementario |
| `05_Enlaces` | Enlaces reales al repositorio, informe y PR publicados |

El historial se reorganizó sin añadir funcionalidades ni modificar código, pruebas o mediciones. Los hashes de las evidencias corresponden a las ejecuciones originales y se conservan sin alterarlos; el historial anterior está respaldado en `historial_antes_10_11_12.bundle`. El paquete de entrega actualizado conserva las tres ramas con 10, 11 y 12 commits propios, respectivamente.

## Informe y scripts

Incorporar [informe-integrante-3.md](informe-integrante-3.md) por secciones en el informe principal, conservando el trabajo de las otras integrantes. Reemplazar los pendientes de 3.3–3.5, añadir conclusiones y recomendaciones, actualizar la responsable e insertar la figura en 3.5.2.

Los scripts Python se ejecutan desde la raíz del repositorio y deben permanecer dentro de `scripts/`:

- `exportar_csv_limpio.py`: exporta el Parquet limpio para validar las mediciones.
- `exportar_csv_vectores.py`: exporta vectores completos y metadatos; no entrena el modelo.
- `verificar_integracion.py`: ejecuta pruebas y comprueba los conteos de carga; `--race` requiere un compilador C compatible.
- `medir_kmeans.py`: ejecuta las repeticiones comparables y conserva resultados en una carpeta nueva.
- `graficar_rendimiento.py`: genera las figuras desde el resumen medido.

No mezclar los tiempos Windows incluidos con nuevas ejecuciones de Colab. El cuaderno complementario se entregó sin ejecutar en Colab. Publicar ramas, crear PR y copiar archivos al Drive son acciones separadas; este documento no acredita que ya se hayan realizado.
