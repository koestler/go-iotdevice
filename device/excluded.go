package device

func IsExcluded(registerName, category string, cfg Config) bool {
	for _, e := range cfg.SkipFields() {
		if e == registerName {
			return true
		}
	}

	for _, e := range cfg.SkipCategories() {
		if e == category {
			return true
		}
	}
	return false

}
