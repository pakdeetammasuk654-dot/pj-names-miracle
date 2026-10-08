package repository

import (
	"context"
	"database/sql"
	"fmt"

	"go-backend/domain"
)

type auspiciousContentRepository struct{ db *sql.DB }

func NewAuspiciousContentRepository(db *sql.DB) domain.AuspiciousContentRepository {
	return &auspiciousContentRepository{db: db}
}

func (r *auspiciousContentRepository) GetTextMap(ctx context.Context) (map[string]string, error) {
	items, err := r.ListTexts(ctx)
	if err != nil {
		return nil, err
	}
	values := make(map[string]string, len(items))
	for _, item := range items {
		values[item.Key] = item.ValueTH
	}
	return values, nil
}

func (r *auspiciousContentRepository) ListTexts(ctx context.Context) ([]domain.AuspiciousEditableText, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT key, group_code, label_th, value_th, help_th, sort_order, updated_at
		FROM auspicious_texts ORDER BY group_code, sort_order, key`)
	if err != nil {
		return nil, fmt.Errorf("query auspicious texts: %w", err)
	}
	defer rows.Close()
	items := make([]domain.AuspiciousEditableText, 0)
	for rows.Next() {
		var item domain.AuspiciousEditableText
		if err := rows.Scan(&item.Key, &item.GroupCode, &item.LabelTH, &item.ValueTH, &item.HelpTH, &item.SortOrder, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan auspicious text: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *auspiciousContentRepository) ListAdminRuleSets(ctx context.Context) ([]domain.AuspiciousAdminRuleSet, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, code, name_th, updated_at FROM auspicious_rule_sets ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("query auspicious rule sets: %w", err)
	}
	defer rows.Close()
	items := make([]domain.AuspiciousAdminRuleSet, 0)
	for rows.Next() {
		var item domain.AuspiciousAdminRuleSet
		if err := rows.Scan(&item.ID, &item.Code, &item.NameTH, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan auspicious rule set: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *auspiciousContentRepository) ListAdminActivities(ctx context.Context) ([]domain.AuspiciousAdminActivity, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT code, name_th, icon, description_th,
		       COALESCE(rule_scope_note_th, ''), COALESCE(safety_note_th, ''), enabled
		FROM auspicious_activities ORDER BY enabled DESC, sort_order, code`)
	if err != nil {
		return nil, fmt.Errorf("query auspicious admin activities: %w", err)
	}
	defer rows.Close()
	items := make([]domain.AuspiciousAdminActivity, 0)
	for rows.Next() {
		var item domain.AuspiciousAdminActivity
		if err := rows.Scan(&item.Code, &item.NameTH, &item.Icon, &item.DescriptionTH, &item.ScopeNoteTH, &item.SafetyNoteTH, &item.Enabled); err != nil {
			return nil, fmt.Errorf("scan auspicious admin activity: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *auspiciousContentRepository) ListAdminRules(ctx context.Context) ([]domain.AuspiciousAdminRule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.id, r.code, rs.name_th, r.name_th, r.description_th,
		       COALESCE(NULLIF(r.meaning_th, ''), r.description_th), r.effect, r.score, r.enabled, r.updated_at
		FROM auspicious_rules r JOIN auspicious_rule_sets rs ON rs.id = r.rule_set_id
		ORDER BY r.enabled DESC, rs.id, r.priority DESC, r.id`)
	if err != nil {
		return nil, fmt.Errorf("query auspicious admin rules: %w", err)
	}
	defer rows.Close()
	items := make([]domain.AuspiciousAdminRule, 0)
	for rows.Next() {
		var item domain.AuspiciousAdminRule
		if err := rows.Scan(&item.ID, &item.Code, &item.RuleSetNameTH, &item.NameTH, &item.DescriptionTH, &item.MeaningTH, &item.Effect, &item.Score, &item.Enabled, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan auspicious admin rule: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *auspiciousContentRepository) UpdateText(ctx context.Context, key, value string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE auspicious_texts SET value_th=$1, updated_at=NOW() WHERE key=$2`, value, key)
	return requireUpdated(result, err, "auspicious text")
}

func (r *auspiciousContentRepository) UpdateRuleSetText(ctx context.Context, id int64, name string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE auspicious_rule_sets SET name_th=$1, updated_at=NOW() WHERE id=$2`, name, id)
	return requireUpdated(result, err, "auspicious rule set")
}

func (r *auspiciousContentRepository) UpdateActivityText(ctx context.Context, code, name, icon, description, scopeNote, safetyNote string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE auspicious_activities SET name_th=$1, icon=$2, description_th=$3,
		       rule_scope_note_th=$4, safety_note_th=$5 WHERE code=$6`, name, icon, description, scopeNote, safetyNote, code)
	return requireUpdated(result, err, "auspicious activity")
}

func (r *auspiciousContentRepository) UpdateRuleText(ctx context.Context, id int64, name, description, meaning string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE auspicious_rules SET name_th=$1, description_th=$2, meaning_th=$3, updated_at=NOW() WHERE id=$4`,
		name, description, meaning, id)
	return requireUpdated(result, err, "auspicious rule")
}

func requireUpdated(result sql.Result, err error, entity string) error {
	if err != nil {
		return fmt.Errorf("update %s: %w", entity, err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read %s update result: %w", entity, err)
	}
	if rows != 1 {
		return fmt.Errorf("%s not found", entity)
	}
	return nil
}
