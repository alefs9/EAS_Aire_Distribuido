package dataset

import (
	"fmt"
	"math"
	"strconv"
)

func ValidateRecord(record []string, columns map[string]int) error {
	for _, field := range []string{"pm10", "pm2.5"} {
		index := columns[field]

		if index >= len(record) {
			return fmt.Errorf("falta el valor de %s", field)
		}

		value, err := strconv.ParseFloat(record[index], 64)
		if err != nil {
			return fmt.Errorf("%s no es un valor numerico: %q", field, record[index])
		}

		if math.IsNaN(value) || math.IsInf(value, 0) {
			return fmt.Errorf("%s contiene un valor no finito", field)
		}

		if value < 0 {
			return fmt.Errorf("%s no puede ser negativo: %v", field, value)
		}
	}

	return nil
}
