package domain

import "time"

// Article is an editorial post displayed in the public knowledge centre.
// ContentHTML is sanitized before it is rendered in a browser.
type Article struct {
	ID          int64
	Slug        string
	Title       string
	Excerpt     string
	Category    string
	ImageURL    string
	PublishedAt time.Time
	UpdatedAt   time.Time
	IsPublished bool
	ContentHTML string
	TitleShort  string
	PinOrder    int
}

type ArticleDashboardStats struct {
	Total     int
	Published int
	Drafts    int
}
