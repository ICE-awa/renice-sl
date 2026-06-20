package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
)

type mockStatRepo struct {
	called string
	err    error

	clickStats []*dtov1.ClickStatItem
	linkStats  []*dtov1.LinkStatItem
	userStats  []*dtov1.UserStatItem
}

func (m *mockStatRepo) GetClickDayStat(context.Context, int) ([]*dtov1.ClickStatItem, error) {
	m.called = "click_day"
	return m.clickStats, m.err
}

func (m *mockStatRepo) GetLinkDayStat(context.Context, int) ([]*dtov1.LinkStatItem, error) {
	m.called = "link_day"
	return m.linkStats, m.err
}

func (m *mockStatRepo) GetUserDayStat(context.Context, int) ([]*dtov1.UserStatItem, error) {
	m.called = "user_day"
	return m.userStats, m.err
}

func (m *mockStatRepo) GetClickHourStat(context.Context, int) ([]*dtov1.ClickStatItem, error) {
	m.called = "click_hour"
	return m.clickStats, m.err
}

func (m *mockStatRepo) GetLinkHourStat(context.Context, int) ([]*dtov1.LinkStatItem, error) {
	m.called = "link_hour"
	return m.linkStats, m.err
}

func (m *mockStatRepo) GetUserHourStat(context.Context, int) ([]*dtov1.UserStatItem, error) {
	m.called = "user_hour"
	return m.userStats, m.err
}

func TestStatService_DispatchesByBucket(t *testing.T) {
	t.Parallel()

	now := time.Now()
	tests := []struct {
		name       string
		call       func(*statService) (any, error)
		wantCalled string
	}{
		{
			name: "link day",
			call: func(s *statService) (any, error) {
				return s.GetLinkStats(context.Background(), &dtov1.GetLinkStatReq{Range: 7, Bucket: "day"})
			},
			wantCalled: "link_day",
		},
		{
			name: "link hour",
			call: func(s *statService) (any, error) {
				return s.GetLinkStats(context.Background(), &dtov1.GetLinkStatReq{Range: 24, Bucket: "hour"})
			},
			wantCalled: "link_hour",
		},
		{
			name: "click day",
			call: func(s *statService) (any, error) {
				return s.GetClickStats(context.Background(), &dtov1.GetClickStatReq{Range: 7, Bucket: "day"})
			},
			wantCalled: "click_day",
		},
		{
			name: "click hour",
			call: func(s *statService) (any, error) {
				return s.GetClickStats(context.Background(), &dtov1.GetClickStatReq{Range: 24, Bucket: "hour"})
			},
			wantCalled: "click_hour",
		},
		{
			name: "user day",
			call: func(s *statService) (any, error) {
				return s.GetUserStats(context.Background(), &dtov1.GetUserStatReq{Range: 7, Bucket: "day"})
			},
			wantCalled: "user_day",
		},
		{
			name: "user hour",
			call: func(s *statService) (any, error) {
				return s.GetUserStats(context.Background(), &dtov1.GetUserStatReq{Range: 24, Bucket: "hour"})
			},
			wantCalled: "user_hour",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockStatRepo{
				clickStats: []*dtov1.ClickStatItem{{Time: now, Count: 1}},
				linkStats:  []*dtov1.LinkStatItem{{Time: now, Count: 2}},
				userStats:  []*dtov1.UserStatItem{{Time: now, Count: 3}},
			}
			svc := &statService{repo: repo}

			resp, err := tt.call(svc)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected stats response")
			}
			if repo.called != tt.wantCalled {
				t.Fatalf("expected repo call %q, got %q", tt.wantCalled, repo.called)
			}
		})
	}
}

func TestStatService_InvalidBucket(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		call func(*statService) error
	}{
		{
			name: "link",
			call: func(s *statService) error {
				_, err := s.GetLinkStats(context.Background(), &dtov1.GetLinkStatReq{Range: 7, Bucket: "week"})
				return err
			},
		},
		{
			name: "click",
			call: func(s *statService) error {
				_, err := s.GetClickStats(context.Background(), &dtov1.GetClickStatReq{Range: 7, Bucket: "week"})
				return err
			},
		},
		{
			name: "user",
			call: func(s *statService) error {
				_, err := s.GetUserStats(context.Background(), &dtov1.GetUserStatReq{Range: 7, Bucket: "week"})
				return err
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &mockStatRepo{}
			svc := &statService{repo: repo}

			err := tt.call(svc)
			if !errors.Is(err, consts.ErrInvalidBucket) {
				t.Fatalf("expected ErrInvalidBucket, got %v", err)
			}
			if repo.called != "" {
				t.Fatalf("invalid bucket should not call repo, got %q", repo.called)
			}
		})
	}
}

func TestStatService_PropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("repo unavailable")
	repo := &mockStatRepo{err: wantErr}
	svc := &statService{repo: repo}

	_, err := svc.GetClickStats(context.Background(), &dtov1.GetClickStatReq{Range: 1, Bucket: "day"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
