package handler

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type adminTemplateExecutor interface {
	ExecuteTemplate(io.Writer, string, any) error
}

const adminNavigation = `<nav class="admin-navigation" aria-label="เมนูหลังบ้าน">
  <a href="/admin">Dashboard</a>
  <a href="/admin/names">จัดการชื่อมงคล</a>
  <a href="/admin/dimension-labels">มิติพลัง 4 ด้าน</a>
  <a href="/admin/auspicious-content">เนื้อหาฤกษ์ยาม</a>
  <a href="/admin/articles">บทความ</a>
  <a href="/admin/articles/new">สร้างบทความ</a>
  <a href="/admin/articles/media">คลังรูป</a>
	<a href="/admin/chat">แชตลูกค้า</a>
  <a href="/" class="admin-navigation-site">ดูหน้าเว็บ ↗</a>
</nav>
<style>
  .admin-navigation{display:flex;align-items:center;gap:4px;flex-wrap:wrap;padding:12px 18px;background:#0b2115;border-bottom:1px solid rgba(212,175,55,.38);position:sticky;top:0;z-index:20}
  .admin-navigation a{padding:8px 11px;border-radius:7px;color:rgba(255,255,255,.8);font-weight:700;text-decoration:none}
  .admin-navigation a:hover{background:rgba(212,175,55,.16);color:#f3e5ab}
  .admin-navigation .admin-navigation-site{margin-left:auto;color:#f3e5ab}
  @media(max-width:560px){.admin-navigation{padding:8px 10px;gap:2px}.admin-navigation a{padding:7px 8px;font-size:.84rem}.admin-navigation .admin-navigation-site{margin-left:0}}
</style>`

func renderAdminPage(c *gin.Context, template adminTemplateExecutor, name string, data any) {
	var page bytes.Buffer
	if err := template.ExecuteTemplate(&page, name, data); err != nil {
		c.Error(err)
		return
	}
	output := strings.Replace(page.String(), "</head>", `<meta name="robots" content="noindex,nofollow,noarchive,nosnippet"><link rel="icon" type="image/png" href="/images/app_logo-256.png"></head>`, 1)
	output = strings.Replace(output, "<body>", "<body>"+adminNavigation, 1)
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(output))
}
