package tokens

type PaginationToken struct {
	Paginate bool
	Page     int
	PerPage  int
	Cursor   string
}
