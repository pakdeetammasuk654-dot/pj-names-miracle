package usecase

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"sync"
	"time"

	"go-backend/domain"
)

const (
	// Exact pools can contain tens of thousands of rows. Keep only a small
	// bounded number so complete ranking does not turn the cache into gigabytes.
	defaultCandidateCacheCapacity = 8
	defaultCandidateCacheTTL      = 10 * time.Minute
	candidateTokenBytes           = 32
)

type cachedCandidatePool struct {
	request          domain.NameSearchRequest
	surnameSatSum    int
	surnameShaSum    int
	goodNumbers      map[string]domain.NumberMeaning
	soundCalibration domain.SoundCalibrationProfile
	semanticCorpus   domain.SemanticCorpusProvenance
	candidates       []domain.NameResultItem
}

type candidateCacheEntry struct {
	pool      cachedCandidatePool
	createdAt time.Time
	expiresAt time.Time
}

// candidatePoolCache is a bounded in-memory cache. The mutex protects both the
// entry map and eviction, while defensive copies keep concurrent reranks from
// sharing mutable slices or request pointers.
type candidatePoolCache struct {
	mu       sync.Mutex
	entries  map[string]candidateCacheEntry
	capacity int
	ttl      time.Duration
	now      func() time.Time
	newToken func() (string, error)
}

func newCandidatePoolCache(capacity int, ttl time.Duration) *candidatePoolCache {
	if capacity <= 0 {
		capacity = defaultCandidateCacheCapacity
	}
	if ttl <= 0 {
		ttl = defaultCandidateCacheTTL
	}
	return &candidatePoolCache{
		entries:  make(map[string]candidateCacheEntry, capacity),
		capacity: capacity,
		ttl:      ttl,
		now:      time.Now,
		newToken: generateCandidateToken,
	}
}

func generateCandidateToken() (string, error) {
	raw := make([]byte, candidateTokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

func (c *candidatePoolCache) Put(pool cachedCandidatePool) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	c.removeExpiredLocked(now)
	if len(c.entries) >= c.capacity {
		c.evictOldestLocked()
	}

	for attempts := 0; attempts < 3; attempts++ {
		token, err := c.newToken()
		if err != nil {
			return "", err
		}
		if _, exists := c.entries[token]; exists {
			continue
		}
		c.entries[token] = candidateCacheEntry{
			pool:      cloneCandidatePool(pool),
			createdAt: now,
			expiresAt: now.Add(c.ttl),
		}
		return token, nil
	}

	return "", errors.New("failed to generate unique candidate token")
}

func (c *candidatePoolCache) Get(token string) (cachedCandidatePool, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	entry, ok := c.entries[token]
	if !ok {
		return cachedCandidatePool{}, false
	}
	if !now.Before(entry.expiresAt) {
		delete(c.entries, token)
		return cachedCandidatePool{}, false
	}
	return cloneCandidatePool(entry.pool), true
}

func (c *candidatePoolCache) removeExpiredLocked(now time.Time) {
	for token, entry := range c.entries {
		if !now.Before(entry.expiresAt) {
			delete(c.entries, token)
		}
	}
}

func (c *candidatePoolCache) evictOldestLocked() {
	var oldestToken string
	var oldestTime time.Time
	for token, entry := range c.entries {
		if oldestToken == "" || entry.createdAt.Before(oldestTime) {
			oldestToken = token
			oldestTime = entry.createdAt
		}
	}
	if oldestToken != "" {
		delete(c.entries, oldestToken)
	}
}

func cloneCandidatePool(pool cachedCandidatePool) cachedCandidatePool {
	cloned := pool
	cloned.request = cloneSearchRequest(pool.request)
	cloned.goodNumbers = make(map[string]domain.NumberMeaning, len(pool.goodNumbers))
	for key, value := range pool.goodNumbers {
		cloned.goodNumbers[key] = value
	}
	cloned.soundCalibration = cloneSoundCalibrationProfile(pool.soundCalibration)
	cloned.candidates = cloneNameResults(pool.candidates)
	return cloned
}

func cloneSoundCalibrationProfile(profile domain.SoundCalibrationProfile) domain.SoundCalibrationProfile {
	cloned := profile
	cloned.Versions = make(map[string]domain.SoundCalibrationStats, len(profile.Versions))
	for version, stats := range profile.Versions {
		cloned.Versions[version] = stats
	}
	return cloned
}

func cloneSearchRequest(req domain.NameSearchRequest) domain.NameSearchRequest {
	cloned := req
	cloned.NumerologyWeight = cloneFloat(req.NumerologyWeight)
	cloned.SemanticWeight = cloneFloat(req.SemanticWeight)
	cloned.SoundWeight = cloneFloat(req.SoundWeight)
	return cloned
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func cloneNameResults(results []domain.NameResultItem) []domain.NameResultItem {
	cloned := make([]domain.NameResultItem, len(results))
	copy(cloned, results)
	for i := range cloned {
		if results[i].RankingReasons != nil {
			cloned[i].RankingReasons = append([]string(nil), results[i].RankingReasons...)
		}
	}
	return cloned
}
