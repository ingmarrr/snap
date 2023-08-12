package snap

type mapper func(Cx) string
type checker func(chs string, chsToCheck string) continueOrEnd
type mapperType int

const (
	Map mapperType = iota
	WordPrefix
	LinePrefix
	Capture
	Undefined
)

type noOp struct {
	body string
}
type rt struct {
	parsed string
	rest   string
}
type continueOrEnd struct {
	continueSearchingForMatchingClosingCharacters bool
	applyParser                                   bool
}

func (n noOp) String() string {
	return n.body
}

func noOpMapper(cx Cx) string {
	return noOp{body: cx.Chs + cx.Body + cx.Chs}.String()
}

func noOpChecker(chs string, chsToCheck string) continueOrEnd {
	return continueOrEnd{
		continueSearchingForMatchingClosingCharacters: false,
		applyParser: false,
	}
}

func noOpMapperCx() mapperCx {
	return mapperCx{
		chs:     "",
		mapper:  noOpMapper,
		ty:      Undefined,
		checker: noOpChecker,
	}
}
