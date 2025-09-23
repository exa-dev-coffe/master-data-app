package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jmoiron/sqlx"
)

func BuildFilterQuery(baseQuery string, params ParamsListRequest) (string, map[string]interface{}) {
	// Implementation here
	args := map[string]interface{}{}
	if params.Search.Field != "" && params.Search.Value != "" {
		if !strings.Contains(strings.ToUpper(baseQuery), "WHERE") {
			baseQuery += " WHERE 1=1 "
		}
		baseQuery += fmt.Sprintf(" AND %s ILIKE :searchValue", params.Search.Field)
		args["searchValue"] = "%" + params.Search.Value + "%"
	}
	if params.Sort.Order != "" && params.Sort.Field != "" {
		baseQuery += fmt.Sprintf(" ORDER BY %s %s ", params.Sort.Field, strings.ToUpper(params.Sort.Order))
	}
	if params.Size > 0 && params.Page > 0 && !params.NoPaginate {
		offset := (params.Page - 1) * params.Size
		baseQuery += "LIMIT :size OFFSET :offset"
		args["size"] = params.Size
		args["offset"] = offset
	}
	return baseQuery, args
}

func BuildCountQuery(baseQuery string, params ParamsListRequest) (string, map[string]interface{}) {
	// Implementation here
	args := map[string]interface{}{}
	if params.Search.Field != "" && params.Search.Value != "" {
		if !strings.Contains(strings.ToUpper(baseQuery), "WHERE") {
			baseQuery += " WHERE  1=1 "
		}
		baseQuery += fmt.Sprintf(" AND %s ILIKE :searchValue", params.Search.Field)
		args["searchValue"] = "%" + params.Search.Value + "%"
	}
	return baseQuery, args
}

func ParseQueryParams(queryParams map[string]string, params *ParamsListRequest) error {
	if page, ok := queryParams["page"]; ok {
		pg, err := strconv.Atoi(page)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid page parameter")
		}
		params.Page = pg
	} else {
		params.Page = 1 // default page
	}
	if size, ok := queryParams["size"]; ok {
		sz, err := strconv.Atoi(size)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid limit parameter")
		}
		params.Size = sz
	} else {
		params.Size = 10 // default size
	}
	if sortField, ok := queryParams["sort"]; ok {
		parts := strings.Split(sortField, ",")
		if len(parts) == 2 {
			params.Sort.Field = parts[0]
			params.Sort.Order = parts[1]
		}
	} else {
		params.Sort.Field = "id"   // default sort field
		params.Sort.Order = "DESC" // default sort order
	}
	if searchField, ok := queryParams["searchKey"]; ok {
		params.Search.Field = searchField
	}
	if searchValue, ok := queryParams["searchValue"]; ok {
		params.Search.Value = searchValue
	}
	if noPaginate, ok := queryParams["noPaginate"]; ok {
		if noPaginate == "true" {
			params.NoPaginate = true
		}
	}
	return nil
}

func WithTransaction[P any](db *sqlx.DB, fn func(tx *sqlx.Tx, args P) error, args P) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// rollback kalau ada panic atau error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // terusin panic biar ga ketelen
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = fn(tx, args)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
