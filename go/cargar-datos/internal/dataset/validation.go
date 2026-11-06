package dataset

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const ExpectedUnit = "µg/m³"

// ValidateRecord comprueba que una medición tenga un contaminante,
// concentración y unidad válidos.
func ValidateRecord(record []string, columns map[string]int) error {
	required := []string{
		"contaminante",
		"concentracion",
		"unidad",
	}

	for _, field := range required {
		index := columns[field]
		if index >= len(record) {
			return fmt.Errorf("falta el valor de %s", field)
		}
	}

	contaminant := strings.TrimSpace(record[columns["contaminante"]])

	if contaminant != "PM10" && contaminant != "PM2.5" {
		return fmt.Errorf("contaminante no reconocido: %q", contaminant)
	}

	unit := strings.TrimSpace(record[columns["unidad"]])

	if unit != ExpectedUnit {
		return fmt.Errorf("unidad no válida para %s: %q (esperada: %s)",
			contaminant, unit, ExpectedUnit)
	}

	rawValue := strings.TrimSpace(record[columns["concentracion"]])

	value, err := strconv.ParseFloat(rawValue, 64)
	if err != nil {
		return fmt.Errorf("concentración no numérica: %q", rawValue)
	}

	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("la concentración contiene un valor no finito")
	}

	if value < 0 {
		return fmt.Errorf("la concentración no puede ser negativa: %v", value)
	}

	return nil
}
