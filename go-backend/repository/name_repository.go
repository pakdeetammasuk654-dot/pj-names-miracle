package repository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
	"go-backend/domain"
)

var (
	vectorCache           = make(map[string]string)
	vectorCacheMu         sync.RWMutex
	randomMeaningsCache   []string
	randomMeaningsCacheMu sync.RWMutex
	randomMeaningsCacheAt time.Time
)

const maxVectorCacheEntries = 256

type NameRepository interface {
	GetGoodNumbersMap(ctx context.Context) (map[string]domain.NumberMeaning, error)
	GetAllNumbersMap(ctx context.Context) (map[string]domain.NumberMeaning, error)
	GetSurnameSums(ctx context.Context, surname string) (satSum int, shaSum int, err error)
	GetNameSumMaximums(ctx context.Context) (maxSatSum int, maxShaSum int, err error)
	GetSoundCalibrationProfile(ctx context.Context) (domain.SoundCalibrationProfile, error)
	GetSemanticCorpusProvenance(ctx context.Context) (domain.SemanticCorpusProvenance, error)
	GetSumsByTHName(ctx context.Context, thname string) (satSum int, shaSum int, meaning string, err error)
	GetMeaningsByNameIDs(ctx context.Context, nameIDs []int) (map[int]string, error)
	SearchCandidateNames(ctx context.Context, targetSatSums []int, targetShaSums []int, dayOfBirth string, wishQuery string, limit int) ([]domain.NameResultItem, error)
	SearchSemanticWishNames(ctx context.Context, targetSatSums []int, targetShaSums []int, wishQuery string, dayOfBirth string, limit int) ([]domain.NameResultItem, error)
	SearchSimilarNames(ctx context.Context, targetSatSums []int, targetShaSums []int, seedName string, dayOfBirth string, limit int) ([]domain.NameResultItem, error)
	SuggestNames(ctx context.Context, seedName string, limit int) ([]domain.NameResultItem, error)
	GetMeaningVectorByTHName(ctx context.Context, thname string) (vectorStr string, meaning string, err error)
	SearchSemanticAndSimilarNames(ctx context.Context, targetSatSums []int, targetShaSums []int, seedName string, meaningQuery string, dayOfBirth string, limit int) ([]domain.NameResultItem, error)
	GetSampleNames(ctx context.Context) ([]domain.Celebrity, error)
	GetRandomMeanings(ctx context.Context, limit int) ([]string, error)
}

func (r *nameRepository) GetMeaningsByNameIDs(ctx context.Context, nameIDs []int) (map[int]string, error) {
	result := make(map[int]string, len(nameIDs))
	if len(nameIDs) == 0 {
		return result, nil
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT name_id, COALESCE(meaning, '')
		FROM names_miracle
		WHERE name_id = ANY($1)`, pq.Array(nameIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch result meanings: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var nameID int
		var meaning string
		if err := rows.Scan(&nameID, &meaning); err != nil {
			return nil, fmt.Errorf("failed to scan result meaning: %w", err)
		}
		result[nameID] = meaning
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading result meanings: %w", err)
	}
	return result, nil
}

// GetNameSumMaximums returns the real search range from names_miracle so the
// combined-mode filter can derive every base sum compatible with a surname.
func (r *nameRepository) GetNameSumMaximums(ctx context.Context) (int, int, error) {
	var maxSatSum, maxShaSum int
	query := `SELECT COALESCE(MAX(sat_sum), 0), COALESCE(MAX(sha_sum), 0) FROM names_miracle`
	err := r.db.QueryRowContext(ctx, query).Scan(&maxSatSum, &maxShaSum)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get name sum maximums: %w", err)
	}
	return maxSatSum, maxShaSum, nil
}

// GetRandomMeanings returns distinct, non-empty meanings from an in-memory cache
// refreshed periodically to eliminate constant DB table scan pressure.
func (r *nameRepository) GetRandomMeanings(ctx context.Context, limit int) ([]string, error) {
	if limit < 1 {
		limit = 4
	}
	if limit > 108 {
		limit = 108
	}

	randomMeaningsCacheMu.RLock()
	hasValidCache := len(randomMeaningsCache) >= limit && time.Since(randomMeaningsCacheAt) < 30*time.Minute
	if hasValidCache {
		pool := make([]string, len(randomMeaningsCache))
		copy(pool, randomMeaningsCache)
		randomMeaningsCacheMu.RUnlock()

		rand.Shuffle(len(pool), func(i, j int) {
			pool[i], pool[j] = pool[j], pool[i]
		})
		return pool[:limit], nil
	}
	randomMeaningsCacheMu.RUnlock()

	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT BTRIM(meaning)
		FROM names_miracle
		WHERE meaning IS NOT NULL AND BTRIM(meaning) <> ''
		LIMIT 500`)
	if err != nil {
		return nil, fmt.Errorf("failed to get random meanings: %w", err)
	}
	defer rows.Close()

	meanings := make([]string, 0, 500)
	for rows.Next() {
		var meaning string
		if err := rows.Scan(&meaning); err != nil {
			return nil, fmt.Errorf("failed to scan random meaning: %w", err)
		}
		meanings = append(meanings, meaning)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to read random meanings: %w", err)
	}

	if len(meanings) > 0 {
		randomMeaningsCacheMu.Lock()
		randomMeaningsCache = meanings
		randomMeaningsCacheAt = time.Now()
		randomMeaningsCacheMu.Unlock()
	}

	pool := make([]string, len(meanings))
	copy(pool, meanings)
	rand.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})
	if len(pool) > limit {
		pool = pool[:limit]
	}
	return pool, nil
}

// GetSoundCalibrationProfile prefers the latest approved immutable snapshot.
// Older development databases may not have the governance schema yet; only in
// that case (or when no approved batch exists) do we use explicitly provisional
// live statistics. The selected profile is cached for stable process behavior.
func (r *nameRepository) GetSoundCalibrationProfile(ctx context.Context) (domain.SoundCalibrationProfile, error) {
	r.soundCalibrationMu.Lock()
	defer r.soundCalibrationMu.Unlock()

	if r.soundCalibration != nil {
		return cloneSoundCalibrationProfile(*r.soundCalibration), nil
	}

	profile, found, err := r.loadApprovedSoundCalibrationProfile(ctx)
	if err != nil && !isGovernanceSchemaUnavailable(err) {
		return domain.SoundCalibrationProfile{}, fmt.Errorf("failed to load approved sound calibration snapshot: %w", err)
	}
	var currentCorpus soundCorpusIdentity
	if found {
		currentCorpus, err = r.loadCurrentSoundCorpusIdentity(ctx)
		if err != nil {
			return domain.SoundCalibrationProfile{}, fmt.Errorf("failed to validate approved sound calibration corpus: %w", err)
		}
		found = approvedProfileMatchesCorpus(profile, currentCorpus)
	}
	if !found {
		profile, currentCorpus, err = r.loadLiveSoundCalibrationProfile(ctx)
		if err != nil {
			return domain.SoundCalibrationProfile{}, err
		}
		if !soundProfileCountsMatchCorpus(profile, currentCorpus) {
			return domain.SoundCalibrationProfile{}, errors.New("sound corpus changed while loading provisional calibration profile")
		}
		profile.CorpusFingerprintSHA256 = currentCorpus.FingerprintSHA256
	}
	profile.LoadedAt = time.Now().UTC().Format(time.RFC3339Nano)

	stored := cloneSoundCalibrationProfile(profile)
	r.soundCalibration = &stored
	return cloneSoundCalibrationProfile(profile), nil
}

// GetSemanticCorpusProvenance verifies that stored candidate vectors have a
// reviewed snapshot compatible with the exact runtime query contract. Missing
// governance schema/approval is an expected unavailable state, not an error.
func (r *nameRepository) GetSemanticCorpusProvenance(ctx context.Context) (domain.SemanticCorpusProvenance, error) {
	r.semanticProvenanceMu.Lock()
	defer r.semanticProvenanceMu.Unlock()

	if r.semanticProvenanceChecked {
		return r.semanticProvenance, nil
	}

	unavailable := domain.SemanticCorpusProvenance{
		Source:               "unavailable_unapproved_vector_corpus",
		EmbeddingModel:       domain.SemanticEmbeddingModel,
		InputTemplateVersion: domain.SemanticInputTemplateVersion,
		VectorDimensions:     domain.SemanticVectorDimensions,
		CorpusFilterVersion:  domain.CorpusFilterVersion,
	}
	query := `
		SELECT snapshot_id, embedding_model, input_template_version, vector_dimensions,
			corpus_filter_version, corpus_row_count, vectorized_row_count,
			corpus_vector_sha256, corpus_membership_sha256,
			source_audit_at, BTRIM(evaluation_version)
		FROM data_governance.semantic_vector_snapshots
		WHERE embedding_model = $1
			AND input_template_version = $2
			AND vector_dimensions = $3
			AND corpus_filter_version = $4
			AND corpus_row_count > 0
			AND vectorized_row_count = corpus_row_count
			AND corpus_vector_sha256 ~ '^[0-9a-f]{64}$'
			AND corpus_membership_sha256 ~ '^[0-9a-f]{64}$'
			AND approved_at IS NOT NULL
			AND NULLIF(BTRIM(approved_by), '') IS NOT NULL
			AND NULLIF(BTRIM(approval_note), '') IS NOT NULL
			AND NULLIF(BTRIM(evaluation_version), '') IS NOT NULL
		ORDER BY source_audit_at DESC, snapshot_id DESC
		LIMIT 1
	`
	var profile domain.SemanticCorpusProvenance
	var sourceAuditAt time.Time
	err := r.db.QueryRowContext(
		ctx,
		query,
		domain.SemanticEmbeddingModel,
		domain.SemanticInputTemplateVersion,
		domain.SemanticVectorDimensions,
		domain.CorpusFilterVersion,
	).Scan(
		&profile.SnapshotID,
		&profile.EmbeddingModel,
		&profile.InputTemplateVersion,
		&profile.VectorDimensions,
		&profile.CorpusFilterVersion,
		&profile.CorpusRowCount,
		&profile.VectorizedRowCount,
		&profile.CorpusVectorSHA256,
		&profile.CorpusMembershipSHA256,
		&sourceAuditAt,
		&profile.EvaluationVersion,
	)
	if errors.Is(err, sql.ErrNoRows) || isGovernanceSchemaUnavailable(err) {
		r.semanticProvenance = ownerConfirmedLegacySemanticCorpus()
		r.semanticProvenanceChecked = true
		return r.semanticProvenance, nil
	}
	if err != nil {
		return domain.SemanticCorpusProvenance{}, fmt.Errorf("failed to verify semantic vector snapshot: %w", err)
	}
	if profile.CorpusRowCount != profile.VectorizedRowCount ||
		!isLowerHexSHA256(profile.CorpusVectorSHA256) ||
		!isLowerHexSHA256(profile.CorpusMembershipSHA256) ||
		strings.TrimSpace(profile.EvaluationVersion) == "" {
		r.semanticProvenance = unavailable
		r.semanticProvenanceChecked = true
		return unavailable, nil
	}
	linked, err := r.semanticSnapshotMatchesCurrentCorpus(ctx, profile)
	if err != nil {
		if isGovernanceSchemaUnavailable(err) {
			r.semanticProvenance = unavailable
			r.semanticProvenanceChecked = true
			return unavailable, nil
		}
		return domain.SemanticCorpusProvenance{}, fmt.Errorf("failed to validate semantic vector snapshot linkage: %w", err)
	}
	if !linked {
		unavailable.Source = "unavailable_stale_vector_corpus"
		r.semanticProvenance = unavailable
		r.semanticProvenanceChecked = true
		return unavailable, nil
	}
	profile.Approved = true
	profile.Source = "approved_snapshot"
	profile.SourceAuditAt = sourceAuditAt.UTC().Format(time.RFC3339Nano)
	r.semanticProvenance = profile
	r.semanticProvenanceChecked = true
	return profile, nil
}

// ownerConfirmedLegacySemanticCorpus enables the existing meaning_vector corpus
// when the product owner has confirmed that it was generated with the current
// embedding contract. Per-row vector presence and dimensions are still checked
// by the semantic candidate query itself.
func ownerConfirmedLegacySemanticCorpus() domain.SemanticCorpusProvenance {
	return domain.SemanticCorpusProvenance{
		Approved:             true,
		Source:               "owner_confirmed_legacy_embedding",
		EmbeddingModel:       domain.SemanticEmbeddingModel,
		InputTemplateVersion: domain.SemanticInputTemplateVersion,
		VectorDimensions:     domain.SemanticVectorDimensions,
		CorpusFilterVersion:  domain.CorpusFilterVersion,
		EvaluationVersion:    "owner_confirmed_text_embedding_3_small",
	}
}

// semanticSnapshotMatchesCurrentCorpus requires every currently eligible name
// and vector to be attached to the approved immutable snapshot. A metadata row
// by itself is deliberately insufficient proof: the nullable linkage column is
// installed with write guards by migration 000022.
func (r *nameRepository) semanticSnapshotMatchesCurrentCorpus(ctx context.Context, profile domain.SemanticCorpusProvenance) (bool, error) {
	query := fmt.Sprintf(`
		SELECT n.name_id, n.meaning_vector_snapshot_id,
			n.meaning_vector IS NOT NULL,
			CASE WHEN n.meaning_vector IS NULL THEN 0 ELSE vector_dims(n.meaning_vector) END
		FROM names_miracle AS n
		WHERE TRUE %s
		ORDER BY n.name_id ASC
	`, candidateEligibilityClause(""))
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	digest := sha256.New()
	var corpusCount, linkedVectorCount, correctDimensionCount int64
	for rows.Next() {
		var nameID int64
		var snapshotID sql.NullInt64
		var hasVector bool
		var vectorDimensions int
		if err := rows.Scan(&nameID, &snapshotID, &hasVector, &vectorDimensions); err != nil {
			return false, err
		}
		_, _ = fmt.Fprintf(digest, "%d\n", nameID)
		corpusCount++
		if snapshotID.Valid && snapshotID.Int64 == profile.SnapshotID && hasVector {
			linkedVectorCount++
			if vectorDimensions == profile.VectorDimensions {
				correctDimensionCount++
			}
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	membershipFingerprint := fmt.Sprintf("%x", digest.Sum(nil))
	return corpusCount == profile.CorpusRowCount &&
		linkedVectorCount == profile.VectorizedRowCount &&
		correctDimensionCount == linkedVectorCount &&
		corpusCount == linkedVectorCount &&
		membershipFingerprint == profile.CorpusMembershipSHA256, nil
}

func isLowerHexSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func (r *nameRepository) loadApprovedSoundCalibrationProfile(ctx context.Context) (domain.SoundCalibrationProfile, bool, error) {
	query := `
		WITH latest_approved AS (
			SELECT MAX(source_audit_at) AS source_audit_at
			FROM data_governance.phonetic_calibration_snapshots
			WHERE formula_version = $1 AND corpus_filter_version = $2
				AND approved_at IS NOT NULL AND NULLIF(BTRIM(evaluation_version), '') IS NOT NULL
				AND corpus_fingerprint_sha256 ~ '^[0-9a-f]{64}$'
		)
		SELECT COALESCE(NULLIF(BTRIM(snapshot.phonetic_version), ''), '<missing>'), snapshot.sample_count,
			snapshot.sound_mean, snapshot.sound_stddev, snapshot.source_audit_at,
			BTRIM(snapshot.evaluation_version), snapshot.corpus_fingerprint_sha256
		FROM data_governance.phonetic_calibration_snapshots AS snapshot
		JOIN latest_approved
			ON latest_approved.source_audit_at = snapshot.source_audit_at
		WHERE snapshot.formula_version = $1 AND snapshot.corpus_filter_version = $2
			AND snapshot.approved_at IS NOT NULL AND NULLIF(BTRIM(snapshot.evaluation_version), '') IS NOT NULL
			AND snapshot.corpus_fingerprint_sha256 ~ '^[0-9a-f]{64}$'
		ORDER BY snapshot.sample_count DESC, snapshot.phonetic_version ASC
	`
	rows, err := r.db.QueryContext(ctx, query, domain.SoundFormulaVersion, domain.CorpusFilterVersion)
	if err != nil {
		return domain.SoundCalibrationProfile{}, false, err
	}
	defer rows.Close()

	profile := domain.SoundCalibrationProfile{
		FormulaVersion: domain.SoundFormulaVersion,
		Source:         "approved_snapshot",
		Approved:       true,
		Versions:       make(map[string]domain.SoundCalibrationStats),
	}
	for rows.Next() {
		var version string
		var count int64
		var mean, stdDev float64
		var sourceAuditAt time.Time
		var evaluationVersion string
		var corpusFingerprint string
		if err := rows.Scan(&version, &count, &mean, &stdDev, &sourceAuditAt, &evaluationVersion, &corpusFingerprint); err != nil {
			return domain.SoundCalibrationProfile{}, false, err
		}
		if profile.EvaluationVersion == "" {
			profile.EvaluationVersion = evaluationVersion
		} else if profile.EvaluationVersion != evaluationVersion {
			return domain.SoundCalibrationProfile{}, false, fmt.Errorf("approved calibration batch mixes evaluation versions %q and %q", profile.EvaluationVersion, evaluationVersion)
		}
		if profile.CorpusFingerprintSHA256 == "" {
			profile.CorpusFingerprintSHA256 = corpusFingerprint
		} else if profile.CorpusFingerprintSHA256 != corpusFingerprint {
			return domain.SoundCalibrationProfile{}, false, fmt.Errorf("approved calibration batch mixes corpus fingerprints")
		}
		stats := domain.SoundCalibrationStats{Count: count, Mean: mean, StdDev: stdDev}
		profile.Versions[version] = stats
		if profile.ReferenceVersion == "" && stats.StdDev > 0 {
			profile.ReferenceVersion = version
			profile.ReferenceMean = stats.Mean
			profile.ReferenceStdDev = stats.StdDev
			profile.SourceAuditAt = sourceAuditAt.UTC().Format(time.RFC3339Nano)
		}
	}
	if err := rows.Err(); err != nil {
		return domain.SoundCalibrationProfile{}, false, err
	}
	return profile, profile.ReferenceVersion != "", nil
}

type soundCorpusIdentity struct {
	FingerprintSHA256 string
	VersionCounts     map[string]int64
	TotalCount        int64
}

// soundCorpusFingerprintRow is encoded as JSON (one object per line) so names
// containing punctuation cannot collide with field delimiters. Struct field
// order is part of the documented fingerprint contract.
type soundCorpusFingerprintRow struct {
	NameID            int64  `json:"name_id"`
	THName            string `json:"thname"`
	PhoneticVersion   string `json:"phonetic_version"`
	EuphonyScore      int    `json:"euphony_score"`
	PronunciationEase int    `json:"pronunciation_ease"`
	RhythmScore       int    `json:"rhythm_score"`
}

func (r *nameRepository) loadCurrentSoundCorpusIdentity(ctx context.Context) (soundCorpusIdentity, error) {
	query := fmt.Sprintf(`
		SELECT name_id, thname,
			COALESCE(NULLIF(BTRIM(phonetic_version), ''), '<missing>') AS version,
			COALESCE(euphony_score, 0), COALESCE(pronunciation_ease, 0), COALESCE(rhythm_score, 0)
		FROM names_miracle AS n
		WHERE TRUE %s
		ORDER BY name_id ASC
	`, candidateEligibilityClause(""))
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return soundCorpusIdentity{}, err
	}
	defer rows.Close()

	identity := soundCorpusIdentity{VersionCounts: make(map[string]int64)}
	digest := sha256.New()
	encoder := json.NewEncoder(digest)
	encoder.SetEscapeHTML(false)
	for rows.Next() {
		var nameID int64
		var thName, version string
		var euphony, pronunciation, rhythm int
		if err := rows.Scan(&nameID, &thName, &version, &euphony, &pronunciation, &rhythm); err != nil {
			return soundCorpusIdentity{}, err
		}
		if err := encoder.Encode(soundCorpusFingerprintRow{
			NameID:            nameID,
			THName:            thName,
			PhoneticVersion:   version,
			EuphonyScore:      euphony,
			PronunciationEase: pronunciation,
			RhythmScore:       rhythm,
		}); err != nil {
			return soundCorpusIdentity{}, fmt.Errorf("encode sound corpus fingerprint row %d: %w", nameID, err)
		}
		identity.VersionCounts[version]++
		identity.TotalCount++
	}
	if err := rows.Err(); err != nil {
		return soundCorpusIdentity{}, err
	}
	identity.FingerprintSHA256 = fmt.Sprintf("%x", digest.Sum(nil))
	return identity, nil
}

func approvedProfileMatchesCorpus(profile domain.SoundCalibrationProfile, current soundCorpusIdentity) bool {
	if !profile.Approved || profile.Source != "approved_snapshot" || profile.EvaluationVersion == "" ||
		!isLowerHexSHA256(profile.CorpusFingerprintSHA256) || current.TotalCount == 0 ||
		profile.CorpusFingerprintSHA256 != current.FingerprintSHA256 ||
		!soundProfileCountsMatchCorpus(profile, current) {
		return false
	}
	for version := range current.VersionCounts {
		stats, ok := profile.Versions[version]
		if !ok || stats.Count <= 1 || stats.StdDev <= 0 {
			return false
		}
	}
	return true
}

func soundProfileCountsMatchCorpus(profile domain.SoundCalibrationProfile, current soundCorpusIdentity) bool {
	if current.TotalCount == 0 || len(profile.Versions) != len(current.VersionCounts) {
		return false
	}
	var profileTotal int64
	for version, currentCount := range current.VersionCounts {
		stats, ok := profile.Versions[version]
		if !ok || stats.Count != currentCount {
			return false
		}
		profileTotal += stats.Count
	}
	return profileTotal == current.TotalCount
}

func isGovernanceSchemaUnavailable(err error) bool {
	var pqErr *pq.Error
	if !errors.As(err, &pqErr) {
		return false
	}
	code := string(pqErr.Code)
	return code == "42P01" || code == "3F000" || code == "42703" // undefined table, schema, or legacy column
}

func (r *nameRepository) loadLiveSoundCalibrationProfile(ctx context.Context) (domain.SoundCalibrationProfile, soundCorpusIdentity, error) {
	query := fmt.Sprintf(`
		SELECT name_id, thname,
			COALESCE(NULLIF(BTRIM(phonetic_version), ''), '<missing>') AS version,
			COALESCE(euphony_score, 0), COALESCE(pronunciation_ease, 0), COALESCE(rhythm_score, 0)
		FROM names_miracle AS n
		WHERE TRUE %s
		ORDER BY name_id ASC
	`, candidateEligibilityClause(""))
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return domain.SoundCalibrationProfile{}, soundCorpusIdentity{}, fmt.Errorf("failed to load live sound calibration corpus: %w", err)
	}
	defer rows.Close()

	profile := domain.SoundCalibrationProfile{
		FormulaVersion: domain.SoundFormulaVersion,
		Source:         "live_repository_distribution_provisional",
		Versions:       make(map[string]domain.SoundCalibrationStats),
	}
	identity := soundCorpusIdentity{VersionCounts: make(map[string]int64)}
	digest := sha256.New()
	encoder := json.NewEncoder(digest)
	encoder.SetEscapeHTML(false)
	accumulators := make(map[string]*soundStatsAccumulator)
	for rows.Next() {
		var nameID int64
		var thName, version string
		var euphony, pronunciation, rhythm int
		if err := rows.Scan(&nameID, &thName, &version, &euphony, &pronunciation, &rhythm); err != nil {
			return domain.SoundCalibrationProfile{}, soundCorpusIdentity{}, fmt.Errorf("failed to scan live sound calibration corpus: %w", err)
		}
		if err := encoder.Encode(soundCorpusFingerprintRow{
			NameID:            nameID,
			THName:            thName,
			PhoneticVersion:   version,
			EuphonyScore:      euphony,
			PronunciationEase: pronunciation,
			RhythmScore:       rhythm,
		}); err != nil {
			return domain.SoundCalibrationProfile{}, soundCorpusIdentity{}, fmt.Errorf("encode sound corpus fingerprint row %d: %w", nameID, err)
		}

		identity.VersionCounts[version]++
		identity.TotalCount++
		soundScore := float64(euphony)*0.45 + float64(pronunciation)*0.35 + float64(rhythm)*0.20
		acc, ok := accumulators[version]
		if !ok {
			acc = &soundStatsAccumulator{}
			accumulators[version] = acc
		}
		acc.Add(soundScore)
	}
	if err := rows.Err(); err != nil {
		return domain.SoundCalibrationProfile{}, soundCorpusIdentity{}, fmt.Errorf("failed while reading live sound calibration corpus: %w", err)
	}
	identity.FingerprintSHA256 = fmt.Sprintf("%x", digest.Sum(nil))

	orderedVersions := make([]string, 0, len(accumulators))
	for version := range accumulators {
		orderedVersions = append(orderedVersions, version)
	}
	sort.Slice(orderedVersions, func(i, j int) bool {
		left := accumulators[orderedVersions[i]]
		right := accumulators[orderedVersions[j]]
		if left.Count != right.Count {
			return left.Count > right.Count
		}
		return orderedVersions[i] < orderedVersions[j]
	})

	for _, version := range orderedVersions {
		acc := accumulators[version]
		stats := domain.SoundCalibrationStats{Count: acc.Count, Mean: acc.Mean()}
		if acc.Count > 1 {
			stats.StdDev = acc.StdDev()
		}
		profile.Versions[version] = stats
		if profile.ReferenceVersion == "" && stats.StdDev > 0 {
			profile.ReferenceVersion = version
			profile.ReferenceMean = stats.Mean
			profile.ReferenceStdDev = stats.StdDev
		}
	}
	return profile, identity, nil
}

type soundStatsAccumulator struct {
	Count int64
	mean  float64
	m2    float64
}

func (a *soundStatsAccumulator) Add(value float64) {
	a.Count++
	delta := value - a.mean
	a.mean += delta / float64(a.Count)
	a.m2 += delta * (value - a.mean)
}

func (a soundStatsAccumulator) Mean() float64 {
	return a.mean
}

func (a soundStatsAccumulator) StdDev() float64 {
	if a.Count <= 1 {
		return 0
	}
	return math.Sqrt(a.m2 / float64(a.Count-1))
}

func cloneSoundCalibrationProfile(profile domain.SoundCalibrationProfile) domain.SoundCalibrationProfile {
	cloned := profile
	cloned.Versions = make(map[string]domain.SoundCalibrationStats, len(profile.Versions))
	for version, stats := range profile.Versions {
		cloned.Versions[version] = stats
	}
	return cloned
}

type nameRepository struct {
	db                        *sql.DB
	soundCalibrationMu        sync.Mutex
	soundCalibration          *domain.SoundCalibrationProfile
	semanticProvenanceMu      sync.Mutex
	semanticProvenanceChecked bool
	semanticProvenance        domain.SemanticCorpusProvenance
}

func NewNameRepository(db *sql.DB) NameRepository {
	return &nameRepository{db: db}
}

// GetGoodNumbersMap returns every D-type dictionary row by exact key. Single
// digit rows such as "6" and leading-zero rows such as "06" are intentionally
// distinct and must never overwrite one another.
func (r *nameRepository) GetGoodNumbersMap(ctx context.Context) (map[string]domain.NumberMeaning, error) {
	all, err := r.GetAllNumbersMap(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]domain.NumberMeaning)
	for pair, meaning := range all {
		if meaning.IsGood() {
			result[pair] = meaning
		}
	}
	return result, nil
}

// GetAllNumbersMap returns the dictionary keyed exactly as authored. Single
// digit keys such as 6 and leading-zero keys such as 06 are intentionally
// distinct domain entries.
func (r *nameRepository) GetAllNumbersMap(ctx context.Context) (map[string]domain.NumberMeaning, error) {
	query := `
		SELECT pairnumber, pairtype, pairpoint,
			COALESCE(miracledesc, ''), COALESCE(miracledetail, '')
		FROM numbers
		ORDER BY BTRIM(pairnumber::text) ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all numbers dictionary: %w", err)
	}
	defer rows.Close()

	result := make(map[string]domain.NumberMeaning, 110)
	for rows.Next() {
		var n domain.NumberMeaning
		if err := rows.Scan(&n.PairNumber, &n.PairType, &n.PairPoint, &n.MiracleDesc, &n.MiracleDetail); err != nil {
			return nil, err
		}
		pair := strings.TrimSpace(n.PairNumber)
		if !isDictionaryPairKey(pair) {
			return nil, fmt.Errorf("invalid dictionary pair number %q", n.PairNumber)
		}
		if _, exists := result[pair]; exists {
			return nil, fmt.Errorf("duplicate dictionary pair number %q", pair)
		}
		n.PairNumber = pair
		result[pair] = n
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading number dictionary: %w", err)
	}
	return result, nil
}

func isDictionaryPairKey(pair string) bool {
	if len(pair) != 1 && len(pair) != 2 {
		return false
	}
	for _, char := range pair {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

type characterNumerology struct {
	satValue int
	shaValue int
	hasSat   bool
	hasSha   bool
}

// GetSurnameSums decodes a surname through the canonical character rule set.
// A row present in either dictionary is explicitly supported; an absent
// dimension contributes zero because the two systems intentionally cover
// different Thai characters. Only absence from both dictionaries is rejected.
func (r *nameRepository) GetSurnameSums(ctx context.Context, surname string) (int, int, error) {
	cleanSurname := strings.TrimSpace(surname)
	if cleanSurname == "" {
		return 0, 0, fmt.Errorf("%w: surname is blank", domain.ErrUnmappedCharacters)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT COALESCE(s.char_key, sh.char_key), s.sat_value, sh.sha_value
		FROM sat_nums s
		FULL OUTER JOIN sha_nums sh ON sh.char_key = s.char_key
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to load character numerology dictionaries: %w", err)
	}
	defer rows.Close()

	dictionary := make(map[rune]characterNumerology)
	for rows.Next() {
		var key string
		var satValue, shaValue sql.NullInt64
		if err := rows.Scan(&key, &satValue, &shaValue); err != nil {
			return 0, 0, fmt.Errorf("failed to scan character numerology dictionary: %w", err)
		}
		runes := []rune(key)
		if len(runes) != 1 {
			return 0, 0, fmt.Errorf("invalid character numerology key %q", key)
		}
		value := characterNumerology{}
		if satValue.Valid {
			value.satValue = int(satValue.Int64)
			value.hasSat = true
		}
		if shaValue.Valid {
			value.shaValue = int(shaValue.Int64)
			value.hasSha = true
		}
		dictionary[runes[0]] = value
	}
	if err := rows.Err(); err != nil {
		return 0, 0, fmt.Errorf("failed while reading character numerology dictionaries: %w", err)
	}

	return calculateTextSums(cleanSurname, dictionary)
}

func calculateTextSums(text string, dictionary map[rune]characterNumerology) (int, int, error) {
	var satSum, shaSum int
	var unknown []rune
	for _, char := range text {
		value, ok := dictionary[char]
		if !ok {
			unknown = append(unknown, char)
			continue
		}
		satSum += value.satValue
		shaSum += value.shaValue
	}
	if len(unknown) > 0 {
		return 0, 0, fmt.Errorf("%w: %q", domain.ErrUnmappedCharacters, string(unknown))
	}
	return satSum, shaSum, nil
}

// GetSumsByTHName retains the editorial meaning of a known row but always
// recalculates its sums from the canonical dictionaries. This prevents future
// callers from reintroducing historical stored-sum anomalies.
func (r *nameRepository) GetSumsByTHName(ctx context.Context, thname string) (int, int, string, error) {
	var meaning string
	err := r.db.QueryRowContext(ctx, "SELECT COALESCE(meaning, '') FROM names_miracle WHERE thname = $1 LIMIT 1", thname).Scan(&meaning)
	if err != nil {
		return 0, 0, "", err
	}
	satSum, shaSum, err := r.GetSurnameSums(ctx, thname)
	if err != nil {
		return 0, 0, "", err
	}
	return satSum, shaSum, meaning, nil
}

// SearchCandidateNames fetches candidates from Postgres names_miracle table
func (r *nameRepository) SearchCandidateNames(ctx context.Context, targetSatSums []int, targetShaSums []int, dayOfBirth string, wishQuery string, limit int) ([]domain.NameResultItem, error) {
	if len(targetSatSums) == 0 || len(targetShaSums) == 0 {
		return []domain.NameResultItem{}, nil
	}
	eligibilityClause := candidateEligibilityClause(dayOfBirth)
	semanticSource := "not_requested"
	if strings.TrimSpace(wishQuery) != "" {
		semanticSource = "unavailable"
	}
	limitClause := ""
	orderClause := ""
	args := []interface{}{pq.Array(targetSatSums), pq.Array(targetShaSums), semanticSource}
	if limit > 0 {
		orderClause = " ORDER BY name_id ASC"
		limitClause = " LIMIT $4"
		args = append(args, limit)
	}
	query := fmt.Sprintf(`
		SELECT name_id, thname, ''::text, sat_sum, sha_sum,
			COALESCE(pronunciation_ease, 0), COALESCE(euphony_score, 0),
			COALESCE(rhythm_score, 0), COALESCE(phonetic_version, ''),
			1.0 AS distance, 1.0 AS similarity_distance, false AS semantic_available, $3::text AS semantic_source,
			''::text AS semantic_query_model, ''::text AS semantic_vector_hash
		FROM names_miracle AS n
		WHERE sat_sum = ANY($1) AND sha_sum = ANY($2) %s%s%s
	`, eligibilityClause, orderClause, limitClause)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query candidate names: %w", err)
	}
	defer rows.Close()

	var results []domain.NameResultItem
	for rows.Next() {
		var item domain.NameResultItem
		if err := scanRankingCandidate(rows, &item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading candidate names: %w", err)
	}
	return results, nil
}

func kalakiniDayClause(dayOfBirth string) string {
	validDays := map[string]string{
		"sunday": "k_sunday", "monday": "k_monday", "tuesday": "k_tuesday",
		"wednesday1": "k_wednesday1", "wednesday2": "k_wednesday2",
		"thursday": "k_thursday", "friday": "k_friday", "saturday": "k_saturday",
	}
	if column, ok := validDays[strings.ToLower(strings.TrimSpace(dayOfBirth))]; ok {
		return fmt.Sprintf(" AND %s = false", column)
	}
	return ""
}

// canonicalStoredSumGuard quarantines the small known set of rows containing แ
// whose stored totals were produced by a historical doubled-character rule.
// The CASE keeps the expensive direct-dictionary check off ordinary rows.
const canonicalStoredSumGuard = `
	AND CASE
		WHEN STRPOS(n.thname, 'แ') = 0 THEN TRUE
		ELSE
			n.sat_sum = COALESCE((
				SELECT SUM(COALESCE(s.sat_value, 0))
				FROM regexp_split_to_table(n.thname, '') AS chars(char_key)
				LEFT JOIN sat_nums AS s ON s.char_key = chars.char_key
			), 0)
			AND n.sha_sum = COALESCE((
				SELECT SUM(COALESCE(sh.sha_value, 0))
				FROM regexp_split_to_table(n.thname, '') AS chars(char_key)
				LEFT JOIN sha_nums AS sh ON sh.char_key = chars.char_key
			), 0)
	END`

// These child honorifics are unambiguous administrative prefixes rather than
// parts of a given name. Broader strings such as นาย/นาง/พระ remain review-only
// because delimiter-free Thai names can legitimately begin with those runes.
const unambiguousTitlePrefixGuard = `
	AND n.thname NOT LIKE 'เด็กหญิง%'
	AND n.thname NOT LIKE 'เด็กชาย%'`

// Build the governed character set once per query, then delete those characters
// from each name. An empty remainder proves full dictionary coverage without
// splitting every row and performing millions of correlated index lookups.
const mappedCharacterGuard = `
	AND translate(
		n.thname,
		(
			SELECT string_agg(char_key, '' ORDER BY char_key)
			FROM (
				SELECT char_key FROM sat_nums
				UNION
				SELECT char_key FROM sha_nums
			) AS mapped_chars
		),
		''
	) = ''`

const lengthGuard = `
	AND char_length(n.thname) >= 2 AND char_length(n.thname) <= 8`

func candidateEligibilityClause(dayOfBirth string) string {
	return kalakiniDayClause(dayOfBirth) + lengthGuard + canonicalStoredSumGuard + unambiguousTitlePrefixGuard + mappedCharacterGuard
}

// SearchSemanticWishNames scores every eligible candidate against the requested
// embedding. It deliberately does not apply a retrieval LIMIT: Top-N is chosen
// only after all ranking dimensions are available in ScoringFilter.
//
// If embedding generation is unavailable, every eligible candidate is still
// returned with semantic_available=false. ScoringFilter then redistributes that
// dimension instead of manufacturing a perfect semantic score.
func (r *nameRepository) SearchSemanticWishNames(ctx context.Context, targetSatSums []int, targetShaSums []int, wishQuery string, dayOfBirth string, limit int) ([]domain.NameResultItem, error) {
	if len(targetSatSums) == 0 || len(targetShaSums) == 0 {
		return []domain.NameResultItem{}, nil
	}

	cleanWish := strings.TrimSpace(wishQuery)
	provenance, err := r.GetSemanticCorpusProvenance(ctx)
	if err != nil {
		return nil, err
	}
	if !provenance.Approved {
		results, err := r.SearchCandidateNames(ctx, targetSatSums, targetShaSums, dayOfBirth, cleanWish, limit)
		return markSemanticUnavailable(results, provenance.Source), err
	}

	vecStr, err := generateOpenAIVector(ctx, cleanWish)
	if err != nil || vecStr == "" {
		results, fallbackErr := r.SearchCandidateNames(ctx, targetSatSums, targetShaSums, dayOfBirth, cleanWish, limit)
		return markSemanticUnavailable(results, "unavailable_query_embedding_error"), fallbackErr
	}

	eligibilityClause := candidateEligibilityClause(dayOfBirth)
	limitClause := ""
	orderClause := ""
	vectorHash := fmt.Sprintf("%x", sha256.Sum256([]byte(vecStr)))
	args := []interface{}{
		pq.Array(targetSatSums),
		pq.Array(targetShaSums),
		vecStr,
		domain.SemanticEmbeddingModel,
		vectorHash,
	}
	semanticAvailableExpr := "meaning_vector IS NOT NULL"
	if provenance.SnapshotID > 0 {
		semanticAvailableExpr = "meaning_vector IS NOT NULL AND meaning_vector_snapshot_id = $6"
		args = append(args, provenance.SnapshotID)
	}
	if limit > 0 {
		orderClause = fmt.Sprintf(" ORDER BY CASE WHEN %s THEN meaning_vector <=> $3::vector ELSE 1.0 END ASC, name_id ASC", semanticAvailableExpr)
		limitClause = fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, limit)
	}
	query := fmt.Sprintf(`
		SELECT name_id, thname, ''::text, sat_sum, sha_sum,
			COALESCE(pronunciation_ease, 0), COALESCE(euphony_score, 0),
			COALESCE(rhythm_score, 0), COALESCE(phonetic_version, ''),
			CASE WHEN %s THEN meaning_vector <=> $3::vector ELSE 1.0 END AS distance,
			1.0 AS similarity_distance,
			(%s) AS semantic_available,
			CASE WHEN %s THEN 'embedding' ELSE 'embedding_missing_or_unapproved_snapshot' END AS semantic_source,
			$4::text AS semantic_query_model, $5::text AS semantic_vector_hash
		FROM names_miracle AS n
		WHERE sat_sum = ANY($1) AND sha_sum = ANY($2)
			AND (meaning_vector IS NULL OR vector_dims(meaning_vector) = %d) %s%s%s
	`, semanticAvailableExpr, semanticAvailableExpr, semanticAvailableExpr,
		domain.SemanticVectorDimensions, eligibilityClause, orderClause, limitClause)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		results, fallbackErr := r.SearchCandidateNames(ctx, targetSatSums, targetShaSums, dayOfBirth, cleanWish, limit)
		if fallbackErr != nil {
			return nil, fmt.Errorf("failed to query complete semantic candidate set: %w", err)
		}
		return markSemanticUnavailable(results, "unavailable_semantic_vector_query_error"), nil
	}
	defer rows.Close()

	results := make([]domain.NameResultItem, 0)
	for rows.Next() {
		var item domain.NameResultItem
		if err := scanRankingCandidate(rows, &item); err != nil {
			return nil, err
		}
		item.CandidateVectorProvenanceVerified = true
		item.CandidateVectorProvenanceSource = provenance.Source
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while reading semantic candidate set: %w", err)
	}
	return results, nil
}

func markSemanticUnavailable(results []domain.NameResultItem, source string) []domain.NameResultItem {
	if source == "" {
		source = "unavailable"
	}
	for i := range results {
		results[i].SemanticAvailable = false
		results[i].SemanticSource = source
		results[i].SemanticQueryModel = ""
		results[i].SemanticVectorHash = ""
		results[i].CandidateVectorProvenanceVerified = false
		results[i].CandidateVectorProvenanceSource = source
	}
	return results
}

// generateOpenAIVector calls OpenAI text-embedding-3-small API with In-Memory Caching (0ms latency for repeated/preset queries)
func generateOpenAIVector(ctx context.Context, text string) (string, error) {
	cleanText := strings.TrimSpace(text)
	cacheKey := domain.SemanticEmbeddingModel + "\x00" + domain.SemanticInputTemplateVersion + "\x00" + cleanText

	// Check In-Memory Cache first
	vectorCacheMu.RLock()
	if cachedVec, found := vectorCache[cacheKey]; found {
		vectorCacheMu.RUnlock()
		return cachedVec, nil
	}
	vectorCacheMu.RUnlock()

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY is not set")
	}

	reqBody, _ := json.Marshal(map[string]interface{}{
		"input": semanticEmbeddingInput(cleanText),
		"model": domain.SemanticEmbeddingModel,
	})

	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai api error status: %d", resp.StatusCode)
	}

	var res struct {
		Model string `json:"model"`
		Data  []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil || len(res.Data) == 0 {
		return "", fmt.Errorf("failed to decode openai response")
	}
	if res.Model != "" && res.Model != domain.SemanticEmbeddingModel {
		return "", fmt.Errorf("unexpected embedding response model %q", res.Model)
	}
	if len(res.Data[0].Embedding) != domain.SemanticVectorDimensions {
		return "", fmt.Errorf(
			"unexpected embedding dimensions: got %d want %d",
			len(res.Data[0].Embedding),
			domain.SemanticVectorDimensions,
		)
	}

	vecStrBytes, _ := json.Marshal(res.Data[0].Embedding)
	vecResult := string(vecStrBytes)

	// Save to In-Memory Cache
	vectorCacheMu.Lock()
	if len(vectorCache) >= maxVectorCacheEntries {
		// Preset wishes are requested repeatedly. Bounding this cache prevents an
		// unbounded stream of custom wishes from growing the API process forever.
		vectorCache = make(map[string]string)
	}
	vectorCache[cacheKey] = vecResult
	vectorCacheMu.Unlock()

	return vecResult, nil
}

func semanticEmbeddingInput(cleanText string) string {
	return "ความหมาย: " + strings.TrimSpace(cleanText)
}

// SearchSimilarNames prepares the eligible similarity pool. It compares both
// raw Thai characters AND
// phonetically-normalized forms (การันต์ stripped, consonant classes unified)
// and takes the higher similarity score so that names sounding alike but spelled
// differently (e.g. กานต์ vs กานดา, ณัฐ vs นัด) also rank well.
func (r *nameRepository) SearchSimilarNames(ctx context.Context, targetSatSums []int, targetShaSums []int, seedName string, dayOfBirth string, limit int) ([]domain.NameResultItem, error) {
	if len(targetSatSums) == 0 || len(targetShaSums) == 0 {
		return []domain.NameResultItem{}, nil
	}
	eligibilityClause := candidateEligibilityClause(dayOfBirth)
	trimmedSeed := strings.TrimSpace(seedName)
	normalizedSeed := normalizeThaiForSimilarity(trimmedSeed)
	limitClause := ""
	orderClause := ""
	args := []interface{}{trimmedSeed, pq.Array(targetSatSums), pq.Array(targetShaSums), normalizedSeed}
	if limit > 0 {
		orderClause = ` ORDER BY GREATEST(
			similarity(thname, $1),
			similarity(normalize_thai_phonetic(thname), $4)
		) DESC, name_id ASC`
		limitClause = " LIMIT $5"
		args = append(args, limit)
	}
	// GREATEST picks the better of raw-character similarity and phonetic
	// similarity so that spelling variants don't penalize the score.
	query := fmt.Sprintf(`
		SELECT name_id, thname, ''::text, sat_sum, sha_sum,
			COALESCE(pronunciation_ease, 0), COALESCE(euphony_score, 0),
			COALESCE(rhythm_score, 0), COALESCE(phonetic_version, ''),
			1.0::double precision AS distance,
			(1.0 - GREATEST(
				similarity(thname, $1),
				similarity(normalize_thai_phonetic(thname), $4)
			)) AS similarity_distance,
			false AS semantic_available, 'semantic_unavailable_trigram_only' AS semantic_source,
			''::text AS semantic_query_model, ''::text AS semantic_vector_hash
		FROM names_miracle AS n
		WHERE sat_sum = ANY($2) AND sha_sum = ANY($3)
			AND (
				thname %% $1
				OR normalize_thai_phonetic(thname) %% $4
			) %s%s%s
	`, eligibilityClause, orderClause, limitClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.NameResultItem
	for rows.Next() {
		var item domain.NameResultItem
		if err := scanRankingCandidate(rows, &item); err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// SuggestNames returns raw text/trigram matches without numerology or shadow energy filtering,
// allowing users to see close names with their actual number pairs (both good and bad).
func (r *nameRepository) SuggestNames(ctx context.Context, seedName string, limit int) ([]domain.NameResultItem, error) {
	trimmedSeed := strings.TrimSpace(seedName)
	if trimmedSeed == "" {
		return []domain.NameResultItem{}, nil
	}
	normalizedSeed := normalizeThaiForSimilarity(trimmedSeed)
	if limit <= 0 {
		limit = 8
	}
	query := `
		SELECT name_id, thname, COALESCE(meaning, ''), sat_sum, sha_sum
		FROM names_miracle
		WHERE thname NOT LIKE 'เด็กหญิง%'
			AND thname NOT LIKE 'เด็กหณิง%'
			AND thname NOT LIKE 'เด็กชาย%'
			AND thname NOT LIKE 'นางสาว%'
			AND thname NOT LIKE 'คุณหญิง%'
			AND (
			thname = $1
			OR thname LIKE $1 || '%'
			OR thname LIKE '%' || $1 || '%'
			OR similarity(thname, $1) >= 0.20
			OR similarity(normalize_thai_phonetic(thname), $2) >= 0.20
		)
		ORDER BY
			CASE
				WHEN thname = $1 THEN 0
				WHEN thname LIKE $1 || '%' THEN 1
				WHEN thname LIKE '%' || $1 || '%' THEN 2
				ELSE 3
			END,
			GREATEST(
				similarity(thname, $1),
				similarity(normalize_thai_phonetic(thname), $2)
			) DESC,
			name_id ASC
		LIMIT $3
	`
	rows, err := r.db.QueryContext(ctx, query, trimmedSeed, normalizedSeed, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.NameResultItem
	for rows.Next() {
		var item domain.NameResultItem
		if err := rows.Scan(&item.NameID, &item.THName, &item.Meaning, &item.SatSum, &item.ShaSum); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

func (r *nameRepository) GetMeaningVectorByTHName(ctx context.Context, thname string) (string, string, error) {
	query := `SELECT meaning_vector::text, COALESCE(meaning, '') FROM names_miracle WHERE thname = $1 LIMIT 1`
	var vectorStr sql.NullString
	var meaning string
	err := r.db.QueryRowContext(ctx, query, thname).Scan(&vectorStr, &meaning)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", "", nil
		}
		return "", "", err
	}
	return vectorStr.String, meaning, nil
}

func (r *nameRepository) SearchSemanticAndSimilarNames(ctx context.Context, targetSatSums []int, targetShaSums []int, seedName string, meaningQuery string, dayOfBirth string, limit int) ([]domain.NameResultItem, error) {
	if len(targetSatSums) == 0 || len(targetShaSums) == 0 {
		return []domain.NameResultItem{}, nil
	}

	provenance, err := r.GetSemanticCorpusProvenance(ctx)
	if err != nil {
		return nil, err
	}
	if !provenance.Approved {
		results, fallbackErr := r.SearchSimilarNames(ctx, targetSatSums, targetShaSums, seedName, dayOfBirth, limit)
		return markSemanticUnavailable(results, provenance.Source), fallbackErr
	}
	semanticVector, err := generateOpenAIVector(ctx, strings.TrimSpace(meaningQuery))
	if err != nil || semanticVector == "" {
		results, fallbackErr := r.SearchSimilarNames(ctx, targetSatSums, targetShaSums, seedName, dayOfBirth, limit)
		return markSemanticUnavailable(results, "unavailable_query_embedding_error"), fallbackErr
	}

	eligibilityClause := candidateEligibilityClause(dayOfBirth)
	trimmedSeed := strings.TrimSpace(seedName)
	normalizedSeed := normalizeThaiForSimilarity(trimmedSeed)
	vectorHash := fmt.Sprintf("%x", sha256.Sum256([]byte(semanticVector)))

	limitClause := ""
	orderClause := ""
	args := []interface{}{
		pq.Array(targetSatSums),
		pq.Array(targetShaSums),
		semanticVector,
		trimmedSeed,
		normalizedSeed,
		domain.SemanticEmbeddingModel,
		vectorHash,
	}
	semanticAvailableExpr := "meaning_vector IS NOT NULL"
	if provenance.SnapshotID > 0 {
		semanticAvailableExpr = "meaning_vector IS NOT NULL AND meaning_vector_snapshot_id = $8"
		args = append(args, provenance.SnapshotID)
	}
	if limit > 0 {
		orderClause = fmt.Sprintf(` ORDER BY CASE WHEN %s THEN meaning_vector <=> $3::vector ELSE 1.0 END ASC,
			GREATEST(similarity(thname, $4), similarity(normalize_thai_phonetic(thname), $5)) DESC,
			name_id ASC`, semanticAvailableExpr)
		limitClause = fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, limit)
	}

	query := fmt.Sprintf(`
		SELECT name_id, thname, ''::text, sat_sum, sha_sum,
			COALESCE(pronunciation_ease, 0), COALESCE(euphony_score, 0),
			COALESCE(rhythm_score, 0), COALESCE(phonetic_version, ''),
			CASE WHEN %s THEN meaning_vector <=> $3::vector ELSE 1.0 END AS distance,
			(1.0 - GREATEST(
				similarity(thname, $4),
				similarity(normalize_thai_phonetic(thname), $5)
			)) AS similarity_distance,
			(%s) AS semantic_available,
			CASE WHEN %s THEN 'embedding' ELSE 'embedding_missing_or_unapproved_snapshot' END AS semantic_source,
			$6::text AS semantic_query_model, $7::text AS semantic_vector_hash
		FROM names_miracle AS n
		WHERE sat_sum = ANY($1) AND sha_sum = ANY($2)
			AND (
				thname %% $4
				OR normalize_thai_phonetic(thname) %% $5
			) %s
			AND (meaning_vector IS NULL OR vector_dims(meaning_vector) = %d)%s%s
	`, semanticAvailableExpr, semanticAvailableExpr, semanticAvailableExpr,
		eligibilityClause, domain.SemanticVectorDimensions, orderClause, limitClause)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []domain.NameResultItem
	for rows.Next() {
		var item domain.NameResultItem
		if err := scanRankingCandidate(rows, &item); err != nil {
			return nil, err
		}
		item.CandidateVectorProvenanceVerified = true
		item.CandidateVectorProvenanceSource = provenance.Source
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// thaiKaranRegex matches a Thai consonant followed by the การันต์ mark (์).
var thaiKaranRegex = regexp.MustCompile(`[ก-ฮ]์`)

// normalizeThaiForSimilarity produces a phonetically-normalized version of a
// Thai name by stripping silent markers (การันต์) and unifying consonant
// classes so that phonetically equivalent names share more trigrams.
//
// Examples:
//
//	กานต์  → กาน
//	ณัฐกร  → นัทกร
//	ศิริษา → สิริสา
func normalizeThaiForSimilarity(name string) string {
	// 1. Strip การันต์: remove consonant + ์ pairs.
	result := thaiKaranRegex.ReplaceAllString(name, "")
	// 2. Unify consonant classes for phonetic equivalence.
	result = strings.NewReplacer(
		"ณ", "น",
		"ศ", "ส",
		"ษ", "ส",
		"ฑ", "ท",
		"ฐ", "ท",
		"ภ", "พ",
		"ธ", "ท",
	).Replace(result)
	return result
}

func scanRankingCandidate(rows *sql.Rows, item *domain.NameResultItem) error {
	return rows.Scan(
		&item.NameID,
		&item.THName,
		&item.Meaning,
		&item.SatSum,
		&item.ShaSum,
		&item.PronunciationEase,
		&item.EuphonyScore,
		&item.RhythmScore,
		&item.PhoneticVersion,
		&item.VectorDistance,
		&item.SimilarityDistance,
		&item.SemanticAvailable,
		&item.SemanticSource,
		&item.SemanticQueryModel,
		&item.SemanticVectorHash,
	)
}

func (r *nameRepository) GetSampleNames(ctx context.Context) ([]domain.Celebrity, error) {
	query := `SELECT name, avatar_url FROM naming_examples WHERE is_celebrity = true ORDER BY sort_order ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []domain.Celebrity
	for rows.Next() {
		var name, avatarUrl string
		if err := rows.Scan(&name, &avatarUrl); err != nil {
			return nil, err
		}

		// Extract filename from avatarUrl (e.g. /uploads/avatars/avatar_123.jpg -> avatar_123.jpg)
		parts := strings.Split(avatarUrl, "/")
		filename := parts[len(parts)-1]

		names = append(names, domain.Celebrity{
			Name:      name,
			AvatarUrl: filename,
		})
	}
	return names, nil
}
