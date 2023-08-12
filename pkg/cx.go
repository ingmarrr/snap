package snap

type (
	Cx struct {
		Chs  string
		Body string
	}
	mapperCx struct {
		chs     string
		mapper  mapper
		ty      mapperType
		checker checker
	}
	parseCx struct {
		chs     string
		buf     string
		rest    string
		mapper  mapper
		checker checker
		ty      mapperType
	}
)
