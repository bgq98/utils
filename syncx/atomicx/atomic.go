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

package atomicx

import "sync/atomic"

// value 是对 atomic.Value 的泛型封装

type Value[T any] struct {
	val atomic.Value
}

func NewValue[T any]() *Value[T] {
	var t T
	return NewValueOf[T](t)
}

func NewValueOf[T any](t T) *Value[T] {
	val := atomic.Value{}
	val.Store(t)
	return &Value[T]{
		val: val,
	}
}

func (v *Value[T]) Load() (val T) {
	data := v.val.Load()
	val = data.(T)
	return val
}

func (v *Value[T]) Store(val T) {
	v.val.Store(val)
}

func (v *Value[T]) Swap(new T) (old T) {
	data := v.val.Swap(new)
	old = data.(T)
	return
}

func (v *Value[T]) CompareAndSwap(old, new T) (swapped bool) {
	return v.val.CompareAndSwap(old, new)
}
