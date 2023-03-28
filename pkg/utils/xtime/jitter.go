/*
 *
 * Copyright 2023 waterdrop authors.
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

package xtime

import (
	"math"
	"math/rand"
	"time"
)

const coefficient = 0.1

// JitterTime simulates jitter by scaling a time.Duration randomly within factor
// The duration d must be greater than zero; and the scaling factor must be
// within the range 0 < factor <= 1.0,  default factor is 0.1
func JitterTime(d time.Duration, factor ...float64) (scale time.Duration) {
	if d <= 0 {
		return d
	}

	f := coefficient
	if len(factor) > 0 {
		if factor[0] > 1.0 || factor[0] <= 0 {
			return d
		}
		f = factor[0]
	}

	min, max := bounds(int64(d), f)
	scale = time.Duration(randRange(min, max))
	return
}

// Jitter simulates jitter by scaling a integer randomly within factor
// The duration d must be greater than zero; and the scaling factor must be
// within the range 0 < factor <= 1.0, default factor is 0.1
func Jitter(n int64, factor ...float64) (scale int64) {
	if n <= 0 {
		return n
	}

	f := coefficient
	if len(factor) > 0 {
		if factor[0] > 1.0 || factor[0] <= 0 {
			return n
		}
		f = factor[0]
	}

	min, max := bounds(n, f)
	scale = randRange(min, max)
	return
}

// bounds returns the min and max values for n after applying scaling factor
// if the max overflow, then truncate and return math.MaxInt64
func bounds(n int64, factor ...float64) (min, max int64) {
	f := coefficient
	if len(factor) > 0 {
		f = factor[0]
	}

	minf := math.Floor(float64(n) * (1 - f))
	maxf := math.Ceil(float64(n) * (1 + f))

	if maxf > math.MaxInt64 {
		return int64(minf), math.MaxInt64
	}
	return int64(minf), int64(maxf)
}

// randRange returns a non-negative pseudo-random number in the half open
// interval [min, max) from the default source
func randRange(min, max int64) int64 {
	if min == max {
		return min
	}
	return rand.Int63n(max-min) + min
}
