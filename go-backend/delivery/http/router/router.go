package router

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go-backend/delivery/http/handler"
	"go-backend/domain"
)

const canonicalPublicHost = "xn--b3cu8e7ah6h.com"

func SetupRouter(
	healthHandler *handler.HealthHandler,
	readinessHandler *handler.ReadinessHandler,
	nameHandler *handler.NameHandler,
	articleHandler *handler.ArticleHandler,
	articleAdminHandler *handler.ArticleAdminHandler,
	articleMediaHandler *handler.ArticleMediaHandler,
	calendarHandler *handler.AuspiciousCalendarHandler,
	adminDashboardHandler *handler.AdminDashboardHandler,
	supportChatHandler *handler.SupportChatHandler,
	dimensionAdminHandler *handler.DimensionAdminHandler,
	auspiciousAdminHandler *handler.AuspiciousAdminHandler,
	nameAdminHandler *handler.NameAdminHandler,
	entitlementHandlers ...*handler.EntitlementHandler,
) *gin.Engine {
	r := gin.Default()
	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(err)
	}
	r.Use(CanonicalHostRedirect(), SearchCrawlerHeaders(), PublicAssetCacheHeaders())

	// Serve avatars, static images, and public web assets
	r.Static("/avatars", "./avatars")
	r.Static("/buddhas", "./thai_buddha")
	r.Static("/images", "./images")
	r.Static("/static", "./static")
	r.Static("/chat-uploads", "./chat_uploads")

	// Web Portal Routes (Store Compliance & Promotion)
	servePage := func(path, file string) {
		r.GET(path, func(c *gin.Context) { c.File(file) })
		r.HEAD(path, func(c *gin.Context) { c.File(file) })
	}

	servePage("/", "./static/index.html")
	servePage("/about", "./static/about.html")
	servePage("/methodology", "./static/methodology.html")
	servePage("/privacy", "./static/privacy.html")
	servePage("/terms", "./static/terms.html")
	servePage("/contact", "./static/contact.html")
	servePage("/download", "./static/download.html")
	servePage("/auspicious-calendar", "./static/auspicious-calendar.html")
	servePage("/isan-auspicious-method", "./static/isan-auspicious-method.html")
	servePage("/taksa-jora-method", "./static/taksa-jora-method.html")
	redirectPermanent := func(from, to string) {
		r.GET(from, func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, to) })
		r.HEAD(from, func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, to) })
	}
	redirectPermanent("/about.html", "/about")
	redirectPermanent("/articles.html", "/articles")
	redirectPermanent("/article", "/articles")
	redirectPermanent("/auspicious-calendar.html", "/auspicious-calendar")
	redirectPermanent("/contact.html", "/contact")
	redirectPermanent("/download.html", "/download")
	redirectPermanent("/isan-auspicious-method.html", "/isan-auspicious-method")
	redirectPermanent("/taksa-jora-method.html", "/taksa-jora-method")
	redirectPermanent("/methodology.html", "/methodology")
	redirectPermanent("/privacy.html", "/privacy")
	redirectPermanent("/privacy-policy", "/privacy")
	redirectPermanent("/privacy-policy.html", "/privacy")
	redirectPermanent("/terms.html", "/terms")
	redirectPermanent("/terms-of-service", "/terms")
	redirectPermanent("/terms-of-service.html", "/terms")
	r.GET("/admin/login", func(c *gin.Context) { c.Status(http.StatusGone) })
	r.HEAD("/admin/login", func(c *gin.Context) { c.Status(http.StatusGone) })
	if supportChatHandler != nil {
		r.GET("/chat", supportChatHandler.CustomerChatPage)
		r.HEAD("/chat", supportChatHandler.CustomerChatPage)
	}
	r.GET("/articles", articleHandler.List)
	r.HEAD("/articles", articleHandler.List)
	r.GET("/articles/:slug", articleHandler.Detail)
	r.HEAD("/articles/:slug", articleHandler.Detail)
	r.GET("/sitemap.xml", articleHandler.Sitemap)
	r.HEAD("/sitemap.xml", articleHandler.Sitemap)
	if adminDashboardHandler != nil {
		r.GET("/admin", adminDashboardHandler.RequireAuth(), adminDashboardHandler.Index)
	}
	if dimensionAdminHandler != nil {
		r.GET("/admin/dimension-labels", dimensionAdminHandler.RequireAuth(), dimensionAdminHandler.Index)
		r.POST("/admin/dimension-labels", dimensionAdminHandler.RequireAuth(), dimensionAdminHandler.Save)
	}
	if auspiciousAdminHandler != nil {
		r.GET("/admin/auspicious-content", auspiciousAdminHandler.RequireAuth(), auspiciousAdminHandler.Index)
		r.POST("/admin/auspicious-content", auspiciousAdminHandler.RequireAuth(), auspiciousAdminHandler.Save)
	}
	if nameAdminHandler != nil {
		r.GET("/admin/names", nameAdminHandler.RequireAuth(), nameAdminHandler.Index)
		r.POST("/admin/names/analyze", nameAdminHandler.RequireAuth(), nameAdminHandler.AnalyzeAPI)
		r.POST("/admin/names/create", nameAdminHandler.RequireAuth(), nameAdminHandler.Create)
	}
	if supportChatHandler != nil {
		r.GET("/admin/chat", supportChatHandler.RequireAuth(), supportChatHandler.AdminIndex)
		r.GET("/admin/chat/summary", supportChatHandler.RequireAuth(), supportChatHandler.AdminInboxStats)
		r.GET("/admin/chat/:id", supportChatHandler.RequireAuth(), supportChatHandler.AdminConversation)
		r.GET("/admin/chat/:id/messages", supportChatHandler.RequireAuth(), supportChatHandler.AdminMessagesJSON)
		r.POST("/admin/chat/:id/messages", supportChatHandler.RequireAuth(), supportChatHandler.AdminSend)
		r.POST("/admin/chat/:id/messages/:messageID/delete", supportChatHandler.RequireAuth(), supportChatHandler.AdminDeleteMessage)
		r.POST("/admin/chat/:id/close", supportChatHandler.RequireAuth(), supportChatHandler.AdminClose)

		registerAdminAPIs := func(group *gin.RouterGroup) {
			group.POST("/auth/login", supportChatHandler.AdminAuthLogin)
			group.GET("/chat/conversations", supportChatHandler.AdminConversationsJSON)
			group.GET("/chat/conversations/:id/messages", supportChatHandler.AdminMessagesJSON)
			group.POST("/chat/conversations/:id/messages", supportChatHandler.AdminSendJSON)
			group.DELETE("/chat/conversations/:id/messages/:messageID", supportChatHandler.AdminDeleteMessageJSON)
			group.POST("/chat/conversations/:id/close", supportChatHandler.AdminCloseJSON)
			group.POST("/chat/device-token", supportChatHandler.AdminRegisterDeviceToken)
		}

		registerAdminAPIs(r.Group("/api/v1/admin", supportChatHandler.RequireAuth()))
		registerAdminAPIs(r.Group("/api/admin", supportChatHandler.RequireAuth()))

		// This page is intentionally a local QA tool, never a public customer portal.
		if os.Getenv("CHAT_TEST_PAGE_ENABLED") == "true" {
			r.GET("/local/chat-test", supportChatHandler.CustomerChatPage)
		}
	}
	if articleAdminHandler != nil {
		adminArticles := r.Group("/admin/articles", articleAdminHandler.RequireAuth())
		adminArticles.GET("", articleAdminHandler.Index)
		adminArticles.GET("/new", articleAdminHandler.New)
		adminArticles.POST("", articleAdminHandler.Create)
		adminArticles.GET("/:id/edit", articleAdminHandler.Edit)
		adminArticles.POST("/:id", articleAdminHandler.Update)
		adminArticles.POST("/:id/delete", articleAdminHandler.Delete)
	}
	if articleMediaHandler != nil {
		adminMedia := r.Group("/admin/articles/media", articleMediaHandler.RequireAuth())
		adminMedia.GET("", articleMediaHandler.Index)
		adminMedia.POST("", articleMediaHandler.Upload)
		adminMedia.POST("/:filename/delete", articleMediaHandler.Delete)
	}
	servePage("/robots.txt", "./static/robots.txt")
	servePage("/site.webmanifest", "./static/site.webmanifest")

	serveApk := func(c *gin.Context) {
		c.Header("Content-Disposition", "attachment; filename=com.numberniceic.app_v1.3.1.apk")
		c.Header("Content-Type", "application/vnd.android.package-archive")
		if _, err := os.Stat("./static/com.numberniceic.app_v1.3.1.apk"); err == nil {
			c.File("./static/com.numberniceic.app_v1.3.1.apk")
		} else if _, err := os.Stat("./static/com.numberniceic.app_v1.3.0.apk"); err == nil {
			c.File("./static/com.numberniceic.app_v1.3.0.apk")
		} else {
			c.File("./static/com.numberniceic.app_v1.1.0.apk")
		}
	}
	r.GET("/download/apk", serveApk)
	r.HEAD("/download/apk", serveApk)

	// Base API route group
	api := r.Group("/api/v1")
	api.Use(APIContractHeaders())
	{
		api.GET("/health", healthHandler.CheckHealth)
		api.GET("/readiness", readinessHandler.CheckReadiness)
		api.GET("/hello", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "Hello World from Gin Clean Architecture with Pipeline Pattern!",
			})
		})

		// Name Search API (Top 100 Best Names)
		searchLimit := RateLimitByClientIP(30, 10)
		api.POST("/names/search", searchLimit, nameHandler.SearchTopNames)
		api.GET("/names/search", searchLimit, nameHandler.SearchTopNames)
		api.GET("/names/suggestions", searchLimit, nameHandler.SuggestNames)
		api.POST("/names/rerank", searchLimit, nameHandler.RerankTopNames)
		if len(entitlementHandlers) > 0 && entitlementHandlers[0] != nil {
			api.POST("/entitlements/verify", RateLimitByClientIP(12, 4), entitlementHandlers[0].VerifyPurchase)
		}

		// Celebrities API
		api.GET("/celebrities", nameHandler.GetCelebrities)

		// Decode & Pair Detail API
		api.GET("/decode", nameHandler.DecodeName)
		api.GET("/pair-detail", nameHandler.GetPairDetail)
		api.GET("/meanings/random", nameHandler.GetRandomMeanings)
		api.GET("/articles", articleHandler.ListJSON)
		if calendarHandler != nil {
			// Auspicious Feature APIs for Web, Android & iOS Native Clients
			api.GET("/auspicious-calendar", calendarHandler.Month)
			api.GET("/auspicious/calendar", calendarHandler.Month)
			api.GET("/auspicious-activities", calendarHandler.Activities)
			api.GET("/auspicious/activities", calendarHandler.Activities)
			api.POST("/auspicious-search", RateLimitByClientIP(30, 10), calendarHandler.Search)
			api.POST("/auspicious/search", RateLimitByClientIP(30, 10), calendarHandler.Search)
			api.GET("/auspicious/taksa", calendarHandler.Taksa)
			api.POST("/auspicious/taksa", calendarHandler.Taksa)
		}
		chatLimit := RateLimitByClientIP(120, 20)
		api.POST("/chat/conversations", chatLimit, supportChatHandler.CreateCustomerConversation)
		api.DELETE("/chat/conversations/:id", chatLimit, supportChatHandler.DeleteCustomerConversation)
		api.GET("/chat/conversations/:id/unread", chatLimit, supportChatHandler.CustomerUnread)
		api.GET("/chat/conversations/:id/messages", chatLimit, supportChatHandler.CustomerMessages)
		api.POST("/chat/conversations/:id/messages", chatLimit, supportChatHandler.CustomerSend)
		api.POST("/chat/conversations/:id/device-token", chatLimit, supportChatHandler.RegisterCustomerDeviceToken)
	}

	return r
}

// CanonicalHostRedirect consolidates the www hostname into the public hostname
// so search engines do not split link signals across duplicate pages.
func CanonicalHostRedirect() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.EqualFold(c.Request.Host, "www."+canonicalPublicHost) {
			c.Redirect(http.StatusMovedPermanently, "https://"+canonicalPublicHost+c.Request.URL.RequestURI())
			c.Abort()
			return
		}
		c.Next()
	}
}

// SearchCrawlerHeaders prevents private, machine-readable, and customer-upload
// endpoints from appearing as low-quality or sensitive search results.
func SearchCrawlerHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/admin") || strings.HasPrefix(path, "/api/") ||
			strings.HasPrefix(path, "/chat-uploads/") || strings.HasPrefix(path, "/local/") {
			c.Header("X-Robots-Tag", "noindex, nofollow, noarchive, nosnippet")
		} else if path == "/chat" {
			c.Header("X-Robots-Tag", "noindex, follow, noarchive, nosnippet")
		}
		c.Next()
	}
}

// PublicAssetCacheHeaders gives versioned UI assets and editorial images a
// longer browser/CDN lifetime without applying caching to HTML or APIs.
func PublicAssetCacheHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/images/articles/") ||
			strings.HasPrefix(path, "/images/app_logo-") {
			c.Header("Cache-Control", "public, max-age=604800, stale-while-revalidate=86400")
		}
		c.Next()
	}
}

// APIContractHeaders lets native clients fail closed when a future backend
// contract is incompatible with the version they were built against.
func APIContractHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-Version", domain.APIVersion)
		c.Header("X-Ranking-Contract-Version", domain.RankingContractVersion)
		c.Next()
	}
}
