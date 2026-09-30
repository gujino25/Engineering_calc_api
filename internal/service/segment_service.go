package service

import (
	"context"
	"encoding/json"
	"enginer/internal/domain"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type systemGetter interface {
	GetByID(ctx context.Context, id string) (domain.System, error)
}

type segmentStore interface {
	Create(ctx context.Context, segment domain.Segment) error
	ListBySystem(ctx context.Context, systemID string) ([]domain.Segment, error)
}

type SegmentService struct {
	segmentStore segmentStore
	systemStore  systemGetter
	cache        *redis.Client
}

func NewSegmentService(segmentStore segmentStore, systemStore systemGetter, cahe *redis.Client) *SegmentService {
	return &SegmentService{
		segmentStore: segmentStore,
		systemStore:  systemStore,
		cache:        cahe,
	}
}

func (s *SegmentService) CreateSegment(ctx context.Context, systemID, name string, shape domain.Shape, rect *domain.RectGeometry, round *domain.RoundGeometry, length float64) (domain.Segment, error) {
	if _, err := s.systemStore.GetByID(ctx, systemID); err != nil {
		return domain.Segment{}, err
	}
	key := fmt.Sprintf("segments:system:%s", systemID)

	segment, err := domain.NewSegment(systemID, name, shape, rect, round, length)
	if err != nil {
		return domain.Segment{}, err
	}

	if err := s.segmentStore.Create(ctx, segment); err != nil {
		return domain.Segment{}, err
	}

	if err := s.cache.Del(ctx, key).Err(); err != nil {
		log.Printf("redis delete cache value failed: %v", err)
	}
	return segment, nil
}

func (s *SegmentService) ListBySystem(ctx context.Context, systemID string) ([]domain.Segment, error) {
	if _, err := s.systemStore.GetByID(ctx, systemID); err != nil {
		return nil, err
	}
	key := fmt.Sprintf("segments:system:%s", systemID)
	cached, err := s.cache.Get(ctx, key).Result()
	if err == nil {
		var segments []domain.Segment
		unmarshalErr := json.Unmarshal([]byte(cached), &segments)
		if unmarshalErr == nil {
			return segments, nil
		}
		log.Printf("unmarshal failed: %v", unmarshalErr)
	} else if !errors.Is(err, redis.Nil) {
		log.Printf("redis get failed: %v", err)
	}

	segments, err := s.segmentStore.ListBySystem(ctx, systemID)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(segments)
	if err != nil {
		log.Printf("marshal segments for cache: %v", err)
		return segments, nil
	}

	if err := s.cache.Set(ctx, key, data, 30*time.Second).Err(); err != nil {
		log.Printf("redis set failed: %v", err)
	}
	return segments, nil
}
