package snap

import (
	"fmt"
	"strings"
)

type String interface {
	String() string
}

type Mapper func(string) fmt.Stringer

type MapperNode struct {
	Chars    string
	Fn       Mapper
	Children []*MapperNode
}

func (n *MapperNode) Insert(chs string, fn Mapper) {
	if len(chs) == 0 {
		n.Fn = fn
		return
	}

	ch := chs[0:1]
	chs = chs[1:]

	for _, child := range n.Children {
		if child.Chars == ch {
			child.Insert(chs, fn)
			return
		}
	}

	child := &MapperNode{
		Chars: ch,
	}
	child.Insert(chs, fn)
	n.Children = append(n.Children, child)
}

func (n *MapperNode) Find(chs string) Mapper {
	if len(chs) == 0 {
		return n.Fn
	}

	ch := chs[0:1]
	chs = chs[1:]

	for _, child := range n.Children {
		if child.Chars == ch {
			return child.Find(chs)
		}
	}

	return nil
}

type Parser struct {
	Lines    []string
	mappers  *MapperNode
	captures *MapperNode
}

func NewParser(s string) Parser {
	lines := strings.Split(s, "\n")
	return Parser{
		Lines:    lines,
		mappers:  &MapperNode{},
		captures: &MapperNode{},
	}
}

func (p *Parser) MapFirst(chs string, fn Mapper) {
	p.mappers.Insert(chs, fn)
}

func (p *Parser) MapCaptures(chs string, fn Mapper) {
	p.captures.Insert(chs, fn)
}

func (p Parser) Parse() []string {
	var results []string
	for _, line := range p.Lines {
		results = append(results, p.parseLine(line))
	}
	return results
}

func (p Parser) parseLine(line string) string {
	var results []string

	if len(line) == 0 {
		return ""
	}

	chs := ""
	tmp := ""
	lastValidMapper := Mapper(nil)
	var lastValidMapperIdx int
	// Iterating over each character in the line
	// as long as we find a valid mapper, we replace the old one
	// with the new one. If we don't find a valid mapper, we
	// break out of the loop and use the last valid mapper
	// to parse the rest of the line.
	for i, ch := range line {
		chs += string(ch)
		fmt.Println("Chs", chs)
		if fn := p.mappers.Find(chs); fn != nil {
			lastValidMapper = fn
			lastValidMapperIdx = i
		} else {
			if lastValidMapper != nil {
				tmp = lastValidMapper(strings.TrimSpace(line[lastValidMapperIdx+1:])).String()
				lastValidMapper = nil
			}
		}
	}

	if tmp != "" {
		line = tmp
	}
	fmt.Println("line", line)

	var mapper Mapper
	inCapture := false
	var capture string
	chs = ""
	var startingPattern string

	for _, ch := range line {

		if p.captures.Find(string(ch)) != nil {
			chs += string(ch)
		} else {
			if chs != "" {
				startingPattern = chs
			}
			chs = ""
		}

		if fn := p.captures.Find(string(chs)); fn != nil {
			mapper = fn

			if startingPattern == chs {
				results = append(results, mapper(string(capture)).String())
				startingPattern = ""
				mapper = nil
				inCapture = false
				capture = ""
				chs = ""
				continue
			}
			continue
		} else {
			if mapper != nil {
				if !inCapture {
					mapper = nil
					inCapture = true
					capture = ""
					chs = ""
				} else {
					results = append(results, mapper(string(capture)).String())
					mapper = nil
					inCapture = false
					capture = ""
					chs = ""
				}
			}
		}

		if inCapture {
			capture += string(ch)
		} else {
			results = append(results, string(ch))
		}
	}

	return strings.Join(results, "")
}
