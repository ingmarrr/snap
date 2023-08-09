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
