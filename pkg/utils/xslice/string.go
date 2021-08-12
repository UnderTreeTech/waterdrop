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

// ContainString check if target is in ss
func ContainString(s []string, target string) bool {
	for _, val := range s {
		if val == target {
			return true
		}
	}
	return false
}

// RemoveString remove empty target elements from ss
func RemoveString(s []string, target string) []string {
	var ret = make([]string, 0)
	for _, val := range s {
		if val != target {
			ret = append(ret, val)
		}
	}
	return ret
}

// ReverseString reverse the input slice
func ReverseString(s []string) []string {
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

// DiffString computes the difference of two slices
// return a slice containing all the entries from s1 but not in s2
func DiffString(s1 []string, s2 []string) []string {
	ret := make([]string, 0)
	for _, val := range s1 {
		if !ContainString(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// IntersectString computes the intersection of two slices
// return a slice containing the entries exist in s1 and s2
func IntersectString(s1 []string, s2 []string) []string {
	ret := make([]string, 0)
	for _, val := range s1 {
		if ContainString(s2, val) {
			ret = append(ret, val)
		}
	}
	return ret
}

// UniqueString Removes duplicate values from slice
func UniqueString(s1 []string) []string {
	unique := make(map[string]interface{})
	for _, val := range s1 {
		unique[val] = nil
	}

	ret := make([]string, 0, len(unique))
	for key := range unique {
		ret = append(ret, key)
	}
	return ret
}

//MergeString merge one or more arrays
func MergeString(s1 []string, s2 ...[]string) []string {
	if len(s2) == 0 {
		return s1
	}
	size := len(s1)
	for _, s := range s2 {
		size += len(s)
	}
	ret := make([]string, 0, size)
	ret = append(ret, s1...)
	for _, s := range s2 {
		ret = append(ret, s...)
	}
	return ret
}

//SortString sort string slice asc
func SortString(s []string) []string {
	sort.Sort(sort.StringSlice(s))
	return s
}
