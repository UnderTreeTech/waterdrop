/*
 *
 * Copyright 2020 waterdrop authors.
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

package xcollection

import (
	"container/heap"
	"math/rand"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPriorityQueue(t *testing.T) {
	c := 100
	pq := NewPriorityQueue(c)

	for i := 0; i < c+1; i++ {
		heap.Push(&pq, &Item{Value: i, Priority: int64(i)})
	}
	assert.Equal(t, pq.Len(), c+1)
	assert.Equal(t, cap(pq), c*2)

	for i := 0; i < c+1; i++ {
		item := heap.Pop(&pq)
		assert.Equal(t, item.(*Item).Value.(int), i)
	}
	assert.Equal(t, cap(pq), c/4)
}

func TestPush(t *testing.T) {
	c := 100
	pq := NewPriorityQueue(c)
	ints := make([]int, 0, c)

	for i := 0; i < c; i++ {
		v := rand.Int()
		ints = append(ints, v)
		heap.Push(&pq, &Item{Value: i, Priority: int64(v)})
	}
	assert.Equal(t, pq.Len(), c)
	assert.Equal(t, cap(pq), c)

	sort.Ints(ints)

	for i := 0; i < c; i++ {
		item, _ := pq.PeekAndShift(int64(ints[len(ints)-1]))
		assert.Equal(t, item.Priority, int64(ints[i]))
	}
}

func TestPop(t *testing.T) {
	c := 100
	pq := NewPriorityQueue(c)

	for i := 0; i < c; i++ {
		v := rand.Int()
		heap.Push(&pq, &Item{Value: "test", Priority: int64(v)})
	}

	for i := 0; i < 10; i++ {
		heap.Remove(&pq, rand.Intn((c-1)-i))
	}

	lastPriority := heap.Pop(&pq).(*Item).Priority
	for i := 0; i < (c - 10 - 1); i++ {
		item := heap.Pop(&pq)
		assert.Equal(t, lastPriority < item.(*Item).Priority, true)
		lastPriority = item.(*Item).Priority
	}
}
