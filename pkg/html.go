package snap

import (
	"fmt"
)

func HtmlParser(s string) string {
	p := NewParser(s)

	results := p.ParseLines()
	return results
}

func heading(body string, lvl int) string {
	return fmt.Sprintf("<h%d>%s</h%d>", lvl, body, lvl)
}

func bold(body string) string {
	return fmt.Sprintf("<b>%s</b>", body)
}

func italic(body string) string {
	return fmt.Sprintf("<em>%s</em>", body)
}

func italicBold(body string) string {
	return fmt.Sprintf("<em><b>%s</b></em>", body)
}

func paragraph(body string) string {
	return fmt.Sprintf("<p>%s</p>", body)
}
