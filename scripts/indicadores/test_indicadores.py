import unittest
import tempfile
from pathlib import Path
import duckdb
from calidad import preparar
from resumen import por_periodo
from temporal import por_anio
from puntos import por_punto
from exportar import exportar

class IndicadoresTest(unittest.TestCase):
    def setUp(self):
        self.con = duckdb.connect()
        self.con.execute('''CREATE TABLE vectores AS SELECT * FROM (VALUES
          (1,'A','E','P',1.,2.,17,'WGS84',DATE '2020-01-01','1 hora',10.,2.,true),
          (2,'A','E','P',1.,2.,17,'WGS84',DATE '2020-01-02','1 hora',30.,6.,true),
          (3,'B','E','P',3.,4.,17,'WGS84',DATE '2021-01-01','24 horas',100.,20.,true),
          (4,'A','E','P',1.,2.,17,'WGS84',NULL,'1 hora',20.,4.,false),
          (5,'A','E','P',1.,2.,17,'WGS84',NULL,'1 hora',-1.,4.,true),
          (6,'A','E','P',1.,2.,17,'WGS84',NULL,'1 hora',NULL,4.,true),
          (7,'A','E','P',1.,2.,17,'WGS84',NULL,'otro',20.,4.,true)
          ) t(vector_id,informe,evaluacion,punto,este,norte,zona,datum,fecha,periodo,"PM10","PM2.5",datum_reconocido)''')
        self.audit = preparar(self.con)
    def tearDown(self):
        self.con.close()
    def test_estadisticos_y_periodos(self):
        row = por_periodo(self.con).filter("periodo='1 hora' AND contaminante='PM10'").fetchone()
        self.assertEqual(row[2:10], (3,10.,15.,20.,20.,25.,29.,30.))
        self.assertEqual(row[-1],1)
        self.assertEqual(por_periodo(self.con).filter("periodo='24 horas' AND contaminante='PM10'").fetchone()[2],1)
    def test_exclusiones(self):
        self.assertEqual(self.audit, {'incluido':4,'concentracion_negativa':1,'concentracion_ausente_o_no_finita':1,'periodo_no_admitido':1})
    def test_ceros_son_observaciones_y_no_se_imputan(self):
        self.con.execute('UPDATE vectores SET "PM10"=0 WHERE vector_id=1')
        row = por_periodo(self.con).filter("periodo='1 hora' AND contaminante='PM10'").fetchone()
        self.assertEqual(row[2],3)
        self.assertEqual(row[10],1)
    def test_fecha_desconocida_no_se_imputa(self):
        rows = por_anio(self.con).fetchall()
        self.assertEqual(sum(r[3] for r in rows),6)
        self.assertEqual({r[2] for r in rows},{2020,2021})
    def test_punto_no_se_mezcla_por_nombre(self):
        rows = por_punto(self.con).fetchall()
        self.assertEqual(len(rows),4)
        self.assertEqual({r[0] for r in rows},{'A','B'})
    def test_no_sobrescribe_evidencias(self):
        with tempfile.TemporaryDirectory() as folder:
            exportar({'resumen':por_periodo(self.con)},Path(folder),{})
            with self.assertRaises(FileExistsError):
                exportar({'resumen':por_periodo(self.con)},Path(folder),{})

if __name__ == '__main__':
    unittest.main()
