"""Exclusiones explícitas y conservación de metadatos desconocidos."""
def preparar(con):
    con.execute('''CREATE TEMP VIEW clasificados AS SELECT *,
      CASE WHEN periodo IS NULL OR periodo NOT IN ('1 hora','24 horas') THEN 'periodo_no_admitido'
      WHEN "PM10" IS NULL OR "PM2.5" IS NULL OR NOT isfinite("PM10") OR NOT isfinite("PM2.5") THEN 'concentracion_ausente_o_no_finita'
      WHEN "PM10" < 0 OR "PM2.5" < 0 THEN 'concentracion_negativa'
      ELSE NULL END AS exclusion
      FROM vectores''')
    con.execute('CREATE TEMP VIEW validos AS SELECT * FROM clasificados WHERE exclusion IS NULL')
    con.execute('''CREATE TEMP VIEW mediciones AS
      SELECT *, 'PM10' AS contaminante, "PM10" AS concentracion FROM validos
      UNION ALL SELECT *, 'PM2.5' AS contaminante, "PM2.5" AS concentracion FROM validos''')
    return dict(con.sql("SELECT coalesce(exclusion,'incluido'), count(*) FROM clasificados GROUP BY exclusion ORDER BY 1").fetchall())
