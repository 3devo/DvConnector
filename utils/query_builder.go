package utils

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/asdine/storm"
	"github.com/asdine/storm/q"
	"github.com/tidwall/gjson"
)

// swagger:parameters GetAllLogFiles GetAllCharts GetAllSheets GetAllWorkspaces
type QueryBuilderParams struct {
	//[{"key": "ID", "value": 1}] Array of values you want to filter
	Filter string `json:"filter"`
	// How many results it should skip
	Skip int `json:"skip"`
	// Max results returned
	Limit int `json:"limit"`
	// "Name, Age" Order by
	OrderBy []string `json:"orderBy"`
	// Reverse Order by
	Reverse bool `json:"reverse"`
}

// QueryBuilder is a method that generates storm query to give a more fine grain control over results
// Some query string examples that can be used
// filter example.com?format=[{key:value}]
// skip example.com?skip=10
// limit example.com?limit=10
// orderBy example.com?orderBy=Name,Age
// orderBy with reverse example.com?orderBy=Name,Age&reverse=true
func QueryBuilder(env *Env, r *http.Request) (storm.Query, error) {
	var selection []q.Matcher
	var query storm.Query
	query = env.Db.Select()
	params := r.URL.Query()

	// Generate equality filter
	if params.Get("filter") != "" {
		result := gjson.Parse(string(params.Get("filter")))
		result.ForEach(func(key, value gjson.Result) bool {
			selection = append(selection, q.Eq(strings.Title(value.Get("key").String()), value.Get("value").Value()))
			return true
		})
		query = env.Db.Select(q.And(selection...))
	}

	// Skip x results
	if params.Get("skip") != "" {
		skip, err := strconv.Atoi(params.Get("skip"))
		if err != nil {
			return nil, err
		}
		query = query.Skip(skip)
	}

	// Limit to x results
	if params.Get("limit") != "" {
		limit, err := strconv.Atoi(params.Get("limit"))
		if err != nil {
			return nil, err
		}
		query = query.Limit(limit)
	}

	// Order by fieldnames
	if params.Get("orderBy") != "" {
		orderBy := string(params.Get("orderBy"))
		tags := strings.Split(orderBy, ",")
		query = query.OrderBy(tags...)
		log.Println(tags)
		reverse, _ := strconv.ParseBool(params.Get("reverse"))
		if reverse {
			query = query.Reverse()
		}
	}
	return query, nil
}
