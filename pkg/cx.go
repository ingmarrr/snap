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
	PCx struct {
		mcx  MapperCx
		buf  string
		rest string
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

func newParseCx(rest string, mCx MapperCx) ParseCx {
	return ParseCx{
		chs:     mCx.Chs,
		buf:     "",
		rest:    rest,
		mapper:  mCx.Mapper,
		checker: mCx.Checker,
		ty:      mCx.Type,
	}
}
