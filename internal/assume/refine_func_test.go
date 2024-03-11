package assume

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty-debug/ctydebug"
	"github.com/zclconf/go-cty/cty"
)

func TestRefineFuncs(t *testing.T) {
	tests := map[string]map[string]struct {
		Args    []cty.Value
		Want    cty.Value
		WantErr string
	}{

		"notnull": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
				},
				Want: cty.DynamicVal,
			},
			"unknown string": {
				Args: []cty.Value{
					cty.UnknownVal(cty.String),
				},
				Want: cty.UnknownVal(cty.String).RefineNotNull(),
			},
			"known string": {
				Args: []cty.Value{
					cty.StringVal("a"),
				},
				Want: cty.StringVal("a"),
			},
			"null string": {
				Args: []cty.Value{
					cty.NullVal(cty.String),
				},
				WantErr: `assumption was not upheld`,
			},
		},

		"stringprefix": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.StringVal("foo-"),
				},
				Want: cty.DynamicVal,
			},
			"unknown string": {
				Args: []cty.Value{
					cty.UnknownVal(cty.String),
					cty.StringVal("foo-"),
				},
				Want: cty.UnknownVal(cty.String).Refine().
					// A dash cannot combine with any subsequent character, so
					// we can assume the entire prefix.
					StringPrefixFull("foo-").
					NewValue(),
			},
			"unknown string with combinable prefix": {
				Args: []cty.Value{
					cty.UnknownVal(cty.String),
					cty.StringVal("foo"),
				},
				Want: cty.UnknownVal(cty.String).Refine().
					// "foo" could potentially combine with a subsequent
					// diacritic, so we can only assume the first two
					// characters of the prefix in practice.
					StringPrefixFull("fo").
					NewValue(),
			},
			"null string": {
				Args: []cty.Value{
					cty.NullVal(cty.String),
					cty.StringVal("foo-"),
				},
				Want: cty.NullVal(cty.String),
			},
			"known string with correct prefix": {
				Args: []cty.Value{
					cty.StringVal("foo-bar"),
					cty.StringVal("foo-"),
				},
				Want: cty.StringVal("foo-bar"),
			},
			"known string with incorrect prefix": {
				Args: []cty.Value{
					cty.StringVal("bar-baz"),
					cty.StringVal("foo-"),
				},
				WantErr: `assumption was not upheld`,
			},
		},

		"listlength": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				Want: cty.DynamicVal,
			},
			"unknown list": {
				Args: []cty.Value{
					cty.UnknownVal(cty.List(cty.String)),
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				Want: cty.UnknownVal(cty.List(cty.String)).Refine().
					CollectionLengthLowerBound(1).
					CollectionLengthUpperBound(2).
					NewValue(),
			},
			"known list with correct length": {
				Args: []cty.Value{
					cty.ListVal([]cty.Value{cty.StringVal("a")}),
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				Want: cty.ListVal([]cty.Value{cty.StringVal("a")}),
			},
			"known list with incorrect length": {
				Args: []cty.Value{
					cty.ListValEmpty(cty.String),
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				WantErr: "assumption was not upheld",
			},
		},
		"listlengthmin": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.NumberIntVal(1),
				},
				Want: cty.DynamicVal,
			},
			"unknown list": {
				Args: []cty.Value{
					cty.UnknownVal(cty.List(cty.String)),
					cty.NumberIntVal(1),
				},
				Want: cty.UnknownVal(cty.List(cty.String)).Refine().
					CollectionLengthLowerBound(1).
					NewValue(),
			},
			"known list with correct length": {
				Args: []cty.Value{
					cty.ListVal([]cty.Value{cty.StringVal("a")}),
					cty.NumberIntVal(1),
				},
				Want: cty.ListVal([]cty.Value{cty.StringVal("a")}),
			},
			"known list with incorrect length": {
				Args: []cty.Value{
					cty.ListValEmpty(cty.String),
					cty.NumberIntVal(1),
				},
				WantErr: "assumption was not upheld",
			},
		},
		"listlengthmax": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.NumberIntVal(1),
				},
				Want: cty.DynamicVal,
			},
			"unknown list": {
				Args: []cty.Value{
					cty.UnknownVal(cty.List(cty.String)),
					cty.NumberIntVal(1),
				},
				Want: cty.UnknownVal(cty.List(cty.String)).Refine().
					CollectionLengthUpperBound(1).
					NewValue(),
			},
			"known list with correct length": {
				Args: []cty.Value{
					cty.ListVal([]cty.Value{cty.StringVal("a")}),
					cty.NumberIntVal(1),
				},
				Want: cty.ListVal([]cty.Value{cty.StringVal("a")}),
			},
			"known list with incorrect length": {
				Args: []cty.Value{
					cty.ListVal([]cty.Value{cty.True, cty.True}),
					cty.NumberIntVal(1),
				},
				WantErr: "assumption was not upheld",
			},
		},

		"setlength": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				Want: cty.DynamicVal,
			},
			"unknown set": {
				Args: []cty.Value{
					cty.UnknownVal(cty.Set(cty.String)),
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				Want: cty.UnknownVal(cty.Set(cty.String)).Refine().
					CollectionLengthLowerBound(1).
					CollectionLengthUpperBound(2).
					NewValue(),
			},
			"known set with correct length": {
				Args: []cty.Value{
					cty.SetVal([]cty.Value{cty.StringVal("a")}),
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				Want: cty.SetVal([]cty.Value{cty.StringVal("a")}),
			},
			"known set with incorrect length": {
				Args: []cty.Value{
					cty.SetValEmpty(cty.String),
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				WantErr: "assumption was not upheld",
			},
		},
		"setlengthmin": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.NumberIntVal(1),
				},
				Want: cty.DynamicVal,
			},
			"unknown set": {
				Args: []cty.Value{
					cty.UnknownVal(cty.Set(cty.String)),
					cty.NumberIntVal(1),
				},
				Want: cty.UnknownVal(cty.Set(cty.String)).Refine().
					CollectionLengthLowerBound(1).
					NewValue(),
			},
			"known set with correct length": {
				Args: []cty.Value{
					cty.SetVal([]cty.Value{cty.StringVal("a")}),
					cty.NumberIntVal(1),
				},
				Want: cty.SetVal([]cty.Value{cty.StringVal("a")}),
			},
			"known set with incorrect length": {
				Args: []cty.Value{
					cty.SetValEmpty(cty.String),
					cty.NumberIntVal(1),
				},
				WantErr: "assumption was not upheld",
			},
		},
		"setlengthmax": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.NumberIntVal(1),
				},
				Want: cty.DynamicVal,
			},
			"unknown set": {
				Args: []cty.Value{
					cty.UnknownVal(cty.Set(cty.String)),
					cty.NumberIntVal(1),
				},
				Want: cty.UnknownVal(cty.Set(cty.String)).Refine().
					CollectionLengthUpperBound(1).
					NewValue(),
			},
			"known set with correct length": {
				Args: []cty.Value{
					cty.SetVal([]cty.Value{cty.StringVal("a")}),
					cty.NumberIntVal(1),
				},
				Want: cty.SetVal([]cty.Value{cty.StringVal("a")}),
			},
			"known set with incorrect length": {
				Args: []cty.Value{
					cty.SetVal([]cty.Value{cty.Zero, cty.NumberIntVal(1)}),
					cty.NumberIntVal(1),
				},
				WantErr: "assumption was not upheld",
			},
		},

		"maplength": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				Want: cty.DynamicVal,
			},
			"unknown map": {
				Args: []cty.Value{
					cty.UnknownVal(cty.Map(cty.String)),
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				Want: cty.UnknownVal(cty.Map(cty.String)).Refine().
					CollectionLengthLowerBound(1).
					CollectionLengthUpperBound(2).
					NewValue(),
			},
			"known map with correct length": {
				Args: []cty.Value{
					cty.MapVal(map[string]cty.Value{"a": cty.StringVal("a")}),
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				Want: cty.MapVal(map[string]cty.Value{"a": cty.StringVal("a")}),
			},
			"known map with incorrect length": {
				Args: []cty.Value{
					cty.MapValEmpty(cty.String),
					cty.NumberIntVal(1),
					cty.NumberIntVal(2),
				},
				WantErr: "assumption was not upheld",
			},
		},
		"maplengthmin": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.NumberIntVal(1),
				},
				Want: cty.DynamicVal,
			},
			"unknown map": {
				Args: []cty.Value{
					cty.UnknownVal(cty.Map(cty.String)),
					cty.NumberIntVal(1),
				},
				Want: cty.UnknownVal(cty.Map(cty.String)).Refine().
					CollectionLengthLowerBound(1).
					NewValue(),
			},
			"known map with correct length": {
				Args: []cty.Value{
					cty.MapVal(map[string]cty.Value{"a": cty.StringVal("a")}),
					cty.NumberIntVal(1),
				},
				Want: cty.MapVal(map[string]cty.Value{"a": cty.StringVal("a")}),
			},
			"known map with incorrect length": {
				Args: []cty.Value{
					cty.MapValEmpty(cty.String),
					cty.NumberIntVal(1),
				},
				WantErr: "assumption was not upheld",
			},
		},
		"maplengthmax": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.NumberIntVal(1),
				},
				Want: cty.DynamicVal,
			},
			"unknown map": {
				Args: []cty.Value{
					cty.UnknownVal(cty.Map(cty.String)),
					cty.NumberIntVal(1),
				},
				Want: cty.UnknownVal(cty.Map(cty.String)).Refine().
					CollectionLengthUpperBound(1).
					NewValue(),
			},
			"known map with correct length": {
				Args: []cty.Value{
					cty.MapVal(map[string]cty.Value{"a": cty.StringVal("a")}),
					cty.NumberIntVal(1),
				},
				Want: cty.MapVal(map[string]cty.Value{"a": cty.StringVal("a")}),
			},
			"known map with incorrect length": {
				Args: []cty.Value{
					cty.MapVal(map[string]cty.Value{"a": cty.Zero, "b": cty.Zero}),
					cty.NumberIntVal(1),
				},
				WantErr: "assumption was not upheld",
			},
		},
	}

	p := NewProvider()
	for funcName, funcTests := range tests {
		t.Run(funcName, func(t *testing.T) {
			f := p.CallStub(funcName)
			for testName, test := range funcTests {
				t.Run(testName, func(t *testing.T) {
					got, gotErr := f(test.Args...)

					if test.WantErr != "" {
						if gotErr == nil {
							t.Fatalf("unexpected success\nwant error: %s", test.WantErr)
						}
						if got, want := gotErr.Error(), test.WantErr; got != want {
							t.Errorf("wrong error\ngot:  %s\nwant: %s", got, want)
						}
						return
					}

					if gotErr != nil {
						t.Fatalf("unexpected error: %s", gotErr)
					}
					if diff := cmp.Diff(test.Want, got, ctydebug.CmpOptions); diff != "" {
						t.Errorf("wrong result\n%s", diff)
					}
				})
			}
		})
	}
}
