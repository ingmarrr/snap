package snap

import (
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

func (p Parser) Parse() string {
	mCx := NoOpMapperCx()
	cx := ParseCx{
		chs:     mCx.Chs,
		buf:     "",
		rest:    p.text,
		mapper:  mCx.Mapper,
		checker: mCx.Checker,
		ty:      mCx.Type,
	}
	rt := p.parse(cx)
	return rt.parsed
}

func (p Parser) ParseLines() string {
	var results []string
	for _, line := range strings.Split(p.text, "\n") {
		results = append(results, p.parseLine(line))
	}
	return strings.Join(results, "\n")
}

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

func (p *Parser) parse(cx ParseCx) Rt {
	currChs := ""
	var mCx MapperCx = NoOpMapperCx()
	var nextChs string

	for {
		if len(cx.rest) == 0 {
			// Reached the end of the string. The only case where we want to return
			// the original string is if we are in a capture mapper. This is because
			// since we reached the end of the string without early exit, that means
			// that we did not find a matching closing character. Therefore we want to
			// return the original string and apply the `noOpMapper` to it.
			if cx.ty == Capture {
				return Rt{
					parsed: noOpMapper(Cx{
						Chs:  "",
						Body: cx.buf + cx.chs,
					}).String(),
					rest: cx.rest,
				}
			}
			// If we are not in a capture mapper, we want to return the parsed string
			// and the rest of the string, since for any other mapper we do not require
			// a matching closing character. [linePrefixes, wordPrefixes, mappers]
			res := cx.mapper(Cx{
				Chs:  cx.chs,
				Body: cx.buf,
			}).String()

			return Rt{
				parsed: res,
				rest:   cx.rest,
			}
		}

		ch := cx.rest[0:1]
		currChs += ch

		check := cx.checker(cx.chs, currChs)

		// If we are in a capture mapper and there is a possibility for a higher
		// priority mapper to be found, we want to continue searching for a matching
		// closing character. The longer the postfix of the mapper is, the higher the
		// priority. Consider the following example:
		// We have a mapper for "*" and "**". If we are looking for "**", we will always
		// find a mapper for "*" first. Therefore we want to continue searching for a
		// matching closing character.
		if check.continueSearchingForMatchingClosingCharacters {
			continue
		}

		// This means that whatever checker we called, be it for a linePrefix, wordPrefix,
		// mapper or capture, we found a character that terminates that specific
		// parsing/transpiling. Therefore we want to apply the mapper to the current
		// string and return the parsed string and the rest of the string.
		if check.applyParser {
			if cx.ty == Capture {
				cx.rest = cx.rest[len(cx.chs):]
			}
			return Rt{
				parsed: cx.mapper(Cx{
					Chs:  cx.chs,
					Body: cx.buf,
				}).String(),
				rest: cx.rest,
			}
		}

		// As long as we find a valid mapper, we want to continue searching for a
		// matching opening character. The longer the prefix of the mapper is, the
		// higher the priority. Consider the following example:
		// We have a mapper for "*" and "**". If we are looking for "**", we will always
		// find a mapper for "*" first. Therefore we want to continue searching for a
		// matching opening character as long as it is necessary.
		if fnMCx := p.ifFn(currChs); fnMCx.Type != Undefined {
			mCx = fnMCx
			// Advancing by a single character here is important since we want to
			// check if the next character is a valid mapper.
			cx.rest = cx.rest[1:]
			nextChs = currChs
			continue
		}

		// If we found a valid mapper, we want to parse the rest of the string recursively.
		// We do this by calling the parse function again with the rest of the string and
		// pass in the required information to the recursive call.
		if mCx.Type != Undefined {
			rt := p.parse(ParseCx{
				chs:     nextChs,
				buf:     "",
				rest:    cx.rest,
				mapper:  mCx.Mapper,
				checker: mCx.Checker,
				ty:      mCx.Type,
			})
			// We just exited the recursive call and therefore we need to update the
			// current context with the parsed string and the rest of the string.
			cx.buf += rt.parsed
			cx.rest = rt.rest

			// We need to reset the current characters and the next characters
			// since we just parsed them and we might encounter another recursive call
			currChs = ""
			nextChs = ""
			mCx = NoOpMapperCx()
		}

		if len(cx.rest) != 0 {
			cx.buf += string(currChs)
			// We need to advance the rest of the string by the length of the current
			// characters since we already checked if the current characters are a valid
			// mapper. If they are not, we need to advance the rest of the string by the
			// length of the current characters, which could be very well more than one.
			cx.rest = cx.rest[len(currChs):]
			currChs = ""
			nextChs = cx.chs
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

func (p *Parser) ifFn(chs string) MapperCx {
	var ty MapperType = Undefined
	var mapper Mapper = noOpMapper
	var checker Checker = noOpChecker
	if fn := p.linePrefixes.Find(chs); fn != nil {
		ty = LinePrefix
		mapper = fn
		checker = lineEndChecker
	}
	if fn := p.wordPrefixes.Find(chs); fn != nil {
		ty = WordPrefix
		mapper = fn
		checker = wordEndChecker
	}
	if fn := p.mappers.Find(chs); fn != nil {
		ty = Map
		mapper = fn
		checker = noCharacterMatchChecker
	}
	if fn := p.captures.Find(chs); fn != nil {
		ty = Capture
		mapper = fn
		checker = captureEndChecker
	}
	return MapperCx{
		Chs:     chs,
		Mapper:  mapper,
		Checker: checker,
		Type:    ty,
	}
}

func captureEndChecker(chs string, chsToCheck string) ContinueOrEnd {
	lnTC := len(chsToCheck)
	lnCh := len(chs)

	if lnTC == 0 || lnCh == 0 {
		return ContinueOrEnd{
			continueSearchingForMatchingClosingCharacters: false,
			applyParser: false,
		}
	}

	if lnTC > lnCh {
		return ContinueOrEnd{
			continueSearchingForMatchingClosingCharacters: false,
			applyParser: false,
		}
	}

	if chs[0:lnTC] == chsToCheck {
		if lnTC == lnCh {
			return ContinueOrEnd{
				continueSearchingForMatchingClosingCharacters: false,
				applyParser: true,
			}
		} else {
			return ContinueOrEnd{
				continueSearchingForMatchingClosingCharacters: true,
				applyParser: false,
			}
		}
	}

	return ContinueOrEnd{
		continueSearchingForMatchingClosingCharacters: false,
		applyParser: false,
	}
}

func lineEndChecker(_ string, chsToCheck string) ContinueOrEnd {
	if len(chsToCheck) == 0 {
		return ContinueOrEnd{
			continueSearchingForMatchingClosingCharacters: false,
			applyParser: true,
		}
	}
	lastChar := chsToCheck[len(chsToCheck)-1:]
	if lastChar == "\n" || lastChar == "\r" {
		return ContinueOrEnd{
			continueSearchingForMatchingClosingCharacters: false,
			applyParser: true,
		}
	}
	return ContinueOrEnd{
		continueSearchingForMatchingClosingCharacters: false,
		applyParser: false,
	}
}

func wordEndChecker(chs string, chsToCheck string) ContinueOrEnd {
	if len(chsToCheck) == 0 {
		return ContinueOrEnd{
			continueSearchingForMatchingClosingCharacters: false,
			applyParser: true,
		}
	}
	lastChar := chsToCheck[len(chsToCheck)-1:]
	if lastChar == " " || lastChar == "\n" || lastChar == "\t" || lastChar == "\r" {
		return ContinueOrEnd{
			continueSearchingForMatchingClosingCharacters: false,
			applyParser: true,
		}
	}
	return ContinueOrEnd{
		continueSearchingForMatchingClosingCharacters: false,
		applyParser: false,
	}
}

func noCharacterMatchChecker(chs string, chsToCheck string) ContinueOrEnd {
	if len(chsToCheck) == 0 {
		return ContinueOrEnd{
			continueSearchingForMatchingClosingCharacters: false,
			applyParser: false,
		}
	}

	if len(chs) != len(chsToCheck) {
		return ContinueOrEnd{
			continueSearchingForMatchingClosingCharacters: false,
			applyParser: false,
		}
	}

	for i, ch := range chs {
		if byte(ch) != chsToCheck[i] {
			return ContinueOrEnd{
				continueSearchingForMatchingClosingCharacters: false,
				applyParser: false,
			}
		}
	}

	return ContinueOrEnd{
		continueSearchingForMatchingClosingCharacters: false,
		applyParser: true,
	}
}
