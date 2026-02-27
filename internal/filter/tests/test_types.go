package tests

type simple struct {
	Name string
	Age  int
}

type nested struct {
	ID     string
	Simple simple
}

type pointer struct {
	ID     string
	Simple *simple
}

type multipleIndirects struct {
	ID     string
	Simple **simple
}

type maxIndirects struct {
	ID     string
	Simple ***simple
}

type unexported struct {
	name string
	age  int
}

type doubleNested struct {
	ID     string
	Simple simple
	Nested nested
}
