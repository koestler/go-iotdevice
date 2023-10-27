package httpServer

type Nameable interface {
	Name() string
}

func getNames[N Nameable](list []N) (ret []string) {
	ret = make([]string, len(list))
	for i, t := range list {
		ret[i] = t.Name()
	}
	return
}
