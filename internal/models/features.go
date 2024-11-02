package models

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Feature struct {
	ID       int
	Loaves   int
	Feature1 int
	Feature2 int
	Feature3 int
	Date     time.Time
}

type FeaturesModel struct {
	DB *sql.DB
}

// func (m *FeaturesModel) GetAll() ([]*Feature, error) {

// }

func (m *FeaturesModel) GetColumns() ([]string, error) {
	stmt := "SHOW COLUMNS FROM features"

	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var columns []string
	var typeVal, null, key, defaultVal, extra sql.NullString
	for rows.Next() {
		var column string
		err := rows.Scan(&column, &typeVal, &null, &key, &defaultVal, &extra)
		if err != nil {
			return nil, err
		}
		columns = append(columns, column)
	}

	return columns[1:], nil
}

func (m *FeaturesModel) Insert(
	date time.Time,
	loaves,
	feature1,
	feature2,
	feature3 int,
	features []string,
) error {

	columns := []string{"date", "loaves"}
	values := []string{"?", "?"}

	for i, feature := range features {
		columns = append(columns, fmt.Sprintf("feature%d_%s", i+1, feature))
		values = append(values, "?")
	}

	stmt := fmt.Sprintf("INSERT INTO features (%s) VALUES (%s)",
		strings.Join(columns, ","),
		strings.Join(values, ","),
	)

	_, err := m.DB.Exec(stmt, date, loaves, feature1, feature2, feature3)
	if err != nil {
		return err
	}

	return nil
}

func (m *FeaturesModel) InsertRecords(
	records [][]string,
	csvToDBMap map[string]int,
	features []string,
) error {
	for _, row := range records {
		dateStr := row[csvToDBMap["date"]]
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("error parsing date: %w", err)
		}

		loavesStr := row[csvToDBMap["loaves"]]
		loaves, err := strconv.Atoi(loavesStr)
		if err != nil {
			return fmt.Errorf("error converting loaves to integer: %w", err)
		}

		// Loop through dynamic features
		feats := make([]int, len(features))
		for i, feature := range features {
			key := fmt.Sprintf("feature%d", i+1)
			featureStr := row[csvToDBMap[key]]
			dynamicFeature, err := strconv.Atoi(featureStr)
			if err != nil {
				return fmt.Errorf("error converting %s to integer: %w", feature, err)
			}
			feats[i] = dynamicFeature
		}

		// Insert the record into the database
		if err := m.Insert(date, loaves, feats[0], feats[1], feats[2], features); err != nil {
			return fmt.Errorf("error inserting record into database: %w", err)
		}
	}
	return nil
}
