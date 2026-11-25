//go:build gui

// Projects the parser's normalized CommonMark and GFM elements onto the HUD.
// Generated markup is read locally. Links and image alternatives require no network access.
package gui

import (
	"bytes"
	"image/color"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var chatMarkdown = goldmark.New(goldmark.WithExtensions(extension.GFM))
var mdBorder = color.RGBA{113, 132, 157, 255}

func layoutMarkdown(source string, width, font int) *mdLayout {
	m := &mdLayout{right: float64(width), base: float64(max(1, font))}
	s := mdStyle{size: m.base}
	var buf bytes.Buffer
	err := chatMarkdown.Convert([]byte(source), &buf)
	if err != nil {
		m.prose(source, s)
		return m
	}
	roots, err := html.ParseFragment(&buf, &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		m.prose(source, s)
		return m
	}
	for _, n := range roots {
		m.node(n, s)
	}
	return m
}
func mdAttr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}
func (m *mdLayout) children(n *html.Node, s mdStyle) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		m.node(c, s)
	}
}
func (m *mdLayout) node(n *html.Node, s mdStyle) {
	if n.Type == html.TextNode {
		m.prose(n.Data, s)
		return
	}
	if n.Type != html.ElementNode {
		return
	}
	switch n.Data {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level := int(n.Data[1] - '0')
		s.size *= 1 + float64(7-level)*.13
		s.bold = true
		fallthrough
	case "p":
		m.finish()
		m.children(n, s)
		m.finish()
		m.y += m.base * .35
	case "strong", "em", "del", "a", "code":
		s.bold = s.bold || n.Data == "strong"
		s.italic = s.italic || n.Data == "em"
		s.strike = s.strike || n.Data == "del"
		s.link = s.link || n.Data == "a"
		s.mono = s.mono || n.Data == "code"
		if n.Data == "code" && n.FirstChild != nil {
			m.word(n.FirstChild.Data, s)
			return
		}
		m.children(n, s)
	case "br":
		m.newline()
	case "hr":
		m.finish()
		m.rectangle(m.left, m.y+3, m.right-m.left, 1, mdBorder)
		m.y += 10
	case "blockquote":
		m.finish()
		start, left := m.y, m.left
		m.left = min(m.left+12, max(left, m.right-32))
		m.x = m.left
		m.children(n, s)
		m.finish()
		m.rectangle(left, start, 2, m.y-start, mdBorder)
		m.left = left
		m.x = left
	case "ul", "ol":
		m.list(n, s)
	case "pre":
		m.finish()
		s.mono = true
		if n.FirstChild != nil && n.FirstChild.FirstChild != nil {
			value := strings.TrimSuffix(n.FirstChild.FirstChild.Data, "\n")
			for _, line := range strings.Split(value, "\n") {
				m.word(strings.ReplaceAll(line, "\t", "    "), s)
				m.newline()
			}
		}
		m.y += 4
	case "table":
		m.table(n, s)
	case "img":
		m.prose(mdAttr(n, "alt"), s)
	case "input":
		checked := false
		for _, a := range n.Attr {
			if a.Key == "checked" {
				checked = true
			}
		}
		box := "☐"
		if checked {
			box = "☑"
		}
		m.word(box, s)
	default:
		m.children(n, s)
	}
}
func (m *mdLayout) list(n *html.Node, s mdStyle) {
	m.finish()
	left := m.left
	index, err := strconv.Atoi(mdAttr(n, "start"))
	if err != nil {
		index = 1
	}
	for li := n.FirstChild; li != nil; li = li.NextSibling {
		if li.Data != "li" || li.Type != html.ElementNode {
			continue
		}
		marker := "•"
		if n.Data == "ol" {
			marker = strconv.Itoa(index) + "."
			index++
		}
		m.word(marker, s)
		m.left = min(m.x+5, max(left, m.right-32))
		m.x = m.left
		first := true
		for c := li.FirstChild; c != nil; c = c.NextSibling {
			if first && c.Type == html.TextNode && strings.TrimSpace(c.Data) == "" {
				continue
			}
			if c.Type == html.ElementNode && c.Data == "p" && first {
				m.children(c, s)
				m.finish()
			} else {
				m.node(c, s)
			}
			first = false
		}
		m.finish()
		m.left = left
		m.x = left
	}
}
func (m *mdLayout) table(n *html.Node, s mdStyle) {
	m.finish()
	left, right := m.left, m.right
	var headers []*html.Node
	for section := n.FirstChild; section != nil; section = section.NextSibling {
		for row := section.FirstChild; row != nil; row = row.NextSibling {
			if row.Data != "tr" {
				continue
			}
			var cells []*html.Node
			for cell := row.FirstChild; cell != nil; cell = cell.NextSibling {
				if cell.Data == "th" || cell.Data == "td" {
					cells = append(cells, cell)
				}
			}
			if len(cells) == 0 {
				continue
			}
			if section.Data == "thead" {
				headers = cells
			}
			top := m.y
			rowTop := top
			bottom := top + s.size*1.35 + 8
			width := (right - left) / float64(len(cells))
			stacked := width < max(34, s.size*2.5)
			if stacked {
				width = right - left
			}
			for i, cell := range cells {
				start := left + float64(i)*width
				if stacked {
					start = left
					top = m.y
				}
				m.left = start + 3
				m.right = start + width - 3
				m.x = m.left
				m.y = top + 4
				style := s
				style.bold = style.bold || cell.Data == "th"
				begin := len(m.runs)
				if stacked && cell.Data == "td" && i < len(headers) {
					heading := s
					heading.bold = true
					m.children(headers[i], heading)
					m.finish()
				}
				m.children(cell, style)
				m.finish()
				align := mdAttr(cell, "align") + mdAttr(cell, "style")
				if !stacked {
					m.alignCell(begin, align)
				}
				bottom = max(bottom, m.y+4)
				if stacked {
					m.rectangle(left, top, width, 1, mdBorder)
					m.y = bottom
				}
			}
			m.rectangle(left, rowTop, right-left, 1, mdBorder)
			if !stacked {
				for i := 0; i <= len(cells); i++ {
					m.rectangle(min(right-1, left+float64(i)*width), top, 1, bottom-top, mdBorder)
				}
			}
			m.rectangle(left, bottom-1, right-left, 1, mdBorder)
			m.left = left
			m.right = right
			m.x = left
			m.y = bottom
		}
	}
	m.y += 4
}
func (m *mdLayout) alignCell(begin int, align string) {
	if !strings.Contains(align, "right") && !strings.Contains(align, "center") {
		return
	}
	for begin < len(m.runs) {
		end := begin + 1
		for end < len(m.runs) && m.runs[end].y == m.runs[begin].y {
			end++
		}
		last := m.runs[end-1]
		offset := max(0, m.right-last.x-last.w)
		if strings.Contains(align, "center") {
			offset /= 2
		}
		for i := begin; i < end; i++ {
			m.runs[i].x += offset
		}
		begin = end
	}
}
