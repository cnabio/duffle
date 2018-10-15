package loader

type InsecureLoader struct {
	source Fetcher
}

func NewInsecureLoader(fetcher Fetcher) *InsecureLoader {
	return &InsecureLoader{
		source: fetcher,
	}
}
