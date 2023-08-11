package snap

import (
	"fmt"
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
	p.MapFirst("#", func(cx Cx) fmt.Stringer {
		return Heading{Body: cx.Body, level: 1}
	})
	p.MapFirst("##", func(cx Cx) fmt.Stringer {
		return Heading{Body: cx.Body, level: 2}
	})
	p.MapFirst("####", func(cx Cx) fmt.Stringer {
		return Heading{Body: cx.Body, level: 4}
	})
	p.MapCaptures("*", func(cx Cx) fmt.Stringer {
		return Italic{Body: cx.Body}
	})
	p.MapCaptures("**", func(cx Cx) fmt.Stringer {
		return Bold{Body: cx.Body}
	})
	fmt.Println("Mapper ::", p.captures.Find("*"))
	results := p.Parse()

	t.Log(results)
	t.Log(p.Lines)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}
