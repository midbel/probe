package probe

import (
	"slices"
)

func flatten(in any, maxDepth int) []any {
	var (
		do  func(any, int)
		tmp []any
	)
	do = func(in any, depth int) {
		switch in := in.(type) {
		case []any:
			for i := range in {
				if maxDepth > 0 && depth >= maxDepth {
					tmp = append(tmp, in[i])
				} else {
					do(in[i], depth+1)
				}
			}
		default:
			tmp = append(tmp, in)
		}
	}
	do(in, 0)
	return tmp
}

func reshape(in []any, rows, cols int) []any {
	if len(in) <= 1 {
		return in
	}
	var (
		arr = flatten(in, 0)
		out = make([]any, rows)
	)
	for i := range rows {
		var (
			start = cols * i
			end   = start + cols
		)
		tmp := make([]any, cols)
		if start < len(arr) {
			if end >= len(arr) {
				end = len(arr)
			}
			copy(tmp, arr[start:end])
		}
		out[i] = tmp
	}
	return out
}

func zip(in []any) []any {
	if len(in) <= 1 {
		return in
	}
	var (
		out   = make([]any, 0, len(in))
		tmp   = make([]any, len(in))
		limit int
	)
	copy(tmp, in)
	for i := range tmp {
		a, ok := tmp[i].([]any)
		if !ok {
			limit = 1
			tmp[i] = []any{tmp[i]}
			continue
		}
		if i == 0 {
			limit = len(a)
		}
		limit = min(limit, len(a))
	}
	for i := 0; i < limit; i++ {
		res := make([]any, 0, limit)
		for j := range tmp {
			a, ok := tmp[j].([]any)
			if !ok {
				res = append(res, nil)
			} else {
				res = append(res, a[i])
			}
		}
		out = append(out, res)
	}
	return out
}

func zipLongest(in []any) []any {
	if len(in) <= 1 {
		return in
	}
	var (
		out   = make([]any, 0, len(in))
		tmp   = make([]any, len(in))
		limit int
	)
	copy(tmp, in)
	for i := range tmp {
		a, ok := tmp[i].([]any)
		if !ok {
			tmp[i] = []any{tmp[i]}
		}
		limit = max(limit, len(a))
	}
	for i := 0; i < limit; i++ {
		res := make([]any, 0, limit)
		for j := range tmp {
			a, ok := tmp[j].([]any)
			if !ok || i >= len(a) {
				res = append(res, nil)
			} else {
				res = append(res, a[i])
			}
		}
		out = append(out, res)
	}
	return out
}

func reverse(in []any) []any {
	if len(in) <= 1 {
		return in
	}
	slices.Reverse(in)
	return in
}
