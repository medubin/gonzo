package url

type URL[params any] struct {
	Values Values
	Params params
}

type Values map[string][]string

func (v Values) GetAll(key string) []string {
	if v == nil {
		return []string{}
	}
	return v[key]
}

func (v Values) Get(key string) string {
	if v == nil {
		return ""
	}
	vs := v[key]
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

// Has checks whether a given key is set.
func (v Values) Has(key string) bool {
	_, ok := v[key]
	return ok
}
