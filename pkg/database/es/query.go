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

package es

import "github.com/olivere/elastic/v7"

// NewBoolQuery creates a new bool query
func NewBoolQuery() *elastic.BoolQuery {
	return elastic.NewBoolQuery()
}

// NewBoostingQuery creates a new boosting query
func NewBoostingQuery() *elastic.BoostingQuery {
	return elastic.NewBoostingQuery()
}

// NewTermQuery creates and initializes a new TermQuery
func NewTermQuery(name string, value interface{}) *elastic.TermQuery {
	return elastic.NewTermQuery(name, value)
}

// NewTermsQuery creates and initializes a new TermsQuery
func NewTermsQuery(name string, values ...interface{}) *elastic.TermsQuery {
	return elastic.NewTermsQuery(name, values...)
}

// NewTermsSetQuery creates and initializes a new TermsSetQuery
func NewTermsSetQuery(name string, values ...interface{}) *elastic.TermsSetQuery {
	return elastic.NewTermsSetQuery(name, values...)
}

// NewWildcardQuery creates and initializes a new WildcardQuery
func NewWildcardQuery(name, wildcard string) *elastic.WildcardQuery {
	return elastic.NewWildcardQuery(name, wildcard)
}

// NewMatchQuery creates and initializes a new MatchQuery
func NewMatchQuery(name string, text interface{}) *elastic.MatchQuery {
	return elastic.NewMatchQuery(name, text)
}

// NewMultiMatchQuery creates and initializes a new MultiMatchQuery
func NewMultiMatchQuery(text interface{}, fields ...string) *elastic.MultiMatchQuery {
	return elastic.NewMultiMatchQuery(text, fields...)
}

// NewMatchAllQuery creates and initializes a new match all query
func NewMatchAllQuery() *elastic.MatchAllQuery {
	return elastic.NewMatchAllQuery()
}

// NewMatchNoneQuery creates and initializes a new match none query
func NewMatchNoneQuery() *elastic.MatchNoneQuery {
	return elastic.NewMatchNoneQuery()
}

// MatchPhraseQuery creates and initializes a new MatchPhraseQuery
func MatchPhraseQuery(name string, value interface{}) *elastic.MatchPhraseQuery {
	return elastic.NewMatchPhraseQuery(name, value)
}

// NewMatchBoolPrefixQuery creates and initializes a new MatchBoolPrefixQuery
func NewMatchBoolPrefixQuery(name string, queryText interface{}) *elastic.MatchBoolPrefixQuery {
	return elastic.NewMatchBoolPrefixQuery(name, queryText)
}

// NewMatchPhrasePrefixQuery creates and initializes a new MatchPhrasePrefixQuery
func NewMatchPhrasePrefixQuery(name string, value interface{}) *elastic.MatchPhrasePrefixQuery {
	return elastic.NewMatchPhrasePrefixQuery(name, value)
}

// NewNestedQuery creates and initializes a new NestedQuery
func NewNestedQuery(path string, query elastic.Query) *elastic.NestedQuery {
	return elastic.NewNestedQuery(path, query)
}

// NewPrefixQuery creates and initializes a new PrefixQuery
func NewPrefixQuery(name string, prefix string) *elastic.PrefixQuery {
	return elastic.NewPrefixQuery(name, prefix)
}

// NewRangeQuery creates and initializes a new RangeQuery
func NewRangeQuery(name string) *elastic.RangeQuery {
	return elastic.NewRangeQuery(name)
}

// NewRegexpQuery creates and initializes a new RegexpQuery
func NewRegexpQuery(name string, regexp string) *elastic.RegexpQuery {
	return elastic.NewRegexpQuery(name, regexp)
}

// NewConstantScoreQuery creates and initializes a new constant score query.
func NewConstantScoreQuery(filter elastic.Query) *elastic.ConstantScoreQuery {
	return elastic.NewConstantScoreQuery(filter)
}

// NewDisMaxQuery creates and initializes a new dis max query
func NewDisMaxQuery() *elastic.DisMaxQuery {
	return elastic.NewDisMaxQuery()
}

// NewDistanceFeatureQuery creates and initializes a new script_score query
func NewDistanceFeatureQuery(field string, origin interface{}, pivot string) *elastic.DistanceFeatureQuery {
	return elastic.NewDistanceFeatureQuery(field, origin, pivot)
}

// NewExistsQuery creates and initializes a new exists query
func NewExistsQuery(name string) *elastic.ExistsQuery {
	return elastic.NewExistsQuery(name)
}

// NewFunctionScoreQuery creates and initializes a new function score query
func NewFunctionScoreQuery() *elastic.FunctionScoreQuery {
	return elastic.NewFunctionScoreQuery()
}

// NewExponentialDecayFunction creates a new ExponentialDecayFunction
func NewExponentialDecayFunction() *elastic.ExponentialDecayFunction {
	return elastic.NewExponentialDecayFunction()
}

// NewFuzzyQuery creates a new fuzzy query
func NewFuzzyQuery(name string, value interface{}) *elastic.FuzzyQuery {
	return elastic.NewFuzzyQuery(name, value)
}

// NewHasChildQuery creates and initializes a new has_child query
func NewHasChildQuery(childType string, query elastic.Query) *elastic.HasChildQuery {
	return elastic.NewHasChildQuery(childType, query)
}

// NewHasParentQuery creates and initializes a new has_parent query
func NewHasParentQuery(parentType string, query elastic.Query) *elastic.HasParentQuery {
	return elastic.NewHasParentQuery(parentType, query)
}

// NewPinnedQuery creates and initializes a new pinned query
func NewPinnedQuery() *elastic.PinnedQuery {
	return elastic.NewPinnedQuery()
}

// NewQueryStringQuery creates and initializes a new QueryStringQuery
func NewQueryStringQuery(queryString string) *elastic.QueryStringQuery {
	return elastic.NewQueryStringQuery(queryString)
}

// NewRawStringQuery initializes a new RawStringQuery
// It is the same as RawStringQuery(q).
func NewRawStringQuery(q string) elastic.RawStringQuery {
	return elastic.NewRawStringQuery(q)
}

// NewSimpleQueryStringQuery creates and initializes a new SimpleQueryStringQuery
func NewSimpleQueryStringQuery(text string) *elastic.SimpleQueryStringQuery {
	return elastic.NewSimpleQueryStringQuery(text)
}

// NewSliceQuery creates a new SliceQuery
func NewSliceQuery() *elastic.SliceQuery {
	return elastic.NewSliceQuery()
}

// NewSpanFirstQuery creates a new SpanFirstQuery
func NewSpanFirstQuery(query elastic.Query, end int) *elastic.SpanFirstQuery {
	return elastic.NewSpanFirstQuery(query, end)
}

// NewSpanNearQuery creates a new SpanNearQuery
func NewSpanNearQuery(clauses ...elastic.Query) *elastic.SpanNearQuery {
	return elastic.NewSpanNearQuery(clauses...)
}

// NewSpanTermQuery creates a new SpanTermQuery. When passing values, the first one
// is used to initialize the value
func NewSpanTermQuery(field string, value ...interface{}) *elastic.SpanTermQuery {
	return elastic.NewSpanTermQuery(field, value...)
}

// NewParentIdQuery creates and initializes a new parent_id query
func NewParentIdQuery(typ, id string) *elastic.ParentIdQuery {
	return elastic.NewParentIdQuery(typ, id)
}

// NewWrapperQuery creates and initializes a new WrapperQuery
func NewWrapperQuery(source string) *elastic.WrapperQuery {
	return elastic.NewWrapperQuery(source)
}

// NewSearchRequest creates a new search request
func NewSearchRequest() *elastic.SearchRequest {
	return elastic.NewSearchRequest()
}

// NewBulkDeleteRequest returns a new BulkDeleteRequest
func NewBulkDeleteRequest() *elastic.BulkDeleteRequest {
	return elastic.NewBulkDeleteRequest()
}

// NewBulkIndexRequest returns a new BulkIndexRequest
// The operation type is "index" by default
func NewBulkIndexRequest() *elastic.BulkIndexRequest {
	return elastic.NewBulkIndexRequest()
}

// NewBulkUpdateRequest returns a new BulkUpdateRequest
func NewBulkUpdateRequest() *elastic.BulkUpdateRequest {
	return elastic.NewBulkUpdateRequest()
}
