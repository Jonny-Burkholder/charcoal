package main

type User struct {
	Name     string
	Age      int
	Password string `charcoal:"-"` // ignore this field when filtering
	Profile  profile
}

type profile struct {
	Occupation string
	Nickname   string
}
