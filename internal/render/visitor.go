package render

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"

	"github.com/benelog/md-lens/internal/ansi"
	"github.com/benelog/md-lens/internal/heading"
	"github.com/benelog/md-lens/internal/highlight"
	"github.com/benelog/md-lens/internal/image"
)

// visitor walks the goldmark AST and writes rich terminal output. Block nodes are visited; inline
// content is rendered by inlineChildren into styled strings that are then wrapped and prefixed.
type visitor struct {
	ctx         *Context
	ansi        *ansi.Ansi
	theme       *highlight.Theme
	highlighter *highlight.Highlighter
	images      *image.Renderer
	headings    *heading.Renderer
	baseDir     string
	source      []byte

	lists      []*listState
	tightDepth int
}

type listState struct {
	ordered   bool
	delimiter string
	number    int
}

func (v *visitor) renderDocument(doc ast.Node) {
	v.visitChildren(doc)
}

func (v *visitor) visitChildren(n ast.Node) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		v.renderNode(c)
	}
}

func (v *visitor) renderNode(n ast.Node) {
	switch node := n.(type) {
	case *ast.Heading:
		v.visitHeading(node)
	case *ast.Paragraph:
		v.visitParagraph(node)
	case *ast.TextBlock:
		// Tight-list item content; rendered like a paragraph (blank suppression via tightDepth).
		v.visitParagraph(node)
	case *ast.FencedCodeBlock:
		v.renderCode(v.codeText(node), v.fencedLang(node))
	case *ast.CodeBlock:
		v.renderCode(v.codeText(node), "")
	case *ast.Blockquote:
		v.visitBlockquote(node)
	case *ast.ThematicBreak:
		v.visitThematicBreak(node)
	case *ast.List:
		v.visitList(node)
	case *ast.ListItem:
		v.visitListItem(node)
	case *east.Table:
		v.renderTable(node)
	case *ast.HTMLBlock:
		// raw HTML block — skipped in v1
	default:
		v.visitChildren(n)
	}
}

// --- block nodes ---------------------------------------------------------

func (v *visitor) visitHeading(h *ast.Heading) {
	v.headings.Render(h.Level, v.collectText(h), v.ctx.out)
	v.ctx.blank()
}

func (v *visitor) visitParagraph(n ast.Node) {
	if img := soleImage(n); img != nil {
		v.renderImageBlock(img)
		if v.tightDepth == 0 {
			v.ctx.blank()
		}
		return
	}
	var sb strings.Builder
	v.inlineChildren(n, &sb)
	for _, line := range wrapText(sb.String(), v.ctx.contentWidth()) {
		v.ctx.line(line)
	}
	if v.tightDepth == 0 {
		v.ctx.blank()
	}
}

func (v *visitor) visitBlockquote(n *ast.Blockquote) {
	bar := v.ansi.Fg3(v.theme.QuoteBar) + "▎" + v.ansi.Reset() + " "
	v.ctx.pushPrefix(bar)
	v.visitChildren(n)
	v.ctx.popPrefix()
}

func (v *visitor) visitThematicBreak(_ *ast.ThematicBreak) {
	v.ctx.line(v.ansi.Fg3(v.theme.Rule) + strings.Repeat("─", v.ctx.contentWidth()) + v.ansi.Reset())
	v.ctx.blank()
}

func (v *visitor) visitList(list *ast.List) {
	var state *listState
	if list.IsOrdered() {
		state = &listState{ordered: true, number: list.Start, delimiter: string(rune(list.Marker))}
	} else {
		state = &listState{ordered: false, number: 0, delimiter: ""}
	}
	v.enterList(state, list.IsTight)
	v.visitChildren(list)
	v.exitList(list.IsTight)
}

func (v *visitor) enterList(state *listState, tight bool) {
	v.lists = append(v.lists, state)
	if tight {
		v.tightDepth++
	}
}

func (v *visitor) exitList(tight bool) {
	v.lists = v.lists[:len(v.lists)-1]
	if tight {
		v.tightDepth--
	}
	if len(v.lists) == 0 {
		v.ctx.blank()
	}
}

func (v *visitor) visitListItem(item *ast.ListItem) {
	var state *listState
	if len(v.lists) > 0 {
		state = v.lists[len(v.lists)-1]
	}

	var markerText, styledMarker string
	switch {
	case taskCheckbox(item) != nil:
		// GFM task list: show a checkbox instead of a bullet.
		cb := taskCheckbox(item)
		if cb.IsChecked {
			markerText = "☑ "
			styledMarker = v.ansi.Fg3(v.theme.TaskDone) + markerText + v.ansi.Reset()
		} else {
			markerText = "☐ "
			styledMarker = markerText
		}
	case state != nil && state.ordered:
		markerText = strconv.Itoa(state.number) + state.delimiter + " "
		state.number++
		styledMarker = v.ansi.Fg3(v.theme.Marker) + markerText + v.ansi.Reset()
	default:
		markerText = bulletGlyph(len(v.lists)-1) + " "
		styledMarker = v.ansi.Fg3(v.theme.Marker) + markerText + v.ansi.Reset()
	}

	v.ctx.pushPrefix(strings.Repeat(" ", ansi.Width(markerText)))
	v.ctx.setPendingMarker(styledMarker)
	v.visitChildren(item)
	v.ctx.popPrefix()
}

// taskCheckbox returns the task-list checkbox of an item, if present. The tasklist extension
// inserts it as the first child of the item's first block (paragraph/text block).
func taskCheckbox(item *ast.ListItem) *east.TaskCheckBox {
	first := item.FirstChild()
	if first == nil {
		return nil
	}
	if cb, ok := first.FirstChild().(*east.TaskCheckBox); ok {
		return cb
	}
	return nil
}

// --- code blocks ---------------------------------------------------------

func (v *visitor) renderCode(literal, info string) {
	code := strings.TrimSuffix(literal, "\n")
	lines := v.styledCodeLines(v.highlighter.Tokenize(code, info))
	gutter := v.ansi.Fg3(v.theme.Dim) + "▏" + v.ansi.Reset() + " "
	v.ctx.pushPrefix(gutter)
	for _, line := range lines {
		v.ctx.line(line)
	}
	v.ctx.popPrefix()
	if v.tightDepth == 0 {
		v.ctx.blank()
	}
}

// styledCodeLines splits the token stream into physical lines, each independently styled.
func (v *visitor) styledCodeLines(tokens []highlight.Token) []string {
	out := []string{""}
	for _, tok := range tokens {
		parts := strings.Split(tok.Text, "\n")
		for i, part := range parts {
			if i > 0 {
				out = append(out, "")
			}
			out[len(out)-1] += v.highlighter.StylePiece(tok.Type, part)
		}
	}
	return out
}

func (v *visitor) codeText(n ast.Node) string {
	var sb strings.Builder
	lines := n.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		sb.Write(seg.Value(v.source))
	}
	return sb.String()
}

func (v *visitor) fencedLang(n *ast.FencedCodeBlock) string {
	if lang := n.Language(v.source); lang != nil {
		return string(lang)
	}
	return ""
}

// --- tables --------------------------------------------------------------

func (v *visitor) renderTable(table *east.Table) {
	var rows [][]string
	var isHeader []bool
	var aligns []east.Alignment
	columns := 0

	for rowNode := table.FirstChild(); rowNode != nil; rowNode = rowNode.NextSibling() {
		_, header := rowNode.(*east.TableHeader)
		var cells []string
		col := 0
		for cellNode := rowNode.FirstChild(); cellNode != nil; cellNode = cellNode.NextSibling() {
			cell, ok := cellNode.(*east.TableCell)
			if !ok {
				continue
			}
			var sb strings.Builder
			if header {
				sb.WriteString(v.ansi.Bold())
			}
			v.inlineChildren(cell, &sb)
			if header {
				sb.WriteString(v.ansi.BoldOff())
			}
			cells = append(cells, sb.String())
			if len(aligns) <= col {
				aligns = append(aligns, cell.Alignment)
			}
			col++
		}
		columns = max(columns, len(cells))
		rows = append(rows, cells)
		isHeader = append(isHeader, header)
	}
	if columns == 0 {
		return
	}

	widths := make([]int, columns)
	for _, row := range rows {
		for c, cell := range row {
			widths[c] = max(widths[c], ansi.Width(cell))
		}
	}

	sep := v.ansi.Fg3(v.theme.Rule) + "│" + v.ansi.Reset()
	for r := range rows {
		row := rows[r]
		var line strings.Builder
		for c := 0; c < columns; c++ {
			if c > 0 {
				line.WriteString(sep)
			}
			cell := ""
			if c < len(row) {
				cell = row[c]
			}
			a := east.AlignNone
			if c < len(aligns) {
				a = aligns[c]
			}
			line.WriteByte(' ')
			line.WriteString(pad(cell, widths[c], a))
			line.WriteByte(' ')
		}
		v.ctx.line(line.String())

		if isHeader[r] && (r+1 >= len(rows) || !isHeader[r+1]) {
			var div strings.Builder
			div.WriteString(v.ansi.Fg3(v.theme.Rule))
			for c := 0; c < columns; c++ {
				if c > 0 {
					div.WriteString("┼")
				}
				div.WriteString(strings.Repeat("─", widths[c]+2))
			}
			div.WriteString(v.ansi.Reset())
			v.ctx.line(div.String())
		}
	}
	if v.tightDepth == 0 {
		v.ctx.blank()
	}
}

func pad(content string, width int, align east.Alignment) string {
	gap := width - ansi.Width(content)
	if gap <= 0 {
		return content
	}
	switch align {
	case east.AlignRight:
		return strings.Repeat(" ", gap) + content
	case east.AlignCenter:
		return strings.Repeat(" ", gap/2) + content + strings.Repeat(" ", gap-gap/2)
	default: // AlignLeft, AlignNone
		return content + strings.Repeat(" ", gap)
	}
}

// --- images --------------------------------------------------------------

func (v *visitor) renderImageBlock(img *ast.Image) {
	alt := v.collectText(img)
	local := v.resolveLocal(string(img.Destination))
	indent := v.ctx.indentWidth()
	if local == "" {
		label := alt
		if strings.TrimSpace(label) == "" {
			label = string(img.Destination)
		}
		v.ctx.line(v.ansi.Dim() + "[image: " + label + "]" + v.ansi.Reset())
	} else {
		v.images.Render(local, alt, indent, v.ctx.out)
	}
}

func (v *visitor) resolveLocal(dest string) string {
	if strings.TrimSpace(dest) == "" || strings.Contains(dest, "://") {
		return ""
	}
	if filepath.IsAbs(dest) {
		return filepath.Clean(dest)
	}
	return filepath.Clean(filepath.Join(v.baseDir, dest))
}

// --- inline rendering ----------------------------------------------------

func (v *visitor) inlineChildren(parent ast.Node, sb *strings.Builder) {
	for c := parent.FirstChild(); c != nil; c = c.NextSibling() {
		v.inlineNode(c, sb)
	}
}

func (v *visitor) inlineNode(n ast.Node, sb *strings.Builder) {
	switch node := n.(type) {
	case *ast.Text:
		sb.Write(node.Segment.Value(v.source))
		switch {
		case node.SoftLineBreak():
			sb.WriteByte(' ')
		case node.HardLineBreak():
			sb.WriteByte('\n')
		}
	case *ast.String:
		sb.Write(node.Value)
	case *ast.Emphasis:
		if node.Level >= 2 {
			sb.WriteString(v.ansi.Bold())
			v.inlineChildren(node, sb)
			sb.WriteString(v.ansi.BoldOff())
		} else {
			sb.WriteString(v.ansi.Italic())
			v.inlineChildren(node, sb)
			sb.WriteString(v.ansi.ItalicOff())
		}
	case *east.Strikethrough:
		sb.WriteString(v.ansi.Strike())
		v.inlineChildren(node, sb)
		sb.WriteString(v.ansi.StrikeOff())
	case *ast.CodeSpan:
		sb.WriteString(v.ansi.Fg3(v.theme.CodeFg))
		sb.WriteString(v.ansi.Bg3(v.theme.CodeBg))
		sb.WriteByte(' ')
		sb.WriteString(v.codeSpanText(node))
		sb.WriteByte(' ')
		sb.WriteString(v.ansi.FgDefault())
		sb.WriteString(v.ansi.BgDefault())
	case *ast.Link:
		v.appendLink(node, sb)
	case *ast.AutoLink:
		v.appendAutoLink(node, sb)
	case *ast.Image:
		alt := v.collectText(node)
		if strings.TrimSpace(alt) == "" {
			alt = "image"
		}
		sb.WriteString(v.ansi.Dim() + "[" + alt + "]" + v.ansi.Reset())
	case *east.TaskCheckBox:
		// Rendered as the list item's marker; ignore it in the inline flow.
	case *ast.RawHTML:
		// raw HTML inline — skipped in v1
	default:
		v.inlineChildren(n, sb)
	}
}

func (v *visitor) appendLink(link *ast.Link, sb *strings.Builder) {
	var label strings.Builder
	v.inlineChildren(link, &label)
	dest := string(link.Destination)
	if v.ctx.hyperlinks {
		styled := v.ansi.Fg3(v.theme.Link) + v.ansi.Underline() + label.String() +
			v.ansi.UnderlineOff() + v.ansi.FgDefault()
		sb.WriteString(ansi.Osc8Link(dest, styled))
	} else {
		sb.WriteString(label.String())
		sb.WriteString(" (")
		sb.WriteString(dest)
		sb.WriteString(")")
	}
}

func (v *visitor) appendAutoLink(n *ast.AutoLink, sb *strings.Builder) {
	url := string(n.URL(v.source))
	label := string(n.Label(v.source))
	if label == "" {
		label = url
	}
	if v.ctx.hyperlinks {
		styled := v.ansi.Fg3(v.theme.Link) + v.ansi.Underline() + label +
			v.ansi.UnderlineOff() + v.ansi.FgDefault()
		sb.WriteString(ansi.Osc8Link(url, styled))
	} else {
		sb.WriteString(label)
		sb.WriteString(" (")
		sb.WriteString(url)
		sb.WriteString(")")
	}
}

func (v *visitor) codeSpanText(n *ast.CodeSpan) string {
	var sb strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch t := c.(type) {
		case *ast.Text:
			sb.Write(t.Segment.Value(v.source))
		case *ast.String:
			sb.Write(t.Value)
		}
	}
	return sb.String()
}

func (v *visitor) collectText(n ast.Node) string {
	var sb strings.Builder
	v.collectTextInto(n, &sb)
	return sb.String()
}

func (v *visitor) collectTextInto(n ast.Node, sb *strings.Builder) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch t := c.(type) {
		case *ast.Text:
			sb.Write(t.Segment.Value(v.source))
		case *ast.String:
			sb.Write(t.Value)
		case *ast.CodeSpan:
			sb.WriteString(v.codeSpanText(t))
		default:
			v.collectTextInto(c, sb)
		}
	}
}

func soleImage(n ast.Node) *ast.Image {
	first := n.FirstChild()
	if first == nil {
		return nil
	}
	if img, ok := first.(*ast.Image); ok && first.NextSibling() == nil {
		return img
	}
	return nil
}
