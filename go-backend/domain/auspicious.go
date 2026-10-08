package domain

import (
	"context"
	"encoding/json"
	"time"
)

type AuspiciousActivity struct {
	Code              string `json:"code"`
	NameTH            string `json:"name_th"`
	Icon              string `json:"icon"`
	Description       string `json:"description"`
	RuleScopeNote     string `json:"rule_scope_note"`
	SafetyNote        string `json:"safety_note,omitempty"`
	SortOrder         int    `json:"sort_order"`
	CategoryCode      string `json:"category_code"`
	CategoryNameTH    string `json:"category_name_th"`
	CategoryIcon      string `json:"category_icon"`
	CategorySortOrder int    `json:"category_sort_order"`
}

// RuleExpression is a small, versioned expression tree stored as JSONB. Leaf
// fields and operators are allow-listed by the evaluator; arbitrary SQL or Go
// expressions are never executed from database content.
type RuleExpression struct {
	All      []RuleExpression `json:"all,omitempty"`
	Any      []RuleExpression `json:"any,omitempty"`
	Not      *RuleExpression  `json:"not,omitempty"`
	Field    string           `json:"field,omitempty"`
	Operator string           `json:"operator,omitempty"`
	MapField string           `json:"map_field,omitempty"`
	Value    json.RawMessage  `json:"value,omitempty"`
}

type AuspiciousRule struct {
	ID            int64          `json:"id"`
	RuleSetCode   string         `json:"rule_set_code"`
	RuleSetName   string         `json:"rule_set_name"`
	SourceURL     string         `json:"source_url"`
	Code          string         `json:"code"`
	SignalGroup   string         `json:"signal_group"`
	NameTH        string         `json:"name_th"`
	DescriptionTH string         `json:"description_th"`
	MeaningTH     string         `json:"meaning_th"`
	Effect        string         `json:"effect"` // exclude or score
	Score         int            `json:"score"`
	Priority      int            `json:"priority"`
	Condition     RuleExpression `json:"condition"`
	Activities    []string       `json:"activities"`
	SourceNote    string         `json:"source_note"`
}

type AuspiciousRuleRepository interface {
	ListActivities(ctx context.Context) ([]AuspiciousActivity, error)
	ListRules(ctx context.Context, activityCode string) ([]AuspiciousRule, error)
}

type AuspiciousEditableText struct {
	Key       string    `json:"key"`
	GroupCode string    `json:"group_code"`
	LabelTH   string    `json:"label_th"`
	ValueTH   string    `json:"value_th"`
	HelpTH    string    `json:"help_th"`
	SortOrder int       `json:"sort_order"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuspiciousAdminRuleSet struct {
	ID        int64
	Code      string
	NameTH    string
	UpdatedAt time.Time
}

type AuspiciousAdminActivity struct {
	Code          string
	NameTH        string
	Icon          string
	DescriptionTH string
	ScopeNoteTH   string
	SafetyNoteTH  string
	Enabled       bool
}

type AuspiciousAdminRule struct {
	ID            int64
	Code          string
	RuleSetNameTH string
	NameTH        string
	DescriptionTH string
	MeaningTH     string
	Effect        string
	Score         int
	Enabled       bool
	UpdatedAt     time.Time
}

type AuspiciousContentRepository interface {
	GetTextMap(ctx context.Context) (map[string]string, error)
	ListTexts(ctx context.Context) ([]AuspiciousEditableText, error)
	ListAdminRuleSets(ctx context.Context) ([]AuspiciousAdminRuleSet, error)
	ListAdminActivities(ctx context.Context) ([]AuspiciousAdminActivity, error)
	ListAdminRules(ctx context.Context) ([]AuspiciousAdminRule, error)
	UpdateText(ctx context.Context, key, value string) error
	UpdateRuleSetText(ctx context.Context, id int64, name string) error
	UpdateActivityText(ctx context.Context, code, name, icon, description, scopeNote, safetyNote string) error
	UpdateRuleText(ctx context.Context, id int64, name, description, meaning string) error
}

type AuspiciousSearchRequest struct {
	ActivityCode  string `json:"activity_code"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	BirthDate     string `json:"birth_date,omitempty"`
	BirthTime     string `json:"birth_time,omitempty"` // เวลาเกิด HH:MM
	DayOfBirth    string `json:"day_of_birth,omitempty"`
	Age           int    `json:"age,omitempty"`
	Gender        string `json:"gender,omitempty"`
	PreferWeekend bool   `json:"prefer_weekend"`
	AvoidWanPhra  bool   `json:"avoid_wan_phra"`
	MonkCeremony  bool   `json:"monk_ceremony"`
}

type AuspiciousReason struct {
	RuleCode    string `json:"rule_code"`
	Name        string `json:"name"`
	Explanation string `json:"explanation"`
	Meaning     string `json:"meaning"`
	Impact      string `json:"impact"`
	Score       int    `json:"score"`
}

// OutfitColor is a display-ready colour from the legacy Nab Ha Sri palette.
// Calculation remains in the usecase layer; clients only render this API data.
type OutfitColor struct {
	NameTH string `json:"name_th"`
	Hex    string `json:"hex"`
}

type OutfitColorSet struct {
	Context       string        `json:"context"`
	LabelTH       string        `json:"label_th"`
	RoleTH        string        `json:"role_th"`
	RoleMeaningTH string        `json:"role_meaning_th"`
	PlanetNameTH  string        `json:"planet_name_th"`
	Number        int           `json:"number"`
	Colors        []OutfitColor `json:"colors"`
}

type OutfitColorAdvice struct {
	Auspicious   []OutfitColorSet `json:"auspicious"`
	Inauspicious []OutfitColorSet `json:"inauspicious"`
}

type AuspiciousDateResult struct {
	Date         string             `json:"date"`
	ThaiDate     string             `json:"thai_date"`
	Weekday      string             `json:"weekday"`
	LunarLabel   string             `json:"lunar_label"`
	LunarPhase   string             `json:"lunar_phase"`
	LunarDay     int                `json:"lunar_day"`
	LunarMonth   int                `json:"lunar_month"`
	IsWanPhra    bool               `json:"is_wan_phra"`
	ImportantDay string             `json:"important_day,omitempty"`
	IsWeekend    bool               `json:"is_weekend"`
	Score        int                `json:"score"`
	Rating       string             `json:"rating"`
	Reasons      []AuspiciousReason `json:"reasons"`
	OutfitAdvice *OutfitColorAdvice `json:"outfit_advice,omitempty"`
}

type AuspiciousSearchResponse struct {
	Activity        AuspiciousActivity     `json:"activity"`
	StartDate       string                 `json:"start_date"`
	EndDate         string                 `json:"end_date"`
	EvaluatedDays   int                    `json:"evaluated_days"`
	ExcludedDays    int                    `json:"excluded_days"`
	Recommendations []AuspiciousDateResult `json:"recommendations"`
	RuleSets        []string               `json:"rule_sets"`
	Disclaimer      string                 `json:"disclaimer"`
	GeneratedAt     time.Time              `json:"generated_at"`
}
