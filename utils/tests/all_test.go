package tests

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"encoding/json"

	"github.com/ioxe/graphql-go/errors"
	"github.com/ioxe/graphql-go/utils/query"
	"github.com/ioxe/graphql-go/utils/schema"
	"github.com/ioxe/graphql-go/utils/validation"
)

type Test struct {
	Name   string
	Rule   string
	Query  string
	Errors []*errors.QueryError
}

func TestAll(t *testing.T) {
	d, err := ioutil.ReadFile("testdata/test.schema")
	if err != nil {
		t.Fatal(err)
	}

	s := schema.New()
	if err := s.Parse(string(d)); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open("testdata/tests.json")
	if err != nil {
		t.Fatal(err)
	}

	var tests []*Test
	if err := json.NewDecoder(f).Decode(&tests); err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			d, err := query.Parse(test.Query)
			if err != nil {
				t.Fatal(err)
			}
			errs := validation.Validate(s, d)
			got := []*errors.QueryError{}
			for _, err := range errs {
				if err.Rule == test.Rule {
					err.Rule = ""
					got = append(got, err)
				}
			}
			if !reflect.DeepEqual(test.Errors, got) {
				t.Errorf("wrong errors\nexpected: %v\ngot:      %v", test.Errors, got)
			}
		})
	}
}
