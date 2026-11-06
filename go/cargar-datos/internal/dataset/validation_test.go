package dataset

import "testing"

func TestValidateRecord(t *testing.T) {
	columns := map[string]int{
		"contaminante":  0,
		"concentracion": 1,
		"unidad":        2,
	}

	tests := []struct {
		name    string
		record  []string
		wantErr bool
	}{
		{
			name:    "PM10 válido",
			record:  []string{"PM10", "25.5", "µg/m³"},
			wantErr: false,
		},
		{
			name:    "PM2.5 válido",
			record:  []string{"PM2.5", "12.8", "µg/m³"},
			wantErr: false,
		},
		{
			name:    "contaminante desconocido",
			record:  []string{"CO2", "25.5", "µg/m³"},
			wantErr: true,
		},
		{
			name:    "unidad incorrecta",
			record:  []string{"PM10", "25.5", "mg/m³"},
			wantErr: true,
		},
		{
			name:    "concentración no numérica",
			record:  []string{"PM10", "abc", "µg/m³"},
			wantErr: true,
		},
		{
			name:    "concentración negativa",
			record:  []string{"PM10", "-5", "µg/m³"},
			wantErr: true,
		},
		{
			name:    "concentración NaN",
			record:  []string{"PM10", "NaN", "µg/m³"},
			wantErr: true,
		},
		{
			name:    "concentración infinita",
			record:  []string{"PM10", "+Inf", "µg/m³"},
			wantErr: true,
		},
		{
			name:    "valor de concentración faltante",
			record:  []string{"PM10"},
			wantErr: true,
		},
		{
			name:    "registro vacío",
			record:  []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRecord(tt.record, columns)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRecord() error = %v, wantErr = %v",
					err, tt.wantErr)
			}
		})
	}
}
