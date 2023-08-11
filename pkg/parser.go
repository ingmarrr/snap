package snap

import (
	"fmt"
	"strings"
)

type String interface {
	String() string
}

type Mapper func(Cx) fmt.Stringer
type Cx struct {
	Chs  string
	Body string
}

type MapperNode struct {
	Chars                 string
	Fn                    Mapper
	Children              []*MapperNode
	maxLenOfCharsThisPath int
}

func (n *MapperNode) Insert(chs string, fn Mapper) {
	n._insert(chs, fn, len(chs))
}

func (n *MapperNode) _insert(chs string, fn Mapper, originalLenOfChars int) {
	if n.maxLenOfCharsThisPath < originalLenOfChars {
		n.maxLenOfCharsThisPath = originalLenOfChars
	}
	if len(chs) == 0 {
		n.Fn = fn
		return
	}

	ch := chs[0:1]
	chs = chs[1:]

	for _, child := range n.Children {
		if child.Chars == ch {
			child.maxLenOfCharsThisPath = originalLenOfChars
			child._insert(chs, fn, originalLenOfChars)
			return
		}
	}

	child := &MapperNode{
		Chars:                 ch,
		maxLenOfCharsThisPath: originalLenOfChars,
	}
	child._insert(chs, fn, originalLenOfChars)
	n.Children = append(n.Children, child)

}

func (n *MapperNode) isWordOfLenPossible(len int) bool {
	if n.maxLenOfCharsThisPath == 0 {
		return false
	}

	return len <= n.maxLenOfCharsThisPath
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
	chsToInsert := strings.Split(chs, "")
	for i := range chsToInsert {
		if i < len(chsToInsert)-1 {
			// If we already have a mapper for the current character, we do not
			// override it. This is to prevent the following scenario:
			// We have a mapper for "*" and "**". If we now try to insert a mapper
			// for "***", we do not want to override the mapper for "**" with the
			// noOpMapper. To prevent this we first check if there is already a
			// mapper for the current characters. If there is, we do not override it.
			if p.captures.Find(chs[0:i+1]) != nil {
				continue
			}
			// If we do not have a mapper for the current character, we insert a
			// noOpMapper. This is important since in the parser we check if we find a mapper
			// For a single character first before starting to add more characters to the
			// string we use for searching. So we need the full path of the string to be present
			// in the tree, but of course if it is not complete yet, we do not want to insert the
			// actual mapper.
			p.captures.Insert(chs[0:i+1], Mapper(noOpMapper))
		} else {
			p.captures.Insert(chs, fn)
		}
	}
}

func noOpMapper(cx Cx) fmt.Stringer {
	return noOp{body: cx.Chs + cx.Body + cx.Chs}
}

type noOp struct {
	body string
}

func (n noOp) String() string {
	return n.body
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
	// Iterating over each character in the line.
	// As long as we find a valid mapper, we replace the old one
	// with the new one. If we don't find a valid mapper, we
	// check if it is possible to find another mapper further down.
	// If not, we break out of the loop and use the last valid mapper
	// to parse the rest of the line.
	for i, ch := range line {
		chs += string(ch)
		if fn := p.mappers.Find(chs); fn != nil {
			lastValidMapper = fn
			lastValidMapperIdx = i
		} else {
			if lastValidMapper != nil {
				tmp = lastValidMapper(Cx{
					Chs:  line[0:lastValidMapperIdx],
					Body: strings.TrimSpace(line[lastValidMapperIdx+1:]),
				}).String()
				lastValidMapper = nil
			}
			if !p.mappers.isWordOfLenPossible(len(chs)) {
				break
			}
		}
	}

	// If we didnt find any mapper and therefore the tmp variable is empty,
	// we just leave the line as is.
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
			// If we do not find a capture mapper for the current character,
			// we take the characters that previously matched a capture mapper
			// and remember them as the starting pattern in order to compare it
			// with the closing pattern.
			if chs != "" {
				startingPattern = chs
			}
			chs = ""
		}

		if fn := p.captures.Find(string(chs)); fn != nil {
			mapper = fn

			if startingPattern == chs {
				results = append(results, mapper(Cx{
					Chs:  startingPattern,
					Body: string(capture),
				}).String())
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
					results = append(results, mapper(Cx{
						Chs:  startingPattern,
						Body: string(capture),
					}).String())
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

	if startingPattern != "" {
		results = append(results, noOpMapper(Cx{
			Chs:  "",
			Body: startingPattern + (capture),
		}).String())
	}

	return strings.Join(results, "")
}
