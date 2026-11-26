"""Gráfico descriptivo de mediana e intervalo intercuartílico, por periodo."""
import argparse
import csv
from pathlib import Path
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--carpeta', type=Path, required=True)
    args = parser.parse_args()
    with (args.carpeta / 'por_periodo.csv').open(encoding='utf-8', newline='') as stream:
        rows = list(csv.DictReader(stream))
    fig, axes = plt.subplots(1, 2, figsize=(10, 4.5), sharey=True)
    for ax, period in zip(axes, ['1 hora','24 horas']):
        subset = [r for r in rows if r['periodo'] == period]
        for i, row in enumerate(subset):
            med, low, high = [float(row[k]) for k in ['mediana','p25','p75']]
            ax.errorbar(i, med, yerr=[[med-low],[high-med]], fmt='o', capsize=8, color='#176855')
            ax.annotate(f'n={int(row["n_observaciones"]):,}', (i, high), xytext=(0,12), textcoords='offset points', ha='center', fontsize=9)
        ax.set_xticks(range(len(subset)), [r['contaminante'] for r in subset])
        ax.set_title(period); ax.set_xlim(-.6,1.6); ax.set_ylim(0,70); ax.grid(axis='y',alpha=.2)
    axes[0].set_ylabel('Concentración original (µg/m³)')
    fig.suptitle('Concentraciones observadas · OEFA EAS')
    fig.text(.5,.025,'Punto: mediana · Barra: percentiles 25–75 · Periodos separados · No representa categorías sanitarias',ha='center',fontsize=9)
    fig.tight_layout(rect=[0,.08,1,.93])
    fig.savefig(args.carpeta/'distribucion_periodos.png',dpi=160)
    plt.close(fig)

if __name__ == '__main__':
    main()
