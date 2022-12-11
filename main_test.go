package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/pashagolub/pgxmock/v2"
)

func TestSlowQueries(t *testing.T) {
	timeout := 1000 * 30

	cases := map[string]struct {
		testURL          string
		expectedSQLQuery string
		dbRows           [][]any
		expectedResponse []SlowQueryLog
	}{
		"no query params": {
			testURL:          "/slow-queries",
			expectedSQLQuery: "^SELECT query, total_exec_time FROM public.pg_stat_statements ORDER BY total_exec_time desc LIMIT 50 OFFSET 0$",
			dbRows:           [][]any{{"select * from users", "12"}, {"select * from users", "14"}},
			expectedResponse: []SlowQueryLog{
				{
					Query:         "select * from users",
					TotalExecTime: "12",
				},
				{
					Query:         "select * from users",
					TotalExecTime: "14",
				},
			},
		},
		"all query params": {
			testURL:          "/slow-queries?page=5&page_size=2&query_type=select&order_by=asc",
			expectedSQLQuery: "^SELECT query, total_exec_time FROM public.pg_stat_statements WHERE query ILIKE 'select%' ORDER BY total_exec_time asc LIMIT 2 OFFSET 10$",
			dbRows:           [][]any{{"select * from users", "14"}, {"select * from users", "12"}},
			expectedResponse: []SlowQueryLog{
				{
					Query:         "select * from users",
					TotalExecTime: "14",
				},
				{
					Query:         "select * from users",
					TotalExecTime: "12",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mock, err := pgxmock.NewPool()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer mock.Close()

			repo := NewPostgresRepo(mock)

			app := Setup(repo)

			rows := mock.NewRows([]string{"query", "total_exec_time"})
			for _, r := range tc.dbRows {
				rows.AddRow(r...)
			}

			mock.ExpectQuery(tc.expectedSQLQuery).WillReturnRows(rows)

			req, _ := http.NewRequest(
				"GET",
				tc.testURL,
				nil,
			)

			res, err := app.Test(req, timeout)

			body, err := io.ReadAll(res.Body)
			expected, err := json.Marshal(tc.expectedResponse)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(expected, body) {
				t.Errorf("expected json: %s\n found %s", expected, body)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations in mock db: %s", err)
			}
		})
	}
}
