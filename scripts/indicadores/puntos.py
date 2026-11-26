"""Identidad contextual: el nombre del punto por sí solo no identifica una estación."""
from resumen import STATS

KEYS = 'informe,evaluacion,punto,este,norte,zona,datum'
def por_punto(con):
    return con.sql(f'''SELECT {KEYS}, periodo, contaminante, {STATS},
      min(fecha) AS fecha_inicial, max(fecha) AS fecha_final,
      count(*) FILTER (WHERE datum_reconocido IS NOT TRUE) AS datum_no_verificado,
      count(*) FILTER (WHERE punto IS NULL OR este IS NULL OR norte IS NULL OR zona IS NULL OR datum IS NULL) AS ubicacion_incompleta
      FROM mediciones GROUP BY {KEYS},periodo,contaminante
      ORDER BY {KEYS},periodo,contaminante''')
