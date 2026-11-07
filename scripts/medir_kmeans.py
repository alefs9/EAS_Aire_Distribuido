"""Mediciones intercaladas, resultados crudos y comparación numérica de K-Means."""
import argparse
import csv
from datetime import datetime, timezone
import hashlib
import json
import math
import os
from pathlib import Path
import platform
import random
import statistics
import subprocess

ROOT = Path(__file__).resolve().parents[1]

def equivalent(a, b):
    for field in ['vectors', 'k', 'seed', 'tolerance', 'max_iterations', 'initial_centroids', 'labels_sha256', 'counts', 'iterations', 'converged']:
        if a[field] != b[field]: raise AssertionError(f'Diferencia en {field}')
    for x, y in zip(sum(a['centroids'], []), sum(b['centroids'], [])):
        if abs(x-y) > 1e-9 + 1e-9*abs(x): raise AssertionError('Centroides distintos')
    if abs(a['inertia']-b['inertia']) > 1e-9 + 1e-9*abs(a['inertia']): raise AssertionError('Inercia distinta')

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--repetitions', type=int, default=5)
    parser.add_argument('--gomaxprocs', type=int, default=4)
    parser.add_argument('--output', type=Path, default=ROOT/'results/rendimiento/ejecucion-local')
    args=parser.parse_args()
    if args.repetitions < 1 or args.gomaxprocs < 1: parser.error('parámetros positivos requeridos')
    # Evita sobreescribir evidencias previas.
    args.output.mkdir(parents=True, exist_ok=False)
    binary=ROOT/'.tools'/('kmeans.exe' if os.name=='nt' else 'kmeans')
    binary.parent.mkdir(exist_ok=True)
    env=dict(os.environ, GOMAXPROCS=str(args.gomaxprocs), CGO_ENABLED='0')
    subprocess.run(['go','build','-o',str(binary),'.'],cwd=ROOT/'go/kmeans_secuencial',env=env,check=True)
    metadata={'timestamp_utc':datetime.now(timezone.utc).isoformat(),
              'commit':subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(),
              'tracked_changes':subprocess.check_output(['git','diff','HEAD','--stat'],cwd=ROOT,text=True),
              'os':platform.platform(),'processor':platform.processor(),'logical_cpus':os.cpu_count(),
              'go':subprocess.check_output(['go','version'],text=True).strip(),
              'gomaxprocs':args.gomaxprocs,'repetitions':args.repetitions,'shuffle_seed':2026,
              'scope':'LoadCSV + RunKMeans; excluye compilación, inicio del proceso, metadatos y serialización',
              'inputs':{}}
    if os.name=='nt':
        ps='@{cpu=(Get-CimInstance Win32_Processor | Select-Object Name,NumberOfCores,NumberOfLogicalProcessors);ram_bytes=(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory} | ConvertTo-Json -Depth 4'
        probe=subprocess.run(['powershell','-NoProfile','-Command',ps],capture_output=True,text=True)
        metadata['hardware_probe']=probe.stdout.strip()
        metadata['hardware_probe_error']=probe.stderr.strip()
    (args.output/'entorno.json').write_text(json.dumps(metadata,indent=2,ensure_ascii=False),encoding='utf-8')
    configurations=[('sequential',1)]+[('concurrent',w) for w in (1,2,4,8)]
    rows=[]
    rng=random.Random(2026)
    for period in ['1_hora','24_horas']:
        path=ROOT/'data/vectors'/f'features_{period}.csv'
        metadata['inputs'][period]={'path':str(path.relative_to(ROOT)),'sha256':hashlib.sha256(path.read_bytes()).hexdigest()}
        def execute(mode,workers,rep):
            command=[str(binary),'-input',str(path),'-period',period,'-mode',mode,'-workers',str(workers),'-k','4','-max-iter','100','-tol','0.001','-seed','42']
            result=subprocess.run(command,check=True,capture_output=True,text=True,env=env)
            record=json.loads(result.stdout)
            record.update(repetition=rep,command=command)
            (args.output/f'{period}-{mode}-w{workers}-r{rep}.json').write_text(json.dumps(record,indent=2),encoding='utf-8')
            return record
        # Una ejecución de calentamiento por configuración, conservada pero excluida del resumen.
        baseline=execute('sequential',1,0)
        for mode,w in configurations[1:]: equivalent(baseline,execute(mode,w,0))
        for rep in range(1,args.repetitions+1):
            order=configurations.copy();rng.shuffle(order)
            for mode,w in order:
                record=execute(mode,w,rep);equivalent(baseline,record);rows.append(record)
                print(period,rep,mode,w,round(record['train_seconds'],6),flush=True)
    fields=['period','mode','workers_effective','repetition','vectors','k','gomaxprocs','iterations','converged','load_seconds','train_seconds','total_seconds','inertia','labels_sha256']
    with (args.output/'mediciones.csv').open('w',newline='',encoding='utf-8') as out:
        writer=csv.DictWriter(out,fieldnames=fields,extrasaction='ignore');writer.writeheader();writer.writerows(rows)
    summary=[]
    for period in ['1_hora','24_horas']:
        baseline=[r for r in rows if r['period']==period and r['mode']=='sequential']
        base_train=statistics.median(r['train_seconds'] for r in baseline)
        base_total=statistics.median(r['total_seconds'] for r in baseline)
        for mode,w in configurations:
            group=[r for r in rows if r['period']==period and r['mode']==mode and r['workers_effective']==w]
            item={'period':period,'mode':mode,'workers':w,'vectors':group[0]['vectors'],'iterations':group[0]['iterations'],'inertia':group[0]['inertia']}
            for field in ['load_seconds','train_seconds','total_seconds']:
                values=[r[field] for r in group]
                item[field]=statistics.median(values)
                item[field+'_min']=min(values);item[field+'_max']=max(values)
            item['speedup_train']=base_train/item['train_seconds']
            item['speedup_total']=base_total/item['total_seconds']
            item['efficiency_workers']=item['speedup_train']/w
            summary.append(item)
    (args.output/'resumen.json').write_text(json.dumps(summary,indent=2),encoding='utf-8')
    metadata['equivalence']='PASS: inicialización, etiquetas SHA256, conteos, iteraciones, convergencia y tolerancias 1e-9 abs + 1e-9 rel'
    (args.output/'entorno.json').write_text(json.dumps(metadata,indent=2,ensure_ascii=False),encoding='utf-8')
    lines=['# Mediciones reales — medianas', '', '| Periodo | Modo | Workers | Carga (s) | ML (s) | Total (s) | Speedup ML | Speedup total |','|---|---|---:|---:|---:|---:|---:|---:|']
    for s in summary:
        lines.append(f"| {s['period']} | {s['mode']} | {s['workers']} | {s['load_seconds']:.6f} | {s['train_seconds']:.6f} | {s['total_seconds']:.6f} | {s['speedup_train']:.3f} | {s['speedup_total']:.3f} |")
    (args.output/'RESUMEN.md').write_text('\n'.join(lines)+'\n',encoding='utf-8')

if __name__=='__main__': main()
