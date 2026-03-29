package paginator

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 50
)

type Params struct {
	Page   int
	Limit  int
	Offset int
}

type Meta struct {
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	TotalItems  int  `json:"total_items"`
	TotalPages  int  `json:"total_pages"`
	HasNextPage bool `json:"has_next_page"`
	HasPrevPage bool `json:"has_prev_page"`
}

type Result[T any] struct {
	Items []T  `json:"items"`
	Meta  Meta `json:"meta"`
}

func FromRaw(pageInput, limitInput string) (Params, error) {
	page := DefaultPage
	if strings.TrimSpace(pageInput) != "" {
		parsedPage, err := strconv.Atoi(strings.TrimSpace(pageInput))
		if err != nil {
			return Params{}, fmt.Errorf("invalid page")
		}
		page = parsedPage
	}

	limit := DefaultLimit
	if strings.TrimSpace(limitInput) != "" {
		parsedLimit, err := strconv.Atoi(strings.TrimSpace(limitInput))
		if err != nil {
			return Params{}, fmt.Errorf("invalid limit")
		}
		limit = parsedLimit
	}

	return NewParams(page, limit)
}

func NewParams(page, limit int) (Params, error) {
	if page <= 0 {
		return Params{}, fmt.Errorf("page must be greater than 0")
	}
	if limit <= 0 {
		return Params{}, fmt.Errorf("limit must be greater than 0")
	}

	effectiveLimit := limit
	if effectiveLimit > MaxLimit {
		effectiveLimit = MaxLimit
	}

	return Params{
		Page:   page,
		Limit:  effectiveLimit,
		Offset: (page - 1) * effectiveLimit,
	}, nil
}

func BuildMeta(params Params, totalItems int) Meta {
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + params.Limit - 1) / params.Limit
	}

	return Meta{
		Page:        params.Page,
		Limit:       params.Limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasNextPage: params.Page < totalPages,
		HasPrevPage: params.Page > 1 && totalPages > 0,
	}
}
