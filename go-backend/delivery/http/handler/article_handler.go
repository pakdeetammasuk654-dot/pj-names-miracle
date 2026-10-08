package handler

import (
	"embed"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"go-backend/domain"
	"go-backend/repository"
	xhtml "golang.org/x/net/html"
)

const publicSiteURL = "https://xn--b3cu8e7ah6h.com"

//go:embed templates/articles.html templates/article_detail.html
var articleTemplateFS embed.FS

type ArticleHandler struct {
	repo     repository.ArticleRepository
	policy   *bluemonday.Policy
	template *template.Template
}

func NewArticleHandler(repo repository.ArticleRepository) (*ArticleHandler, error) {
	templates, err := template.ParseFS(articleTemplateFS, "templates/articles.html", "templates/article_detail.html")
	if err != nil {
		return nil, fmt.Errorf("parse article templates: %w", err)
	}
	return &ArticleHandler{repo: repo, policy: articleSanitizationPolicy(), template: templates}, nil
}

func (h *ArticleHandler) List(c *gin.Context) {
	articles, err := h.repo.ListPublished(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดบทความได้")
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := h.template.ExecuteTemplate(c.Writer, "articles.html", articleListPage{Articles: articles, JSONLD: renderArticleListJSONLD(articles)}); err != nil {
		c.Error(err)
	}
}

func (h *ArticleHandler) Detail(c *gin.Context) {
	slug, err := url.PathUnescape(c.Param("slug"))
	if err != nil || strings.TrimSpace(slug) == "" {
		c.String(http.StatusBadRequest, "ลิงก์บทความไม่ถูกต้อง")
		return
	}
	article, err := h.repo.GetPublishedBySlug(c.Request.Context(), slug)
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดบทความได้")
		return
	}
	if article == nil {
		c.String(http.StatusNotFound, "ไม่พบบทความ")
		return
	}
	articles, err := h.repo.ListPublished(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถโหลดบทความที่เกี่ยวข้องได้")
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	content := enhanceArticleImages(h.policy.Sanitize(article.ContentHTML), article.Title)
	content += articleEditorialNote
	content += renderArticleShare(publicSiteURL+"/articles/"+url.PathEscape(article.Slug), article.Title)
	content += renderRelatedArticles(*article, articles)
	content += articleDownloadCTA
	content += renderArticleJSONLD(*article)
	if err := h.template.ExecuteTemplate(c.Writer, "article_detail.html", articleDetailPage{
		Article:        *article,
		CanonicalURL:   publicSiteURL + "/articles/" + url.PathEscape(article.Slug),
		OpenGraphImage: openGraphImageURL(*article),
		MetaTitle:      articleSEOTitle(*article),
		ContentHTML:    template.HTML(content), // article content is sanitized; SEO and navigation components are trusted, shared components
	}); err != nil {
		c.Error(err)
	}
}

func (h *ArticleHandler) Sitemap(c *gin.Context) {
	articles, err := h.repo.ListPublished(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "ไม่สามารถสร้าง sitemap ได้")
		return
	}
	urls := []sitemapURL{
		{Location: publicSiteURL + "/", LastModified: "2026-08-26", ChangeFrequency: "weekly", Priority: "1.0"},
		{Location: publicSiteURL + "/articles", LastModified: "2026-08-26", ChangeFrequency: "daily", Priority: "0.9"},
		{Location: publicSiteURL + "/download", LastModified: "2026-08-26", ChangeFrequency: "monthly", Priority: "0.9"},
		{Location: publicSiteURL + "/auspicious-calendar", LastModified: "2026-08-26", ChangeFrequency: "weekly", Priority: "0.8"},
		{Location: publicSiteURL + "/isan-auspicious-method", LastModified: "2026-08-26", ChangeFrequency: "monthly", Priority: "0.8"},
		{Location: publicSiteURL + "/methodology", LastModified: "2026-08-26", ChangeFrequency: "monthly", Priority: "0.8"},
		{Location: publicSiteURL + "/about", LastModified: "2026-08-26", ChangeFrequency: "monthly", Priority: "0.8"},
		{Location: publicSiteURL + "/contact", LastModified: "2026-08-26", ChangeFrequency: "yearly", Priority: "0.5"},
		{Location: publicSiteURL + "/privacy", LastModified: "2026-08-26", ChangeFrequency: "yearly", Priority: "0.4"},
		{Location: publicSiteURL + "/terms", LastModified: "2026-08-26", ChangeFrequency: "yearly", Priority: "0.4"},
	}
	for _, article := range articles {
		urls = append(urls, sitemapURL{
			Location:        publicSiteURL + "/articles/" + url.PathEscape(article.Slug),
			LastModified:    articleLastModified(article).UTC().Format("2006-01-02"),
			ChangeFrequency: "monthly",
			Priority:        "0.8",
		})
	}
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Status(http.StatusOK)
	_, _ = c.Writer.Write([]byte(xml.Header))
	if err := xml.NewEncoder(c.Writer).Encode(sitemap{XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9", URLs: urls}); err != nil {
		c.Error(err)
	}
}

// ListJSON exposes only card-safe editorial metadata for the landing page.
func (h *ArticleHandler) ListJSON(c *gin.Context) {
	articles, err := h.repo.ListPublished(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถโหลดบทความได้"})
		return
	}
	items := make([]articleSummary, 0, len(articles))
	for _, article := range articles {
		items = append(items, articleSummary{
			Slug: article.Slug, Title: article.Title, TitleShort: article.TitleShort,
			Excerpt: article.Excerpt, Category: article.Category, ImageURL: article.ImageURL,
		})
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

type articleListPage struct {
	Articles []domain.Article
	JSONLD   template.HTML
}

type articleDetailPage struct {
	Article        domain.Article
	CanonicalURL   string
	OpenGraphImage string
	MetaTitle      string
	ContentHTML    template.HTML
}

const articleEditorialNote = `<aside class="article-editorial-note"><strong>เกี่ยวกับบทความนี้</strong><p>เรียบเรียงโดย อ.ทญา ว่องไว ผู้พัฒนาระบบชื่อดี 4 มิติ เนื้อหาด้านเลขศาสตร์และฤกษ์ยามเป็นความเชื่อตามตำรา ใช้เพื่อประกอบการพิจารณา ไม่ใช่คำรับรองผลลัพธ์หรือคำแนะนำทางการแพทย์ กฎหมาย และการเงิน</p><a href="/about">อ่านประวัติและแนวทางการทำงานของผู้เขียน →</a></aside>`

func articleSEOTitle(article domain.Article) string {
	title := strings.TrimSpace(article.Title)
	const brandSuffix = " | ชื่อดี"
	if utf8RuneCount(title)+utf8RuneCount(brandSuffix) <= 60 {
		return title + brandSuffix
	}
	short := strings.TrimSpace(article.TitleShort)
	if short != "" {
		return "วิเคราะห์ชื่อ " + short + " ด้วยเลขศาสตร์ | ชื่อดี"
	}
	runes := []rune(title)
	if len(runes) > 52 {
		title = string(runes[:52])
	}
	return title + brandSuffix
}

func enhanceArticleImages(content, articleTitle string) string {
	document, err := xhtml.Parse(strings.NewReader("<!doctype html><html><body>" + content + "</body></html>"))
	if err != nil {
		return content
	}
	var body *xhtml.Node
	var walk func(*xhtml.Node)
	walk = func(node *xhtml.Node) {
		if node.Type == xhtml.ElementNode && node.Data == "body" {
			body = node
		}
		if node.Type == xhtml.ElementNode && node.Data == "img" {
			setHTMLAttribute(node, "loading", "lazy")
			setHTMLAttribute(node, "decoding", "async")
			alt := htmlAttribute(node, "alt")
			lowerAlt := strings.ToLower(strings.TrimSpace(alt))
			if lowerAlt == "" || strings.HasSuffix(lowerAlt, ".jpg") || strings.HasSuffix(lowerAlt, ".jpeg") || strings.HasSuffix(lowerAlt, ".png") || strings.HasSuffix(lowerAlt, ".webp") {
				setHTMLAttribute(node, "alt", "ภาพประกอบบทความ "+articleTitle)
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(document)
	if body == nil {
		return content
	}
	var rendered strings.Builder
	for child := body.FirstChild; child != nil; child = child.NextSibling {
		if err := xhtml.Render(&rendered, child); err != nil {
			return content
		}
	}
	return rendered.String()
}

func htmlAttribute(node *xhtml.Node, name string) string {
	for _, attribute := range node.Attr {
		if attribute.Key == name {
			return attribute.Val
		}
	}
	return ""
}

func setHTMLAttribute(node *xhtml.Node, name, value string) {
	for index := range node.Attr {
		if node.Attr[index].Key == name {
			node.Attr[index].Val = value
			return
		}
	}
	node.Attr = append(node.Attr, xhtml.Attribute{Key: name, Val: value})
}

func absoluteArticleURL(path string) string {
	if strings.TrimSpace(path) == "" {
		return publicSiteURL + "/images/app_logo-512.png"
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	return publicSiteURL + "/" + strings.TrimPrefix(path, "/")
}

// openGraphImageURL keeps one canonical image URL for social crawlers. Facebook
// caches this URL reliably and refreshes it when its Sharing Debugger re-scrapes.
func openGraphImageURL(article domain.Article) string {
	return absoluteArticleURL(article.ImageURL)
}

func articleLastModified(article domain.Article) time.Time {
	if article.UpdatedAt.IsZero() {
		return article.PublishedAt
	}
	return article.UpdatedAt
}

func renderRelatedArticles(current domain.Article, articles []domain.Article) string {
	return relatedArticleHTML(selectRelatedArticles(current, articles))
}

type relatedArticleCandidate struct {
	article domain.Article
	score   int
	index   int
}

func selectRelatedArticles(current domain.Article, articles []domain.Article) []domain.Article {
	currentTerms := articleSimilarityTerms(current)
	candidates := make([]relatedArticleCandidate, 0, len(articles))
	for index, article := range articles {
		if sameArticle(current, article) {
			continue
		}
		sharedTerms := sharedTermCount(currentTerms, articleSimilarityTerms(article))
		if sharedTerms > 8 {
			sharedTerms = 8
		}
		score := sharedTerms * 8
		if current.Category != "" && current.Category == article.Category {
			score += 100
		}
		candidates = append(candidates, relatedArticleCandidate{article: article, score: score, index: index})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].index < candidates[j].index
		}
		return candidates[i].score > candidates[j].score
	})
	if len(candidates) > 3 {
		candidates = candidates[:3]
	}
	related := make([]domain.Article, 0, len(candidates))
	for _, candidate := range candidates {
		related = append(related, candidate.article)
	}
	return related
}

func sameArticle(left, right domain.Article) bool {
	if left.ID != 0 || right.ID != 0 {
		return left.ID == right.ID
	}
	return left.Slug == right.Slug
}

func articleSimilarityTerms(article domain.Article) map[string]struct{} {
	text := strings.ToLower(article.Title + " " + article.TitleShort + " " + article.Excerpt)
	text = strings.NewReplacer("ชื่อดี", "", "เลขศาสตร์", "", "พลังเงา", "", "วิเคราะห์", "", "บทความ", "").Replace(text)
	terms := make(map[string]struct{})
	for _, word := range strings.FieldsFunc(text, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		if utf8RuneCount(word) >= 3 {
			terms[word] = struct{}{}
		}
	}
	thaiRunes := []rune(text)
	for index := 0; index+2 < len(thaiRunes); index++ {
		trigram := thaiRunes[index : index+3]
		if unicode.Is(unicode.Thai, trigram[0]) && unicode.Is(unicode.Thai, trigram[1]) && unicode.Is(unicode.Thai, trigram[2]) {
			terms[string(trigram)] = struct{}{}
		}
	}
	return terms
}

func utf8RuneCount(value string) int {
	return len([]rune(value))
}

func sharedTermCount(left, right map[string]struct{}) int {
	count := 0
	for term := range left {
		if _, ok := right[term]; ok {
			count++
		}
	}
	return count
}

func relatedArticleHTML(articles []domain.Article) string {
	if len(articles) == 0 {
		return ""
	}
	var content strings.Builder
	content.WriteString(`<section class="related-articles" aria-labelledby="related-articles-title"><h2 id="related-articles-title">บทความที่เกี่ยวข้อง</h2><div class="related-articles-grid">`)
	for _, article := range articles {
		imageURL := article.ImageURL
		if strings.TrimSpace(imageURL) == "" {
			imageURL = "/images/app_logo-256.png?v=4"
		}
		content.WriteString(`<a class="related-article-card" href="/articles/`)
		content.WriteString(html.EscapeString(url.PathEscape(article.Slug)))
		content.WriteString(`"><img src="`)
		content.WriteString(html.EscapeString(imageURL))
		content.WriteString(`" alt="`)
		content.WriteString(html.EscapeString(article.Title))
		content.WriteString(`" width="640" height="360" loading="lazy" decoding="async"><span>`)
		content.WriteString(html.EscapeString(article.Category))
		content.WriteString(`</span><h3>`)
		content.WriteString(html.EscapeString(article.Title))
		content.WriteString(`</h3></a>`)
	}
	content.WriteString(`</div></section>`)
	return content.String()
}

func renderArticleJSONLD(article domain.Article) string {
	type articleSchema struct {
		Context          string `json:"@context"`
		Type             string `json:"@type"`
		MainEntityOfPage string `json:"mainEntityOfPage"`
		Headline         string `json:"headline"`
		Description      string `json:"description"`
		Image            string `json:"image"`
		DatePublished    string `json:"datePublished"`
		DateModified     string `json:"dateModified"`
		Author           struct {
			Type string `json:"@type"`
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"author"`
		Publisher struct {
			Type string `json:"@type"`
			Name string `json:"name"`
			Logo struct {
				Type string `json:"@type"`
				URL  string `json:"url"`
			} `json:"logo"`
		} `json:"publisher"`
	}
	data := articleSchema{Context: "https://schema.org", Type: "Article", MainEntityOfPage: publicSiteURL + "/articles/" + url.PathEscape(article.Slug), Headline: article.Title, Description: article.Excerpt, Image: absoluteArticleURL(article.ImageURL), DatePublished: article.PublishedAt.UTC().Format(time.RFC3339), DateModified: articleLastModified(article).UTC().Format(time.RFC3339)}
	data.Author.Type, data.Author.Name, data.Author.URL = "Person", "ทญา ว่องไว", publicSiteURL+"/about"
	data.Publisher.Type, data.Publisher.Name = "Organization", "ชื่อดี-ตั้งชื่อมงคล"
	data.Publisher.Logo.Type, data.Publisher.Logo.URL = "ImageObject", publicSiteURL+"/images/app_logo-512.png"
	payload, err := json.Marshal(data)
	if err != nil {
		return ""
	}
	breadcrumbPayload, _ := json.Marshal(map[string]any{
		"@context": "https://schema.org",
		"@type":    "BreadcrumbList",
		"itemListElement": []map[string]any{
			{"@type": "ListItem", "position": 1, "name": "หน้าแรก", "item": publicSiteURL + "/"},
			{"@type": "ListItem", "position": 2, "name": "บทความ", "item": publicSiteURL + "/articles"},
			{"@type": "ListItem", "position": 3, "name": article.Title, "item": publicSiteURL + "/articles/" + url.PathEscape(article.Slug)},
		},
	})
	return `<script type="application/ld+json">` + string(payload) + `</script><script type="application/ld+json">` + string(breadcrumbPayload) + `</script>`
}

func renderArticleListJSONLD(articles []domain.Article) template.HTML {
	items := make([]map[string]any, 0, len(articles))
	for index, article := range articles {
		items = append(items, map[string]any{
			"@type":    "ListItem",
			"position": index + 1,
			"url":      publicSiteURL + "/articles/" + url.PathEscape(article.Slug),
			"name":     article.Title,
		})
	}
	payload, err := json.Marshal(map[string]any{
		"@context":   "https://schema.org",
		"@type":      "CollectionPage",
		"name":       "บทความชื่อมงคล เลขศาสตร์ และพลังเงา",
		"url":        publicSiteURL + "/articles",
		"inLanguage": "th-TH",
		"mainEntity": map[string]any{"@type": "ItemList", "itemListElement": items},
	})
	if err != nil {
		return ""
	}
	breadcrumbPayload, _ := json.Marshal(map[string]any{
		"@context": "https://schema.org",
		"@type":    "BreadcrumbList",
		"itemListElement": []map[string]any{
			{"@type": "ListItem", "position": 1, "name": "หน้าแรก", "item": publicSiteURL + "/"},
			{"@type": "ListItem", "position": 2, "name": "บทความ", "item": publicSiteURL + "/articles"},
		},
	})
	return template.HTML(`<script type="application/ld+json">` + string(payload) + `</script><script type="application/ld+json">` + string(breadcrumbPayload) + `</script>`)
}

func formatArticleDate(date time.Time) string {
	return date.Format("2 January 2006")
}

type sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

type sitemapURL struct {
	Location        string `xml:"loc"`
	LastModified    string `xml:"lastmod,omitempty"`
	ChangeFrequency string `xml:"changefreq"`
	Priority        string `xml:"priority"`
}

type articleSummary struct {
	Slug       string `json:"slug"`
	Title      string `json:"title"`
	TitleShort string `json:"title_short"`
	Excerpt    string `json:"excerpt"`
	Category   string `json:"category"`
	ImageURL   string `json:"image_url"`
}
