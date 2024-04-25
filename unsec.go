package mfctx

var unsecure bool = false

func SetUnsecure(unsec bool) {
	unsecure = unsec
}

func GetUnsecure() (unsec bool) {
	return unsecure
}

// WithUnsec adds jsonify value and adds it when unsecure is true
func (c *Crumps) WithUnsec(name string, value any) *Crumps {
	if !unsecure {
		return c
	}

	return c.With(name, value)
}

// LogUnsec logs when unsecure is true
func (c *Crumps) LogUnsec(level LogLevel, message string) {
	if !unsecure {
		return
	}
	c.Log(level, message)
}
