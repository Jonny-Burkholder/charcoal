package tokens

type PaginationToken struct {
	Paginate bool
	Page     int
	Limit    int
	Cursor   string
}
