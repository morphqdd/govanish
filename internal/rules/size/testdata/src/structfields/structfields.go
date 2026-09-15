package structfields

type Big struct { // want `struct Big has 11 fields, limit is 10`
	A, B, C int
	D, E, F int
	G, H, I int
	J       int
	K       int
}

type AtLimit struct {
	A, B, C int
	D, E, F int
	G, H, I int
	J       int
}

type Small struct {
	A int
}

type Empty struct{}
