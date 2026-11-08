"""Actualizar inventarios existentes sin archivos eliminados ni autorreferencias."""
from pathlib import Path
import csv
import hashlib

ROOT = Path(__file__).resolve().parents[1]
FIELDS = ["archivo", "bytes", "filas_datos", "sha256"]

def digest(path):
    result = hashlib.sha256()
    with path.open("rb") as stream:
        for block in iter(lambda: stream.read(1024 * 1024), b""):
            result.update(block)
    return result.hexdigest()

def update_inventory(name):
    inventory = ROOT / "results" / name / "inventario.csv"
    with inventory.open(encoding="utf-8-sig", newline="") as stream:
        previous = list(csv.DictReader(stream))
    rows = []
    removed = 0
    for entry in previous:
        path = (ROOT / entry["archivo"]).resolve()
        if not path.is_relative_to(ROOT) or path == inventory.resolve() or not path.is_file():
            removed += 1
            continue
        sha = digest(path)
        count = entry.get("filas_datos", "")
        if path.suffix.lower() == ".csv":
            with path.open(encoding="utf-8-sig", newline="") as stream:
                count = str(max(sum(1 for _ in csv.reader(stream)) - 1, 0))
        elif sha != entry.get("sha256"):
            count = ""  # No heredar un conteo sin verificar después de un cambio.
        rows.append({"archivo": path.relative_to(ROOT).as_posix(),
                     "bytes": str(path.stat().st_size),
                     "filas_datos": count, "sha256": sha})
    temporary = inventory.with_suffix(".csv.tmp")
    try:
        with temporary.open("w", encoding="utf-8", newline="") as stream:
            writer = csv.DictWriter(stream, fieldnames=FIELDS, lineterminator="\n")
            writer.writeheader()
            writer.writerows(rows)
        temporary.replace(inventory)
    finally:
        temporary.unlink(missing_ok=True)
    print(f"{name}: {len(rows)} archivos; {removed} referencias retiradas.")

if __name__ == "__main__":
    for category in ("eda", "limpieza", "vectores"):
        update_inventory(category)
