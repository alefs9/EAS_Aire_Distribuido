"""Regenerar el CSV completo y sus metadatos desde el Parquet de vectores."""
from pathlib import Path
import argparse,hashlib,json,os,shutil,uuid
import duckdb

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output-dir',type=Path,help='Carpeta de salida; por defecto data/vectors')
    parser.add_argument('--overwrite',action='store_true',help='Reemplazar expresamente salidas existentes')
    args=parser.parse_args()
    root=Path(__file__).resolve().parents[1]
    manifest=json.loads((root/'data/vectors/manifest.json').read_text(encoding='utf-8'))
    source=root/manifest['parquet']
    target=args.output_dir.resolve() if args.output_dir else root/'data/vectors'
    target.mkdir(parents=True,exist_ok=True)
    names=['vectores_PM10_PM25.csv','metadatos_1_hora.csv','metadatos_24_horas.csv']
    if not args.overwrite:
        for name in names:
            if (target/name).exists():raise FileExistsError(f'Ya existe {target/name}. Usa --overwrite para reemplazarlo.')
    if shutil.disk_usage(target).free<250_000_000:raise RuntimeError('Se requieren al menos 250 MB libres para las exportaciones CSV.')
    h=hashlib.sha256()
    with source.open('rb') as f:
        for b in iter(lambda:f.read(1024*1024),b''):h.update(b)
    if h.hexdigest()!=manifest['sha256_parquet']:raise ValueError('El Parquet no coincide con el manifiesto.')
    qi=lambda s:'"'+s.replace('"','""')+'"'
    metadata=','.join(qi(s) for s in ['vector_id']+manifest['clave_observacion']+['datum_reconocido'])
    con=duckdb.connect()
    try:
        con.execute("SET memory_limit='2GB'")
        con.execute(f'SET threads={min(os.cpu_count() or 2,4)}')
        con.read_parquet(str(source)).create_view('vectores')
        count=con.execute('SELECT count(*) FROM vectores').fetchone()[0]
        if count!=manifest['pares_reales']:raise ValueError('El conteo no coincide con el manifiesto.')
        exports=[('vectores_PM10_PM25.csv','SELECT * FROM vectores ORDER BY vector_id',count)]
        for item in manifest['por_periodo']:
            period=item['periodo'];tag='1_hora' if period=='1 hora' else '24_horas'
            query='SELECT '+metadata+" FROM vectores WHERE periodo='"+period+"' ORDER BY vector_id"
            exports.append(('metadatos_'+tag+'.csv',query,item['filas']))
        for name,query,expected in exports:
            temp=target/(name+'.'+uuid.uuid4().hex+'.part')
            try:
                relation=con.sql(query)
                if relation.count('*').fetchone()[0]!=expected:raise ValueError('Conteo incorrecto: '+name)
                relation.write_csv(str(temp),header=True,sep=',')
                os.replace(temp,target/name)
                print(f'{name}: {expected:,} filas; {(target/name).stat().st_size:,} bytes')
            finally:temp.unlink(missing_ok=True)
    finally:con.close()

if __name__=='__main__':main()
