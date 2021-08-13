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

func TestFloat64(t *testing.T) {
	f1 := []float64{3.12, 3.3, 3.0, 4, 7.222222}
	f2 := []float64{3.12, 3.3, 3.0, 4, 8.222222}

	assert.Equal(t, true, ContainFloat64(f1, 3.3))
	assert.ElementsMatch(t, []float64{7.222222}, DiffFloat64(f1, f2))
	assert.ElementsMatch(t, []float64{8.222222}, DiffFloat64(f2, f1))
	assert.ElementsMatch(t, []float64{3.12, 3.3, 3.0, 4}, IntersectFloat64(f1, f2))
	assert.ElementsMatch(t, []float64{3.12, 3.3, 3.0, 4, 7.222222, 3.12, 3.3, 3.0, 4, 8.222222}, MergeFloat64(f1, f2))
	assert.ElementsMatch(t, []float64{3.12, 3.3, 4, 8.222222}, RemoveFloat64(f2, 3.0))
	assert.ElementsMatch(t, []float64{8.222222, 4, 3.0, 3.3, 3.12}, ReverseFloat64(f2))
	assert.ElementsMatch(t, []float64{3.0, 3.12, 3.3, 4, 7.222222}, SortFloat64(f1))
	assert.ElementsMatch(t, []float64{3.0, 3.0, 3.12, 3.12, 3.3, 3.3, 4, 4, 7.222222, 8.222222}, SortFloat64(MergeFloat64(f1, f2)))
	assert.ElementsMatch(t, []float64{3.0, 3.12, 3.3, 4, 7.222222, 8.222222}, UniqueFloat64(MergeFloat64(f1, f2)))
}
