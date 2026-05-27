package repository

import (
	"awesomeProject4/backend/repository/dao"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const resumeCacheTTL = 24 * time.Hour

type ResumeRepository struct {
	resumeDAO   *dao.ResumeDAO
	redisClient redis.Cmdable
}

func NewResumeRepository(resumeDAO *dao.ResumeDAO, redisClient redis.Cmdable) *ResumeRepository {
	return &ResumeRepository{
		resumeDAO:   resumeDAO,
		redisClient: redisClient,
	}
}

func (r *ResumeRepository) Create(ctx context.Context, resume *dao.Resume) error {
	if err := r.resumeDAO.Create(ctx, resume); err != nil {
		return err
	}
	return r.setResumeCache(ctx, resume)
}

func (r *ResumeRepository) GetByID(ctx context.Context, id uint64) (*dao.Resume, error) {
	if resume, err := r.getResumeCache(ctx, id); err == nil && resume != nil {
		return resume, nil
	}
	resume, err := r.resumeDAO.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := r.setResumeCache(ctx, resume); err != nil {
		return resume, nil
	}
	return resume, nil
}

func (r *ResumeRepository) GetByIDAndUserID(ctx context.Context, id uint64, userID uint) (*dao.Resume, error) {
	if resume, err := r.getResumeCache(ctx, id); err == nil && resume != nil && resume.UserID == userID {
		return resume, nil
	}
	resume, err := r.resumeDAO.GetByIDAndUserID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if err := r.setResumeCache(ctx, resume); err != nil {
		return resume, nil
	}
	return resume, nil
}

func (r *ResumeRepository) ListByUserID(ctx context.Context, userID uint) ([]dao.Resume, error) {
	return r.resumeDAO.ListByUserID(ctx, userID)
}

func (r *ResumeRepository) GetDefaultByUserID(ctx context.Context, userID uint) (*dao.Resume, error) {
	return r.resumeDAO.GetDefaultByUserID(ctx, userID)
}

func (r *ResumeRepository) Delete(ctx context.Context, id uint64, userID uint) error {
	if _, err := r.redisClient.Exists(ctx, resumeCacheKey(id)).Result(); err == nil {
		if err := r.redisClient.Del(ctx, resumeCacheKey(id)).Err(); err != nil {
			return err
		}
	}
	return r.resumeDAO.Delete(ctx, id, userID)
}

func (r *ResumeRepository) SetDefault(ctx context.Context, id uint64, userID uint) error {
	if err := r.resumeDAO.ClearDefault(ctx, userID); err != nil {
		return err
	}
	if _, err := r.redisClient.Exists(ctx, resumeCacheKey(id)).Result(); err == nil {
		_ = r.redisClient.Del(ctx, resumeCacheKey(id)).Err()
	}
	return r.resumeDAO.SetDefault(ctx, id, userID)
}

func (r *ResumeRepository) getResumeCache(ctx context.Context, id uint64) (*dao.Resume, error) {
	val, err := r.redisClient.Get(ctx, resumeCacheKey(id)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}
	var resume dao.Resume
	if err := json.Unmarshal([]byte(val), &resume); err != nil {
		return nil, err
	}
	return &resume, nil
}

func (r *ResumeRepository) setResumeCache(ctx context.Context, resume *dao.Resume) error {
	if resume == nil {
		return nil
	}
	payload, err := json.Marshal(resume)
	if err != nil {
		return err
	}
	return r.redisClient.Set(ctx, resumeCacheKey(resume.ID), payload, resumeCacheTTL).Err()
}

func resumeCacheKey(id uint64) string {
	return fmt.Sprintf("resume:%d", id)
}
