package testtypes

import "strconv"

// This file contains code copied from github.com/gogo/protobuf/proto
// so that we can avoid that dependency.

// ProtoEnumName is a helper function to simplify printing protocol buffer enums
// by name.  Given an enum map and a value, it returns a useful string.
func ProtoEnumName(m map[int32]string, v int32) string {
	s, ok := m[v]
	if ok {
		return s
	}
	return strconv.Itoa(int(v))
}

type ProtoMessage struct{}
