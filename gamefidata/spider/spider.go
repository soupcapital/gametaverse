package spider

type Spider interface {
	Init() (err error)
	Run()
}
