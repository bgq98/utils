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

type Set[T comparable] interface {
	Add(key T)
	Delete(key T)
	Exist(key T) bool
	Keys() []T
}

type MapSet[T comparable] struct {
	m map[T]struct{}
}

func NewMapSet[T comparable](size int) *MapSet[T] {
	return &MapSet[T]{
		m: make(map[T]struct{}, size),
	}
}

func (m *MapSet[T]) Add(key T) {
	m.m[key] = struct{}{}
}

func (m *MapSet[T]) Delete(key T) {
	delete(m.m, key)
}

func (m *MapSet[T]) Exist(key T) bool {
	_, ok := m.m[key]
	return ok
}

func (m *MapSet[T]) Keys() []T {
	ans := make([]T, 0, len(m.m))
	for keys := range m.m {
		ans = append(ans, keys)
	}
	return ans
}
