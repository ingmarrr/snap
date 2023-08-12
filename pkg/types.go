package snap

import "fmt"

type Mapper func(Cx) fmt.Stringer
type Checker func(chs string, chsToCheck string) ContinueOrEnd
type MapperParser func(cx ParseCx) Rt
type MapperType int

const (
	Map MapperType = iota
	WordPrefix
	LinePrefix
	Capture
	Undefined
)

type noOp struct {
	body string
}
type Rt struct {
	parsed string
	rest   string
}
type ContinueOrEnd struct {
	continueSearchingForMatchingClosingCharacters bool
	applyParser                                   bool
}

func (n noOp) String() string {
	return n.body
}

func noOpMapper(cx Cx) fmt.Stringer {
	return noOp{body: cx.Chs + cx.Body + cx.Chs}
}

func noOpChecker(chs string, chsToCheck string) ContinueOrEnd {
	return ContinueOrEnd{
		continueSearchingForMatchingClosingCharacters: false,
		applyParser: false,
	}
}
