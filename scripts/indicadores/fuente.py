"""Lectura verificada del Parquet original de vectores."""
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]

def abrir(con, root=ROOT):
    manifest_path = root / 'data/vectors/manifest.json'
    manifest = json.loads(manifest_path.read_text(encoding='utf-8'))
    source = root / manifest['parquet']
    digest = hashlib.file_digest(source.open('rb'), 'sha256').hexdigest()
    if digest != manifest['sha256_parquet']:
        raise ValueError('El Parquet no coincide con el manifiesto.')
    con.read_parquet(str(source)).create_view('vectores')
    count, unique = con.sql('SELECT count(*), count(DISTINCT vector_id) FROM vectores').fetchone()
    if count != manifest['pares_reales'] or unique != count:
        raise ValueError('Volumen o vector_id inconsistente.')
    if manifest['unidad'] != 'µg/m³' or manifest['escalado'] is not None:
        raise ValueError('Se requieren concentraciones originales en µg/m³.')
    for item in manifest['por_periodo']:
        actual = con.execute('SELECT count(*) FROM vectores WHERE periodo=?', [item['periodo']]).fetchone()[0]
        if actual != item['filas']:
            raise ValueError('Conteo por periodo inconsistente.')
    return {'archivo': manifest['parquet'], 'sha256': digest, 'run_id': manifest['run_id'],
            'unidad': manifest['unidad'], 'pares_entrada': count,
            'mediciones_sin_par_excluidas_previamente': manifest['mediciones_sin_par']}
