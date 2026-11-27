const decimal = new Intl.NumberFormat('es-PE', {minimumFractionDigits: 2, maximumFractionDigits: 2});
export function formatDecimal(value) { return Number.isFinite(value) ? decimal.format(value) : 'No disponible'; }
export function formatTolerance(value) { return value !== 0 && Math.abs(value) < 0.01 ? value.toExponential(2) : formatDecimal(value); }
export function calculationMethod(origin) { return /^cl[uú]ster\b/i.test(origin) ? 'K-Means distribuido' : 'K-Means'; }
