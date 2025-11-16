package models

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type DataResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ListResponse struct {
	Message string        `json:"message"`
	Data    []interface{} `json:"data"`
	Count   int           `json:"count"`
}

type PaginatedResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Paging  PagingInfo  `json:"paging"`
}

type PagingInfo struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}
