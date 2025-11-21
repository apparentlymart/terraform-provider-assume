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

		"equal": {
			"dynamicval": {
				Args: []cty.Value{
					cty.DynamicVal,
					cty.StringVal("hi"),
				},
				Want: cty.StringVal("hi"),
			},
			"unknown string": {
				Args: []cty.Value{
					cty.UnknownVal(cty.String),
					cty.StringVal("hi"),
				},
				Want: cty.StringVal("hi"),
			},
			"known string, correct": {
				Args: []cty.Value{
					cty.StringVal("hi"),
					cty.StringVal("hi"),
				},
				Want: cty.StringVal("hi"),
			},
			"known string, incorrect": {
				Args: []cty.Value{
					cty.StringVal("hello"),
					cty.StringVal("hi"),
				},
				WantErr: `the actual value "hello" does not match the assumed value`,
			},
			"unknown map compared to object": {
				Args: []cty.Value{
					cty.UnknownVal(cty.Map(cty.String)),
					cty.ObjectVal(map[string]cty.Value{"greeting": cty.StringVal("hello")}),
				},
				Want: cty.ObjectVal(map[string]cty.Value{"greeting": cty.StringVal("hello")}),
			},
			"known map compared to object, correct": {
				Args: []cty.Value{
					cty.MapVal(map[string]cty.Value{"greeting": cty.StringVal("hello")}),
					cty.ObjectVal(map[string]cty.Value{"greeting": cty.StringVal("hello")}),
				},
				Want: cty.ObjectVal(map[string]cty.Value{"greeting": cty.StringVal("hello")}),
			},
			"known map compared to object, incorrect": {
				Args: []cty.Value{
					cty.MapVal(map[string]cty.Value{"greeting": cty.StringVal("howdy")}),
					cty.ObjectVal(map[string]cty.Value{"greeting": cty.StringVal("hello")}),
				},
				WantErr: `the actual value does not match the assumed value`,
			},
			"mismatching types with unknown value, convertable": {
				Args: []cty.Value{
					cty.UnknownVal(cty.Bool),
					cty.StringVal("true"),
				},
				Want: cty.StringVal("true"),
			},
			"mismatching types with known value, convertable": {
				Args: []cty.Value{
					cty.True,
					cty.StringVal("true"),
				},
				Want: cty.StringVal("true"),
			},
			"mismatching types with known value, unconvertable": {
				Args: []cty.Value{
					cty.ListValEmpty(cty.String),
					cty.StringVal("true"),
				},
				WantErr: `actual value type list of string does not match assumed value type string`,
			},
			"early failure due to other refinements": {
				Args: []cty.Value{
					cty.UnknownVal(cty.String).Refine().StringPrefix("arn:").NewValue(),
					cty.StringVal("does not start with arn:"),
				},
				// We can predict that the given value definitely won't equal
				// the expected value based on its known string prefix, even
				// though we don't know the entire string yet.
				WantErr: `the actual value (a string starting with "arn:") does not match the assumed value`,
			},
			"partially-unknown input that might match": {
				Args: []cty.Value{
					cty.ListVal([]cty.Value{cty.StringVal("a"), cty.UnknownVal(cty.String)}),
					cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
				},
				Want: cty.ListVal([]cty.Value{cty.StringVal("a"), cty.StringVal("b")}),
			},
			"partially-unknown input that cannot match": {
				Args: []cty.Value{
					cty.ListVal([]cty.Value{cty.StringVal("a"), cty.UnknownVal(cty.String)}),
					cty.ListVal([]cty.Value{cty.StringVal("not a"), cty.StringVal("b")}),
				},
				WantErr: `the actual value does not match the assumed value`,
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
