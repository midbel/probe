package probe

import (
	"errors"
	"fmt"
)

var errArgument = errors.New("invalid number of argument given")

func invalidArgs(msg string, n int) error {
	return fmt.Errorf("%w: %s - %d given", errArgument, msg, n)
}

var builtins map[string]func(any, []Expr) (any, error)

var predicables = []string{
	"eq",
	"ne",
	"lt",
	"le",
	"gt",
	"ge",
	"between",
	"literal",
	"number",
	"string",
	"boolean",
	"object",
	"array",
	"in",
}

func init() {
	builtins = map[string]func(any, []Expr) (any, error){
		"as":         runAs,
		"len":        runLen,
		"at":         runAt,
		"first":      runFirst,
		"last":       runLast,
		"range":      runRange,
		"filter":     runFilter,
		"some":       runSome,
		"every":      runEvery,
		"entries":    runEntries,
		"keys":       runKeys,
		"values":     runValues,
		"flatten":    runFlatten,
		"reshape":    runReshape,
		"zip":        runZip,
		"ziplongest": runZipLongest,
		"reverse":    runReverse,
		"default":    runDefault,
		"not":        runNot,
		"eq":         runEqual,
		"ne":         runNotEqual,
		"lt":         runLesserThan,
		"le":         runLesserEq,
		"gt":         runGreaterThan,
		"ge":         runGreaterEq,
		"between":    runBetween,
		"literal":    runLiteral,
		"number":     runNumber,
		"string":     runString,
		"boolean":    runBoolean,
		"object":     runObject,
		"array":      runArray,
		"in":         runIn,
		"ifeq":       runIfEqual,
		"ifne":       runIfNotEqual,
		"ifexists":   runIfExists,
		"exists":     runExists,
		"empty":      runEmpty,
		"null":       runNull,
	}
}

// :as()
func runAs(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("as takes only one argument", len(args))
	}
	target, err := getStrFromExpr(args[0])
	if err != nil {
		return nil, err
	}
	switch target {
	case "string":
		val, err = castToString(val)
	case "number":
		val, err = castToNumber(val)
	case "bool":
		val, err = castToBool(val)
	default:
		return nil, fmt.Errorf("%s: unknown target type", target)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: value can not be converted to target type %s", ErrType, target)
	}
	return val, nil
}

// :len, :length
func runLen(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("length takes no argument(s)", len(args))
	}
	var x int
	switch arr := val.(type) {
	case []any:
		x = len(arr)
	case map[string]any:
		x = len(arr)
	default:
		return nil, compositeExpected("length")
	}
	return float64(x), nil
}

// :at()
func runAt(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("at takes only one argument", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, arrayExpected("at")
	}
	ix, err := getIntFromExpr(args[0])
	if err != nil {
		return nil, err
	}
	if ix < 0 || ix >= len(arr) {
		return missed, nil
	}
	return arr[ix], nil
}

func runRange(val any, args []Expr) (any, error) {
	if len(args) != 2 {
		return nil, invalidArgs("range takes only two arguments", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, arrayExpected("at")
	}
	fix, err := getIntFromExpr(args[0])
	if err != nil {
		return nil, err
	}
	if fix < 0 || fix >= len(arr) {
		return missed, nil
	}
	tix, err := getIntFromExpr(args[1])
	if err != nil {
		return nil, err
	}
	if tix < 0 || tix >= len(arr) || fix > tix {
		return missed, nil
	}
	return arr[fix:tix], nil
}

// arr:filter()
func runFilter(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("filter takes one argument", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, arrayExpected("filter")
	}
	fn, ok := args[0].(lambda)
	if !ok {
		return nil, fmt.Errorf("argument to filter should be a lambda")
	}
	var tmp []any
	for i := range arr {
		ok = fn.isDefined(arr[i])
		if ok {
			tmp = append(tmp, arr[i])
		}
	}
	return tmp, nil
}

// returns val if some values in val pass the predicate otherwise discarded is returned
// not the same as classic "some" where a boolean value is returned if some values pass
func runSome(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("some takes one argument", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, arrayExpected("some")
	}
	fn, ok := args[0].(lambda)
	if !ok {
		return nil, fmt.Errorf("argument to some should be a lambda")
	}
	for i := range arr {
		ok = fn.isDefined(arr[i])
		if ok {
			break
		}
	}
	if !ok {
		return discarded, nil
	}
	return val, nil
}

// returns val if all values in val pass the predicate otherwise discarded is returned
// not the same as classic "all" where a boolean value is returned if all values pass
func runEvery(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("every takes one argument", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, arrayExpected("every")
	}
	fn, ok := args[0].(lambda)
	if !ok {
		return nil, fmt.Errorf("argument to some should be a lambda")
	}
	for i := range arr {
		ok = fn.isDefined(arr[i])
		if !ok {
			return discarded, nil
		}
	}
	return val, nil
}

func runEntries(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("entries takes not argument(s)", len(args))
	}
	var res any
	switch arr := val.(type) {
	case []any:
		val := make([]any, 0, len(arr))
		for i, v := range arr {
			val = append(val, []any{float64(i), v})
		}
		res = val
	case map[string]any:
		val := make([]any, 0, len(arr))
		for k, v := range arr {
			val = append(val, []any{k, v})
		}
		res = val
	default:
		return nil, compositeExpected("entries")
	}
	return res, nil
}

func runKeys(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("keys takes not argument(s)", len(args))
	}
	var res any
	switch arr := val.(type) {
	case []any:
		val := make([]any, 0, len(arr))
		for i := range arr {
			val = append(val, float64(i))
		}
		res = val
	case map[string]any:
		val := make([]any, 0, len(arr))
		for k := range arr {
			val = append(val, k)
		}
		res = val
	default:
		return nil, compositeExpected("keys")
	}
	return res, nil
}

func runValues(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("values takes not argument(s)", len(args))
	}
	var res any
	switch arr := val.(type) {
	case []any:
		val := make([]any, 0, len(arr))
		for i := range arr {
			val = append(val, arr[i])
		}
		res = val
	case map[string]any:
		val := make([]any, 0, len(arr))
		for k := range arr {
			val = append(val, arr[k])
		}
		res = val
	default:
		return nil, compositeExpected("values")
	}
	return res, nil
}

func runFlatten(val any, args []Expr) (any, error) {
	if len(args) > 1 {
		return nil, invalidArgs("flatten takes zero or one argument", len(args))
	}
	var maxDepth int
	if len(args) == 1 {
		md, err := getIntFromExpr(args[0])
		if err != nil {
			return nil, err
		}
		maxDepth = md
	}
	return flatten(val, maxDepth), nil
}

// :reshape(rows, cols)
func runReshape(val any, args []Expr) (any, error) {
	if len(args) != 2 {
		return nil, invalidArgs("reshape takes two arguments", len(args))
	}
	rows, err := getIntFromExpr(args[0])
	if err != nil {
		return nil, err
	}
	if rows <= 0 {
		return nil, fmt.Errorf("reshape: rows must be a positive integer")
	}
	cols, err := getIntFromExpr(args[1])
	if err != nil {
		return nil, err
	}
	if cols <= 0 {
		return nil, fmt.Errorf("reshape: cols must be a positive integer")
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, arrayExpected("reshape")
	}
	return reshape(arr, rows, cols), nil
}

// :zip()
func runZip(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("zip takes no argument(s)", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, arrayExpected("zip")
	}
	return zip(arr), nil
}

// :ziplongest()
func runZipLongest(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("ziplongest takes no argument(s)", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, arrayExpected("ziplongest")
	}
	return zipLongest(arr), nil
}

func runReverse(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("reverse takes no argument(s)", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		return nil, arrayExpected("reverse")
	}
	return reverse(arr), nil
}

// :first()
func runFirst(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("first takes not argument(s)", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		if isMissing(val) {
			return val, nil
		}
		return nil, arrayExpected("first")
	}
	if len(arr) == 0 {
		return missed, nil
	}
	return arr[0], nil
}

// :last()
func runLast(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("last takes not argument(s)", len(args))
	}
	arr, ok := val.([]any)
	if !ok {
		if isMissing(val) {
			return val, nil
		}
		return nil, arrayExpected("last")
	}
	if len(arr) == 0 {
		return missed, nil
	}
	return arr[len(arr)-1], nil
}

// :default()
func runDefault(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("default takes only one argument", len(args))
	}
	if isDefined(val) {
		return val, nil
	}
	return getAnyFromExpr(args[0])
}

// :not()
func runNot(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("not takes no argument(s)", len(args))
	}
	return !isDefined(val), nil
}

// :eq()
func runEqual(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("eq takes only one argument", len(args))
	}
	ok := isEqual(val, args[0])
	if !ok || isMissing(val) {
		return discarded, nil
	}
	return val, nil
}

// :ne()
func runNotEqual(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("ne takes only one argument", len(args))
	}
	if isMissing(val) {
		return discarded, nil
	}
	ok := !isEqual(val, args[0])
	if !ok {
		return discarded, nil
	}
	return val, nil
}

// :lt
func runLesserThan(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("lt takes only one argument", len(args))
	}
	if isMissing(val) {
		return discarded, nil
	}
	ok := isLess(val, args[0])
	if !ok {
		return discarded, nil
	}
	return val, nil
}

// :le
func runLesserEq(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("le takes only one argument", len(args))
	}
	if isMissing(val) {
		return discarded, nil
	}
	ok := isLess(val, args[0]) || isEqual(val, args[0])
	if !ok {
		return discarded, nil
	}
	return val, nil
}

// :gt
func runGreaterThan(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("gt takes only one argument", len(args))
	}
	if isMissing(val) {
		return discarded, nil
	}
	ok := !isLess(val, args[0]) && !isEqual(val, args[0])
	if !ok {
		return discarded, nil
	}
	return val, nil
}

// :ge
func runGreaterEq(val any, args []Expr) (any, error) {
	if len(args) != 1 {
		return nil, invalidArgs("ge takes only one argument", len(args))
	}
	if isMissing(val) {
		return discarded, nil
	}
	ok := isEqual(val, args[0]) || !isLess(val, args[0])
	if !ok {
		return discarded, nil
	}
	return val, nil
}

// :between
func runBetween(val any, args []Expr) (any, error) {
	if len(args) != 2 {
		return nil, invalidArgs("between takes two arguments", len(args))
	}
	if isMissing(val) {
		return discarded, nil
	}
	if isEqual(val, args[0]) || isEqual(val, args[1]) {
		return val, nil
	}
	if isLess(val, args[1]) && !isLess(val, args[0]) {
		return val, nil
	}
	return discarded, nil
}

// :in
func runIn(val any, args []Expr) (any, error) {
	if len(args) == 0 {
		return nil, invalidArgs("in takes at least one argument", len(args))
	}
	if isMissing(val) {
		return discarded, nil
	}
	for i := range args {
		if isEqual(val, args[i]) {
			return val, nil
		}
	}
	return discarded, nil
}

// :ifeq
func runIfEqual(val any, args []Expr) (any, error) {
	if len(args) != 3 {
		return nil, invalidArgs("ifeq takes exactly three arguments", len(args))
	}
	if isEqual(val, args[0]) {
		return getAnyFromExpr(args[1])
	}
	return getAnyFromExpr(args[2])
}

// :ifne
func runIfNotEqual(val any, args []Expr) (any, error) {
	if len(args) != 3 {
		return nil, invalidArgs("ifne takes exactly three arguments", len(args))
	}
	if !isEqual(val, args[0]) {
		return getAnyFromExpr(args[1])
	}
	return getAnyFromExpr(args[2])
}

// :ifexists
func runIfExists(val any, args []Expr) (any, error) {
	if len(args) < 2 {
		return nil, invalidArgs("ifexists takes at least two arguments", len(args))
	}
	if len(args) == 2 {
		if isDefined(val) {
			return getAnyFromExpr(args[0])
		}
		return getAnyFromExpr(args[1])
	}
	var ok bool
	switch arr := val.(type) {
	case []any:
		ix, err := getIntFromExpr(args[0])
		if err != nil {
			return nil, err
		}
		ok = ix >= 0 && ix <= len(arr)
	case map[string]any:
		key, err := getStrFromExpr(args[0])
		if err != nil {
			return nil, err
		}
		_, ok = arr[key]
	default:
		return nil, compositeExpected("ifexists")
	}
	if ok {
		return getAnyFromExpr(args[1])
	}
	return getAnyFromExpr(args[2])
}

// :exists
func runExists(val any, args []Expr) (any, error) {
	if len(args) == 0 {
		return isDefined(val), nil
	}
	if len(args) != 1 {
		return nil, invalidArgs("exists takes zero or one argument(s)", len(args))
	}
	switch arr := val.(type) {
	case []any:
		ix, err := getIntFromExpr(args[0])
		if err != nil {
			return nil, err
		}
		if ix >= 0 && ix < len(arr) {
			return arr[ix], nil
		}
		return discarded, nil
	case map[string]any:
		key, err := getStrFromExpr(args[0])
		if err != nil {
			return nil, err
		}
		val, ok := arr[key]
		if ok {
			return val, nil
		}
		return discarded, nil
	default:
		return nil, compositeExpected("exists")
	}
}

func runLiteral(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("literal takes no argument(s)", len(args))
	}
	switch val.(type) {
	case string:
	case float64:
	default:
		return discarded, nil
	}
	return val, nil
}

func runNumber(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("number takes no argument(s)", len(args))
	}
	if _, ok := val.(float64); !ok {
		return discarded, nil
	}
	return val, nil
}

func runString(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("string takes no argument(s)", len(args))
	}
	if _, ok := val.(string); !ok {
		return discarded, nil
	}
	return val, nil
}

func runBoolean(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("boolean takes no argument(s)", len(args))
	}
	if _, ok := val.(bool); !ok {
		return discarded, nil
	}
	return val, nil
}

func runObject(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("object takes no argument(s)", len(args))
	}
	if _, ok := val.(map[string]any); !ok {
		return discarded, nil
	}
	return val, nil
}

func runArray(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("array takes no argument(s)", len(args))
	}
	if _, ok := val.([]any); !ok {
		return discarded, nil
	}
	return val, nil
}

// :empty
func runEmpty(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("empty takes no argument(s)", len(args))
	}
	switch arr := val.(type) {
	case []any:
		if len(arr) == 0 {
			return arr, nil
		}
		return discarded, nil
	case map[string]any:
		if len(arr) == 0 {
			return arr, nil
		}
		return discarded, nil
	case nil:
		return nil, nil
	default:
		return discarded, nil
	}
}

// :null
func runNull(val any, args []Expr) (any, error) {
	if len(args) != 0 {
		return nil, invalidArgs("null takes not argument(s)", len(args))
	}
	if val == nil {
		return val, nil
	}
	return discarded, nil
}
