"""CSV ordenados, manifiesto de ejecución y hashes de las salidas."""
import hashlib
import json

def exportar(tablas, destino, procedencia):
    destino.mkdir(parents=True, exist_ok=True)
    names = [name + '.csv' for name in tablas] + ['procedencia.json']
    if any((destino / name).exists() for name in names):
        raise FileExistsError('Usa una carpeta nueva para conservar la ejecución anterior.')
    outputs = {}
    for name, tabla in tablas.items():
        path = destino / (name + '.csv')
        tabla.write_csv(str(path), header=True)
        with path.open('rb') as stream:
            digest = hashlib.file_digest(stream, 'sha256').hexdigest()
        outputs[path.name] = {'filas': tabla.count('*').fetchone()[0], 'sha256': digest}
    (destino / 'procedencia.json').write_text(json.dumps({**procedencia, 'salidas': outputs}, ensure_ascii=False, indent=2), encoding='utf-8')
