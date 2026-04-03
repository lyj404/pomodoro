package storage

import (
	"database/sql"
	"strconv"

	"github.com/lyj404/pomodoro/internal/model"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Load() (model.Settings, error) {
	settings := model.DefaultSettings()
	rows, err := r.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return settings, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return settings, err
		}

		switch key {
		case "work_minutes":
			settings.WorkMinutes, _ = strconv.Atoi(value)
		case "short_break_minutes":
			settings.ShortBreakMinutes, _ = strconv.Atoi(value)
		case "long_break_minutes":
			settings.LongBreakMinutes, _ = strconv.Atoi(value)
		case "long_break_interval":
			settings.LongBreakInterval, _ = strconv.Atoi(value)
		case "auto_start_next_phase":
			settings.AutoStartNextPhase = value == "1"
		case "theme":
			settings.Theme = value
		case "language":
			settings.Language = value
		}
	}

	return settings, rows.Err()
}

func (r *SettingsRepository) Save(settings model.Settings) error {
	values := map[string]string{
		"work_minutes":          strconv.Itoa(settings.WorkMinutes),
		"short_break_minutes":   strconv.Itoa(settings.ShortBreakMinutes),
		"long_break_minutes":    strconv.Itoa(settings.LongBreakMinutes),
		"long_break_interval":   strconv.Itoa(settings.LongBreakInterval),
		"auto_start_next_phase": boolToString(settings.AutoStartNextPhase),
		"theme":                 settings.Theme,
		"language":              settings.Language,
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for key, value := range values {
		if _, err := tx.Exec(`
			INSERT INTO settings(key, value)
			VALUES(?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value
		`, key, value); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func boolToString(v bool) string {
	if v {
		return "1"
	}
	return "0"
}
