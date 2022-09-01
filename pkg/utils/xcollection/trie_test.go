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

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_trie_Query(t1 *testing.T) {
	tests := []struct {
		name         string
		t            *Trie
		text         string
		wantSentence string
		wantKeywords []string
		wantExist    bool
	}{
		{
			name:         "test trie",
			t:            NewTrie([]string{"AV", "色情", "习近平", "李克强"}),
			text:         "青少年易受AV色情影视影响；高举中国特色社会主义旗帜，紧密团结在以习近平同志的党中央，李克强同志劳苦功高。",
			wantSentence: "青少年易受****影视影响；高举中国特色社会主义旗帜，紧密团结在以***同志的党中央，***同志劳苦功高。",
			wantKeywords: []string{"AV", "色情", "习近平", "李克强"},
			wantExist:    true,
		},
		{
			name:         "test trie",
			t:            NewTrie([]string{"AV", "色情", "习近平", "李克强"}, WithMask('#')),
			text:         "青少年易受AV色情影视影响；高举中国特色社会主义旗帜，紧密团结在以习近平同志的党中央，李克强同志劳苦功高。",
			wantSentence: "青少年易受####影视影响；高举中国特色社会主义旗帜，紧密团结在以###同志的党中央，###同志劳苦功高。",
			wantKeywords: []string{"AV", "色情", "习近平", "李克强"},
			wantExist:    true,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := tt.t
			gotSentence, gotKeywords, gotExist := t.Query(tt.text)
			assert.Equal(t1, gotSentence, tt.wantSentence)
			assert.ElementsMatch(t1, gotKeywords, tt.wantKeywords)
			assert.Equal(t1, gotExist, tt.wantExist)
		})
	}
}

func Test_trie_Add(t1 *testing.T) {
	tests := []struct {
		name         string
		t            *Trie
		keywords     []string
		wantKeywords []string
	}{
		{
			name:         "add",
			t:            NewTrie(nil),
			keywords:     []string{"AV", "色情", "习近平", "李克强"},
			wantKeywords: []string{"AV", "色情", "习近平", "李克强"},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			for _, keyword := range tt.keywords {
				tt.t.Add(keyword)
			}
			gotKeywords := tt.t.QueryAll()
			assert.ElementsMatch(t1, gotKeywords, tt.wantKeywords)
		})
	}
}

func Test_trie_Delete(t1 *testing.T) {
	tests := []struct {
		name         string
		t            *Trie
		keywords     []string
		wantKeywords []string
	}{
		{
			name:         "delete",
			t:            NewTrie([]string{"AV", "色情", "习近平", "李克强"}),
			wantKeywords: []string{"色情", "习近平", "李克强"},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			tt.t.Delete("AV")
			gotKeywords := tt.t.QueryAll()
			assert.ElementsMatch(t1, gotKeywords, tt.wantKeywords)
		})
	}
}

func Test_trie_DeleteAll(t1 *testing.T) {
	tests := []struct {
		name         string
		t            *Trie
		keywords     []string
		wantKeywords []string
	}{
		{
			name:         "delete",
			t:            NewTrie([]string{"AV", "色情", "习近平", "李克强"}),
			wantKeywords: nil,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			tt.t.DeleteAll()
			gotKeywords := tt.t.QueryAll()
			assert.ElementsMatch(t1, gotKeywords, tt.wantKeywords)
		})
	}
}

func Test_trie_QueryAll(t1 *testing.T) {
	tests := []struct {
		name         string
		t            *Trie
		keywords     []string
		wantKeywords []string
	}{
		{
			name:         "delete",
			t:            NewTrie([]string{"AV", "色情", "习近平", "李克强"}),
			wantKeywords: []string{"AV", "色情", "习近平", "李克强"},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			gotKeywords := tt.t.QueryAll()
			assert.ElementsMatch(t1, gotKeywords, tt.wantKeywords)
		})
	}
}

func BenchmarkTrie(b *testing.B) {
	b.ReportAllocs()

	t := NewTrie([]string{"AV", "色情", "习近平", "李克强"})
	for i := 0; i < b.N; i++ {
		t.Query("青少年易受AV色情影视影响；高举中国特色社会主义旗帜，紧密团结在以习近平同志的党中央，李克强同志劳苦功高。")
	}
}
