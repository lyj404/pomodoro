package service

import (
	"time"

	"github.com/lyj404/pomodoro/internal/model"
	"github.com/lyj404/pomodoro/internal/storage"
)

type StatsService struct {
	sessionRepo *storage.SessionRepository
}

func NewStatsService(sessionRepo *storage.SessionRepository) *StatsService {
	return &StatsService{sessionRepo: sessionRepo}
}

func (s *StatsService) TodayStats() (storage.TodayStats, error) {
	return s.sessionRepo.GetTodayStats(time.Now())
}

func (s *StatsService) RecentSessions(limit int) ([]model.Session, error) {
	return s.sessionRepo.ListRecent(limit)
}

func (s *StatsService) RecentSessionsPaginated(offset, limit int) ([]model.Session, int, error) {
	sessions, err := s.sessionRepo.ListRecentPaginated(offset, limit)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.sessionRepo.CountTotal()
	if err != nil {
		return nil, 0, err
	}
	return sessions, total, nil
}

func (s *StatsService) CountTotalSessions() (int, error) {
	return s.sessionRepo.CountTotal()
}

func (s *StatsService) DeleteSession(id int64) error {
	return s.sessionRepo.DeleteByID(id)
}

func (s *StatsService) DeleteSessions(ids []int64) error {
	return s.sessionRepo.DeleteByIDs(ids)
}
