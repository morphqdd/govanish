package underscores

var bad_name = 1 // want `identifier bad_name contains an underscore`

func Bad_Func() {} // want `identifier Bad_Func contains an underscore`

func Good() {
	good_local := 1 // want `identifier good_local contains an underscore`
	_ = good_local
}

type Good_Type struct { // want `identifier Good_Type contains an underscore`
	Bad_Field int // want `identifier Bad_Field contains an underscore`
}

func Blank() {
	_, value := 1, 2
	_ = value
}
