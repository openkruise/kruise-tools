/*
Copyright 2025 The Kruise Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package utils

import "testing"

func TestInt32Min(t *testing.T) {
	testCases := []struct {
		name     string
		a        int32
		items    []int32
		expected int32
	}{
		{
			name:     "No extra items",
			a:        10,
			items:    []int32{},
			expected: 10,
		},
		{
			name:     "All positive numbers",
			a:        10,
			items:    []int32{5, 20, 12},
			expected: 5,
		},
		{
			name:     "With negative numbers",
			a:        -5,
			items:    []int32{10, -2, -10},
			expected: -10,
		},
		{
			name:     "With zero",
			a:        1,
			items:    []int32{5, 0, 2},
			expected: 0,
		},
		{
			name:     "All numbers are the same",
			a:        7,
			items:    []int32{7, 7, 7},
			expected: 7,
		},
		{
			name:     "Initial value 'a' is the minimum",
			a:        3,
			items:    []int32{10, 5, 8},
			expected: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Int32Min(tc.a, tc.items...); got != tc.expected {
				t.Errorf("Int32Min() = %v, want %v", got, tc.expected)
			}
		})
	}
}
