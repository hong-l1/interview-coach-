package handler

import (
	"awesomeProject4/backend/api/validate"
	"awesomeProject4/backend/api/handler/interview_run"
	"awesomeProject4/backend/repository/dao"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func BenchmarkBuildInterviewEntry(b *testing.B) {
	deps := newBenchmarkInterviewEntryDeps()
	req := &validate.InterviewQuestionRequest{
		InterviewType: "specialized",
		Domain:        "backend",
		ResumeID:      1,
		Difficulty:    "medium",
		Company:       "ByteDance",
		Position:      "Go Backend",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry, err := buildInterviewEntry(context.Background(), req, 1, deps)
		if err != nil {
			b.Fatal(err)
		}
		if entry.session.SessionID == "" {
			b.Fatal("session id should not be empty")
		}
	}
}

func BenchmarkBuildInterviewEntryParallel(b *testing.B) {
	deps := newBenchmarkInterviewEntryDeps()
	req := newInterviewEntryBenchmarkRequest()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			entry, err := buildInterviewEntry(context.Background(), req, 1, deps)
			if err != nil {
				b.Fatal(err)
			}
			if entry.session.RecordID == 0 {
				b.Fatal("record id should not be zero")
			}
		}
	})
}

func BenchmarkBuildInterviewEntryIntegration(b *testing.B) {
	cfg := mustLoadInterviewEntryIntegrationConfig(b)
	db := mustOpenInterviewEntryBenchmarkDB(b, cfg)
	redisClient := mustOpenInterviewEntryBenchmarkRedis(b, cfg)
	deps, cleanupSession := newIntegrationInterviewEntryDeps(b, db, redisClient, cfg)
	req := newInterviewEntryBenchmarkRequest()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry, err := buildInterviewEntry(context.Background(), req, 1, deps)
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		cleanupSession(entry.session.SessionID)
		b.StartTimer()
	}
}

func BenchmarkBuildInterviewEntryIntegrationParallel(b *testing.B) {
	cfg := mustLoadInterviewEntryIntegrationConfig(b)
	db := mustOpenInterviewEntryBenchmarkDB(b, cfg)
	redisClient := mustOpenInterviewEntryBenchmarkRedis(b, cfg)
	deps, cleanupSession := newIntegrationInterviewEntryDeps(b, db, redisClient, cfg)
	req := newInterviewEntryBenchmarkRequest()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			entry, err := buildInterviewEntry(context.Background(), req, 1, deps)
			if err != nil {
				b.Fatal(err)
			}
			b.StopTimer()
			cleanupSession(entry.session.SessionID)
			b.StartTimer()
		}
	})
}

func BenchmarkBuildInterviewEntryIntegrationBreakdown(b *testing.B) {
	cfg := mustLoadInterviewEntryIntegrationConfig(b)
	db := mustOpenInterviewEntryBenchmarkDB(b, cfg)
	redisClient := mustOpenInterviewEntryBenchmarkRedis(b, cfg)
	deps, cleanupSession, metrics := newBreakdownInterviewEntryDeps(b, db, redisClient, cfg)
	req := newInterviewEntryBenchmarkRequest()

	b.ReportAllocs()
	b.ResetTimer()
	startAll := time.Now()
	for i := 0; i < b.N; i++ {
		entry, err := buildInterviewEntry(context.Background(), req, 1, deps)
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		cleanupSession(entry.session.SessionID)
		b.StartTimer()
	}
	total := time.Since(startAll)

	mysqlTotal := time.Duration(atomic.LoadInt64(&metrics.mysqlNanos))
	redisTotal := time.Duration(atomic.LoadInt64(&metrics.redisNanos))
	otherTotal := total - mysqlTotal - redisTotal
	if otherTotal < 0 {
		otherTotal = 0
	}

	totalMs := float64(total.Nanoseconds()) / 1e6
	mysqlMs := float64(mysqlTotal.Nanoseconds()) / 1e6
	redisMs := float64(redisTotal.Nanoseconds()) / 1e6
	otherMs := float64(otherTotal.Nanoseconds()) / 1e6

	b.ReportMetric(totalMs/float64(b.N), "total_ms/op")
	b.ReportMetric(mysqlMs/float64(b.N), "mysql_ms/op")
	b.ReportMetric(redisMs/float64(b.N), "redis_ms/op")
	b.ReportMetric(otherMs/float64(b.N), "other_ms/op")
	if totalMs > 0 {
		b.ReportMetric(mysqlMs/totalMs*100, "mysql_pct")
		b.ReportMetric(redisMs/totalMs*100, "redis_pct")
		b.ReportMetric(otherMs/totalMs*100, "other_pct")
	}
}

func BenchmarkSubmitAnswerPath(b *testing.B) {
	deps := newBenchmarkSubmitAnswerDeps()
	req := newSubmitAnswerBenchmarkRequest()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := runSubmitAnswerPath(context.Background(), req, deps); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSubmitAnswerPathParallel(b *testing.B) {
	deps := newBenchmarkSubmitAnswerDeps()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := newSubmitAnswerBenchmarkRequest()
			if err := runSubmitAnswerPath(context.Background(), req, deps); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkSubmitAnswerPathIntegration(b *testing.B) {
	cfg := mustLoadInterviewEntryIntegrationConfig(b)
	redisClient := mustOpenInterviewEntryBenchmarkRedis(b, cfg)
	deps, cleanupSession := newIntegrationSubmitAnswerDeps(b, redisClient, cfg)

	b.ReportAllocs()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		req := newSubmitAnswerBenchmarkRequest()
		if err := deps.prepare(context.Background(), req.SessionID); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()
		if err := runSubmitAnswerPath(context.Background(), req, deps.submitAnswerDeps); err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		cleanupSession(req.SessionID)
	}
}

func BenchmarkSubmitAnswerPathIntegrationBreakdown(b *testing.B) {
	cfg := mustLoadInterviewEntryIntegrationConfig(b)
	redisClient := mustOpenInterviewEntryBenchmarkRedis(b, cfg)
	deps, cleanupSession, metrics := newBreakdownSubmitAnswerDeps(b, redisClient, cfg)

	b.ReportAllocs()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		req := newSubmitAnswerBenchmarkRequest()
		prepareStart := time.Now()
		if err := deps.prepare(context.Background(), req.SessionID); err != nil {
			b.Fatal(err)
		}
		atomic.AddInt64(&metrics.prepareNanos, time.Since(prepareStart).Nanoseconds())
		runStart := time.Now()
		b.StartTimer()
		if err := runSubmitAnswerPath(context.Background(), req, deps.submitAnswerDeps); err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		atomic.AddInt64(&metrics.totalNanos, time.Since(runStart).Nanoseconds())
		b.StopTimer()
		cleanupSession(req.SessionID)
	}

	total := time.Duration(atomic.LoadInt64(&metrics.totalNanos))
	prepareTotal := time.Duration(atomic.LoadInt64(&metrics.prepareNanos))
	redisTotal := time.Duration(atomic.LoadInt64(&metrics.redisNanos))
	sendTotal := time.Duration(atomic.LoadInt64(&metrics.sendNanos))
	otherTotal := total - redisTotal - sendTotal
	if otherTotal < 0 {
		otherTotal = 0
	}

	totalMs := float64(total.Nanoseconds()) / 1e6
	prepareMs := float64(prepareTotal.Nanoseconds()) / 1e6
	redisMs := float64(redisTotal.Nanoseconds()) / 1e6
	sendMs := float64(sendTotal.Nanoseconds()) / 1e6
	otherMs := float64(otherTotal.Nanoseconds()) / 1e6

	b.ReportMetric(totalMs/float64(b.N), "total_ms/op")
	b.ReportMetric(prepareMs/float64(b.N), "prepare_ms/op")
	b.ReportMetric(redisMs/float64(b.N), "redis_ms/op")
	b.ReportMetric(sendMs/float64(b.N), "send_ms/op")
	b.ReportMetric(otherMs/float64(b.N), "other_ms/op")
	if totalMs > 0 {
		b.ReportMetric(redisMs/totalMs*100, "redis_pct")
		b.ReportMetric(sendMs/totalMs*100, "send_pct")
		b.ReportMetric(otherMs/totalMs*100, "other_pct")
	}
}

func newBenchmarkInterviewEntryDeps() interviewEntryDeps {
	var recordID uint64
	var sessionID uint64
	var sessionStore sync.Map

	return interviewEntryDeps{
		createRecord: func(_ context.Context, record *dao.InterviewRecord) error {
			record.ID = atomic.AddUint64(&recordID, 1)
			return nil
		},
		saveSession: func(_ context.Context, session *interviewSession, _ time.Duration) error {
			sessionStore.Store(session.SessionID, *session)
			return nil
		},
		newSessionID: func() string {
			id := atomic.AddUint64(&sessionID, 1)
			return fmt.Sprintf("bench-session-%d", id)
		},
		now: time.Now,
	}
}

func newInterviewEntryBenchmarkRequest() *validate.InterviewQuestionRequest {
	return &validate.InterviewQuestionRequest{
		InterviewType: "specialized",
		Domain:        "backend",
		ResumeID:      1,
		Difficulty:    "medium",
		Company:       "ByteDance",
		Position:      "Go Backend",
	}
}

func newSubmitAnswerBenchmarkRequest() *validate.SubmitAnswerRequest {
	return &validate.SubmitAnswerRequest{
		SessionID: fmt.Sprintf("bench-submit-%s", uuid.NewString()),
		Answer:    "This is a benchmark answer.",
	}
}

type submitAnswerDeps struct {
	evalSession func(ctx context.Context, sessionID string, now int64, ttl time.Duration) (int, error)
	sendAnswer  func(ctx context.Context, sessionID string, answer string) error
}

func runSubmitAnswerPath(ctx context.Context, req *validate.SubmitAnswerRequest, deps submitAnswerDeps) error {
	now := time.Now().Unix()
	result, err := deps.evalSession(ctx, req.SessionID, now, interviewSessionTTL)
	if err != nil {
		return err
	}

	switch result {
	case 0:
		return deps.sendAnswer(ctx, req.SessionID, req.Answer)
	case 1:
		return errors.New("session not found")
	case 2:
		return errors.New("interview session status is invalid")
	default:
		return errors.New("unknown submit answer result")
	}
}

func newBenchmarkSubmitAnswerDeps() submitAnswerDeps {
	var sessionStore sync.Map

	return submitAnswerDeps{
		evalSession: func(_ context.Context, sessionID string, now int64, _ time.Duration) (int, error) {
			value, ok := sessionStore.Load(sessionID)
			if !ok {
				sessionStore.Store(sessionID, benchmarkSessionState{
					Status:       interviewStatusActive,
					CurrentIndex: 1,
					UpdatedAt:    now,
				})
				return 0, nil
			}
			session := value.(benchmarkSessionState)
			if session.Status != interviewStatusActive {
				return 2, nil
			}
			session.CurrentIndex++
			session.UpdatedAt = now
			sessionStore.Store(sessionID, session)
			return 0, nil
		},
		sendAnswer: func(_ context.Context, _ string, _ string) error {
			return nil
		},
	}
}

type benchmarkSessionState struct {
	Status       string
	CurrentIndex int64
	UpdatedAt    int64
}

type integrationSubmitAnswerDeps struct {
	submitAnswerDeps submitAnswerDeps
	prepare          func(ctx context.Context, sessionID string) error
}

type breakdownSubmitAnswerMetrics struct {
	totalNanos   int64
	prepareNanos int64
	redisNanos int64
	sendNanos  int64
}

type interviewEntryIntegrationConfig struct {
	mysqlDSN      string
	redisAddr     string
	redisPassword string
	redisDB       int
	sessionPrefix string
}

func mustLoadInterviewEntryIntegrationConfig(b *testing.B) interviewEntryIntegrationConfig {
	b.Helper()

	if os.Getenv("BENCH_ENTRY_INTEGRATION") != "1" {
		b.Skip("set BENCH_ENTRY_INTEGRATION=1 to enable real Redis/MySQL benchmark")
	}

	mysqlDSN := os.Getenv("BENCH_MYSQL_DSN")
	if mysqlDSN == "" {
		b.Skip("set BENCH_MYSQL_DSN to a writable benchmark database")
	}

	redisAddr := os.Getenv("BENCH_REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}

	redisDB := 0
	if rawDB := os.Getenv("BENCH_REDIS_DB"); rawDB != "" {
		parsed, err := strconv.Atoi(rawDB)
		if err != nil {
			b.Fatalf("invalid BENCH_REDIS_DB: %v", err)
		}
		redisDB = parsed
	}

	sessionPrefix := os.Getenv("BENCH_SESSION_PREFIX")
	if sessionPrefix == "" {
		sessionPrefix = "bench-mianshi-session"
	}

	return interviewEntryIntegrationConfig{
		mysqlDSN:      mysqlDSN,
		redisAddr:     redisAddr,
		redisPassword: os.Getenv("BENCH_REDIS_PASSWORD"),
		redisDB:       redisDB,
		sessionPrefix: sessionPrefix,
	}
}

func mustOpenInterviewEntryBenchmarkDB(b *testing.B, cfg interviewEntryIntegrationConfig) *gorm.DB {
	b.Helper()

	db, err := gorm.Open(mysql.Open(cfg.mysqlDSN), &gorm.Config{
		DisableAutomaticPing:   true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		b.Fatalf("open mysql: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		b.Fatalf("get sql db: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		b.Fatalf("ping mysql: %v", err)
	}
	if err := db.AutoMigrate(&dao.InterviewRecord{}); err != nil {
		b.Fatalf("auto migrate interview_records: %v", err)
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	b.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func mustOpenInterviewEntryBenchmarkRedis(b *testing.B, cfg interviewEntryIntegrationConfig) *redis.Client {
	b.Helper()

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.redisAddr,
		Password: cfg.redisPassword,
		DB:       cfg.redisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Fatalf("ping redis: %v", err)
	}

	b.Cleanup(func() {
		_ = client.Close()
	})

	return client
}

func newIntegrationInterviewEntryDeps(
	b *testing.B,
	db *gorm.DB,
	redisClient *redis.Client,
	cfg interviewEntryIntegrationConfig,
) (interviewEntryDeps, func(sessionID string)) {
	b.Helper()

	cleanupSession := func(sessionID string) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := redisClient.Del(ctx, interviewSessionKey(sessionID)).Err(); err != nil {
			b.Fatalf("cleanup redis session: %v", err)
		}
	}

	return interviewEntryDeps{
		createRecord: func(ctx context.Context, record *dao.InterviewRecord) error {
			tx := db.WithContext(ctx).Begin()
			if tx.Error != nil {
				return tx.Error
			}
			if err := tx.Create(record).Error; err != nil {
				_ = tx.Rollback().Error
				return err
			}
			return tx.Rollback().Error
		},
		saveSession: func(ctx context.Context, session *interviewSession, ttl time.Duration) error {
			key := interviewSessionKey(session.SessionID)
			values := map[string]any{
				"session_id":     session.SessionID,
				"user_id":        session.UserID,
				"record_id":      session.RecordID,
				"resume_id":      session.ResumeID,
				"interview_type": session.InterviewType,
				"domain":         session.Domain,
				"company":        session.Company,
				"position":       session.Position,
				"status":         session.Status,
				"difficulty":     session.Difficulty,
				"current_index":  session.CurrentIndex,
				"created_at":     session.CreatedAt,
				"updated_at":     session.UpdatedAt,
				"ended_at":       session.EndedAt,
			}
			if err := redisClient.HSet(ctx, key, values).Err(); err != nil {
				return err
			}
			return redisClient.Expire(ctx, key, ttl).Err()
		},
		newSessionID: func() string {
			return fmt.Sprintf("%s-%s", cfg.sessionPrefix, uuid.NewString())
		},
		now: time.Now,
	}, cleanupSession
}

type breakdownInterviewEntryMetrics struct {
	mysqlNanos int64
	redisNanos int64
}

func newBreakdownInterviewEntryDeps(
	b *testing.B,
	db *gorm.DB,
	redisClient *redis.Client,
	cfg interviewEntryIntegrationConfig,
) (interviewEntryDeps, func(sessionID string), *breakdownInterviewEntryMetrics) {
	b.Helper()

	metrics := &breakdownInterviewEntryMetrics{}
	deps, cleanupSession := newIntegrationInterviewEntryDeps(b, db, redisClient, cfg)

	deps.createRecord = func(ctx context.Context, record *dao.InterviewRecord) error {
		start := time.Now()
		defer func() {
			atomic.AddInt64(&metrics.mysqlNanos, time.Since(start).Nanoseconds())
		}()
		tx := db.WithContext(ctx).Begin()
		if tx.Error != nil {
			return tx.Error
		}
		if err := tx.Create(record).Error; err != nil {
			_ = tx.Rollback().Error
			return err
		}
		return tx.Rollback().Error
	}

	deps.saveSession = func(ctx context.Context, session *interviewSession, ttl time.Duration) error {
		start := time.Now()
		defer func() {
			atomic.AddInt64(&metrics.redisNanos, time.Since(start).Nanoseconds())
		}()
		key := interviewSessionKey(session.SessionID)
		values := map[string]any{
			"session_id":     session.SessionID,
			"user_id":        session.UserID,
			"record_id":      session.RecordID,
			"resume_id":      session.ResumeID,
			"interview_type": session.InterviewType,
			"domain":         session.Domain,
			"company":        session.Company,
			"position":       session.Position,
			"status":         session.Status,
			"difficulty":     session.Difficulty,
			"current_index":  session.CurrentIndex,
			"created_at":     session.CreatedAt,
			"updated_at":     session.UpdatedAt,
			"ended_at":       session.EndedAt,
		}
		if err := redisClient.HSet(ctx, key, values).Err(); err != nil {
			return err
		}
		return redisClient.Expire(ctx, key, ttl).Err()
	}

	return deps, cleanupSession, metrics
}

func newIntegrationSubmitAnswerDeps(
	b *testing.B,
	redisClient *redis.Client,
	cfg interviewEntryIntegrationConfig,
) (integrationSubmitAnswerDeps, func(sessionID string)) {
	b.Helper()

	cleanupSession := func(sessionID string) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := redisClient.Del(ctx, interviewSessionKey(sessionID)).Err(); err != nil {
			b.Fatalf("cleanup redis session: %v", err)
		}
		interview_run.Store.DeleteInterviewRuntime(sessionID)
	}

	prepare := func(ctx context.Context, sessionID string) error {
		now := time.Now().Unix()
		values := map[string]any{
			"session_id":     sessionID,
			"user_id":        1,
			"record_id":      1,
			"resume_id":      1,
			"interview_type": "specialized",
			"domain":         "backend",
			"company":        "ByteDance",
			"position":       "Go Backend",
			"status":         interviewStatusActive,
			"difficulty":     "medium",
			"current_index":  0,
			"created_at":     now,
			"updated_at":     now,
			"ended_at":       0,
		}
		if err := redisClient.HSet(ctx, interviewSessionKey(sessionID), values).Err(); err != nil {
			return err
		}
		if err := redisClient.Expire(ctx, interviewSessionKey(sessionID), interviewSessionTTL).Err(); err != nil {
			return err
		}
		interview_run.Store.SetInterviewRuntime(sessionID, &interview_run.InterviewRuntime{})
		return nil
	}

	return integrationSubmitAnswerDeps{
		submitAnswerDeps: submitAnswerDeps{
			evalSession: func(ctx context.Context, sessionID string, now int64, ttl time.Duration) (int, error) {
				return redisClient.Eval(
					ctx,
					subAnswerScript,
					[]string{interviewSessionKey(sessionID)},
					interviewStatusActive,
					now,
					int64(ttl/time.Second),
				).Int()
			},
			sendAnswer: func(ctx context.Context, sessionID string, answer string) error {
				runtime := interview_run.Store.GetInterviewRuntime(sessionID)
				if runtime == nil {
					return errors.New("interview runtime not found")
				}
				return nil
			},
		},
		prepare: prepare,
	}, cleanupSession
}

func newBreakdownSubmitAnswerDeps(
	b *testing.B,
	redisClient *redis.Client,
	cfg interviewEntryIntegrationConfig,
) (integrationSubmitAnswerDeps, func(sessionID string), *breakdownSubmitAnswerMetrics) {
	b.Helper()

	baseDeps, cleanupSession := newIntegrationSubmitAnswerDeps(b, redisClient, cfg)
	metrics := &breakdownSubmitAnswerMetrics{}

	baseDeps.submitAnswerDeps.evalSession = func(ctx context.Context, sessionID string, now int64, ttl time.Duration) (int, error) {
		start := time.Now()
		defer func() {
			atomic.AddInt64(&metrics.redisNanos, time.Since(start).Nanoseconds())
		}()
		return redisClient.Eval(
			ctx,
			subAnswerScript,
			[]string{interviewSessionKey(sessionID)},
			interviewStatusActive,
			now,
			int64(ttl/time.Second),
		).Int()
	}

	baseDeps.submitAnswerDeps.sendAnswer = func(ctx context.Context, sessionID string, answer string) error {
		start := time.Now()
		defer func() {
			atomic.AddInt64(&metrics.sendNanos, time.Since(start).Nanoseconds())
		}()
		runtime := interview_run.Store.GetInterviewRuntime(sessionID)
		if runtime == nil {
			return errors.New("interview runtime not found")
		}
		return nil
	}

	return baseDeps, cleanupSession, metrics
}
