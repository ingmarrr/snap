package snap

import (
	"fmt"
	"strings"
)

type Parser struct {
	text         string
	linePrefixes *MapperNode
	wordPrefixes *MapperNode
	mappers      *MapperNode
	captures     *MapperNode
}

func NewParser(s string) Parser {
	return Parser{
		text:         s,
		linePrefixes: &MapperNode{},
		wordPrefixes: &MapperNode{},
		mappers:      &MapperNode{},
		captures:     &MapperNode{},
	}
}

func (p *Parser) Map(chs string, fn Mapper) {
	p.mappers.Insert(chs, fn)
}

func (p *Parser) MapLinePrefix(chs string, fn Mapper) {
	p.linePrefixes.Insert(chs, fn)
}

func (p *Parser) MapWordPrefix(chs string, fn Mapper) {
	p.wordPrefixes.Insert(chs, fn)
}

func (p *Parser) MapCapture(chs string, fn Mapper) {
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

func (p Parser) Parse() string {
	rt := p.parse(newParseCx("", p.text, noOpMapper))
	return rt.parsed
}

func (p Parser) ParseLines() string {
	var results []string
	for _, line := range strings.Split(p.text, "\n") {
		results = append(results, p.parseLine(line))
	}
	return strings.Join(results, "\n")
}

// ParseBlocks parses the text block by block. A block is simply text
// separated by two newlines. This is useful for parsing markdown
// since markdown uses two newlines to separate blocks.
//
// Example:
//
// 1 - # The Clone Wars
//
// 2 - ...
//
// 3 - Hello There ... *General Kenobi*.. what a pleasant surprise **to see you here**
//
// 4 - ...
//
// 5 - ## The Reupublic
func (p Parser) ParseBlocks() string {
	blocks := strings.Split(p.text, "\n\n")
	var lines []string
	for _, block := range blocks {
		lines = append(lines, strings.Join(strings.Split(block, "\n"), ""))
	}
	var results []string
	for _, line := range lines {
		results = append(results, p.parseLine(line))
	}

	return strings.Join(results, "\n")
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
		if fn := p.linePrefixes.Find(chs); fn != nil {
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
			if !p.linePrefixes.isWordOfLenPossible(len(chs)) {
				break
			}
		}
	}

	// If we didnt find any mapper and therefore the tmp variable is empty,
	// we just leave the line as is.
	if tmp != "" {
		line = tmp
	}
	// fmt.Println("line", line)

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

type ParseCx struct {
	chs    string
	buf    string
	rest   string
	mapper Mapper
}

func newParseCx(chs string, rest string, mapper Mapper) ParseCx {
	return ParseCx{
		chs:    chs,
		buf:    "",
		rest:   rest,
		mapper: mapper,
	}
}

func (p *Parser) parse(cx ParseCx) Rt {
	currChs := ""
	var mapper Mapper
	var nextChs string

	for {
		if len(cx.rest) == 0 {
			return Rt{
				parsed: "",
				rest:   "",
			}
		}

		ch := cx.rest[0:1]
		currChs += ch

		conOrIs := containsOrIs(cx.chs, currChs)

		if conOrIs.contains {
			if conOrIs.is {
				cx.rest = cx.rest[len(cx.chs):]
				return Rt{
					parsed: cx.mapper(Cx{
						Chs:  cx.chs,
						Body: cx.buf,
					}).String(),
					rest: cx.rest,
				}
			} else {
				continue
			}
		}

		if fn := p.ifFn(currChs); fn != nil {
			mapper = fn
			cx.rest = cx.rest[1:]
			nextChs = currChs
			continue
		}

		if mapper != nil {
			rt := p.parse(newParseCx(nextChs, cx.rest, mapper))
			cx.buf += rt.parsed
			cx.rest = rt.rest
			currChs = ""
			nextChs = ""
		}

		if len(cx.rest) != 0 {
			cx.buf += string(currChs)
			cx.rest = cx.rest[1:]
			currChs = ""
			continue
		}

		return Rt{
			parsed: cx.mapper(Cx{
				Chs:  cx.chs,
				Body: cx.buf,
			}).String(),
			rest: cx.rest,
		}
	}
}

func (p *Parser) ifFn(chs string) Mapper {
	if fn := p.linePrefixes.Find(chs); fn != nil {
		return fn
	}
	if fn := p.wordPrefixes.Find(chs); fn != nil {
		return fn
	}
	if fn := p.mappers.Find(chs); fn != nil {
		return fn
	}
	if fn := p.captures.Find(chs); fn != nil {
		return fn
	}
	return nil
}

func (p *Parser) parseCapture(chs string, s string, fn Mapper) Rt {
	closingChs := ""
	buf := ""
	rest := ""
	for i, ch := range s {
		closingChs += string(ch)
		cOrIs := containsOrIs(chs, closingChs)
		if cOrIs.contains {
			if cOrIs.is {
				rest = s[i+1:]
				break
			} else {
				continue
			}
		}

		buf += string(ch)
		closingChs = ""
	}

	return Rt{
		parsed: fn(Cx{
			Chs:  chs,
			Body: buf,
		}).String(),
		rest: rest,
	}
}

func containsOrIs(chs string, chsToCheck string) ConOrIs {
	lnTC := len(chsToCheck)
	lnCh := len(chs)

	if lnTC == 0 || lnCh == 0 {
		return ConOrIs{
			contains: false,
			is:       false,
		}
	}

	if lnTC > lnCh {
		return ConOrIs{
			contains: false,
			is:       false,
		}
	}

	if chs[0:lnTC] == chsToCheck {
		if lnTC == lnCh {
			return ConOrIs{
				contains: true,
				is:       true,
			}
		} else {
			return ConOrIs{
				contains: true,
				is:       false,
			}
		}
	}

	return ConOrIs{
		contains: false,
		is:       false,
	}
}

func parseLinePrefix(prefix string, s string, fn Mapper) Rt {
	buf := ""
	rest := ""
	for i, ch := range s {
		if ch == '\n' {
			buf += string(ch)
			rest = s[i+1:]
		}
	}
	return Rt{
		parsed: fn(Cx{
			Chs:  prefix,
			Body: buf,
		}).String(),
		rest: rest,
	}
}

func parseWordPrefix(prefix string, s string, fn Mapper) Rt {
	buf := ""
	rest := ""
	for i, ch := range s {
		if ch == ' ' || ch == '\n' || ch == '\t' || ch == '\r' {
			buf += string(ch)
			rest = s[i+1:]
		}
	}
	return Rt{
		parsed: fn(Cx{
			Chs:  prefix,
			Body: buf,
		}).String(),
		rest: rest,
	}
}

type Rt struct {
	parsed string
	rest   string
}

type ConOrIs struct {
	contains bool
	is       bool
}
