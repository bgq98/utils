/*
   Copyright 2023 bgq98

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

package slice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	testCase := []struct {
		name string
		src  []int
		dst  int
		want bool
	}{
		{
			name: "dst exist",
			src:  []int{1, 2, 3, 4, 5},
			dst:  1,
			want: true,
		},
		{
			name: "dst not exist",
			src:  []int{1, 2, 3, 4, 5},
			dst:  6,
			want: false,
		},
		{
			name: "src nil",
			dst:  3,
			want: false,
		},
		{
			name: "length of src is 0",
			src:  []int{},
			dst:  4,
			want: false,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, Contains[int](tc.src, tc.dst))
		})
	}
}

func TestContainsFunc(t *testing.T) {
	testCase := []struct {
		name string
		src  []int
		dst  int
		want bool
	}{
		{
			name: "dst exist",
			src:  []int{1, 2, 3, 4, 5},
			dst:  1,
			want: true,
		},
		{
			name: "dst not exist",
			src:  []int{1, 2, 3, 4, 5},
			dst:  6,
			want: false,
		},
		{
			name: "src nil",
			dst:  3,
			want: false,
		},
		{
			name: "length of src is 0",
			src:  []int{},
			dst:  4,
			want: false,
		},
	}
	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, ContainsFunc[int](tc.src, func(src int) bool {
				return src == tc.dst
			}))
		})
	}
}
