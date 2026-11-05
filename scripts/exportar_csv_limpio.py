"""Regenerar el CSV local desde el Parquet limpio publicado."""
from pathlib import Path
import argparse
import hashlib
import json
import os
import shutil
import uuid
import duckdb

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--output',type=Path,help='Ruta de salida; por defecto data/processed/PM25_PM10_limpio.csv')
    parser.add_argument('--overwrite',action='store_true',help='Reemplazar expresamente un CSV existente')
    args=parser.parse_args()
    root=Path(__file__).resolve().parents[1]
    manifest=json.loads((root/'data/processed/ultima_ejecucion.json').read_text(encoding='utf-8'))
    source=root/manifest['parquet']
    target=args.output.resolve() if args.output else root/manifest['csv']
    if not source.is_file():raise FileNotFoundError(source)
    if target.exists() and not args.overwrite:raise FileExistsError(f'La salida ya existe: {target}. Usa --overwrite para reemplazarla.')
    target.parent.mkdir(parents=True,exist_ok=True)
    if shutil.disk_usage(target.parent).free<350_000_000:raise RuntimeError('Se requieren al menos 350 MB libres para exportar el CSV.')
    h=hashlib.sha256()
    with source.open('rb') as f:
        for block in iter(lambda:f.read(1024*1024),b''):h.update(block)
    if h.hexdigest()!=manifest['sha256_parquet']:raise ValueError('El Parquet no coincide con el manifiesto publicado.')
    temp=target.with_name(target.name+'.'+uuid.uuid4().hex+'.part')
    con=duckdb.connect()
    try:
        con.execute("SET memory_limit='3GB'")
        con.execute(f'SET threads={min(os.cpu_count() or 2,4)}')
        relation=con.read_parquet(str(source))
        rows=relation.count('*').fetchone()[0]
        if rows!=manifest['filas']:raise ValueError('Cantidad de mediciones distinta de la declarada.')
        relation.write_csv(str(temp),header=True,sep=',')
        os.replace(temp,target)
        print(f'CSV generado: {target}')
        print(f'Mediciones: {rows:,}; bytes: {target.stat().st_size:,}')
    finally:
        con.close()
        temp.unlink(missing_ok=True)

if __name__=='__main__':main()
