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
  `
	p := NewParser(s)
	p.MapFirst("#", func(s string) fmt.Stringer {
		return Heading{Body: s, level: 1}
	})
	p.MapFirst("##", func(s string) fmt.Stringer {
		return Heading{Body: s, level: 2}
	})
	p.MapFirst("####", func(s string) fmt.Stringer {
		return Heading{Body: s, level: 4}
	})
	p.MapCaptures("*", func(s string) fmt.Stringer {
		return Italic{Body: s}
	})
	p.MapCaptures("**", func(s string) fmt.Stringer {
		return Bold{Body: s}
	})
	results := p.Parse()

	t.Log(results)
	t.Log(p.Lines)
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}
