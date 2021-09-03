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

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	es7 "github.com/olivere/elastic/v7"
)

// Config client config
type Config struct {
	// Username es auth username
	Username string
	// Password es auth password
	Password string
	// Schema http/https, default http
	Schema string
	// URLs endpoint urls
	URLs []string
	// Plugins enabled plugins
	Plugins []string
}

// Client es client struct
type Client struct {
	client *es7.Client
	config *Config
}

// NewClient returns es7 client instance
func NewClient(config *Config) *Client {
	if "" == config.Schema {
		config.Schema = "http"
	}

	es7, _ := es7.NewClient(
		es7.SetHttpClient(&http.Client{
			Transport: NewTransport(config),
		}),
		es7.SetBasicAuth(config.Username, config.Password),
		es7.SetURL(config.URLs...),
		es7.SetScheme(config.Schema),
		es7.SetRequiredPlugins(config.Plugins...),
	)
	client := &Client{
		config: config,
		client: es7,
	}
	return client
}

// Ping checks if an Elasticsearch server on a given URL is alive
// If the server responds with HTTP Status code 200 OK, the server is alive
func (c *Client) Ping(ctx context.Context) (alive bool, err error) {
	_, status, err := c.client.Ping(c.config.URLs[0]).Do(ctx)
	alive = status == http.StatusOK
	return
}

// CreateIndex create an index
func (c *Client) CreateIndex(ctx context.Context, index string, mapping string) (err error) {
	_, err = c.client.CreateIndex(index).Body(mapping).Do(ctx)
	return
}

// ExistIndex check if index exists
func (c *Client) ExistIndex(ctx context.Context, index string) (exist bool, err error) {
	exist, err = c.client.IndexExists(index).Do(ctx)
	return
}

// DeleteIndex delete an index
func (c *Client) DeleteIndex(ctx context.Context, index string) (err error) {
	_, err = c.client.DeleteIndex(index).Do(ctx)
	return
}

// CreateDoc insert a doc to index
// If don't assign a doc id, es will automatically generate one and assign it to _id
func (c *Client) CreateDoc(ctx context.Context, index string, doc interface{}, docId ...string) (id string, err error) {
	idx := c.client.Index().Index(index).BodyJson(doc)
	if len(docId) > 0 {
		idx.Id(docId[0])
	}
	reply, err := idx.Do(ctx)
	if err != nil {
		return
	}
	id = reply.Id
	return
}

// CreateDocs batch insert a doc to index
// If don't assign a doc ids, es will automatically generate ids and assign them to _id separately
func (c *Client) CreateDocs(ctx context.Context, index string, docs []interface{}, docIds ...string) (ids []string, err error) {
	var withDocId bool
	if len(docIds) > 0 {
		if len(docIds) != len(docs) {
			return nil, errors.New("docs size must equal ids size")
		}
		withDocId = true
	}

	bulk := c.client.Bulk()
	for idx, doc := range docs {
		req := es7.NewBulkIndexRequest().Index(index).Doc(doc)
		if withDocId {
			req.Id(docIds[idx])
		}
		bulk = bulk.Add(req)
	}

	reply, err := bulk.Do(ctx)
	if err != nil {
		return
	}

	items := reply.Indexed()
	ids = make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.Id)
	}
	return
}

// UpdateDoc update a doc by id
func (c *Client) UpdateDoc(ctx context.Context, index string, id string, doc interface{}) (err error) {
	_, err = c.client.Update().Index(index).Id(id).Doc(doc).Do(ctx)
	return
}

// UpdateDocs batch update docs
func (c *Client) UpdateDocs(ctx context.Context, index string, ids []string, docs []interface{}) (num int64, err error) {
	bulk := c.client.Bulk()
	for idx, id := range ids {
		req := es7.NewBulkUpdateRequest().Index(index).Id(id).Doc(docs[idx])
		bulk = bulk.Add(req)
	}
	reply, err := bulk.Do(ctx)
	if err != nil {
		return
	}
	num = int64(len(reply.Updated()))
	return
}

// DeleteDoc delete doc by id
func (c *Client) DeleteDoc(ctx context.Context, index string, id string) (err error) {
	_, err = c.client.Delete().Index(index).Id(id).Do(ctx)
	return
}

// DeleteDocs batch delete docs by doc ids
func (c *Client) DeleteDocs(ctx context.Context, index string, ids []string) (num int64, err error) {
	bulk := c.client.Bulk()
	for _, id := range ids {
		req := es7.NewBulkDeleteRequest().Index(index).Id(id)
		bulk = bulk.Add(req)
	}
	reply, err := bulk.Do(ctx)
	if err != nil {
		return
	}

	num = int64(len(reply.Deleted()))
	return
}

// DeleteDocByQuery delete doc by query
func (c *Client) DeleteDocByQuery(ctx context.Context, index string, filter es7.Query) (num int64, err error) {
	reply, err := c.client.DeleteByQuery(index).Query(filter).Do(ctx)
	if err != nil {
		return
	}
	num = reply.Deleted
	return
}

// GetDoc retrieve doc by id
// hit means doc exists or not when request success(return nil)
func (c *Client) GetDoc(ctx context.Context, index string, id string, reply interface{}) (hit bool, err error) {
	result, err := c.client.Get().Index(index).Id(id).Do(ctx)
	if err != nil {
		return
	}

	if result.Found {
		hit = true
		err = json.Unmarshal(result.Source, reply)
	}
	return
}

// GetDocs retrieve docs by ids
func (c *Client) GetDocs(ctx context.Context, index string, ids []string) (docs [][]byte, err error) {
	mget := c.client.MultiGet()
	for _, id := range ids {
		item := es7.NewMultiGetItem().Index(index).Id(id)
		mget = mget.Add(item)
	}
	items, err := mget.Do(ctx)
	if err != nil {
		return
	}

	for _, item := range items.Docs {
		if item.Found {
			docs = append(docs, item.Source)
		}
	}
	return
}

// Count index documents
func (c *Client) Count(ctx context.Context, index string, query es7.Query) (num int64, err error) {
	num, err = c.client.Count().Index(index).Query(query).Do(ctx)
	return
}

// Refresh asks Elasticsearch to refresh one or more indices
func (c *Client) Refresh(ctx context.Context, indices ...string) (err error) {
	_, err = c.client.Refresh(indices...).Do(ctx)
	return
}

// Flush asks Elasticsearch to free memory from the index and flush data to disk
func (c *Client) Flush(ctx context.Context, indices ...string) (err error) {
	_, err = c.client.Flush(indices...).Do(ctx)
	return
}

// NewSearch is the entry point for searches
func (c *Client) NewSearch(index ...string) (search *es7.SearchService) {
	return c.client.Search(index...)
}

// NewScroll is the entry point for scroll searches
func (c *Client) NewScroll(index ...string) (search *es7.ScrollService) {
	return c.client.Scroll(index...)
}

// NewBulk is the entry point to mass insert/update/delete documents
func (c *Client) NewBulk() *es7.BulkService {
	return c.client.Bulk()
}
