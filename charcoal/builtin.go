package charcoal

import (
	"charcoal/internal/query"
	"charcoal/internal/tokens"
)

// builtin.go contains logic for working with the built-in data layer integrations

type sqlAdapter interface {
	TokensToSql(tokens tokens.Tokens) (string, error)
}

type mongoAdapter interface {
	TokensToMongo(tokens tokens.Tokens) (string, error)
}

type elasticAdapter interface {
	TokensToElastic(tokens tokens.Tokens) (string, error)
}

type redisAdapter interface {
	TokensToRedis(tokens tokens.Tokens) (string, error)
}

type graphQLAdapter interface {
	TokensToGraphQL(tokens tokens.Tokens) (string, error)
}

type builtin struct {
	sqlAdapter
	mongoAdapter
	elasticAdapter
	redisAdapter
	graphQLAdapter
}

func (c charcoal) ToSql(queryString string) (string, error) {
	// TODO: use built-in integrations to convert tokens to SQL
	toks, err := query.Parse(queryString, c.Filter.Fields)
	if err != nil {
		return "", err
	}
	return c.sqlAdapter.TokensToSql(toks)
}

func (c charcoal) ToMongo(query string) (string, error) {
	return "", nil
}

func (c charcoal) ToElastic(query string) (string, error) {
	return "", nil
}

func (c charcoal) ToRedis(query string) (string, error) {
	return "", nil
}

func (c charcoal) ToGraphQL(query string) (string, error) {
	return "", nil
}
