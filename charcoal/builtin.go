package charcoal

// builtin.go contains logic for working with the built-in data layer integrations

type integrations struct {
}

func (c charcoal) ToSql(res Result) (string, error) {
	// TODO: use built-in integrations to convert tokens to SQL
	// I need to rework this a little bit - the integration needs to
	// be able to have knowledge of the data schema. Ideally they could
	// just be stored in the charcoal struct, but I can't really define
	// an interface that's satisfactory for all the disparate data types
	return "", nil
}

func (c charcoal) ToMongo(res Result) (string, error) {
	return "", nil
}

func (c charcoal) ToElastic(res Result) (string, error) {
	return "", nil
}

func (c charcoal) ToRedis(res Result) (string, error) {
	return "", nil
}

func (c charcoal) ToGraphQL(res Result) (string, error) {
	return "", nil
}
