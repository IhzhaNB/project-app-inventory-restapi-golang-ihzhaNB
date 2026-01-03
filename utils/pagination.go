package utils

type Pagination struct {
	Page       int `json:"page" query:"page"`
	Limit      int `json:"limit" query:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func NewPagination(page, limit int) Pagination {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10 // default limit
	}
	if limit > 100 {
		limit = 100 // max limit
	}

	return Pagination{
		Page:  page,
		Limit: limit,
	}
}

func (p *Pagination) Offset() int {
	return (p.Page - 1) * p.Limit
}

func (p *Pagination) SetTotal(total int) {
	p.Total = total
	if p.Limit > 0 {
		p.TotalPages = (total + p.Limit - 1) / p.Limit
	}
}
