import csv
import hashlib
import json
import math
import pathlib
import struct
import sys

ABS = 1e-8
REL = 1e-10
evidence = pathlib.Path(sys.argv[1])
vectors = pathlib.Path(sys.argv[2])


def read(path):
    return json.loads(path.read_text(encoding="utf-8-sig"))


def assign(points, centers):
    counts = [0] * len(centers)
    labels = []
    for x, y in points:
        label = min(range(len(centers)), key=lambda i: (
            (x - centers[i][0]) ** 2 + (y - centers[i][1]) ** 2
        ))
        counts[label] += 1
        labels.append(label)
    return counts, labels


def compare(ref, actual, points, workers, k, case):
    errors = []
    def check(condition, label):
        if not condition:
            errors.append(label)

    expected_reason = "converged" if ref["converged"] else "max_iterations"
    check(actual["status"] == "completed", "status")
    cfg = actual["config"]
    check((cfg["k"], cfg["seed"], cfg["tolerance"], cfg["max_iterations"]) ==
          (k, 42, 0.0001, 100), "configuration")
    check((ref["k"], ref["seed"], ref["tolerance"], ref["max_iterations"], ref["mode"]) ==
          (k, 42, 0.0001, 100, "sequential"), "reference_configuration")
    check(cfg["file"] == ref["input"], "same_input")
    check(actual["reason"] == expected_reason and actual["iterations"] == ref["iterations"], "termination")
    check(actual["rows"] == ref["vectors"] == len(points), "rows")
    check(actual["initial_centroids"] == ref["initial_centroids"], "initial_centroids")
    check(len(actual["counts"]) == k and actual["counts"] == ref["counts"], "counts")
    check(sum(actual["counts"]) == len(points), "count_conservation")
    boundaries = actual["partitions"]
    check(len(boundaries) == len(actual["participants"]) == cfg["nodes"] == workers and
          len(set(actual["participants"])) == workers, "participants")
    end = 0
    for i, p in enumerate(boundaries):
        check(p["id"] == i and p["node_id"] == actual["participants"][i] and
              p["start"] == end and 0 <= end < p["end"] <= len(points), f"partition_{i}")
        end = p["end"]
    check(end == len(points), "partition_coverage")
    independent, labels = assign(points, actual["centroids"])
    reference_counts, reference_labels = assign(points, ref["centroids"])
    check(independent == actual["counts"], "independent_counts")
    check(reference_counts == ref["counts"], "reference_independent_counts")
    digest = hashlib.sha256()
    for label in reference_labels:
        digest.update(struct.pack("<Q", label))
    check(digest.hexdigest() == ref["labels_sha256"], "reference_labels_sha256")
    mismatches = [i for i, (a, b) in enumerate(zip(labels, reference_labels)) if a != b]
    check(not mismatches, "row_assignments")
    check(len(actual["centroids"]) == len(ref["centroids"]) == k, "centroid_dimensions")
    diffs = []
    for i, (expected, received) in enumerate(zip(ref["centroids"], actual["centroids"])):
        check(len(expected) == len(received) == 2, f"center_dimension_{i}")
        delta = [abs(a - b) for a, b in zip(expected, received)]
        bound = [ABS + REL * abs(x) for x in expected]
        relative = [d / abs(x) if x else None for d, x in zip(delta, expected)]
        check(all(math.isfinite(d) and d <= limit for d, limit in zip(delta, bound)), f"center_{i}")
        diffs.append(dict(cluster=i, absolute_difference=delta, relative_difference=relative, allowed=bound))
    if case == "sintetico":
        check(sorted(zip(map(tuple, actual["centroids"]), actual["counts"])) ==
              [((1, 1), 2), ((11, 11), 2)], "known_centroids_and_counts")
    return dict(nodes=workers, passed=not errors, failures=errors, partitions=boundaries,
                assignment_mismatches=len(mismatches), first_mismatching_rows=mismatches[:10],
                independently_assigned_counts=independent, reference_counts=ref["counts"],
                actual_counts=actual["counts"], reference_reason=expected_reason,
                actual_reason=actual["reason"], iterations=actual["iterations"],
                reference_iterations=ref["iterations"], differences=diffs)


report = {"absolute_tolerance": ABS, "relative_tolerance": REL, "cases": {}}
for case, path, k in (
    ("sintetico", evidence / "sintetico.csv", 2),
    ("1_hora", vectors / "features_1_hora.csv", 4),
    ("24_horas", vectors / "features_24_horas.csv", 4),
):
    with path.open(newline="", encoding="utf-8") as stream:
        rows = csv.reader(stream)
        assert next(rows) == ["PM10", "PM2.5"]
        points = [tuple(map(float, row)) for row in rows]
    ref = read(evidence / f"{case}-secuencial.json")
    checks = [compare(ref, read(evidence / f"{case}-{n}n.json"), points, n, k, case) for n in (1, 2)]
    report["cases"][case] = dict(input=path.name, sha256=hashlib.sha256(path.read_bytes()).hexdigest(),
                                 rows=len(points), k=k, comparisons=checks)
(evidence / "comparaciones.json").write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
for case, entry in report["cases"].items():
    for item in entry["comparisons"]:
        delta = max(d for center in item["differences"] for d in center["absolute_difference"])
        print(f"{case} {item['nodes']} nodo(s): {item['passed']} N={entry['rows']} "
              f"iter={item['iterations']} motivo={item['actual_reason']} max_abs={delta:.12g} "
              f"fallos={item['failures']}")
sys.exit(0 if all(x["passed"] for entry in report["cases"].values() for x in entry["comparisons"]) else 1)
