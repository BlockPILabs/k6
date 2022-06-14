package blockpi

func Register(m map[string]interface{}) map[string]interface{} {
	m["pi"] = New()
	return m
}
