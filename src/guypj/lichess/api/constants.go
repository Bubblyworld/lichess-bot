package api

type User struct {
	ID     string
	Name   string
	Title  string
	Rating int64

	Online bool
}

type Clock struct {
	Initial   int64
	Increment int64
}

type Variant struct {
	Key  string
	Name string
}
