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
	"sort"
)

// ContainFloat64 check if target is in ss
func ContainFloat64(s []float64, target float64) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveFloat64 remove empty target elements from ss
func RemoveFloat64(s []float64, target float64) []float64 {
	var ret = make([]float64, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseFloat64 reverse the input slice
func ReverseFloat64(s []float64) []float64 {
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

// DiffFloat64 computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffFloat64(s1 []float64, s2 []float64) []float64 {
	ret := make([]float64, 0)
	for _, val := range s1 {
		if !ContainFloat64(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectFloat64 computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectFloat64(s1 []float64, s2 []float64) []float64 {
	ret := make([]float64, 0)
	for _, val := range s1 {
		if ContainFloat64(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueFloat64 Removes duplicate values from slice
func UniqueFloat64(s1 []float64) []float64 {
	unique := make(map[float64]struct{})
	ret := make([]float64, 0, len(s1))
	for _, val := range s1 {
		if _, ok := unique[val]; ok {
			continue
		}
		unique[val] = struct{}{}
		ret = append(ret, val)
	}
	return ret
}

// MergeFloat64 merge one or more arrays
func MergeFloat64(s1 []float64, s2 ...[]float64) []float64 {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]float64, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

// SortFloat64 sort float64 slice asc
func SortFloat64(s []float64) []float64 {
	sort.Sort(sort.Float64Slice(s))
	return s
}
