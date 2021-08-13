/*
 *
 * Copyright 2021 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package xslice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt8(t *testing.T) {
	s1 := []int8{-1, 3, 4, 127}
	s2 := []int8{-1, 3, 5, 0}

	assert.Equal(t, true, ContainInt8(s1, 3))
	assert.Equal(t, []int8{4, 127}, DiffInt8(s1, s2))
	assert.Equal(t, []int8{5, 0}, DiffInt8(s2, s1))
	assert.Equal(t, []int8{-1, 3}, IntersectInt8(s1, s2))
	assert.Equal(t, []int8{-1, 3, 4, 127, -1, 3, 5, 0}, MergeInt8(s1, s2))
	assert.Equal(t, []int8{-1, 5, 0}, RemoveInt8(s2, 3))
	assert.Equal(t, []int8{0, 5, 3, -1}, ReverseInt8(s2))
	assert.Equal(t, []int8{-1, 3, 4, 127}, SortInt8(s1))
	assert.Equal(t, []int8{-1, -1, 0, 3, 3, 4, 5, 127}, SortInt8(MergeInt8(s1, s2)))
	assert.ElementsMatch(t, []int8{-1, 0, 3, 4, 5, 127}, UniqueInt8(MergeInt8(s1, s2)))
}

func TestInt16(t *testing.T) {
	s1 := []int16{-1, 3, 4, 127}
	s2 := []int16{-1, 3, 5, 0}

	assert.Equal(t, true, ContainInt16(s1, 3))
	assert.Equal(t, []int16{4, 127}, DiffInt16(s1, s2))
	assert.Equal(t, []int16{5, 0}, DiffInt16(s2, s1))
	assert.Equal(t, []int16{-1, 3}, IntersectInt16(s1, s2))
	assert.Equal(t, []int16{-1, 3, 4, 127, -1, 3, 5, 0}, MergeInt16(s1, s2))
	assert.Equal(t, []int16{-1, 5, 0}, RemoveInt16(s2, 3))
	assert.Equal(t, []int16{0, 5, 3, -1}, ReverseInt16(s2))
	assert.Equal(t, []int16{-1, 3, 4, 127}, SortInt16(s1))
	assert.Equal(t, []int16{-1, -1, 0, 3, 3, 4, 5, 127}, SortInt16(MergeInt16(s1, s2)))
	assert.ElementsMatch(t, []int16{-1, 0, 3, 4, 5, 127}, UniqueInt16(MergeInt16(s1, s2)))
}

func TestInt(t *testing.T) {
	s1 := []int{-1, 3, 4, 127}
	s2 := []int{-1, 3, 5, 0}

	assert.Equal(t, true, ContainInt(s1, 3))
	assert.Equal(t, []int{4, 127}, DiffInt(s1, s2))
	assert.Equal(t, []int{5, 0}, DiffInt(s2, s1))
	assert.Equal(t, []int{-1, 3}, IntersectInt(s1, s2))
	assert.Equal(t, []int{-1, 3, 4, 127, -1, 3, 5, 0}, MergeInt(s1, s2))
	assert.Equal(t, []int{-1, 5, 0}, RemoveInt(s2, 3))
	assert.Equal(t, []int{0, 5, 3, -1}, ReverseInt(s2))
	assert.Equal(t, []int{-1, 3, 4, 127}, SortInt(s1))
	assert.Equal(t, []int{-1, -1, 0, 3, 3, 4, 5, 127}, SortInt(MergeInt(s1, s2)))
	assert.ElementsMatch(t, []int{-1, 0, 3, 4, 5, 127}, UniqueInt(MergeInt(s1, s2)))
}

func TestInt32(t *testing.T) {
	s1 := []int32{-1, 3, 4, 127}
	s2 := []int32{-1, 3, 5, 0}

	assert.Equal(t, true, ContainInt32(s1, 3))
	assert.Equal(t, []int32{4, 127}, DiffInt32(s1, s2))
	assert.Equal(t, []int32{5, 0}, DiffInt32(s2, s1))
	assert.Equal(t, []int32{-1, 3}, IntersectInt32(s1, s2))
	assert.Equal(t, []int32{-1, 3, 4, 127, -1, 3, 5, 0}, MergeInt32(s1, s2))
	assert.Equal(t, []int32{-1, 5, 0}, RemoveInt32(s2, 3))
	assert.Equal(t, []int32{0, 5, 3, -1}, ReverseInt32(s2))
	assert.Equal(t, []int32{-1, 3, 4, 127}, SortInt32(s1))
	assert.Equal(t, []int32{-1, -1, 0, 3, 3, 4, 5, 127}, SortInt32(MergeInt32(s1, s2)))
	assert.ElementsMatch(t, []int32{-1, 0, 3, 4, 5, 127}, UniqueInt32(MergeInt32(s1, s2)))
}

func TestInt64(t *testing.T) {
	s1 := []int64{-1, 3, 4, 127}
	s2 := []int64{-1, 3, 5, 0}

	assert.Equal(t, true, ContainInt64(s1, 3))
	assert.Equal(t, []int64{4, 127}, DiffInt64(s1, s2))
	assert.Equal(t, []int64{5, 0}, DiffInt64(s2, s1))
	assert.Equal(t, []int64{-1, 3}, IntersectInt64(s1, s2))
	assert.Equal(t, []int64{-1, 3, 4, 127, -1, 3, 5, 0}, MergeInt64(s1, s2))
	assert.Equal(t, []int64{-1, 5, 0}, RemoveInt64(s2, 3))
	assert.Equal(t, []int64{0, 5, 3, -1}, ReverseInt64(s2))
	assert.Equal(t, []int64{-1, 3, 4, 127}, SortInt64(s1))
	assert.Equal(t, []int64{-1, -1, 0, 3, 3, 4, 5, 127}, SortInt64(MergeInt64(s1, s2)))
	assert.ElementsMatch(t, []int64{-1, 0, 3, 4, 5, 127}, UniqueInt64(MergeInt64(s1, s2)))
}

func TestUint8(t *testing.T) {
	s1 := []uint8{1, 3, 4, 127}
	s2 := []uint8{1, 3, 5, 0}

	assert.Equal(t, true, ContainUint8(s1, 3))
	assert.Equal(t, []uint8{4, 127}, DiffUint8(s1, s2))
	assert.Equal(t, []uint8{5, 0}, DiffUint8(s2, s1))
	assert.Equal(t, []uint8{1, 3}, IntersectUint8(s1, s2))
	assert.Equal(t, []uint8{1, 3, 4, 127, 1, 3, 5, 0}, MergeUint8(s1, s2))
	assert.Equal(t, []uint8{1, 5, 0}, RemoveUint8(s2, 3))
	assert.Equal(t, []uint8{0, 5, 3, 1}, ReverseUint8(s2))
	assert.Equal(t, []uint8{1, 3, 4, 127}, SortUint8(s1))
	assert.Equal(t, []uint8{0, 1, 1, 3, 3, 4, 5, 127}, SortUint8(MergeUint8(s1, s2)))
	assert.ElementsMatch(t, []uint8{1, 0, 3, 4, 5, 127}, UniqueUint8(MergeUint8(s1, s2)))
}

func TestUint16(t *testing.T) {
	s1 := []uint16{1, 3, 4, 127}
	s2 := []uint16{1, 3, 5, 0}

	assert.Equal(t, true, ContainUint16(s1, 3))
	assert.Equal(t, []uint16{4, 127}, DiffUint16(s1, s2))
	assert.Equal(t, []uint16{5, 0}, DiffUint16(s2, s1))
	assert.Equal(t, []uint16{1, 3}, IntersectUint16(s1, s2))
	assert.Equal(t, []uint16{1, 3, 4, 127, 1, 3, 5, 0}, MergeUint16(s1, s2))
	assert.Equal(t, []uint16{1, 5, 0}, RemoveUint16(s2, 3))
	assert.Equal(t, []uint16{0, 5, 3, 1}, ReverseUint16(s2))
	assert.Equal(t, []uint16{1, 3, 4, 127}, SortUint16(s1))
	assert.Equal(t, []uint16{0, 1, 1, 3, 3, 4, 5, 127}, SortUint16(MergeUint16(s1, s2)))
	assert.ElementsMatch(t, []uint16{1, 0, 3, 4, 5, 127}, UniqueUint16(MergeUint16(s1, s2)))
}

func TestUint(t *testing.T) {
	s1 := []uint{1, 3, 4, 127}
	s2 := []uint{1, 3, 5, 0}

	assert.Equal(t, true, ContainUint(s1, 3))
	assert.Equal(t, []uint{4, 127}, DiffUint(s1, s2))
	assert.Equal(t, []uint{5, 0}, DiffUint(s2, s1))
	assert.Equal(t, []uint{1, 3}, IntersectUint(s1, s2))
	assert.Equal(t, []uint{1, 3, 4, 127, 1, 3, 5, 0}, MergeUint(s1, s2))
	assert.Equal(t, []uint{1, 5, 0}, RemoveUint(s2, 3))
	assert.Equal(t, []uint{0, 5, 3, 1}, ReverseUint(s2))
	assert.Equal(t, []uint{1, 3, 4, 127}, SortUint(s1))
	assert.Equal(t, []uint{0, 1, 1, 3, 3, 4, 5, 127}, SortUint(MergeUint(s1, s2)))
	assert.ElementsMatch(t, []uint{1, 0, 3, 4, 5, 127}, UniqueUint(MergeUint(s1, s2)))
}

func TestUint32(t *testing.T) {
	s1 := []uint32{1, 3, 4, 127}
	s2 := []uint32{1, 3, 5, 0}

	assert.Equal(t, true, ContainUint32(s1, 3))
	assert.Equal(t, []uint32{4, 127}, DiffUint32(s1, s2))
	assert.Equal(t, []uint32{5, 0}, DiffUint32(s2, s1))
	assert.Equal(t, []uint32{1, 3}, IntersectUint32(s1, s2))
	assert.Equal(t, []uint32{1, 3, 4, 127, 1, 3, 5, 0}, MergeUint32(s1, s2))
	assert.Equal(t, []uint32{1, 5, 0}, RemoveUint32(s2, 3))
	assert.Equal(t, []uint32{0, 5, 3, 1}, ReverseUint32(s2))
	assert.Equal(t, []uint32{1, 3, 4, 127}, SortUint32(s1))
	assert.Equal(t, []uint32{0, 1, 1, 3, 3, 4, 5, 127}, SortUint32(MergeUint32(s1, s2)))
	assert.ElementsMatch(t, []uint32{1, 0, 3, 4, 5, 127}, UniqueUint32(MergeUint32(s1, s2)))
}

func TestUint64(t *testing.T) {
	s1 := []uint64{1, 3, 4, 127}
	s2 := []uint64{1, 3, 5, 0}

	assert.Equal(t, true, ContainUint64(s1, 3))
	assert.Equal(t, []uint64{4, 127}, DiffUint64(s1, s2))
	assert.Equal(t, []uint64{5, 0}, DiffUint64(s2, s1))
	assert.Equal(t, []uint64{1, 3}, IntersectUint64(s1, s2))
	assert.Equal(t, []uint64{1, 3, 4, 127, 1, 3, 5, 0}, MergeUint64(s1, s2))
	assert.Equal(t, []uint64{1, 5, 0}, RemoveUint64(s2, 3))
	assert.Equal(t, []uint64{0, 5, 3, 1}, ReverseUint64(s2))
	assert.Equal(t, []uint64{1, 3, 4, 127}, SortUint64(s1))
	assert.Equal(t, []uint64{0, 1, 1, 3, 3, 4, 5, 127}, SortUint64(MergeUint64(s1, s2)))
	assert.ElementsMatch(t, []uint64{1, 0, 3, 4, 5, 127}, UniqueUint64(MergeUint64(s1, s2)))
}
