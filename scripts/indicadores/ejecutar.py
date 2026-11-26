"""python scripts/indicadores/ejecutar.py --salida results/indicadores/ejecucion"""
import argparse
from pathlib import Path
import duckdb
from fuente import abrir, ROOT
from calidad import preparar
from resumen import por_periodo
from temporal import por_anio
from puntos import por_punto
from exportar import exportar

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--salida', type=Path, required=True)
    args = parser.parse_args()
    with duckdb.connect() as con:
        con.execute("SET threads=1")
        source = abrir(con)
        source['clasificacion_pares'] = preparar(con)
        source['duckdb'] = duckdb.__version__
        source['denominador'] = 'Un par completo por observación y periodo; una concentración por contaminante. No población.'
        source['fecha_ausente_excluida_solo_del_resumen_anual'] = con.sql('SELECT count(*) FROM validos WHERE fecha IS NULL').fetchone()[0]
        source['scripts_sha256'] = {}
        import hashlib
        for path in sorted((ROOT / 'scripts/indicadores').glob('*.py')):
            source['scripts_sha256'][path.name] = hashlib.sha256(path.read_bytes()).hexdigest()
        exportar({'por_periodo':por_periodo(con), 'por_anio':por_anio(con), 'por_punto':por_punto(con)}, args.salida, source)
        print('Indicadores exportados:', args.salida)

if __name__ == '__main__':
    main()
