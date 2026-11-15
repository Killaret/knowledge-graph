package web

// Stage A of URL-HEADING-1: deterministic page structure extraction.
// No model is involved — container detection, noise removal, outline and
// Markdown rendering, section budget and title candidates are all rules.
// See docs/tasks/URL-HEADING-1-heading-extraction.md for the spec and the
// probe evidence behind each rule.

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	importer "knowledge-graph/internal/application/import"

	"golang.org/x/net/html"
)

const defaultContentMaxRunes = 20000

// contentContainerClasses is the A1 fallback list: the probe (вывод 6) found
// Confluence, Keycloak docs and jsonformatter.org serve the body without
// <main>/<article>, so the id/class check is required.
// Note: `_` is a word character, so \b never fires inside BEM names like
// `lc-page__content` — separators here are space, _, -, ., / explicitly.
var contentContainerClasses = regexp.MustCompile(`(?i)(?:^|[\s_\-./])(content|page-content|post-content|entry-content|markdown-body|mw-parser-output|documentation|article)(?:[\s_\-./]|$)`)

// chromeIDClass marks elements whose class merely CONTAINS a container word:
// modal-content, innermenu-content, base-search-container. They look like
// content containers but are UI chrome — the fallback skips them.
var chromeIDClass = regexp.MustCompile(`(?i)(modal|dialog|popup|popover|menu|nav|search|cookie|banner|dropdown|drawer|tooltip|toast|overlay|flyout|cart|auth|login|signup|breadcrumb|pagination|filter|gallery|slider|carousel|player|subscribe|promo|advert|extra[-_]sale)`)

// noiseIDClass matches promotional/navigational blocks that live INSIDE
// <main>/<article> (probe, вывод 5: рекламные h2 у Практикума, Skillbox,
// Coursera).
var noiseIDClass = regexp.MustCompile(`(?i)(promo|subscribe|newsletter|related|sidebar|rail|popular|comments?|share|social|breadcrumb|cookie|banner|advert|menu|\bads?\b)`)

// noiseIDClassAllowed cancels a noise match on BEM-style modifiers that only
// DESCRIBE the element: Habr's `tm-page__main_has-sidebar` wraps the whole
// article — it is not a sidebar.
var noiseIDClassAllowed = regexp.MustCompile(`(?i)(has|with|without|no)[-_](sidebar|rail|promo|banner|comments?|ads?|share|social|popular|related|subscribe|newsletter|breadcrumb|cookie|advert)\b`)

// noiseTags are dropped wherever they appear inside the content container.
var noiseTags = map[string]bool{
	"nav": true, "aside": true, "footer": true, "form": true,
	"script": true, "style": true, "noscript": true, "template": true,
	"iframe": true, "svg": true, "button": true, "input": true,
	"select": true, "textarea": true,
}

var noiseRoles = map[string]bool{
	"navigation": true, "complementary": true, "banner": true, "contentinfo": true,
}

var headingTags = map[string]int{
	"h1": 1, "h2": 2, "h3": 3, "h4": 4, "h5": 5, "h6": 6,
}

const maxRelatedLinks = 20

// block is one Markdown-renderable unit inside a section. Cutting a section
// happens only at a block boundary — never inside a list or a code fence.
type block struct {
	text string
}

type section struct {
	heading *importer.OutlineEntry // nil for the lead section before the first heading
	blocks  []block
}

// extractPage runs the stage-A pipeline over a parsed document.
// maxRunes is the section budget for the rendered body (A4).
func extractPage(doc *html.Node, rawURL string, maxRunes int) *importer.ExtractedPage {
	if maxRunes <= 0 {
		maxRunes = defaultContentMaxRunes
	}
	out := &importer.ExtractedPage{TitleSource: "rule"}

	// A1 — content container.
	container, containerIsBody := findContentContainer(doc)

	// Question 3 — links inside noise areas are kept as metadata.related_links.
	out.RelatedLinks = collectRelatedLinks(doc, container)

	// A2 — remove noise subtrees inside the container.
	out.NoiseDropped = dropNoise(container, containerIsBody)

	// A3 — outline + Markdown sections.
	sections, outline := buildSections(container)
	out.Outline = outline

	// A4 — section budget.
	out.Text, out.SectionsDropped = renderWithBudget(sections, maxRunes)

	// A5 — title candidates and the stage-A selection rule.
	out.TitleCandidates, out.Title = chooseTitle(container, doc, rawURL)
	return out
}

// findContentContainer returns the first match of the A1 chain and whether the
// fallback ended on <body> (which flips the <header> rule in A2).
func findContentContainer(doc *html.Node) (*html.Node, bool) {
	body := findFirst(doc, "body")
	if body == nil {
		return doc, true
	}
	if n := findFirst(body, "main"); n != nil {
		return n, false
	}
	if n := findFirst(body, "article"); n != nil {
		return n, false
	}
	if n := findFirstMatch(body, func(n *html.Node) bool {
		return attrOf(n, "role") == "main"
	}); n != nil {
		return n, false
	}
	// Class fallback: several elements can carry a content-ish class
	// (modal-content, innermenu-content, search containers). Chrome-looking
	// candidates are skipped, and among the rest the one holding the most
	// text wins — first-match alone grabs a modal or a menu (probe pages
	// 08/18/19).
	var candidates []*html.Node
	walkAll(body, func(n *html.Node) {
		attrs := attrOf(n, "id") + " " + attrOf(n, "class")
		if !contentContainerClasses.MatchString(attrs) || chromeIDClass.MatchString(attrs) {
			return
		}
		candidates = append(candidates, n)
	})
	// When a signature is shared by 2+ siblings, their parent is the real
	// container: Practicum splits an article into sibling
	// <section id="post-content-text"> blocks.
	byParentSig := map[*html.Node]map[string]int{}
	for _, n := range candidates {
		sig := attrOf(n, "id") + "|" + attrOf(n, "class")
		if n.Parent == nil {
			continue
		}
		if byParentSig[n.Parent] == nil {
			byParentSig[n.Parent] = map[string]int{}
		}
		byParentSig[n.Parent][sig]++
	}
	bestText := 0
	var best *html.Node
	for parent, sigs := range byParentSig {
		for _, cnt := range sigs {
			if cnt >= 2 {
				if l := utf8.RuneCountInString(extractText(parent)); l > bestText {
					best, bestText = parent, l
				}
			}
		}
	}
	for _, n := range candidates {
		sig := attrOf(n, "id") + "|" + attrOf(n, "class")
		if n.Parent != nil && byParentSig[n.Parent][sig] >= 2 {
			continue // covered by the promoted parent
		}
		if l := utf8.RuneCountInString(extractText(n)); l > bestText {
			best, bestText = n, l
		}
	}
	if best != nil {
		// Guard: a widget-sized winner holding none of the document's
		// headings is not the article — Karpov's t978__content menu popup
		// won by text while #allrecords carried every h1–h3. Falling back
		// to <body> lets the noise pass strip the chrome instead of
		// cropping the extraction to it.
		if countHeadings(best) == 0 && countHeadings(body) >= 3 {
			return body, true
		}
		return best, false
	}
	return body, true
}

func countHeadings(n *html.Node) int {
	c := 0
	walkAll(n, func(x *html.Node) {
		if _, ok := headingTags[x.Data]; ok {
			c++
		}
	})
	return c
}

func walkAll(n *html.Node, f func(*html.Node)) {
	if n.Type == html.ElementNode {
		f(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walkAll(c, f)
	}
}

// dropNoise removes noise subtrees from the container in place and returns the
// number of dropped elements. A <header> counts as noise only when it is a
// direct child of the container AND the container is <body> — inside an
// article container the header carries the article h1 (probe, вывод 3:
// Wikipedia and Atlassian keep h1 in a header-like block).
func dropNoise(container *html.Node, containerIsBody bool) int {
	dropped := 0
	var next *html.Node
	for c := container.FirstChild; c != nil; c = next {
		next = c.NextSibling
		if c.Type == html.ElementNode && isNoiseElement(c, containerIsBody) {
			dropped += countElements(c)
			container.RemoveChild(c)
			continue
		}
		// Only the direct children of a <body> fallback get the <header>
		// rule — deeper levels keep article headers that carry the page h1.
		dropped += dropNoise(c, false)
	}
	return dropped
}

func isNoiseElement(n *html.Node, isDirectChildOfBodyContainer bool) bool {
	if n.Type != html.ElementNode {
		return false
	}
	if n.Data == "header" {
		return isDirectChildOfBodyContainer
	}
	if noiseTags[n.Data] ||
		noiseRoles[attrOf(n, "role")] ||
		attrOf(n, "aria-hidden") == "true" {
		return true
	}
	attrs := attrOf(n, "id") + " " + attrOf(n, "class")
	return noiseIDClass.MatchString(attrs) && !noiseIDClassAllowed.MatchString(attrs)
}

func countElements(n *html.Node) int {
	c := 0
	if n.Type == html.ElementNode {
		c = 1
	}
	for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
		c += countElements(ch)
	}
	return c
}

// collectRelatedLinks gathers (text, href) anchors from nav/aside/footer and
// page-level headers — areas that never reach the body text (question 3:
// keep, don't show). http(s) only, ≤ maxRelatedLinks.
func collectRelatedLinks(doc, container *html.Node) []importer.RelatedLink {
	var links []importer.RelatedLink
	var walk func(n *html.Node, inNoise bool)
	walk = func(n *html.Node, inNoise bool) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "nav", "aside", "footer":
				inNoise = true
			case "header":
				if !isDescendantOf(n, container) {
					inNoise = true
				}
			}
			if inNoise && n.Data == "a" {
				if href := attrOf(n, "href"); importer.IsAllowedURL(href) {
					if text := cleanText(extractText(n)); text != "" {
						links = append(links, importer.RelatedLink{Text: text, URL: href})
					}
				}
			}
		}
		for c := n.FirstChild; c != nil && len(links) < maxRelatedLinks; c = c.NextSibling {
			walk(c, inNoise)
		}
	}
	walk(doc, false)
	return links
}

func isDescendantOf(n, ancestor *html.Node) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if p == ancestor {
			return true
		}
	}
	return false
}

// buildSections walks the cleaned container and splits it into sections at
// every h1–h6. Outline levels are normalized: the first heading becomes level
// 2 (`##` — the `## [title](url)` first line is added later by BuildContent),
// the rest shift relative to it, clamped to [2, 6].
func buildSections(container *html.Node) ([]*section, []importer.OutlineEntry) {
	var sections []*section
	var outline []importer.OutlineEntry
	cur := &section{}
	firstLevel := 0

	pushBlock := func(text string) {
		if strings.TrimSpace(text) != "" {
			cur.blocks = append(cur.blocks, block{text: text})
		}
	}

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			pushBlock(n.Data)
			return
		}
		if n.Type != html.ElementNode {
			return
		}
		if lvl, ok := headingTags[n.Data]; ok {
			// Heading text is taken plainly: emphasis markers are redundant
			// inside a `###` line and corrupt outline entries (Skillbox wraps
			// h2s in <strong>).
			if text := headingText(n); text != "" {
				if firstLevel == 0 {
					firstLevel = lvl
				}
				norm := lvl - firstLevel + 2
				if norm < 2 {
					norm = 2
				}
				if norm > 6 {
					norm = 6
				}
				if len(cur.blocks) > 0 || cur.heading != nil {
					sections = append(sections, cur)
				}
				entry := importer.OutlineEntry{Level: norm, Text: text}
				cur = &section{heading: &entry}
				outline = append(outline, entry)
			}
			return
		}
		switch n.Data {
		case "p":
			pushBlock(cleanText(renderInline(n)))
		case "ul", "ol":
			pushBlock(renderList(n, 0))
		case "pre", "code":
			pushBlock(renderCodeBlock(n))
		case "table":
			pushBlock(renderTable(n))
		case "blockquote":
			pushBlock(renderBlockquote(n))
		case "img", "video", "figure", "picture", "hr", "br":
			// skipped entirely (A3)
		case "a":
			// Links keep only their text (A3): the source URL is already the
			// note's first line.
			pushBlock(cleanText(renderInline(n)))
		default:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}
	walk(container)
	if len(cur.blocks) > 0 || cur.heading != nil {
		sections = append(sections, cur)
	}
	return sections, outline
}

// headingText returns the heading's text on a single line with collapsed
// whitespace — no Markdown emphasis (see buildSections).
func headingText(n *html.Node) string {
	return strings.Join(strings.Fields(extractText(n)), " ")
}

// renderInline renders a node's inline content preserving the typographic
// marks the owner asked for (strong/em/code), with links flattened to text.
func renderInline(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		if x.Type == html.TextNode {
			b.WriteString(x.Data)
			return
		}
		if x.Type != html.ElementNode {
			return
		}
		switch x.Data {
		case "strong", "b":
			b.WriteString("**")
			walkChildren(x, &walk)
			b.WriteString("**")
		case "em", "i":
			b.WriteString("*")
			walkChildren(x, &walk)
			b.WriteString("*")
		case "code":
			b.WriteString("`")
			b.WriteString(extractText(x))
			b.WriteString("`")
		case "script", "style", "noscript", "template", "img", "svg":
			// dropped
		default:
			walkChildren(x, &walk)
		}
	}
	walk(n)
	return b.String()
}

func walkChildren(n *html.Node, walk *func(*html.Node)) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		(*walk)(c)
	}
}

// renderList renders ul/ol with nesting as indented Markdown list lines.
func renderList(n *html.Node, depth int) string {
	var lines []string
	idx := 0
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type != html.ElementNode {
			continue
		}
		switch c.Data {
		case "ul", "ol":
			lines = append(lines, renderList(c, depth+1))
		case "li":
			idx++
			marker := "- "
			if n.Data == "ol" {
				marker = strconv.Itoa(idx) + ". "
			}
			var inline strings.Builder
			for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
				if cc.Type == html.ElementNode && (cc.Data == "ul" || cc.Data == "ol") {
					continue
				}
				inline.WriteString(renderInline(cc))
			}
			if text := strings.Join(strings.Fields(inline.String()), " "); text != "" {
				lines = append(lines, strings.Repeat("  ", depth)+marker+text)
			}
			for cc := c.FirstChild; cc != nil; cc = cc.NextSibling {
				if cc.Type == html.ElementNode && (cc.Data == "ul" || cc.Data == "ol") {
					lines = append(lines, renderList(cc, depth+1))
				}
			}
		}
	}
	return strings.Join(nonEmpty(lines), "\n")
}

// renderCodeBlock renders pre/code as a fenced block; fence parity is a
// mutation-guarded invariant (criterion 2).
func renderCodeBlock(n *html.Node) string {
	code := strings.Trim(extractText(n), "\n")
	if strings.TrimSpace(code) == "" {
		return ""
	}
	return "```\n" + code + "\n```"
}

// renderTable renders a ≤6-column table as Markdown rows; wider tables fall
// back to one paragraph per row (A3).
func renderTable(n *html.Node) string {
	var rows [][]string
	var walk func(*html.Node)
	walk = func(x *html.Node) {
		if x.Type == html.ElementNode && x.Data == "tr" {
			var cells []string
			for c := x.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
					cells = append(cells, strings.Join(strings.Fields(cleanText(renderInline(c))), " "))
				}
			}
			if len(cells) > 0 {
				rows = append(rows, cells)
			}
			return
		}
		for c := x.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	if len(rows) == 0 {
		return ""
	}
	width := 0
	for _, r := range rows {
		if len(r) > width {
			width = len(r)
		}
	}
	if width > 6 {
		var paras []string
		for _, r := range rows {
			paras = append(paras, strings.Join(nonEmpty(r), " "))
		}
		return strings.Join(paras, "\n\n")
	}
	var lines []string
	for i, r := range rows {
		for len(r) < width {
			r = append(r, "")
		}
		lines = append(lines, "| "+strings.Join(r, " | ")+" |")
		if i == 0 {
			lines = append(lines, "|"+strings.Repeat(" --- |", width))
		}
	}
	return strings.Join(lines, "\n")
}

func renderBlockquote(n *html.Node) string {
	text := cleanText(renderInline(n))
	if text == "" {
		return ""
	}
	var lines []string
	for _, l := range strings.Split(text, "\n") {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, "> "+l)
		}
	}
	return strings.Join(lines, "\n")
}

// renderWithBudget renders sections under a rune budget (A4): a section is
// added whole when it fits; when it does not, the cut lands on a block
// boundary — never inside a list or a code fence — and the section together
// with every following one counts as dropped. The trailer line reports the
// count and feeds metadata.import_truncated.sections_dropped.
func renderWithBudget(sections []*section, maxRunes int) (string, int) {
	var parts []string
	used := 0
	dropped := 0
	stop := false

	emit := func(text string) {
		parts = append(parts, text)
		used += utf8.RuneCountInString(text) + 2 // blank-line separator
	}

	for _, s := range sections {
		if stop {
			dropped++
			continue
		}
		head := ""
		if s.heading != nil {
			head = strings.Repeat("#", s.heading.Level) + " " + s.heading.Text
		}
		// Whole-section fit check.
		size := 0
		if head != "" {
			size += utf8.RuneCountInString(head) + 2
		}
		for _, b := range s.blocks {
			size += utf8.RuneCountInString(b.text) + 2
		}
		if used+size <= maxRunes {
			if head != "" {
				emit(head)
			}
			for _, b := range s.blocks {
				emit(b.text)
			}
			continue
		}
		// Partial: heading first, then complete blocks while they fit.
		if head != "" && used+utf8.RuneCountInString(head)+2 <= maxRunes {
			emit(head)
		}
		for _, b := range s.blocks {
			if used+utf8.RuneCountInString(b.text)+2 > maxRunes {
				break
			}
			emit(b.text)
		}
		dropped++
		stop = true
	}
	if dropped > 0 {
		parts = append(parts, "_(обрезано: "+strconv.Itoa(dropped)+" разделов не вошли)_")
	}
	return cleanMarkdown(strings.Join(parts, "\n\n")), dropped
}

// chooseTitle implements A5: candidates in order — every h1 inside the
// cleaned container, og:title, <title> minus the site suffix, last URL path
// segment — then the first candidate of 3–200 runes that is not the site
// name wins.
func chooseTitle(container, doc *html.Node, rawURL string) ([]string, string) {
	// "Имя сайта" in the A5 rule means a short site identity — og:site_name
	// and the domain label. Two real cases forced the length guard and the
	// exclusion of application-name: Practicum mislabels og:site_name with
	// the full article title (~95 runes), and Keycloak's docs-api sets
	// application-name to the document title — strict equality rejection
	// would fall through to the URL segment ("index"). A genuine site name
	// ("jwt.io", "MDN", "Хабр") is short, so long values are not treated as
	// site names. application-name is used only for suffix stripping.
	rawOgSiteName := metaContent(doc, "og:site_name")
	ogSiteName := rawOgSiteName
	if utf8.RuneCountInString(ogSiteName) > 50 {
		ogSiteName = ""
	}
	appName := metaContent(doc, "application-name")
	suffixSiteNames := []string{rawOgSiteName, appName}
	domainLabel := ""
	host := ""
	if u, err := url.Parse(rawURL); err == nil {
		host = strings.TrimPrefix(strings.ToLower(u.Hostname()), "www.")
		if i := strings.Index(host, "."); i > 0 {
			domainLabel = host[:i]
		} else {
			domainLabel = host
		}
	}

	var candidates []string
	add := func(s string) {
		s = strings.Join(strings.Fields(s), " ") // single line, collapsed
		if s == "" {
			return
		}
		for _, c := range candidates {
			if c == s {
				return
			}
		}
		if len(candidates) < 5 {
			candidates = append(candidates, s)
		}
	}

	// 1. All h1 inside the container, in document order.
	var walkH1 func(n *html.Node)
	walkH1 = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "h1" {
			add(headingText(n))
		}
		walkChildren(n, &walkH1)
	}
	walkH1(container)

	// 2. og:title.
	add(metaContent(doc, "og:title"))

	// 3. <title> minus the site suffix.
	add(stripSiteSuffix(extractTitle(doc), suffixSiteNames, domainLabel, host))

	// 4. Last URL path segment, cleaned.
	if u, err := url.Parse(rawURL); err == nil {
		seg := lastPathSegment(u.Path)
		seg = strings.ReplaceAll(seg, "-", " ")
		seg = strings.ReplaceAll(seg, "_", " ")
		add(seg)
	}

	isSiteName := func(s string) bool {
		ls := strings.ToLower(strings.TrimSpace(s))
		return (ogSiteName != "" && ls == strings.ToLower(ogSiteName)) ||
			(domainLabel != "" && ls == domainLabel)
	}
	for _, c := range candidates {
		if l := utf8.RuneCountInString(c); l >= 3 && l <= 200 && !isSiteName(c) {
			return candidates, c
		}
	}
	return candidates, rawURL
}

// stripSiteSuffix removes trailing site-name segment(s) from a <title> like
// "JSON Web Tokens - jwt.io" (probe, вывод 1: suffix on almost every page).
// Separators per spec: " — ", " – ", " - ", " | ", " · ", " : ".
func stripSiteSuffix(title string, siteNames []string, domainLabel string, host string) string {
	seps := []string{" — ", " – ", " - ", " | ", " · ", " : "}
	for {
		best := -1
		var bestSep string
		for _, sep := range seps {
			if i := strings.LastIndex(title, sep); i > best {
				best = i
				bestSep = sep
			}
		}
		if best < 0 {
			return title
		}
		last := strings.ToLower(strings.TrimSpace(title[best+len(bestSep):]))
		match := (domainLabel != "" && last == domainLabel) ||
			(host != "" && last == host) // " - jwt.io" style: host with TLD
		for _, sn := range siteNames {
			if sn != "" && last == strings.ToLower(sn) {
				match = true
			}
		}
		if match {
			title = strings.TrimSpace(title[:best])
			continue
		}
		return title
	}
}

func lastPathSegment(path string) string {
	path = strings.TrimRight(path, "/")
	if i := strings.LastIndex(path, "/"); i >= 0 {
		path = path[i+1:]
	}
	if i := strings.LastIndex(path, "."); i > 0 {
		path = path[:i] // drop .html / .php etc.
	}
	if dec, err := url.PathUnescape(path); err == nil {
		path = dec
	}
	return path
}

func metaContent(doc *html.Node, property string) string {
	var found string
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		if found != "" {
			return
		}
		if n.Type == html.ElementNode && n.Data == "meta" &&
			(attrOf(n, "property") == property || attrOf(n, "name") == property) {
			found = attrOf(n, "content")
		}
		walkChildren(n, &walk)
	}
	walk(doc)
	return strings.TrimSpace(found)
}

func findFirst(n *html.Node, tag string) *html.Node {
	return findFirstMatch(n, func(x *html.Node) bool {
		return x.Type == html.ElementNode && x.Data == tag
	})
}

func findFirstMatch(n *html.Node, pred func(*html.Node) bool) *html.Node {
	if n.Type == html.ElementNode && pred(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if r := findFirstMatch(c, pred); r != nil {
			return r
		}
	}
	return nil
}

func attrOf(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// cleanMarkdown collapses >1 consecutive blank lines and trims outer space,
// preserving list indentation and fences.
func cleanMarkdown(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	blank := false
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			if blank {
				continue
			}
			blank = true
			out = append(out, "")
			continue
		}
		blank = false
		out = append(out, strings.TrimRight(l, " \t"))
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func nonEmpty(in []string) []string {
	var out []string
	for _, s := range in {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}
