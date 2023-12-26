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

// FilterMap 执行过滤并且转化
// 如果 m 的第二个返回值是 false ,那么我们会忽略第一个返回值
func FilterMap[Src interface{}, Dst interface{}](src []Src, m func(idx int, src Src) (Dst, bool)) []Dst {
	res := make([]Dst, 0, len(src))
	for i, s := range src {
		dst, ok := m(i, s)
		if ok {
			res = append(res, dst)
		}
	}
	return res
}

// Map 执行转化
func Map[Src interface{}, Dst interface{}](src []Src, m func(idx int, src Src) Dst) []Dst {
	res := make([]Dst, len(src))
	for i, s := range src {
		res[i] = m(i, s)
	}
	return res
}

// toMap 构造 map
func toMap[T comparable](src []T) map[T]struct{} {
	var dataMap = make(map[T]struct{}, len(src))
	for _, v := range src {
		dataMap[v] = struct{}{}
	}
	return dataMap
}

func deduplicateFunc[T interface{}](data []T, equal equalFunc[T]) []T {
	var newData = make([]T, 0, len(data))
	for k, v := range data {
		if !ContainsFunc[T](data[k+1:], func(src T) bool {
			return equal(src, v)
		}) {
			newData = append(newData, v)
		}
	}
	return newData
}

func deduplicate[T comparable](data []T) []T {
	dataMap := toMap[T](data)
	var newData = make([]T, 0, len(dataMap))
	for k := range dataMap {
		newData = append(newData, k)
	}
	return newData
}
