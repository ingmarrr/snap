package snap

type (
	Cx struct {
		Chs  string
		Body string
	}
	MapperCx struct {
		Chs     string
		Mapper  Mapper
		Type    MapperType
		Checker Checker
	}
	ParseCx struct {
		chs     string
		buf     string
		rest    string
		mapper  Mapper
		checker Checker
		ty      MapperType
	}
)

func NoOpMapperCx() MapperCx {
	return MapperCx{
		Chs:     "",
		Mapper:  noOpMapper,
		Type:    Undefined,
		Checker: noOpChecker,
	}
}
