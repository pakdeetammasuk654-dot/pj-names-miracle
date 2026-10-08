package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
	"go-backend/domain"
)

type auspiciousRuleRepository struct {
	db *sql.DB
}

func NewAuspiciousRuleRepository(db *sql.DB) domain.AuspiciousRuleRepository {
	return &auspiciousRuleRepository{db: db}
}

func (r *auspiciousRuleRepository) ListActivities(ctx context.Context) ([]domain.AuspiciousActivity, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.code, a.name_th, a.icon, a.description_th,
		       COALESCE(a.rule_scope_note_th, ''), COALESCE(a.safety_note_th, ''), a.sort_order,
		       COALESCE(c.code, 'other'), COALESCE(c.name_th, 'กิจกรรมอื่น'),
		       COALESCE(c.icon, '✦'), COALESCE(c.sort_order, 999)
		FROM auspicious_activities a
		LEFT JOIN auspicious_activity_categories c ON c.code = a.category_code AND c.enabled = TRUE
		WHERE a.enabled = TRUE
		ORDER BY COALESCE(c.sort_order, 999), a.sort_order, a.code
	`)
	if err != nil {
		return nil, fmt.Errorf("query auspicious activities: %w", err)
	}
	defer rows.Close()

	activities := make([]domain.AuspiciousActivity, 0)
	for rows.Next() {
		var item domain.AuspiciousActivity
		if err := rows.Scan(
			&item.Code, &item.NameTH, &item.Icon, &item.Description,
			&item.RuleScopeNote, &item.SafetyNote, &item.SortOrder,
			&item.CategoryCode, &item.CategoryNameTH, &item.CategoryIcon, &item.CategorySortOrder,
		); err != nil {
			return nil, fmt.Errorf("scan auspicious activity: %w", err)
		}
		activities = append(activities, item)
	}
	return activities, rows.Err()
}

func (r *auspiciousRuleRepository) ListRules(ctx context.Context, activityCode string) ([]domain.AuspiciousRule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, rs.code, rs.name_th, rs.source_url, r.code, r.signal_group,
		       r.name_th, r.description_th, COALESCE(NULLIF(r.meaning_th, ''), r.description_th),
		       r.effect, r.score, r.priority,
		       r.condition_json, COALESCE(r.source_note, ''),
		       COALESCE(array_agg(ra.activity_code ORDER BY ra.activity_code)
		         FILTER (WHERE ra.activity_code IS NOT NULL), ARRAY[]::text[])
		FROM auspicious_rules r
		JOIN auspicious_rule_sets rs ON rs.id = r.rule_set_id
		LEFT JOIN auspicious_rule_activities ra ON ra.rule_id = r.id
		WHERE r.enabled = TRUE AND rs.enabled = TRUE
		  AND (
		    NOT EXISTS (SELECT 1 FROM auspicious_rule_activities scope WHERE scope.rule_id = r.id)
		    OR EXISTS (SELECT 1 FROM auspicious_rule_activities scope WHERE scope.rule_id = r.id AND scope.activity_code = $1)
		  )
		GROUP BY r.id, rs.id
		ORDER BY r.priority DESC, r.id
	`, activityCode)
	if err != nil {
		return nil, fmt.Errorf("query auspicious rules: %w", err)
	}
	defer rows.Close()

	rules := make([]domain.AuspiciousRule, 0)
	for rows.Next() {
		var rule domain.AuspiciousRule
		var conditionJSON []byte
		if err := rows.Scan(
			&rule.ID, &rule.RuleSetCode, &rule.RuleSetName, &rule.SourceURL,
			&rule.Code, &rule.SignalGroup, &rule.NameTH, &rule.DescriptionTH, &rule.MeaningTH,
			&rule.Effect, &rule.Score, &rule.Priority, &conditionJSON,
			&rule.SourceNote, pq.Array(&rule.Activities),
		); err != nil {
			return nil, fmt.Errorf("scan auspicious rule: %w", err)
		}
		if err := json.Unmarshal(conditionJSON, &rule.Condition); err != nil {
			return nil, fmt.Errorf("decode condition for rule %s: %w", rule.Code, err)
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}
