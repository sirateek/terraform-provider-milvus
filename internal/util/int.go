// Copyright Siratee K. 2026
// SPDX-License-Identifier: MIT

package util

import "fmt"

// ToInt64 is a helper to convert any to int64.
func ToInt64(v any) (int64, error) {
	switch val := v.(type) {
	case int64:
		return val, nil
	case int:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case string:
		var i int64
		_, err := fmt.Sscanf(val, "%d", &i)
		return i, err
	default:
		return 0, fmt.Errorf("cannot convert %v to int64", v)
	}
}
