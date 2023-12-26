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

package events

type InconsistentEvent struct {
	Id        int64
	Direction string // 以哪个源为准 SRC 源表为准 DST 目标表为准
	Type      string
}

const (
	// InconsistentEventTypeTargetMissing target 中没有数据
	InconsistentEventTypeTargetMissing = "target_missing"
	// InconsistentEventTypeNotEqual 目标表和源表的数据不相等
	InconsistentEventTypeNotEqual = "neq"
	// InconsistentEventTypeBaseMissing base 中没有数据
	InconsistentEventTypeBaseMissing = "base_missing"
)
