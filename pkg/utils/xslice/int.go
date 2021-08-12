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

import "sort"

// ContainInt8 check if target is in ss
func ContainInt8(s []int8, target int8) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveInt8 remove empty target elements from ss
func RemoveInt8(s []int8, target int8) []int8 {
	var ret = make([]int8, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseInt8 reverse the input slice
func ReverseInt8(s []int8) []int8 {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffInt8 computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffInt8(s1 []int8, s2 []int8) []int8 {
	ret := make([]int8, 0)
	for _, val := range s1 {
		if !ContainInt8(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectInt8 computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectInt8(s1 []int8, s2 []int8) []int8 {
	ret := make([]int8, 0)
	for _, val := range s1 {
		if ContainInt8(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueInt8 Removes duplicate values from slice
func UniqueInt8(s1 []int8) []int8 {
	unique := make(map[int8]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]int8, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeInt8 merge one or more arrays
func MergeInt8(s1 []int8, s2 ...[]int8) []int8 {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]int8, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortInt8 sort int8 slice asc
func SortInt8(s []int8) []int8 {
	intSlice := make([]int, 0, len(s))
	for _, val := range s {
		intSlice = append(intSlice, int(val))
	}
	sort.Sort(sort.IntSlice(intSlice))
	for index, val := range intSlice {
		s[index] = int8(val)
	}
	return s
}

// ContainUint8 check if target is in ss
func ContainUint8(s []uint8, target uint8) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveUint8 remove empty target elements from ss
func RemoveUint8(s []uint8, target uint8) []uint8 {
	var ret = make([]uint8, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseUint8 reverse the input slice
func ReverseUint8(s []uint8) []uint8 {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffUint8 computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffUint8(s1 []uint8, s2 []uint8) []uint8 {
	ret := make([]uint8, 0)
	for _, val := range s1 {
		if !ContainUint8(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectUint8 computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectUint8(s1 []uint8, s2 []uint8) []uint8 {
	ret := make([]uint8, 0)
	for _, val := range s1 {
		if ContainUint8(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueUint8 Removes duplicate values from slice
func UniqueUint8(s1 []uint8) []uint8 {
	unique := make(map[uint8]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]uint8, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeUint8 merge one or more arrays
func MergeUint8(s1 []uint8, s2 ...[]uint8) []uint8 {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]uint8, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortUint8 sort uint8 slice asc
func SortUint8(s []uint8) []uint8 {
	intSlice := make([]int, 0, len(s))
	for _, val := range s {
		intSlice = append(intSlice, int(val))
	}
	sort.Sort(sort.IntSlice(intSlice))
	for index, val := range intSlice {
		s[index] = uint8(val)
	}
	return s
}

// ContainInt16 check if target is in ss
func ContainInt16(s []int16, target int16) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveInt16 remove empty target elements from ss
func RemoveInt16(s []int16, target int16) []int16 {
	var ret = make([]int16, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseInt16 reverse the input slice
func ReverseInt16(s []int16) []int16 {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffInt16 computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffInt16(s1 []int16, s2 []int16) []int16 {
	ret := make([]int16, 0)
	for _, val := range s1 {
		if !ContainInt16(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectInt16 computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectInt16(s1 []int16, s2 []int16) []int16 {
	ret := make([]int16, 0)
	for _, val := range s1 {
		if ContainInt16(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueInt16 Removes duplicate values from slice
func UniqueInt16(s1 []int16) []int16 {
	unique := make(map[int16]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]int16, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeInt16 merge one or more arrays
func MergeInt16(s1 []int16, s2 ...[]int16) []int16 {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]int16, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortInt16 sort int16 slice asc
func SortInt16(s []int16) []int16 {
	intSlice := make([]int, 0, len(s))
	for _, val := range s {
		intSlice = append(intSlice, int(val))
	}
	sort.Sort(sort.IntSlice(intSlice))
	for index, val := range intSlice {
		s[index] = int16(val)
	}
	return s
}

// ContainUint16 check if target is in ss
func ContainUint16(s []uint16, target uint16) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveUint16 remove empty target elements from ss
func RemoveUint16(s []uint16, target uint16) []uint16 {
	var ret = make([]uint16, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseUint16 reverse the input slice
func ReverseUint16(s []uint16) []uint16 {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffUint16 computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffUint16(s1 []uint16, s2 []uint16) []uint16 {
	ret := make([]uint16, 0)
	for _, val := range s1 {
		if !ContainUint16(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectUint16 computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectUint16(s1 []uint16, s2 []uint16) []uint16 {
	ret := make([]uint16, 0)
	for _, val := range s1 {
		if ContainUint16(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueUint16 Removes duplicate values from slice
func UniqueUint16(s1 []uint16) []uint16 {
	unique := make(map[uint16]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]uint16, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeUint16 merge one or more arrays
func MergeUint16(s1 []uint16, s2 ...[]uint16) []uint16 {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]uint16, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortUint16 sort uint16 slice asc
func SortUint16(s []uint16) []uint16 {
	intSlice := make([]int, 0, len(s))
	for _, val := range s {
		intSlice = append(intSlice, int(val))
	}
	sort.Sort(sort.IntSlice(intSlice))
	for index, val := range intSlice {
		s[index] = uint16(val)
	}
	return s
}

// ContainInt check if target is in ss
func ContainInt(s []int, target int) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveInt remove empty target elements from ss
func RemoveInt(s []int, target int) []int {
	var ret = make([]int, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseInt reverse the input slice
func ReverseInt(s []int) []int {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffInt computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffInt(s1 []int, s2 []int) []int {
	ret := make([]int, 0)
	for _, val := range s1 {
		if !ContainInt(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectInt computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectInt(s1 []int, s2 []int) []int {
	ret := make([]int, 0)
	for _, val := range s1 {
		if ContainInt(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueInt Removes duplicate values from slice
func UniqueInt(s1 []int) []int {
	unique := make(map[int]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]int, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeInt merge one or more arrays
func MergeInt(s1 []int, s2 ...[]int) []int {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]int, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortInt sort int slice asc
func SortInt(s []int) []int {
	sort.Sort(sort.IntSlice(s))
	return s
}

// ContainUint check if target is in ss
func ContainUint(s []uint, target uint) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveUint remove empty target elements from ss
func RemoveUint(s []uint, target uint) []uint {
	var ret = make([]uint, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseUint reverse the input slice
func ReverseUint(s []uint) []uint {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffUint computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffUint(s1 []uint, s2 []uint) []uint {
	ret := make([]uint, 0)
	for _, val := range s1 {
		if !ContainUint(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectUint computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectUint(s1 []uint, s2 []uint) []uint {
	ret := make([]uint, 0)
	for _, val := range s1 {
		if ContainUint(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueUint Removes duplicate values from slice
func UniqueUint(s1 []uint) []uint {
	unique := make(map[uint]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]uint, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeUint merge one or more arrays
func MergeUint(s1 []uint, s2 ...[]uint) []uint {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]uint, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortUint sort uint slice asc
func SortUint(s []uint) []uint {
	intSlice := make([]int, 0, len(s))
	for _, val := range s {
		intSlice = append(intSlice, int(val))
	}
	sort.Sort(sort.IntSlice(intSlice))
	for index, val := range intSlice {
		s[index] = uint(val)
	}
	return s
}

// ContainInt32 check if target is in ss
func ContainInt32(s []int32, target int32) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveInt32 remove empty target elements from ss
func RemoveInt32(s []int32, target int32) []int32 {
	var ret = make([]int32, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseInt32 reverse the input slice
func ReverseInt32(s []int32) []int32 {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffInt32 computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffInt32(s1 []int32, s2 []int32) []int32 {
	ret := make([]int32, 0)
	for _, val := range s1 {
		if !ContainInt32(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectInt32 computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectInt32(s1 []int32, s2 []int32) []int32 {
	ret := make([]int32, 0)
	for _, val := range s1 {
		if ContainInt32(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueInt32 Removes duplicate values from slice
func UniqueInt32(s1 []int32) []int32 {
	unique := make(map[int32]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]int32, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeInt32 merge one or more arrays
func MergeInt32(s1 []int32, s2 ...[]int32) []int32 {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]int32, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortInt32 sort int32 slice asc
func SortInt32(s []int32) []int32 {
	intSlice := make([]int, 0, len(s))
	for _, val := range s {
		intSlice = append(intSlice, int(val))
	}
	sort.Sort(sort.IntSlice(intSlice))
	for index, val := range intSlice {
		s[index] = int32(val)
	}
	return s
}

// ContainUint32 check if target is in ss
func ContainUint32(s []uint32, target uint32) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveUint32 remove empty target elements from ss
func RemoveUint32(s []uint32, target uint32) []uint32 {
	var ret = make([]uint32, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseUint32 reverse the input slice
func ReverseUint32(s []uint32) []uint32 {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffUint32 computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffUint32(s1 []uint32, s2 []uint32) []uint32 {
	ret := make([]uint32, 0)
	for _, val := range s1 {
		if !ContainUint32(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectUint32 computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectUint32(s1 []uint32, s2 []uint32) []uint32 {
	ret := make([]uint32, 0)
	for _, val := range s1 {
		if ContainUint32(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueUint32 Removes duplicate values from slice
func UniqueUint32(s1 []uint32) []uint32 {
	unique := make(map[uint32]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]uint32, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeUint32 merge one or more arrays
func MergeUint32(s1 []uint32, s2 ...[]uint32) []uint32 {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]uint32, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortUint32 sort uint32 slice asc
func SortUint32(s []uint32) []uint32 {
	intSlice := make([]int, 0, len(s))
	for _, val := range s {
		intSlice = append(intSlice, int(val))
	}
	sort.Sort(sort.IntSlice(intSlice))
	for index, val := range intSlice {
		s[index] = uint32(val)
	}
	return s
}

// ContainInt64 check if target is in ss
func ContainInt64(s []int64, target int64) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveInt64 remove empty target elements from ss
func RemoveInt64(s []int64, target int64) []int64 {
	var ret = make([]int64, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseInt64 reverse the input slice
func ReverseInt64(s []int64) []int64 {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffInt64 computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffInt64(s1 []int64, s2 []int64) []int64 {
	ret := make([]int64, 0)
	for _, val := range s1 {
		if !ContainInt64(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectInt64 computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectInt64(s1 []int64, s2 []int64) []int64 {
	ret := make([]int64, 0)
	for _, val := range s1 {
		if ContainInt64(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueInt64 Removes duplicate values from slice
func UniqueInt64(s1 []int64) []int64 {
	unique := make(map[int64]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]int64, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeInt64 merge one or more arrays
func MergeInt64(s1 []int64, s2 ...[]int64) []int64 {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]int64, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortInt64 sort int64 slice asc
func SortInt64(s []int64) []int64 {
	intSlice := make([]int, 0, len(s))
	for _, val := range s {
		intSlice = append(intSlice, int(val))
	}
	sort.Sort(sort.IntSlice(intSlice))
	for index, val := range intSlice {
		s[index] = int64(val)
	}
	return s
}

// ContainUint64 check if target is in ss
func ContainUint64(s []uint64, target uint64) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveUint64 remove empty target elements from ss
func RemoveUint64(s []uint64, target uint64) []uint64 {
	var ret = make([]uint64, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseUint64 reverse the input slice
func ReverseUint64(s []uint64) []uint64 {
	if len(s) < 2 {
		return s
	}

	start := 0
	end := len(s) - 1
	for start < end {
		tmp := s[start]
		s[start] = s[end]
		s[end] = tmp
		start++
		end--
	}
	return s
}

// DiffUint64 computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffUint64(s1 []uint64, s2 []uint64) []uint64 {
	ret := make([]uint64, 0)
	for _, val := range s1 {
		if !ContainUint64(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectUint64 computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectUint64(s1 []uint64, s2 []uint64) []uint64 {
	ret := make([]uint64, 0)
	for _, val := range s1 {
		if ContainUint64(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueUint64 Removes duplicate values from slice
func UniqueUint64(s1 []uint64) []uint64 {
	unique := make(map[uint64]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]uint64, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

// MergeUint64 merge one or more arrays
func MergeUint64(s1 []uint64, s2 ...[]uint64) []uint64 {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]uint64, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortUint64 sort int64 slice asc
func SortUint64(s []uint64) []uint64 {
	intSlice := make([]int, 0, len(s))
	for _, val := range s {
		intSlice = append(intSlice, int(val))
	}
	sort.Sort(sort.IntSlice(intSlice))
	for index, val := range intSlice {
		s[index] = uint64(val)
	}
	return s
}
