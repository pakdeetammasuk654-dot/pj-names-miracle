package repository

import (
	"context"
	"database/sql"
	"fmt"

	"go-backend/domain"
)

type ArticleRepository interface {
	ListPublished(ctx context.Context) ([]domain.Article, error)
	GetPublishedBySlug(ctx context.Context, slug string) (*domain.Article, error)
	ListAll(ctx context.Context) ([]domain.Article, error)
	GetByID(ctx context.Context, id int64) (*domain.Article, error)
	Create(ctx context.Context, article *domain.Article) error
	Update(ctx context.Context, article *domain.Article) error
	Delete(ctx context.Context, id int64) error
	IsImageReferenced(ctx context.Context, imageURL string) (bool, error)
	GetDashboardStats(ctx context.Context) (domain.ArticleDashboardStats, error)
}

type articleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) ArticleRepository {
	return &articleRepository{db: db}
}

const articleFields = `
	art_id, slug, title, excerpt, category, image_url, published_at,
	updated_at, is_published, content, title_short, COALESCE(pin_order, 0)`

func (r *articleRepository) ListPublished(ctx context.Context) ([]domain.Article, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+articleFields+`
		FROM articles
		WHERE is_published = true
		ORDER BY CASE WHEN COALESCE(pin_order, 0) > 0 THEN pin_order ELSE 999 END, published_at DESC, art_id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list published articles: %w", err)
	}
	defer rows.Close()

	articles := make([]domain.Article, 0)
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate published articles: %w", err)
	}
	return articles, nil
}

func (r *articleRepository) GetPublishedBySlug(ctx context.Context, slug string) (*domain.Article, error) {
	article, err := scanArticle(r.db.QueryRowContext(ctx, `
		SELECT `+articleFields+`
		FROM articles
		WHERE slug = $1 AND is_published = true`, slug))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) ListAll(ctx context.Context) ([]domain.Article, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+articleFields+`
		FROM articles
		ORDER BY CASE WHEN COALESCE(pin_order, 0) > 0 THEN pin_order ELSE 999 END, published_at DESC, art_id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()
	articles := make([]domain.Article, 0)
	for rows.Next() {
		article, err := scanArticle(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	return articles, rows.Err()
}

func (r *articleRepository) GetByID(ctx context.Context, id int64) (*domain.Article, error) {
	article, err := scanArticle(r.db.QueryRowContext(ctx, `SELECT `+articleFields+` FROM articles WHERE art_id = $1`, id))
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (r *articleRepository) Create(ctx context.Context, article *domain.Article) error {
	return r.db.QueryRowContext(ctx, `
		INSERT INTO articles (slug, title, excerpt, category, image_url, published_at, is_published, content, title_short, pin_order)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING art_id`,
		article.Slug, article.Title, article.Excerpt, article.Category, article.ImageURL, article.PublishedAt,
		article.IsPublished, article.ContentHTML, article.TitleShort, article.PinOrder,
	).Scan(&article.ID)
}

func (r *articleRepository) Update(ctx context.Context, article *domain.Article) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE articles SET slug=$1, title=$2, excerpt=$3, category=$4, image_url=$5, published_at=$6,
		updated_at=NOW(), is_published=$7, content=$8, title_short=$9, pin_order=$10 WHERE art_id=$11`,
		article.Slug, article.Title, article.Excerpt, article.Category, article.ImageURL, article.PublishedAt,
		article.IsPublished, article.ContentHTML, article.TitleShort, article.PinOrder, article.ID,
	)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *articleRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM articles WHERE art_id = $1`, id)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil {
		return err
	} else if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *articleRepository) IsImageReferenced(ctx context.Context, imageURL string) (bool, error) {
	var referenced bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM articles
			WHERE image_url = $1 OR position($1 in content) > 0
		)`, imageURL).Scan(&referenced); err != nil {
		return false, fmt.Errorf("check image references: %w", err)
	}
	return referenced, nil
}

func (r *articleRepository) GetDashboardStats(ctx context.Context) (domain.ArticleDashboardStats, error) {
	var stats domain.ArticleDashboardStats
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE is_published), COUNT(*) FILTER (WHERE NOT is_published)
		FROM articles`).Scan(&stats.Total, &stats.Published, &stats.Drafts); err != nil {
		return domain.ArticleDashboardStats{}, fmt.Errorf("get article dashboard stats: %w", err)
	}
	return stats, nil
}

type articleScanner interface {
	Scan(dest ...any) error
}

func scanArticle(scanner articleScanner) (domain.Article, error) {
	var article domain.Article
	if err := scanner.Scan(
		&article.ID, &article.Slug, &article.Title, &article.Excerpt,
		&article.Category, &article.ImageURL, &article.PublishedAt, &article.UpdatedAt,
		&article.IsPublished, &article.ContentHTML, &article.TitleShort,
		&article.PinOrder,
	); err != nil {
		return domain.Article{}, fmt.Errorf("scan article: %w", err)
	}
	return article, nil
}
