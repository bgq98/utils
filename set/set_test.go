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

package set

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetx_Add(t *testing.T) {
	Addvals := []int{1, 2, 3, 1}
	s := NewMapSet[int](10)
	t.Run("Add", func(t *testing.T) {
		for _, val := range Addvals {
			s.Add(val)
		}
		assert.Equal(t, s.m, map[int]struct{}{
			1: {},
			2: {},
			3: {},
		})
	})
}

func TestSetx_Delete(t *testing.T) {
	testCases := []struct {
		name    string
		delVal  int
		setSet  map[int]struct{}
		wantSet map[int]struct{}
		isExist bool
	}{
		{
			name:   "delete val",
			delVal: 2,
			setSet: map[int]struct{}{
				2: struct{}{},
			},
			wantSet: map[int]struct{}{},
			isExist: true,
		},
		{
			name:   "deleted val not found",
			delVal: 3,
			setSet: map[int]struct{}{
				2: struct{}{},
			},
			wantSet: map[int]struct{}{
				2: struct{}{},
			},
			isExist: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := NewMapSet[int](10)
			s.m = tc.setSet
			s.Delete(tc.delVal)
			assert.Equal(t, tc.wantSet, s.m)
		})
	}
}

func TestSetx_IsExist(t *testing.T) {
	s := NewMapSet[int](10)
	s.Add(1)
	testcases := []struct {
		name    string
		val     int
		isExist bool
	}{
		{
			name:    "found",
			val:     1,
			isExist: true,
		},
		{
			name:    "not fonud",
			val:     2,
			isExist: false,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ok := s.Exist(tc.val)
			assert.Equal(t, tc.isExist, ok)
		})
	}
}

func TestSetx_Values(t *testing.T) {
	s := NewMapSet[int](10)
	testcases := []struct {
		name    string
		setSet  map[int]struct{}
		wantval map[int]struct{}
	}{
		{
			name: "found values",
			setSet: map[int]struct{}{
				1: struct{}{},
				2: struct{}{},
				3: struct{}{},
			},
			wantval: map[int]struct{}{
				1: struct{}{},
				2: struct{}{},
				3: struct{}{},
			},
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s.m = tc.setSet
			vals := s.Keys()
			ok := equal(vals, tc.wantval)
			assert.Equal(t, true, ok)
		})
	}
}

func equal(nums []int, m map[int]struct{}) bool {
	for _, num := range nums {
		_, ok := m[num]
		if !ok {
			return false
		}
		delete(m, num)
	}
	return true && len(m) == 0
}
