package dataset

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
)

const MaxReportedErrors = 10

// CSVRecord representa un registro leído del CSV.
type CSVRecord struct {
	Number int64
	Values []string
	Err    error
}

// Stats almacena los resultados del procesamiento.
type Stats struct {
	Read     int64
	Accepted int64
	Rejected int64
	PM10     int64
	PM25     int64
	Errors   []string
}

// processRecord valida y contabiliza un registro.
func processRecord(record CSVRecord, columns map[string]int, stats *Stats) {
	stats.Read++

	if record.Err != nil {
		stats.Rejected++

		if len(stats.Errors) < MaxReportedErrors {
			stats.Errors = append(stats.Errors,
				fmt.Sprintf("Registro %d: error de lectura: %v", record.Number, record.Err))
		}
		return
	}

	if err := ValidateRecord(record.Values, columns); err != nil {
		stats.Rejected++

		if len(stats.Errors) < MaxReportedErrors {
			stats.Errors = append(stats.Errors,
				fmt.Sprintf("Registro %d: %v", record.Number, err))
		}
		return
	}

	stats.Accepted++

	switch strings.TrimSpace(record.Values[columns["contaminante"]]) {
	case "PM10":
		stats.PM10++
	case "PM2.5":
		stats.PM25++
	}
}

// ProcessCSVSequential procesa los registros uno por uno.
// Se conserva como referencia para comparar el rendimiento.
func ProcessCSVSequential(reader *CSVReader) (Stats, error) {
	var stats Stats
	var number int64

	for {
		record, err := reader.Read()

		if errors.Is(err, io.EOF) {
			break
		}

		number++

		if err != nil {
			processRecord(CSVRecord{
				Number: number,
				Err:    err,
			}, reader.Columns, &stats)

			// Los errores de lectura del CSV se contabilizan.
			// Si son errores de formato recuperables, se continúa.
			var parseErr *csv.ParseError
			if !errors.As(err, &parseErr) {
				return stats, fmt.Errorf("error de lectura del CSV: %w", err)
			}
			continue
		}

		processRecord(CSVRecord{
			Number: number,
			Values: record,
		}, reader.Columns, &stats)
	}

	return stats, nil
}

// ProcessCSVConcurrent procesa el CSV mediante lotes y workers.
func ProcessCSVConcurrent(reader *CSVReader, workers, batchSize int) (Stats, error) {
	if workers <= 0 {
		return Stats{}, fmt.Errorf("el número de workers debe ser mayor que cero")
	}

	if batchSize <= 0 {
		return Stats{}, fmt.Errorf("el tamaño del lote debe ser mayor que cero")
	}

	jobs := make(chan []CSVRecord)
	results := make(chan Stats, workers)

	var wg sync.WaitGroup

	// Inicia los workers.
	for i := 0; i < workers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			var localStats Stats

			for batch := range jobs {
				for _, record := range batch {
					processRecord(record, reader.Columns, &localStats)
				}
			}

			results <- localStats
		}()
	}

	// Lee el CSV y distribuye los lotes.
	var batch []CSVRecord
	var number int64
	var readErr error

	for {
		record, err := reader.Read()

		if errors.Is(err, io.EOF) {
			break
		}

		number++

		batch = append(batch, CSVRecord{
			Number: number,
			Values: record,
			Err:    err,
		})

		if len(batch) == batchSize {
			jobs <- batch
			batch = nil
		}

		if err != nil {
			var parseErr *csv.ParseError
			if !errors.As(err, &parseErr) {
				readErr = fmt.Errorf("error de lectura del CSV: %w", err)
				break
			}
		}
	}

	if len(batch) > 0 {
		jobs <- batch
	}

	close(jobs)
	wg.Wait()
	close(results)

	// Consolida los resultados de los workers.
	var total Stats

	for result := range results {
		total.Read += result.Read
		total.Accepted += result.Accepted
		total.Rejected += result.Rejected
		total.PM10 += result.PM10
		total.PM25 += result.PM25

		for _, message := range result.Errors {
			if len(total.Errors) >= MaxReportedErrors {
				break
			}
			total.Errors = append(total.Errors, message)
		}
	}

	return total, readErr
}
