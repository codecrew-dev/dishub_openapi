package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type PaginationParams struct {
	Page  int64
	Limit int64
	Skip  int64
}

func parsePagination(c *gin.Context) PaginationParams {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.ParseInt(pageStr, 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	return PaginationParams{
		Page:  page,
		Limit: limit,
		Skip:  (page - 1) * limit,
	}
}

func calcTotalPages(total, limit int64) int64 {
	if limit == 0 {
		return 0
	}
	pages := total / limit
	if total%limit != 0 {
		pages++
	}
	return pages
}
