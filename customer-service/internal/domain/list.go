package domain

type ListFilter struct {
	Status string
	Search string
	Limit  int
	Cursor string
}

type ListResult struct {
	Customers []Customer
	NextToken string
}
