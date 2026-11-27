"""Resúmenes anuales; un año sin registros no se convierte en cero."""
from resumen import STATS

def por_anio(con):
    return con.sql(f'''SELECT periodo, contaminante, year(fecha) AS anio, {STATS}
      FROM mediciones WHERE fecha IS NOT NULL
      GROUP BY periodo,contaminante,anio ORDER BY periodo,contaminante,anio''')
