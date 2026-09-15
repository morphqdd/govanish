package bad

var loaded bool

func init() {
	loaded = true
}

func Loaded() bool {
	return loaded
}
