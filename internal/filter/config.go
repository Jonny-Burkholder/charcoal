package filter

type Config struct {
	SkipHiddenFields bool
	MaxIndirects     int
}

func defaultConfig() Config {
	return Config{
		SkipHiddenFields: true,
		MaxIndirects:     2,
	}
}

func (cfg Config) validate() error {
	// TODO: validate config/figure out if it's worth validating

	return nil
}
