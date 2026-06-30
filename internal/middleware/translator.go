package middleware

import (
	"go-server-starter/internal/constant"
	"go-server-starter/internal/i18n"
	"go-server-starter/pkg/translator"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

var matcher = language.NewMatcher([]language.Tag{
	language.English, // Default
	language.Chinese, // zh
})

// localeCache caches Accept-Language header → normalized locale.
var localeCache sync.Map

// Translations 翻译中间件,根据请求头的 locale 设置翻译器
func Translations(t *translator.Translator) gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := i18n.NormalizeLocale(extractLocale(c))
		trans, _ := t.GetTranslator(locale)
		c.Set(constant.CTX_KEY_OF_TRANSLATOR, trans)
		c.Set(constant.CTX_KEY_OF_LOCALE, locale)
		c.Next()
	}
}

// extractLocale 从请求头中提取并规范化语言环境
func extractLocale(c *gin.Context) string {
	// 首先，检查自定义的 "locale" 请求头
	locale := c.GetHeader("locale")
	if locale != "" {
		return locale
	}

	// Check cache
	acceptLang := c.GetHeader("Accept-Language")
	if cached, ok := localeCache.Load(acceptLang); ok {
		return cached.(string)
	}

	// Parse and match（解析 Accept-Language，自动按 q 值排序）
	tags, _, err := language.ParseAcceptLanguage(acceptLang)
	if err != nil {
		// Parsing failed; return empty string and let NormalizeLocale fall back to default
		return ""
	}
	// 匹配到最佳支持的语言
	tag, _, _ := matcher.Match(tags...)
	// Match always returns a tag (the first in the supported list when no match found),
    // so ignoring its confidence return value is intentional.
	result := tag.String()

	localeCache.Store(acceptLang, result)
	return result
}
