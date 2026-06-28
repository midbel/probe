package probe

import "fmt"

type Type int8

const (
	Scalar Type = 1 << iota
	Object
	Array
)

type PathInfo struct {
	Path  string
	Depth int
	Type  Type
}

func Discover(in any) []PathInfo {
	root := PathInfo{
		Depth: 0,
	}
	return discover(in, root)
}

func discover(in any, parent PathInfo) []PathInfo {
	switch in := in.(type) {
	default:
		parent.Type = Scalar
		return []PathInfo{parent}
	case []any:
		parent.Type = Array
		return discoverArray(in, parent)
	case map[string]any:
		parent.Type = Object
		return discoverObject(in, parent)
	}
}

func discoverObject(in map[string]any, parent PathInfo) []PathInfo {
	var (
		seen  = make(map[string]struct{})
		paths []PathInfo
	)
	for k, v := range in {
		if _, ok := v.([]any); ok {
			k = fmt.Sprintf("%s[*]", k)
		}
		var (
			root = PathInfo{
				Path: k,
				Depth: parent.Depth+1,
			}
			others = discover(v, root)
		)
		for i := range others {
			others[i].Path = joinPath(parent.Path, others[i].Path)
		}
		paths = append(paths, uniquePaths(others, seen)...)
	}
	return paths
}

func discoverArray(in []any, parent PathInfo) []PathInfo {
	var (
		seen  = make(map[string]struct{})
		paths []PathInfo
	)
	for i := range in {
		others := discover(in[i], parent)
		paths = append(paths, uniquePaths(others, seen)...)
	}
	return paths
}

func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func uniquePaths(paths []PathInfo, seen map[string]struct{}) []PathInfo {
	var all []PathInfo
	for _, p := range paths {
		if _, ok := seen[p.Path]; ok {
			continue
		}
		seen[p.Path] = struct{}{}
		all = append(all, p)
	}
	return all
}
