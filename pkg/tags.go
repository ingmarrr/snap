package snap

import "fmt"

type Heading struct {
	level int
	Body  string
}

func (h Heading) String() string {
	return fmt.Sprintf("<h%d>%s</h%d>", h.level, h.Body, h.level)
}

type Italic struct {
	Body string
}

func (i Italic) String() string {
	return fmt.Sprintf("<i>%s</i>", i.Body)
}

type Bold struct {
	Body string
}

func (b Bold) String() string {
	return fmt.Sprintf("<b>%s</b>", b.Body)
}

type NavbarComponent struct {
	Body string
}

func (n NavbarComponent) String() string {
	return `<nav>
	<a href="/html/">HTML</a> |
	<a href="/css/">CSS</a> |
	<a href="/js/">JavaScript</a> |
	<a href="/python/">Python</a>
  </nav>`
}
