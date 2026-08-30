package zeopoxaexporter

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"
)

type ZeopoxaTrack struct {
	Id        int64
	StartTime time.Time
	Duration  time.Duration
	Points    []ZeopoxaPoint
}

type ZeopoxaPoint struct {
	Created   time.Time
	Latidude  float64
	Longitude float64
	// In meter
	Elevation float64
	// In km/h
	Speed     float64
	HeartRate float64
}

func Parse(ctx context.Context, db *sql.DB, timezone *time.Location) ([]ZeopoxaTrack, error) {
	const (
		op         = "zeopoxaexporter.Parse"
		BATCH_SIZE = 500
	)
	output := make([]ZeopoxaTrack, 0)
	totalCount := 0

	query := fmt.Sprintf(`SELECT
		mt.id, 
		CONCAT(
			mt.year, '-', mt.month, '-', mt.day, 'T',
			CASE WHEN mt.start_time = '' THEN '00:00' ELSE mt.start_time END,
			':00'
		) as created,
		mt.time_milisec,
		mt.latlon_array,
		mt.speed_array,
		mt.elevation_array,
		mt.heart_rate_array
	FROM main_table mt
	LIMIT %d OFFSET ?;
	`, BATCH_SIZE)

	for offset := 0; ; offset += BATCH_SIZE {
		slog.Debug("fetch tracks", slog.Int("limit", BATCH_SIZE), slog.Int("offset", offset))
		count := 0

		rows, err := db.QueryContext(ctx, query, offset)
		if err != nil {
			return nil, fmt.Errorf("%s: query: %w", op, err)
		}

		for rows.Next() {
			count++
			track, err := parseRow(rows, timezone)
			if err != nil {
				_ = rows.Close()
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			output = append(output, *track)
		}

		_ = rows.Close()

		totalCount += count

		if count < BATCH_SIZE {
			break
		}
	}

	slog.Info("tracks processed", slog.Int("count", totalCount))

	return output, nil
}

func parseRow(rows *sql.Rows, timezone *time.Location) (*ZeopoxaTrack, error) {
	const op = "parseRow"

	var (
		track             ZeopoxaTrack
		start             string
		durationRaw       float64
		latlonArrayRaw    string
		speedArrayRaw     string
		elevationArrayRaw string
		heartRateArrayRaw string
	)

	type point struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}

	latlonArray := make([]point, 0)
	speedArray := make([]float64, 0)
	elevationArray := make([]float64, 0)
	heartRateArray := make([]float64, 0)
	err := rows.Scan(
		&track.Id,
		&start,
		&durationRaw,
		&latlonArrayRaw,
		&speedArrayRaw,
		&elevationArrayRaw,
		&heartRateArrayRaw,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: scan: %w", op, err)
	}

	track.Duration = time.Duration(durationRaw) * time.Millisecond

	track.StartTime, err = time.ParseInLocation("2006-1-2T15:04:05", start, timezone)
	if err != nil {
		return nil, fmt.Errorf("%s: start time: %w", op, err)
	}

	err = json.Unmarshal([]byte(latlonArrayRaw), &latlonArray)
	if err != nil {
		return nil, fmt.Errorf("%s: latlon points: %w", op, err)
	}
	err = json.Unmarshal([]byte(speedArrayRaw), &speedArray)
	if err != nil {
		return nil, fmt.Errorf("%s: speed array: %w", op, err)
	}
	err = json.Unmarshal([]byte(elevationArrayRaw), &elevationArray)
	if err != nil {
		return nil, fmt.Errorf("%s: elevation array: %w", op, err)
	}
	err = json.Unmarshal([]byte(heartRateArrayRaw), &heartRateArray)
	if err != nil {
		return nil, fmt.Errorf("%s: heart rate array: %w", op, err)
	}

	totalPoints := len(latlonArray)
	track.Points = make([]ZeopoxaPoint, 0, totalPoints)

	for idx := range totalPoints {

		// Linear approximation
		delta := time.Duration(
			int64(idx) * (int64(track.Duration) / int64(totalPoints)),
		)
		created := track.StartTime.Add(delta)
		p := ZeopoxaPoint{
			Created:   created,
			Latidude:  latlonArray[idx].Latitude,
			Longitude: latlonArray[idx].Longitude,
		}
		if len(elevationArray) == totalPoints {
			p.Elevation = elevationArray[idx]
		}
		if len(speedArray) == totalPoints {
			p.Speed = speedArray[idx]
		}
		if len(heartRateArray) == totalPoints {
			p.HeartRate = heartRateArray[idx]
		}

		track.Points = append(track.Points, p)
	}

	slog.Debug(
		"track processed successfully",
		slog.Int64("id", track.Id),
		slog.Time("started", track.StartTime),
		slog.Duration("duration", track.Duration),
		slog.Int("points_count", totalPoints),
	)

	return &track, nil
}
