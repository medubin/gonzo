package url

type URL[params any, pathParams any] struct {
	PathParams pathParams
	Params     params
}
