"""Genera figuras exportables a partir de medianas y rangos realmente medidos."""
import argparse
import json
from pathlib import Path
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument('directory',type=Path)
    args=parser.parse_args()
    rows=json.loads((args.directory/'resumen.json').read_text(encoding='utf-8'))
    fig, axes=plt.subplots(1,2,figsize=(12,4.6),layout='constrained')
    for ax,period in zip(axes,['1_hora','24_horas']):
        group=[r for r in rows if r['period']==period]
        values=[r['train_seconds']*1000 for r in group]
        lower=[(r['train_seconds']-r['train_seconds_min'])*1000 for r in group]
        upper=[(r['train_seconds_max']-r['train_seconds'])*1000 for r in group]
        labels=['Secuencial' if r['mode']=='sequential' else f"Conc. {r['workers']}w" for r in group]
        ax.bar(labels,values,yerr=[lower,upper],capsize=4,color=['#52677e']+['#087f8c']*4)
        ax.set_title(f"{period.replace('_',' ')} · {group[0]['vectors']:,} vectores")
        ax.set_ylabel('Entrenamiento (ms)');ax.grid(axis='y',alpha=.2);ax.set_axisbelow(True)
    fig.suptitle('K-Means: mediana de 5 repeticiones; barras de error mínimo–máximo\nGOMAXPROCS=4 · k=4 · semilla=42 · sin escalado',fontsize=12)
    fig.savefig(args.directory/'tiempos.png',dpi=180)
    fig.savefig(args.directory/'tiempos.svg')
    plt.close(fig)

if __name__=='__main__':main()
