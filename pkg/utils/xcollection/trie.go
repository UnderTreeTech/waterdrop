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

package xcollection

import "sync"

// rune 42 stands for *
const defaultMask = 42

// rune 0 stands for ""
const emptyMask = 0

type (
	TrieOption func(root *Trie)

	trieNode struct {
		children map[rune]*trieNode
		end      bool
	}

	Trie struct {
		root *trieNode
		mu   sync.RWMutex
		mask rune
	}
)

// newTrieNode returns a pointer of trieNode
func newTrieNode() *trieNode {
	return &trieNode{
		children: make(map[rune]*trieNode),
	}
}

// NewTrie return a pointer of trie
func NewTrie(keywords []string, opts ...TrieOption) *Trie {
	n := &trieNode{
		children: make(map[rune]*trieNode),
	}

	t := &Trie{
		root: n,
	}

	if len(opts) == 0 {
		t.mask = defaultMask
	}

	for _, opt := range opts {
		opt(t)
	}

	for _, keyword := range keywords {
		t.Add(keyword)
	}

	return t
}

// Add add a keyword to trie
func (t *Trie) Add(keyword string) {
	chars := []rune(keyword)
	if len(chars) == 0 {
		return
	}

	t.mu.Lock()
	node := t.root
	for _, char := range chars {
		if _, ok := node.children[char]; !ok {
			node.children[char] = newTrieNode()
		}
		node = node.children[char]
	}
	node.end = true
	t.mu.Unlock()
}

// Delete delete a keyword from trie, be aware of that Delete execute soft delete
func (t *Trie) Delete(keyword string) {
	chars := []rune(keyword)
	if len(chars) == 0 {
		return
	}

	node := t.root
	size := len(chars)

	t.mu.Lock()
	for i := 0; i < size; i++ {
		child, ok := node.children[chars[i]]
		if !ok {
			t.mu.Unlock()
			return
		}
		node = child
		if i == size-1 {
			node.end = false
		}
	}
	t.mu.Unlock()
}

// DeleteAll delete all keywords from trie
func (t *Trie) DeleteAll() {
	t.mu.Lock()
	t.root = newTrieNode()
	t.mu.Unlock()
}

// Query find sensitive keywords and return them.
func (t *Trie) Query(text string) (sanitize string, keywords []string, exist bool) {
	chars := []rune(text)
	txtLen := len(chars)

	if txtLen == 0 {
		return
	}

	node := t.root
	t.mu.RLock()
	for i := 0; i < txtLen; i++ {
		if _, ok := node.children[chars[i]]; !ok {
			continue
		}

		node = node.children[chars[i]]
		for j := i + 1; j < txtLen; j++ {
			if _, ok := node.children[chars[j]]; !ok {
				break
			}
			node = node.children[chars[j]]
			if node.end {
				keywords = append(keywords, string(chars[i:j+1]))
				t.replaceWithMask(chars, i, j+1)
			}
		}
		node = t.root
	}

	if len(keywords) > 0 {
		exist = true
	}
	sanitize = string(chars)
	t.mu.RUnlock()

	return
}

// QueryAll return all the keywords
func (t *Trie) QueryAll() (keywords []string) {
	t.mu.RLock()
	keywords = t.deepRead(t.root, keywords, "")
	t.mu.RUnlock()
	return
}

func (t *Trie) deepRead(node *trieNode, words []string, parentWord string) (keywords []string) {
	for char, child := range node.children {
		if child.end {
			words = append(words, parentWord+string(char))
		}
		if len(child.children) > 0 {
			words = t.deepRead(child, words, parentWord+string(char))
		}
	}
	return words
}

// replaceWithMask replace keyword with mask
func (t *Trie) replaceWithMask(chars []rune, start, end int) {
	// if mask equal "" need not to replace
	if t.mask == emptyMask {
		return
	}

	for i := start; i < end; i++ {
		chars[i] = t.mask
	}
}

// WithMask mask option
func WithMask(mask rune) TrieOption {
	return func(root *Trie) {
		root.mask = mask
	}
}
