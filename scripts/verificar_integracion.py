"""Comprueba el cargador real y los modelos existentes sin duplicar sus módulos."""
import argparse
import json
import os
from pathlib import Path
import re
import subprocess
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parents[1]

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--race', action='store_true', help='Requiere CGO_ENABLED=1 y compilador C compatible')
    args = parser.parse_args()
    target = ROOT / 'results/pruebas'
    target.mkdir(parents=True, exist_ok=True)
    suffix = 'race' if args.race else 'normal'
    records = []
    env = dict(os.environ, KMEANS_REAL_DATA='1', GOMAXPROCS='4')
    def run(command, cwd, name):
        result = subprocess.run(command, cwd=cwd, env=env, text=True, encoding='utf-8', errors='replace', capture_output=True)
        log = '$ ' + ' '.join(command) + '\n' + result.stdout + result.stderr
        (target / f'{name}-{suffix}.txt').write_text(log, encoding='utf-8')
        records.append({'command': command, 'cwd': str(cwd.relative_to(ROOT)), 'exit_code': result.returncode, 'log': f'{name}-{suffix}.txt'})
        print(name, 'exit=', result.returncode, flush=True)
        if result.returncode:
            raise RuntimeError(log)
        return result.stdout
    metadata = {'timestamp_utc': datetime.now(timezone.utc).isoformat(),
                'commit': subprocess.check_output(['git', 'rev-parse', 'HEAD'], cwd=ROOT, text=True).strip(),
                'go': subprocess.check_output(['go', 'version'], text=True).strip(),
                'gomaxprocs': 4, 'cgo_enabled': env.get('CGO_ENABLED'), 'cc': env.get('CC'), 'runs': records}
    try:
        for module in ['cargar-datos', 'kmeans_secuencial']:
            command = ['go', 'test', '-v', '-count=1', '-timeout=300s']
            if args.race: command.append('-race')
            run(command + ['./...'], ROOT/'go'/module, module)
        if not args.race:
            path = ROOT/'data/processed/PM25_PM10_limpio.csv'
            if not path.is_file():
                raise FileNotFoundError('Regenera primero el CSV con scripts/exportar_csv_limpio.py')
            expected = {'Registros leídos': 1062897, 'Registros aceptados': 1062897,
                        'Registros rechazados': 0, 'Mediciones PM10 aceptadas': 598502,
                        'Mediciones PM2.5 aceptadas': 464395}
            for mode in ['sequential', 'concurrent']:
                output = run(['go', 'run', './cmd/cargar-datos', '-mode='+mode, '-workers=4', '-batch=1000', str(path)], ROOT/'go/cargar-datos', 'carga-'+mode)
                for label, count in expected.items():
                    match = re.search(re.escape(label)+r': (\d+)', output)
                    if not match or int(match[1]) != count:
                        raise AssertionError(f'{label}: se esperaba {count}')
                if 'Conteos reconciliados: true' not in output:
                    raise AssertionError('conteos no reconciliados')
        metadata['status'] = 'PASS'
    except Exception as exc:
        metadata['status'] = 'FAIL'
        metadata['error'] = str(exc)
        raise
    finally:
        (target/f'ejecucion-{suffix}.json').write_text(json.dumps(metadata, indent=2, ensure_ascii=False), encoding='utf-8')

if __name__ == '__main__': main()
