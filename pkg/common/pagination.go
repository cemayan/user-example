package common

import "strings"

type SortingColumnName int64
type SortingColumnType int64

const (
	Country SortingColumnName = iota
)

const (
	ASC SortingColumnType = iota
	DESC
)

// Pagination is representation of payload which is given request
// SColumn represents sorting column in DB
// SType represents sorting type in DB such as ASC,DESC
// ConditionQuery represents  query in DB such as "country = ?"
// ConditionValue represents  query value in DB such as "UK"
type Pagination struct {
	Limit      int               `json:"limit,omitempty"`
	Page       int               `json:"page,omitempty"`
	SColumn    SortingColumnName `json:"sColumn,omitempty"`
	SType      SortingColumnType `json:"sType,omitempty"`
	CQuery     string            `json:"cQuery,omitempty"`
	CValue     string            `json:"cVal,omitempty"`
	TotalRows  int64             `json:"total_rows"`
	TotalPages int               `json:"total_pages"`
	Rows       interface{}       `json:"rows"`
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

// GetSort is returns sorting query which is wanted sorting column and type
// To prevent SQL injection it is used for ENUM variable
// https://gorm.io/docs/security.html#SQL-injection-Methods
func (p *Pagination) GetSort() string {
	var sb strings.Builder
	switch p.SColumn {
	case Country:
		sb.WriteString("country ")
	}

	switch p.SType {
	case ASC:
		sb.WriteString("asc")
	case DESC:
		sb.WriteString("desc")
	}

	return sb.String()
}
