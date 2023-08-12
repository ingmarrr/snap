package snap

import (
	"strings"
	"testing"
)

func TestMap(t *testing.T) {
	s := `# Hello World
## Hello World
Hello There ... *General Kenobi*.. what a pleasant surprise **to see you here**
#### Hello World
Hello there * General Kenobi
  `
	p := NewParser(s)
	p.MapLinePrefix("#", func(cx Cx) string {
		return Heading{Body: cx.Body, level: 1}.String()
	})
	p.MapLinePrefix("##", func(cx Cx) string {
		return Heading{Body: cx.Body, level: 2}.String()
	})
	p.MapLinePrefix("####", func(cx Cx) string {
		return Heading{Body: cx.Body, level: 4}.String()
	})
	p.MapCapture("*", func(cx Cx) string {
		return Italic{Body: cx.Body}.String()
	})
	p.MapCapture("**", func(cx Cx) string {
		return Bold{Body: cx.Body}.String()
	})
	results := p.ParseLines()

	t.Log(results)
	t.Log(strings.Split(p.text, "\n"))
}

func testParseImpl(t *testing.T) {
	s := `**Hello World** 
##Hello World
# Hello *there* what is up
@navbarComponent`

	p := NewParser(s)
	p.MapCapture("*", func(cx Cx) string {
		return Italic{Body: cx.Body}.String()
	})
	p.MapCapture("**", func(cx Cx) string {
		return Bold{Body: cx.Body}.String()
	})
	p.MapLinePrefix("#", func(cx Cx) string {
		return Heading{Body: cx.Body, level: 1}.String()
	})
	p.MapWordPrefix("##", func(cx Cx) string {
		return Heading{Body: cx.Body, level: 2}.String()
	})
	p.MapWordPrefix("@", func(cx Cx) string {
		return NavbarComponent{Body: cx.Body}.String()
	})

	results := p.Parse()
	t.Log("Results", results)
	t.Error("Not implemented")
}

func testParseTwoConsecutiveRecusive(t *testing.T) {
	s := `***Hello World***`
	p := NewParser(s)
	p.MapCapture("*", func(cx Cx) string {
		return Italic{Body: cx.Body}.String()
	})
	p.MapCapture("**", func(cx Cx) string {
		return Bold{Body: cx.Body}.String()
	})
	results := p.Parse()
	t.Log("Results", results)
	t.Error("Not implemented")
}

func TestNotClosingSth(t *testing.T) {
	s := `**Hello World`
	p := NewParser(s)
	p.MapCapture("*", func(cx Cx) string {
		return Italic{Body: cx.Body}.String()
	})
	p.MapCapture("**", func(cx Cx) string {
		return Bold{Body: cx.Body}.String()
	})
	results := p.Parse()
	t.Log("Results", results)
	t.Error("Not implemented")
}
