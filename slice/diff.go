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

// DiffSet 差集,只支持 comparable 类型。已去重
func DiffSet[T comparable](src, dst []T) []T {
	srcMap := toMap[T](src)
	for _, val := range dst {
		delete(srcMap, val)
	}
	var res = make([]T, 0, len(srcMap))
	for keys := range srcMap {
		res = append(res, keys)
	}
	return res
}

// DiffSetFunc 差集。已去重
// 优先使用 DiffSet
func DiffSetFunc[T interface{}](src, dst []T, equal equalFunc[T]) []T {
	var res = make([]T, 0, len(src))
	for _, val := range src {
		if !ContainsFunc[T](dst, func(src T) bool {
			return equal(src, val)
		}) {
			res = append(res, val)
		}
	}
	return deduplicateFunc[T](res, equal)
}
