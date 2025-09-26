package common

import (
	"fmt"
	"strconv"
	"strings"

	"eka-dev.com/master-data/lib"
	"eka-dev.com/master-data/middleware"
	"eka-dev.com/master-data/utils/response"
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
			return response.BadRequest("Invalid page parameter", nil)
		}
		params.Page = pg
	} else {
		params.Page = 1 // default page
	}
	if size, ok := queryParams["size"]; ok {
		sz, err := strconv.Atoi(size)
		if err != nil {
			return response.BadRequest("Invalid size parameter", nil)
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

func GetClaimsFromLocals(c *fiber.Ctx) (*middleware.Claims, error) {
	user := c.Locals("user")
	claims, ok := user.(*middleware.Claims)
	if !ok {
		return nil, response.InternalServerError("Failed to get user from token", nil)
	}
	return claims, nil
}

func GetOneDataRequest(c *fiber.Ctx) (*OneRequest, error) {
	var request OneRequest
	err := c.QueryParser(&request)
	if err != nil {
		return nil, response.BadRequest("Invalid query parameters: "+err.Error(), nil)
	}
	if request.Id <= 0 {
		return nil, response.BadRequest("Invalid id parameter", nil)
	}

	err = lib.ValidateRequest(request)

	if err != nil {
		return nil, err
	}

	return &request, nil
}
