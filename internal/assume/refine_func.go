package assume

import (
	"math"
	"math/big"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var notnullFunc = makeRefineFunc(
	cty.DynamicPseudoType,
	"Assume that the given value will never be null.",
	nil,
	func(args []cty.Value, b *cty.RefinementBuilder) *cty.RefinementBuilder {
		return b.NotNull()
	},
)

var stringprefixFunc = makeRefineFunc(
	cty.String,
	"Assume that the given string will always have a fixed prefix.",
	nil,
	func(args []cty.Value, b *cty.RefinementBuilder) *cty.RefinementBuilder {
		prefix := args[0].AsString()
		return b.StringPrefix(prefix)
	},
	function.Parameter{
		Name:        "prefix",
		Type:        cty.String,
		Description: "The prefix to assume.",
	},
)

var listlengthFunc = makeCollectionLengthBoundsFunc(cty.List, "list")
var listlengthminFunc = makeCollectionLengthLowerBoundFunc(cty.List, "list")
var listlengthmaxFunc = makeCollectionLengthUpperBoundFunc(cty.List, "list")
var setlengthFunc = makeCollectionLengthBoundsFunc(cty.Set, "set")
var setlengthminFunc = makeCollectionLengthLowerBoundFunc(cty.Set, "set")
var setlengthmaxFunc = makeCollectionLengthUpperBoundFunc(cty.Set, "set")
var maplengthFunc = makeCollectionLengthBoundsFunc(cty.Map, "map")
var maplengthminFunc = makeCollectionLengthLowerBoundFunc(cty.Map, "map")
var maplengthmaxFunc = makeCollectionLengthUpperBoundFunc(cty.Map, "map")

func makeRefineFunc(typeConstraint cty.Type, desc string, checkArgs func([]cty.Value) error, refine func(args []cty.Value, b *cty.RefinementBuilder) *cty.RefinementBuilder, params ...function.Parameter) *function.Spec {
	spec := &function.Spec{
		Description: desc,
		Params: []function.Parameter{
			{
				Name:         "value",
				Type:         typeConstraint,
				Description:  "The value to make the assumption about.",
				AllowNull:    true,
				AllowUnknown: true,
			},
		},
		Type: function.StaticReturnType(typeConstraint),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			v := args[0]
			if checkArgs != nil {
				err := checkArgs(args[1:])
				if err, ok := err.(*function.ArgError); ok {
					err.Index++ // to account for the always-present extra "value" argument
					return cty.UnknownVal(v.Type()), err
				}
			}
			realRefine := func(b *cty.RefinementBuilder) *cty.RefinementBuilder {
				return refine(args[1:], b)
			}
			ret, ok := tryApplyRefinement(v, realRefine)
			if !ok {
				return cty.UnknownVal(v.Type()), function.NewArgErrorf(0, "assumption was not upheld")
			}
			return ret, nil
		},
	}
	spec.Params = append(spec.Params, params...)
	return spec
}

func makeCollectionLengthBoundsFunc(kind func(cty.Type) cty.Type, noun string) *function.Spec {
	return makeRefineFunc(
		kind(cty.DynamicPseudoType),
		"Assume that the given "+noun+" will have a length in the given bounds.",
		func(args []cty.Value) error {
			for i, v := range args {
				if v, acc := v.AsBigFloat().Int64(); acc != big.Exact || v >= math.MaxInt {
					return function.NewArgErrorf(i, "must be a whole number between 0 and %d", math.MaxInt)
				}
			}
			return nil
		},
		func(args []cty.Value, b *cty.RefinementBuilder) *cty.RefinementBuilder {
			// Our argument validator above already guaranteed that the two
			// arguments are whole numbers that can fit into an int.
			lower, _ := args[0].AsBigFloat().Int64()
			upper, _ := args[1].AsBigFloat().Int64()
			return b.CollectionLengthLowerBound(int(lower)).CollectionLengthUpperBound(int(upper))
		},
		function.Parameter{
			Name:        "min_length",
			Type:        cty.Number,
			Description: "The minimum possible " + noun + " length.",
		},
		function.Parameter{
			Name:        "max_length",
			Type:        cty.Number,
			Description: "The maximum possible " + noun + " length.",
		},
	)
}

func makeCollectionLengthLowerBoundFunc(kind func(cty.Type) cty.Type, noun string) *function.Spec {
	return makeRefineFunc(
		kind(cty.DynamicPseudoType),
		"Assume that the given "+noun+" will have a length of at least the given number.",
		func(args []cty.Value) error {
			if v, acc := args[0].AsBigFloat().Int64(); acc != big.Exact || v >= math.MaxInt {
				return function.NewArgErrorf(0, "must be a whole number between 0 and %d", math.MaxInt)
			}
			return nil
		},
		func(args []cty.Value, b *cty.RefinementBuilder) *cty.RefinementBuilder {
			// Our argument validator above already guaranteed that the
			// argument is a whole number that can fit into an int.
			bound, _ := args[0].AsBigFloat().Int64()
			return b.CollectionLengthLowerBound(int(bound))
		},
		function.Parameter{
			Name:        "min_length",
			Type:        cty.Number,
			Description: "The minimum possible " + noun + " length.",
		},
	)
}

func makeCollectionLengthUpperBoundFunc(kind func(cty.Type) cty.Type, noun string) *function.Spec {
	return makeRefineFunc(
		kind(cty.DynamicPseudoType),
		"Assume that the given "+noun+" will have a length of at most the given number.",
		func(args []cty.Value) error {
			if v, acc := args[0].AsBigFloat().Int64(); acc != big.Exact || v >= math.MaxInt {
				return function.NewArgErrorf(0, "must be a whole number between 0 and %d", math.MaxInt)
			}
			return nil
		},
		func(args []cty.Value, b *cty.RefinementBuilder) *cty.RefinementBuilder {
			// Our argument validator above already guaranteed that the
			// argument is a whole number that can fit into an int.
			bound, _ := args[0].AsBigFloat().Int64()
			return b.CollectionLengthUpperBound(int(bound))
		},
		function.Parameter{
			Name:        "max_length",
			Type:        cty.Number,
			Description: "The maximum possible " + noun + " length.",
		},
	)
}

func tryApplyRefinement(v cty.Value, refine func(b *cty.RefinementBuilder) *cty.RefinementBuilder) (result cty.Value, ok bool) {
	defer func() {
		if bad := recover(); bad != nil {
			result = cty.DynamicVal
			ok = false
		}
	}()

	// The following will panic if the given refinement isn't applicable to
	// the given value. Our defer function above will then recover and
	// arrange for this function to return false as its second result.
	return v.RefineWith(refine), true
}
