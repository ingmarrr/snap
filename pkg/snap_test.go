package snap

import (
	"fmt"
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
	p.MapLinePrefix("#", func(cx Cx) fmt.Stringer {
		return Heading{Body: cx.Body, level: 1}
	})
	p.MapLinePrefix("##", func(cx Cx) fmt.Stringer {
		return Heading{Body: cx.Body, level: 2}
	})
	p.MapLinePrefix("####", func(cx Cx) fmt.Stringer {
		return Heading{Body: cx.Body, level: 4}
	})
	p.MapCapture("*", func(cx Cx) fmt.Stringer {
		return Italic{Body: cx.Body}
	})
	p.MapCapture("**", func(cx Cx) fmt.Stringer {
		return Bold{Body: cx.Body}
	})
	results := p.ParseLines()

	t.Log(results)
	t.Log(strings.Split(p.text, "\n"))
}

func TestParseImpl(t *testing.T) {
	s := `**Hello |ok whatever| World**`
	p := NewParser(s)
	p.MapCapture("**", func(cx Cx) fmt.Stringer {
		return Bold{Body: cx.Body}
	})
	p.MapCapture("|", func(cx Cx) fmt.Stringer {
		return Italic{Body: cx.Body}
	})

	results := p.Parse()
	t.Log("Results", results)
	t.Error("Not implemented")
}
