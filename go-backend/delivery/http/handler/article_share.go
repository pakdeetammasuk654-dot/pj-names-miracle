package handler

import (
	"html"
	"net/url"
)

func renderArticleShare(articleURL, title string) string {
	escapedURL := html.EscapeString(articleURL)
	facebookURL := "https://www.facebook.com/sharer/sharer.php?u=" + url.QueryEscape(articleURL)
	xURL := "https://twitter.com/intent/tweet?url=" + url.QueryEscape(articleURL) + "&text=" + url.QueryEscape(title)
	lineURL := "https://social-plugins.line.me/lineit/share?url=" + url.QueryEscape(articleURL)
	return `<section class="article-share" aria-labelledby="article-share-title"><h2 id="article-share-title">แชร์บทความนี้</h2><div class="article-share-actions"><a class="article-share-button article-share-facebook" href="` + html.EscapeString(facebookURL) + `" target="_blank" rel="noopener noreferrer">Facebook</a><a class="article-share-button article-share-x" href="` + html.EscapeString(xURL) + `" target="_blank" rel="noopener noreferrer">X</a><a class="article-share-button article-share-line" href="` + html.EscapeString(lineURL) + `" target="_blank" rel="noopener noreferrer">LINE</a><button class="article-share-button article-share-copy" type="button" data-share-url="` + escapedURL + `" data-share-title="` + html.EscapeString(title) + `">คัดลอกลิงก์</button><button class="article-share-button article-share-native" type="button" data-share-url="` + escapedURL + `" data-share-title="` + html.EscapeString(title) + `">แชร์…</button></div></section>`
}
