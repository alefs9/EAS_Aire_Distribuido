"""Estadísticos descriptivos sin mezclar contaminantes ni periodos."""
STATS = '''count(*) AS n_observaciones, min(concentracion) AS minimo,
 quantile_cont(concentracion,0.25) AS p25, median(concentracion) AS mediana,
 avg(concentracion) AS media, quantile_cont(concentracion,0.75) AS p75,
 quantile_cont(concentracion,0.95) AS p95, max(concentracion) AS maximo,
 count(*) FILTER (WHERE concentracion=0) AS valores_cero'''

def por_periodo(con):
    return con.sql(f'''SELECT periodo, contaminante, {STATS},
      min(fecha) AS fecha_inicial, max(fecha) AS fecha_final,
      count(*) FILTER (WHERE fecha IS NULL) AS fechas_ausentes
      FROM mediciones GROUP BY periodo,contaminante ORDER BY periodo,contaminante''')
