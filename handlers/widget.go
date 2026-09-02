package handlers

import (
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"dishub_openapi/database"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ── Badge helpers ──────────────────────────────────────────────

var badgeLabelI18n = map[string]map[string]string{
	"status":  {"ko": "상태", "en": "Status", "ja": "状態", "zh": "状态"},
	"servers": {"ko": "서버 수", "en": "Servers", "ja": "サーバー", "zh": "服务器"},
	"votes":   {"ko": "좋아요", "en": "Votes", "ja": "投票", "zh": "推荐"},
	"members": {"ko": "멤버 수", "en": "Members", "ja": "メンバー", "zh": "成员"},
}

var statusValueI18n = map[string]map[string]string{
	"online":    {"ko": "온라인", "en": "Online", "ja": "オンライン", "zh": "在线"},
	"idle":      {"ko": "자리비움", "en": "Idle", "ja": "離席中", "zh": "空闲"},
	"dnd":       {"ko": "방해금지", "en": "DND", "ja": "取り込み中", "zh": "请勿打扰"},
	"streaming": {"ko": "방송중", "en": "Streaming", "ja": "配信中", "zh": "直播中"},
	"offline":   {"ko": "오프라인", "en": "Offline", "ja": "オフライン", "zh": "离线"},
}

// textWidthEst approximates pixel width for Verdana ~11px.
// CJK/wide characters count as 13px, ASCII as 6.5px.
func textWidthEst(s string) float64 {
	w := 0.0
	for _, r := range s {
		if r > 0x2E7F {
			w += 13.0
		} else {
			w += 6.5
		}
	}
	return w
}

func formatCount(n int) string {
	s := fmt.Sprintf("%d", n)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	return b.String()
}

func isFreshBotStatsValue(raw any) bool {
	var last time.Time
	switch v := raw.(type) {
	case primitive.DateTime:
		last = v.Time()
	case time.Time:
		last = v
	case string:
		parsed, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return false
		}
		last = parsed
	default:
		return false
	}
	return time.Since(last) <= 24*time.Hour
}

func botStatusWithFreshness(bot bson.M) (string, bool) {
	if !isFreshBotStatsValue(bot["lastStatsAt"]) {
		return "offline", false
	}
	status, _ := bot["status"].(string)
	online, _ := bot["online"].(bool)
	return status, online
}

func badgeImageDataURI(imageURL string) string {
	if imageURL == "" || !strings.HasPrefix(imageURL, "https://") {
		return ""
	}

	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil || len(body) == 0 {
		return ""
	}

	return fmt.Sprintf("data:%s;base64,%s", contentType, base64.StdEncoding.EncodeToString(body))
}

// badgeLogoW: 3px left pad + 14px logo + 3px right pad
const badgeLogoW = 20

func makeBadgeSVG(label, value, valueBg, avatarUrl, name string) string {
	lw := textWidthEst(label)
	vw := textWidthEst(value)

	// label section = logo column + text + 10px padding each side
	lArea := badgeLogoW + int(lw) + 20
	vArea := int(vw) + 20
	total := lArea + vArea

	// Text centers in 10x scaled coordinates
	lMid := int((float64(badgeLogoW+10) + lw/2) * 10)
	vMid := int((float64(lArea+10) + vw/2) * 10)
	lLen := int(lw * 10)
	vLen := int(vw * 10)
	if lLen < 1 {
		lLen = 1
	}
	if vLen < 1 {
		vLen = 1
	}

	le := html.EscapeString(label)
	ve := html.EscapeString(value)

	var defsExtra, logoHTML string
	avatarDataURI := badgeImageDataURI(avatarUrl)
	if avatarDataURI != "" {
		defsExtra = `<clipPath id="ai"><rect x="3" y="3" rx="3" width="14" height="14"/></clipPath>`
		logoHTML = fmt.Sprintf(`<image href="%s" x="3" y="3" width="14" height="14" clip-path="url(#ai)" preserveAspectRatio="xMidYMid slice"/>`, html.EscapeString(avatarDataURI))
	} else {
		initial := "?"
		if runes := []rune(name); len(runes) > 0 {
			initial = strings.ToUpper(string(runes[:1]))
		}
		defsExtra = ""
		logoHTML = fmt.Sprintf(
			`<rect x="3" y="3" width="14" height="14" rx="3" fill="#5865F2"/>`+
				`<text x="10" y="13.5" text-anchor="middle" font-family="-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif" font-size="9" font-weight="700" fill="#fff">%s</text>`,
			html.EscapeString(initial),
		)
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img">`, total)
	fmt.Fprintf(&sb, `<title>%s: %s</title>`, le, ve)
	sb.WriteString(`<defs>`)
	sb.WriteString(`<linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#bbb" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient>`)
	sb.WriteString(defsExtra)
	fmt.Fprintf(&sb, `<clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath>`, total)
	sb.WriteString(`</defs>`)
	sb.WriteString(`<g clip-path="url(#r)">`)
	fmt.Fprintf(&sb, `<rect width="%d" height="20" fill="#555"/>`, lArea)
	fmt.Fprintf(&sb, `<rect x="%d" width="%d" height="20" fill="%s"/>`, lArea, vArea, valueBg)
	fmt.Fprintf(&sb, `<rect width="%d" height="20" fill="url(#s)"/>`, total)
	sb.WriteString(logoHTML)
	sb.WriteString(`</g>`)
	sb.WriteString(`<g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="110">`)
	fmt.Fprintf(&sb, `<text x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>`, lMid, lLen, le)
	fmt.Fprintf(&sb, `<text x="%d" y="140" transform="scale(.1)" textLength="%d">%s</text>`, lMid, lLen, le)
	fmt.Fprintf(&sb, `<text x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>`, vMid, vLen, ve)
	fmt.Fprintf(&sb, `<text x="%d" y="140" transform="scale(.1)" textLength="%d">%s</text>`, vMid, vLen, ve)
	sb.WriteString(`</g></svg>`)
	return sb.String()
}

func badgeSVGResponse(c *gin.Context, label, value, valueBg, avatarUrl, name string) {
	svg := makeBadgeSVG(label, value, valueBg, avatarUrl, name)
	c.Header("Cache-Control", "public, max-age=300")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Data(http.StatusOK, "image/svg+xml; charset=utf-8", []byte(svg))
}

func getBadgeLang(_ *gin.Context) string {
	return "ko"
}

func getBadgeLabel(key, lang string) string {
	if m, ok := badgeLabelI18n[key]; ok {
		if v, ok2 := m[lang]; ok2 {
			return v
		}
		return m["ko"]
	}
	return key
}

// ── Bot badge handlers ─────────────────────────────────────────

func botID(c *gin.Context) string {
	return c.Param("id")
}

func serverID(c *gin.Context) string {
	return c.Param("id")
}

func GetBotStatusBadge(c *gin.Context) {
	id := botID(c)
	lang := getBadgeLang(c)

	col := database.GetCollection("bots")
	var bot bson.M
	if err := col.FindOne(c.Request.Context(), bson.M{"botId": id}).Decode(&bot); err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	name, _ := bot["name"].(string)
	status, online := botStatusWithFreshness(bot)
	avatarUrl, _ := bot["avatar"].(string)

	s := strings.ToLower(status)
	var statusKey, color string
	switch {
	case s == "streaming":
		statusKey, color = "streaming", "#7c3aed"
	case s == "online" || online:
		statusKey, color = "online", "#22c55e"
	case s == "idle":
		statusKey, color = "idle", "#eab308"
	case s == "dnd":
		statusKey, color = "dnd", "#ef4444"
	default:
		statusKey, color = "offline", "#6b7280"
	}

	label := getBadgeLabel("status", lang)
	value := statusValueI18n[statusKey][lang]
	badgeSVGResponse(c, label, value, color, avatarUrl, name)
}

func GetBotServersBadge(c *gin.Context) {
	id := botID(c)
	lang := getBadgeLang(c)

	col := database.GetCollection("bots")
	var bot bson.M
	if err := col.FindOne(c.Request.Context(), bson.M{"botId": id}).Decode(&bot); err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	name, _ := bot["name"].(string)
	serverCount, _ := bot["serverCount"].(int32)
	avatarUrl, _ := bot["avatar"].(string)
	badgeSVGResponse(c, getBadgeLabel("servers", lang), formatCount(int(serverCount)), "#5865f2", avatarUrl, name)
}

func GetBotVotesBadge(c *gin.Context) {
	id := botID(c)
	lang := getBadgeLang(c)

	col := database.GetCollection("bots")
	var bot bson.M
	if err := col.FindOne(c.Request.Context(), bson.M{"botId": id}).Decode(&bot); err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	name, _ := bot["name"].(string)
	hearts, _ := bot["hearts"].(int32)
	avatarUrl, _ := bot["avatar"].(string)
	badgeSVGResponse(c, getBadgeLabel("votes", lang), formatCount(int(hearts)), "#ec4899", avatarUrl, name)
}

// ── Server badge handlers ──────────────────────────────────────

func serverIconUrl(server bson.M) string {
	icon, _ := server["icon"].(string)
	serverId, _ := server["serverId"].(string)
	if icon == "" || serverId == "" {
		return ""
	}
	if strings.HasPrefix(icon, "http") {
		return icon
	}
	ext := "png"
	if strings.HasPrefix(icon, "a_") {
		ext = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.%s?size=64", serverId, icon, ext)
}

func GetServerMembersBadge(c *gin.Context) {
	id := serverID(c)
	lang := getBadgeLang(c)

	col := database.GetCollection("servers")
	var server bson.M
	if err := col.FindOne(c.Request.Context(), bson.M{"serverId": id}).Decode(&server); err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	name, _ := server["name"].(string)
	memberCount, _ := server["memberCount"].(int32)
	badgeSVGResponse(c, getBadgeLabel("members", lang), formatCount(int(memberCount)), "#5865f2", serverIconUrl(server), name)
}

func GetServerVotesBadge(c *gin.Context) {
	id := serverID(c)
	lang := getBadgeLang(c)

	col := database.GetCollection("servers")
	var server bson.M
	if err := col.FindOne(c.Request.Context(), bson.M{"serverId": id}).Decode(&server); err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	name, _ := server["name"].(string)
	hearts, _ := server["hearts"].(int32)
	badgeSVGResponse(c, getBadgeLabel("votes", lang), formatCount(int(hearts)), "#ec4899", serverIconUrl(server), name)
}

var widgetBotI18n = map[string][2]string{
	"ko": {"자세히 보기", "초대하기"},
	"en": {"View More", "Invite"},
	"ja": {"詳しく見る", "招待する"},
	"zh": {"查看详情", "邀请"},
}

var widgetServerI18n = map[string][2]string{
	"ko": {"자세히 보기", "참가하기"},
	"en": {"View More", "Join"},
	"ja": {"詳しく見る", "参加する"},
	"zh": {"查看详情", "加入"},
}

var botTagLabels = map[string]map[string]string{
	"ko": {"manage": "관리", "music": "뮤직", "game": "게임", "gambling": "도박", "leveling": "레벨링", "utility": "유틸리티", "chat": "채팅", "webDashboard": "웹 대시보드"},
	"en": {"manage": "Manage", "music": "Music", "game": "Game", "gambling": "Gambling", "leveling": "Leveling", "utility": "Utility", "chat": "Chat", "webDashboard": "Web Dashboard"},
	"ja": {"manage": "管理", "music": "ミュージック", "game": "ゲーム", "gambling": "ギャンブル", "leveling": "レベリング", "utility": "ユーティリティ", "chat": "チャット", "webDashboard": "Webダッシュボード"},
	"zh": {"manage": "管理", "music": "音乐", "game": "游戏", "gambling": "赌博", "leveling": "等级", "utility": "实用", "chat": "聊天", "webDashboard": "网页面板"},
}

var serverTagLabels = map[string]map[string]string{
	"ko": {"community": "커뮤니티", "gaming": "게임", "music": "음악", "anime": "애니메이션", "art": "예술/창작", "tech": "기술/개발", "education": "교육", "social": "소셜/친목"},
	"en": {"community": "Community", "gaming": "Gaming", "music": "Music", "anime": "Anime", "art": "Art/Creative", "tech": "Tech/Dev", "education": "Education", "social": "Social"},
	"ja": {"community": "コミュニティ", "gaming": "ゲーム", "music": "音楽", "anime": "アニメ", "art": "アート/創作", "tech": "技術/開発", "education": "教育", "social": "ソーシャル"},
	"zh": {"community": "社区", "gaming": "游戏", "music": "音乐", "anime": "动漫", "art": "艺术/创作", "tech": "技术/开发", "education": "教育", "social": "社交"},
}

func translateTags(tags []string, labelMap map[string]map[string]string, lang string) []string {
	m, ok := labelMap[lang]
	if !ok {
		m = labelMap["ko"]
	}
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		if label, ok := m[t]; ok {
			result = append(result, label)
		} else {
			result = append(result, t)
		}
	}
	return result
}

func pickDesc(desc map[string]string, lang string) string {
	for _, k := range []string{lang, "ko", "en", "ja", "zh"} {
		if v, ok := desc[k]; ok && strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func widgetHTML(c *gin.Context, body string) {
	c.Header("X-Frame-Options", "ALLOWALL")
	c.Header("Content-Security-Policy", "frame-ancestors *")
	c.Header("Cache-Control", "public, max-age=300")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(body))
}

var widgetCSS = `
  * { box-sizing: border-box; margin: 0; padding: 0; }
  html, body { width: 100%; height: 100%; }
  body { font-family: Pretendard, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: transparent; color: #f5f6f8; overflow: hidden; }
  .card { position: relative; display: flex; flex-direction: column; height: 100%; padding: 13px; gap: 8px; border: 1px solid rgba(255,255,255,.09); border-radius: 14px; background: linear-gradient(145deg, #202632 0%, #171b23 100%); box-shadow: inset 0 1px 0 rgba(255,255,255,.035); }
  .card::before { content: ''; position: absolute; inset: 0; border-radius: inherit; background: radial-gradient(circle at 90% 0%, rgba(109,120,247,.13), transparent 42%); pointer-events: none; }
  .top, .desc, .tags, .btns, .footer { position: relative; z-index: 1; }
  .top { display: flex; align-items: center; gap: 11px; min-width: 0; }
  .avatar-wrap { position: relative; flex-shrink: 0; }
  .avatar { width: 44px; height: 44px; border-radius: 12px; object-fit: cover; background: linear-gradient(135deg, #7c86ff, #5865f2); display: flex; align-items: center; justify-content: center; font-weight: 800; font-size: 18px; color: #fff; flex-shrink: 0; box-shadow: 0 5px 14px rgba(0,0,0,.22); }
  .dot { position: absolute; bottom: -2px; right: -2px; width: 12px; height: 12px; border-radius: 50%; border: 2px solid #1b2029; box-shadow: 0 0 0 1px rgba(0,0,0,.12); }
  .info { flex: 1; min-width: 0; }
  .name { font-size: 15px; font-weight: 800; letter-spacing: -.02em; color: #f8fafc; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; display: flex; align-items: center; gap: 5px; }
  .badge { width: 15px; height: 15px; display: inline-flex; align-items: center; justify-content: center; border-radius: 50%; background: #6d78f7; color: #fff; font-size: 10px; flex-shrink: 0; }
  .stats { display: flex; gap: 9px; font-size: 10px; color: #aeb4c0; margin-top: 5px; flex-wrap: nowrap; overflow: hidden; }
  .stat { display: flex; align-items: center; gap: 3px; }
  .desc { font-size: 11px; color: #adb4c1; line-height: 1.45; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .tags { display: flex; gap: 4px; flex-wrap: nowrap; overflow: hidden; min-height: 19px; }
  .tag { background: rgba(109,120,247,.1); color: #b4baff; border: 1px solid rgba(109,120,247,.18); border-radius: 6px; padding: 2px 6px; font-size: 9px; font-weight: 700; white-space: nowrap; }
  .btns { display: flex; gap: 6px; margin-top: auto; }
  .btn { flex: 1; min-height: 29px; display: flex; align-items: center; justify-content: center; border-radius: 8px; font-size: 11px; font-weight: 750; text-decoration: none; transition: background .15s, border-color .15s, transform .15s; }
  .btn:hover { transform: translateY(-1px); }
  .btn-outline { border: 1px solid rgba(255,255,255,.1); color: #bdc3ce; background: rgba(255,255,255,.025); }
  .btn-outline:hover { border-color: rgba(109,120,247,.45); color: #fff; }
  .btn-fill { background: #6d78f7; color: #fff; box-shadow: 0 4px 12px rgba(88,101,242,.2); }
  .btn-fill:hover { background: #7b85fa; }
  .footer { font-size: 8px; letter-spacing: .03em; color: #667080; text-align: right; line-height: 1; }
`

func GetBotWidget(c *gin.Context) {
	id := c.Param("id")
	lang := c.DefaultQuery("lang", "ko")

	labels, ok := widgetBotI18n[lang]
	if !ok {
		labels = widgetBotI18n["ko"]
	}

	col := database.GetCollection("bots")
	var bot bson.M
	err := col.FindOne(c.Request.Context(), bson.M{"botId": id, "verified": true}).Decode(&bot)
	if err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	name, _ := bot["name"].(string)
	avatarUrl, _ := bot["avatar"].(string)
	inviteUrl, _ := bot["inviteUrl"].(string)
	badge, _ := bot["badge"].(bool)
	rating, _ := bot["rating"].(float64)
	hearts, _ := bot["hearts"].(int32)
	serverCount, _ := bot["serverCount"].(int32)
	status, online := botStatusWithFreshness(bot)

	descRaw, _ := bot["description"].(bson.M)
	descMap := make(map[string]string)
	for k, v := range descRaw {
		if s, ok := v.(string); ok {
			descMap[k] = s
		}
	}
	desc := pickDesc(descMap, lang)

	var rawTags []string
	if t, ok := bot["tags"].(bson.A); ok {
		for _, v := range t {
			if s, ok := v.(string); ok {
				rawTags = append(rawTags, s)
			}
		}
	}
	if len(rawTags) > 3 {
		rawTags = rawTags[:3]
	}
	tags := translateTags(rawTags, botTagLabels, lang)

	statusColor := "#6b7280"
	s := strings.ToLower(status)
	switch {
	case s == "streaming":
		statusColor = "#7c3aed"
	case s == "online" || online:
		statusColor = "#22c55e"
	case s == "idle":
		statusColor = "#eab308"
	case s == "dnd":
		statusColor = "#ef4444"
	}

	pageUrl := fmt.Sprintf("https://dishub.codecrew.kr/bots/%s", id)

	avatarHTML := fmt.Sprintf(`<div class="avatar">%s</div>`, html.EscapeString(string([]rune(name)[:1])))
	if avatarUrl != "" {
		avatarHTML = fmt.Sprintf(`<img class="avatar" src="%s" alt="%s">`, html.EscapeString(avatarUrl), html.EscapeString(name))
	}

	badgeHTML := ""
	if badge {
		badgeHTML = `<span class="badge">✓</span>`
	}

	descHTML := ""
	if desc != "" {
		descHTML = fmt.Sprintf(`<div class="desc">%s</div>`, html.EscapeString(desc))
	}

	tagsHTML := ""
	if len(tags) > 0 {
		var sb strings.Builder
		sb.WriteString(`<div class="tags">`)
		for _, t := range tags {
			sb.WriteString(fmt.Sprintf(`<span class="tag">#%s</span>`, html.EscapeString(t)))
		}
		sb.WriteString(`</div>`)
		tagsHTML = sb.String()
	}

	inviteHTML := ""
	if inviteUrl != "" {
		inviteHTML = fmt.Sprintf(`<a class="btn btn-fill" href="%s" target="_blank" rel="noreferrer">%s</a>`,
			html.EscapeString(inviteUrl), html.EscapeString(labels[1]))
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>%s
  .dot { background: %s; }
</style>
</head>
<body>
<div class="card">
  <div class="top">
    <div class="avatar-wrap">
      %s
      <div class="dot"></div>
    </div>
    <div class="info">
      <div class="name">%s %s</div>
      <div class="stats">
        <span class="stat">🖥 %d</span>
        <span class="stat">❤️ %d</span>
        <span class="stat">⭐ %.1f</span>
      </div>
    </div>
  </div>
  %s
  %s
  <div class="btns">
    <a class="btn btn-outline" href="%s" target="_blank" rel="noreferrer">%s</a>
    %s
  </div>
  <div class="footer">powered by DisHub</div>
</div>
</body>
</html>`,
		lang, widgetCSS, statusColor,
		avatarHTML,
		html.EscapeString(name), badgeHTML,
		serverCount, hearts, rating,
		descHTML, tagsHTML,
		html.EscapeString(pageUrl), html.EscapeString(labels[0]),
		inviteHTML,
	)

	widgetHTML(c, body)
}

func GetServerWidget(c *gin.Context) {
	id := c.Param("id")
	lang := c.DefaultQuery("lang", "ko")

	labels, ok := widgetServerI18n[lang]
	if !ok {
		labels = widgetServerI18n["ko"]
	}

	col := database.GetCollection("servers")
	var server bson.M
	err := col.FindOne(c.Request.Context(), bson.M{"serverId": id}).Decode(&server)
	if err != nil {
		c.String(http.StatusNotFound, "Not found")
		return
	}

	name, _ := server["name"].(string)
	icon, _ := server["icon"].(string)
	serverId, _ := server["serverId"].(string)
	inviteUrl, _ := server["inviteUrl"].(string)
	badge, _ := server["badge"].(bool)
	rating, _ := server["rating"].(float64)
	hearts, _ := server["hearts"].(int32)
	memberCount, _ := server["memberCount"].(int32)

	descRaw, _ := server["description"].(bson.M)
	descMap := make(map[string]string)
	for k, v := range descRaw {
		if s, ok := v.(string); ok {
			descMap[k] = s
		}
	}
	desc := pickDesc(descMap, lang)

	var rawTags []string
	if t, ok := server["tags"].(bson.A); ok {
		for _, v := range t {
			if s, ok := v.(string); ok {
				rawTags = append(rawTags, s)
			}
		}
	}
	if len(rawTags) > 3 {
		rawTags = rawTags[:3]
	}
	tags := translateTags(rawTags, serverTagLabels, lang)

	iconUrl := ""
	if icon != "" {
		if strings.HasPrefix(icon, "http") {
			iconUrl = icon
		} else {
			ext := "png"
			if strings.HasPrefix(icon, "a_") {
				ext = "gif"
			}
			iconUrl = fmt.Sprintf("https://cdn.discordapp.com/icons/%s/%s.%s?size=128", serverId, icon, ext)
		}
	}

	if inviteUrl != "" && !strings.HasPrefix(inviteUrl, "http") {
		inviteUrl = "https://discord.gg/" + inviteUrl
	}

	pageUrl := fmt.Sprintf("https://dishub.codecrew.kr/servers/%s", serverId)

	avatarHTML := fmt.Sprintf(`<div class="avatar">%s</div>`, html.EscapeString(string([]rune(name)[:1])))
	if iconUrl != "" {
		avatarHTML = fmt.Sprintf(`<img class="avatar" src="%s" alt="%s">`, html.EscapeString(iconUrl), html.EscapeString(name))
	}

	badgeHTML := ""
	if badge {
		badgeHTML = `<span class="badge">✓</span>`
	}

	descHTML := ""
	if desc != "" {
		descHTML = fmt.Sprintf(`<div class="desc">%s</div>`, html.EscapeString(desc))
	}

	tagsHTML := ""
	if len(tags) > 0 {
		var sb strings.Builder
		sb.WriteString(`<div class="tags">`)
		for _, t := range tags {
			sb.WriteString(fmt.Sprintf(`<span class="tag">#%s</span>`, html.EscapeString(t)))
		}
		sb.WriteString(`</div>`)
		tagsHTML = sb.String()
	}

	inviteHTML := ""
	if inviteUrl != "" {
		inviteHTML = fmt.Sprintf(`<a class="btn btn-fill" href="%s" target="_blank" rel="noreferrer">%s</a>`,
			html.EscapeString(inviteUrl), html.EscapeString(labels[1]))
	}

	body := fmt.Sprintf(`<!DOCTYPE html>
<html lang="%s">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<style>%s</style>
</head>
<body>
<div class="card">
  <div class="top">
    %s
    <div class="info">
      <div class="name">%s %s</div>
      <div class="stats">
        <span class="stat">👥 %d</span>
        <span class="stat">❤️ %d</span>
        <span class="stat">⭐ %.1f</span>
      </div>
    </div>
  </div>
  %s
  %s
  <div class="btns">
    <a class="btn btn-outline" href="%s" target="_blank" rel="noreferrer">%s</a>
    %s
  </div>
  <div class="footer">powered by DisHub</div>
</div>
</body>
</html>`,
		lang, widgetCSS,
		avatarHTML,
		html.EscapeString(name), badgeHTML,
		memberCount, hearts, rating,
		descHTML, tagsHTML,
		html.EscapeString(pageUrl), html.EscapeString(labels[0]),
		inviteHTML,
	)

	widgetHTML(c, body)
}
