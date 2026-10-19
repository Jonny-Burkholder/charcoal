package tokens

type PaginationToken struct {
	Paginate bool
	Limit    uint
	Offset   int
}
