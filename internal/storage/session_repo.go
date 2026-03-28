package storage

import (
	"database/sql"
	"time"

	"github.com/lyj404/pomodoro/internal/model"
)

type SessionRepository struct {
	db *sql.DB
}

type TodayStats struct {
	CompletedPomodoros int
	FocusSeconds       int
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(session model.Session) error {
	_, err := r.db.Exec(`
		INSERT INTO sessions(mode, planned_seconds, actual_seconds, started_at, ended_at, completed)
		VALUES(?, ?, ?, ?, ?, ?)
	`,
		string(session.Mode),
		session.PlannedSeconds,
		session.ActualSeconds,
		session.StartedAt,
		session.EndedAt,
		boolToInt(session.Completed),
	)
	return err
}

func (r *SessionRepository) ListRecent(limit int) ([]model.Session, error) {
	rows, err := r.db.Query(`
		SELECT id, mode, planned_seconds, actual_seconds, started_at, ended_at, completed
		FROM sessions
		ORDER BY started_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []model.Session
	for rows.Next() {
		var s model.Session
		var mode string
		var endedAt sql.NullTime
		var completed int
		if err := rows.Scan(&s.ID, &mode, &s.PlannedSeconds, &s.ActualSeconds, &s.StartedAt, &endedAt, &completed); err != nil {
			return nil, err
		}
		s.Mode = model.SessionMode(mode)
		s.Completed = completed == 1
		if endedAt.Valid {
			value := endedAt.Time
			s.EndedAt = &value
		}
		sessions = append(sessions, s)
	}

	return sessions, rows.Err()
}

func (r *SessionRepository) DeleteByID(id int64) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (r *SessionRepository) DeleteByIDs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`DELETE FROM sessions WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.Exec(id); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *SessionRepository) GetTodayStats(now time.Time) (TodayStats, error) {
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	row := r.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN mode = 'work' AND completed = 1 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN mode = 'work' AND completed = 1 THEN actual_seconds ELSE 0 END), 0)
		FROM sessions
		WHERE started_at >= ? AND started_at < ?
	`, startOfDay, endOfDay)

	var stats TodayStats
	err := row.Scan(&stats.CompletedPomodoros, &stats.FocusSeconds)
	return stats, err
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
