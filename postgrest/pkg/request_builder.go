package postgrest_go

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RequestError represents an error response from the PostgREST server.
type RequestError struct {
	Message        string `json:"message"`
	Details        string `json:"details"`
	Hint           string `json:"hint"`
	Code           string `json:"code"`
	HTTPStatusCode int    `json:"-"`
}

func (rq *RequestError) Error() string {
	return fmt.Sprintf("%s: %s", rq.Code, rq.Message)
}

// RequestBuilder represents a builder for PostgREST requests.
type RequestBuilder struct {
	client *Client
	path   string
	params url.Values
	header http.Header
}

// Select starts building a SELECT request with the specified columns.
func (b *RequestBuilder) Select(columns ...string) *SelectRequestBuilder {
	b.params.Set("select", strings.Join(columns, ","))
	return &SelectRequestBuilder{
		FilterRequestBuilder{
			QueryRequestBuilder: QueryRequestBuilder{
				client:     b.client,
				path:       b.path,
				httpMethod: "GET",
				header:     b.header,
				params:     b.params,
			},
			negateNext: false,
		},
	}
}

// Insert starts building an INSERT request with the provided JSON data.
func (b *RequestBuilder) Insert(json interface{}) *QueryRequestBuilder {
	b.header.Set("Prefer", "return=representation")
	return &QueryRequestBuilder{
		client:     b.client,
		path:       b.path,
		httpMethod: http.MethodPost,
		json:       json,
		params:     b.params,
		header:     b.header,
	}
}

// Upsert starts building an UPSERT request with the provided JSON data.
func (b *RequestBuilder) Upsert(json interface{}) *QueryRequestBuilder {
	b.header.Set("Prefer", "return=representation,resolution=merge-duplicates")
	return &QueryRequestBuilder{
		client:     b.client,
		path:       b.path,
		httpMethod: http.MethodPost,
		json:       json,
		params:     b.params,
		header:     b.header,
	}
}

// Update starts building an UPDATE request with the provided JSON data.
func (b *RequestBuilder) Update(json interface{}) *FilterRequestBuilder {
	b.header.Set("Prefer", "return=representation")
	return &FilterRequestBuilder{
		QueryRequestBuilder: QueryRequestBuilder{
			client:     b.client,
			path:       b.path,
			httpMethod: http.MethodPatch,
			json:       json,
			params:     b.params,
			header:     b.header,
		},
		negateNext: false,
	}
}

// Delete starts building a DELETE request.
func (b *RequestBuilder) Delete() *FilterRequestBuilder {
	return &FilterRequestBuilder{
		QueryRequestBuilder: QueryRequestBuilder{
			client:     b.client,
			path:       b.path,
			httpMethod: http.MethodDelete,
			json:       nil,
			params:     b.params,
			header:     b.header,
		},
		negateNext: false,
	}
}

// QueryRequestBuilder represents a builder for query requests.
type QueryRequestBuilder struct {
	client     *Client
	params     url.Values
	header     http.Header
	path       string
	httpMethod string
	json       interface{}
	isCount    bool
}

// Execute sends the query request and unmarshals the response JSON into the provided object.
func (b *QueryRequestBuilder) Execute(r interface{}) error {
	return b.ExecuteWithContext(context.Background(), r)
}

// ExecuteWithContext sends the query request with the provided context and unmarshals the response JSON into the provided object.
func (b *QueryRequestBuilder) ExecuteWithContext(ctx context.Context, r interface{}) error {
	data, err := json.Marshal(b.json)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, b.httpMethod, b.path, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	query, err := url.QueryUnescape(b.params.Encode())

	if err != nil {
		return err
	}

	req.URL.RawQuery = query

	req.Header = b.client.Headers()

	// Inject/override custom headers
	for key, vals := range b.header {
		for _, val := range vals {
			req.Header.Set(key, val)
		}
	}

	req.URL.Path = req.URL.Path[1:]
	req.URL = b.client.Transport.baseURL.ResolveReference(req.URL)

	resp, err := b.client.session.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	statusOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	if !statusOK {
		reqError := RequestError{HTTPStatusCode: resp.StatusCode}

		if err = json.Unmarshal(body, &reqError); err != nil {
			return err
		}

		return &reqError
	}

	if resp.StatusCode != http.StatusNoContent && r != nil {
		if b.isCount {
			contentRange := resp.Header.Get("Content-Range")
			contentRangeParts := strings.Split(contentRange, "/")
			if len(contentRangeParts) != 2 {
				return errors.New("invalid content range returned from count request")
			}
			return json.Unmarshal([]byte(contentRangeParts[1]), r)
		}

		if err = json.Unmarshal(body, r); err != nil {
			return err
		}
	}

	return nil
}

// FilterRequestBuilder represents a builder for filter requests.
type FilterRequestBuilder struct {
	QueryRequestBuilder
	negateNext bool
}

// Not negates the next filter condition.
func (b *FilterRequestBuilder) Not() *FilterRequestBuilder {
	b.negateNext = true
	return b
}

// Filter adds a filter condition to the request.
func (b *FilterRequestBuilder) Filter(column, operator, criteria string) *FilterRequestBuilder {
	if b.negateNext {
		b.negateNext = false
		operator = "not." + operator
	}
	b.params.Add(column, operator+"."+criteria)
	return b
}

// Eq adds an equality filter condition to the request.
func (b *FilterRequestBuilder) Eq(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "eq", SanitizeParam(value))
}

// Neq adds a not-equal filter condition to the request.
func (b *FilterRequestBuilder) Neq(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "neq", SanitizeParam(value))
}

// Gt adds a greater-than filter condition to the request.
func (b *FilterRequestBuilder) Gt(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "gt", SanitizeParam(value))
}

// Gte adds a greater-than-or-equal filter condition to the request.
func (b *FilterRequestBuilder) Gte(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "gte", SanitizeParam(value))
}

// Lt adds a less-than filter condition to the request.
func (b *FilterRequestBuilder) Lt(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "lt", SanitizeParam(value))
}

// Lte adds a less-than-or-equal filter condition to the request.
func (b *FilterRequestBuilder) Lte(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "lte", SanitizeParam(value))
}

// Is adds an IS filter condition to the request.
func (b *FilterRequestBuilder) Is(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "is", SanitizeParam(value))
}

// Like adds a LIKE filter condition to the request.
func (b *FilterRequestBuilder) Like(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "like", SanitizeParam(value))
}

// Ilike adds a ILIKE filter condition to the request.
func (b *FilterRequestBuilder) Ilike(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "ilike", SanitizeParam(value))
}

// Fts adds a full-text search filter condition to the request.
func (b *FilterRequestBuilder) Fts(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "fts", SanitizeParam(value))
}

// Plfts adds a phrase-level full-text search filter condition to the request.
func (b *FilterRequestBuilder) Plfts(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "plfts", SanitizeParam(value))
}

// Phfts adds a phrase-headline-level full-text search filter condition to the request.
func (b *FilterRequestBuilder) Phfts(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "phfts", SanitizeParam(value))
}

// Wfts adds a word-level full-text search filter condition to the request.
func (b *FilterRequestBuilder) Wfts(column, value string) *FilterRequestBuilder {
	return b.Filter(column, "wfts", SanitizeParam(value))
}

// In adds an IN filter condition to the request.
func (b *FilterRequestBuilder) In(column string, values []string) *FilterRequestBuilder {
	sanitized := make([]string, len(values))
	for i, value := range values {
		sanitized[i] = SanitizeParam(value)
	}
	return b.Filter(column, "in", fmt.Sprintf("(%s)", strings.Join(sanitized, ",")))
}

// Cs adds a contains set filter condition to the request.
func (b *FilterRequestBuilder) Cs(column string, values []string) *FilterRequestBuilder {
	sanitized := make([]string, len(values))
	for i, value := range values {
		sanitized[i] = SanitizeParam(value)
	}
	return b.Filter(column, "cs", fmt.Sprintf("{%s}", strings.Join(sanitized, ",")))
}

// Cd adds a contained by set filter condition to the request.
func (b *FilterRequestBuilder) Cd(column string, values []string) *FilterRequestBuilder {
	sanitized := make([]string, len(values))
	for i, value := range values {
		sanitized[i] = SanitizeParam(value)
	}
	return b.Filter(column, "cd", fmt.Sprintf("{%s}", strings.Join(sanitized, ",")))
}

// Ov adds an overlaps set filter condition to the request.
func (b *FilterRequestBuilder) Ov(column string, values []string) *FilterRequestBuilder {
	sanitized := make([]string, len(values))
	for i, value := range values {
		sanitized[i] = SanitizeParam(value)
	}
	return b.Filter(column, "ov", fmt.Sprintf("{%s}", strings.Join(sanitized, ",")))
}

// Sl adds a strictly left of filter condition to the request.
func (b *FilterRequestBuilder) Sl(column string, from, to int) *FilterRequestBuilder {
	return b.Filter(column, "sl", fmt.Sprintf("(%d,%d)", from, to))
}

// Sr adds a strictly right of filter condition to the request.
func (b *FilterRequestBuilder) Sr(column string, from, to int) *FilterRequestBuilder {
	return b.Filter(column, "sr", fmt.Sprintf("(%d,%d)", from, to))
}

// Nxl adds a not strictly left of filter condition to the request.
func (b *FilterRequestBuilder) Nxl(column string, from, to int) *FilterRequestBuilder {
	return b.Filter(column, "nxl", fmt.Sprintf("(%d,%d)", from, to))
}

// Nxr adds a not strictly right of filter condition to the request.
func (b *FilterRequestBuilder) Nxr(column string, from, to int) *FilterRequestBuilder {
	return b.Filter(column, "nxr", fmt.Sprintf("(%d,%d)", from, to))
}

// Ad adds an adjacent to filter condition to the request.
func (b *FilterRequestBuilder) Ad(column string, values []string) *FilterRequestBuilder {
	sanitized := make([]string, len(values))
	for i, value := range values {
		sanitized[i] = SanitizeParam(value)
	}
	return b.Filter(column, "ad", fmt.Sprintf("{%s}", strings.Join(sanitized, ",")))
}

// IsNull adds a is null filter condition to the request.
func (b *FilterRequestBuilder) IsNull(column string) *FilterRequestBuilder {
	return b.Filter(column, "is", "null")
}

// FilterRequestBuilder represents a builder for SELECT requests.
type SelectRequestBuilder struct {
	FilterRequestBuilder
}

// OrderBy sets the ordering column and direction for the SELECT request.
func (b *SelectRequestBuilder) OrderBy(column, direction string) *SelectRequestBuilder {
	b.params.Set("order", column+"."+direction)
	return b
}

// Range sets the range of rows to be returned for the SELECT request.
func (b *SelectRequestBuilder) Range(from, to int) *SelectRequestBuilder {
	b.params.Set("range", fmt.Sprintf("%d-%d", from, to))
	return b
}

// SingleRow sets the single row behavior for the SELECT request.
func (b *SelectRequestBuilder) SingleRow() *SelectRequestBuilder {
	b.params.Set("single-row", "true")
	return b
}

// OnlyPayload sets the only payload behavior for the SELECT request.
func (b *SelectRequestBuilder) OnlyPayload() *SelectRequestBuilder {
	b.params.Set("only-payload", "true")
	return b
}

// WithoutCount sets the without count behavior for the SELECT request.
func (b *SelectRequestBuilder) WithoutCount() *SelectRequestBuilder {
	b.params.Set("without-count", "true")
	return b
}

// SingleValue sets the single value behavior for the SELECT request.
func (b *SelectRequestBuilder) SingleValue() *SelectRequestBuilder {
	b.params.Set("single-value", "true")
	return b
}

// Limit will restrict the number of results via the Range header.
func (b *SelectRequestBuilder) Limit(size int) *SelectRequestBuilder {
	return b.LimitWithOffset(size, 0)
}

// LimitWithOffset is essentially pagination by providing a start and end index.
func (b *SelectRequestBuilder) LimitWithOffset(size int, start int) *SelectRequestBuilder {
	b.header.Set("Range-Unit", "items")
	b.header.Set("Range", fmt.Sprintf("%d-%d", start, start+size-1))
	return b
}

func (b *SelectRequestBuilder) Single() *SelectRequestBuilder {
	b.header.Set("Accept", "application/vnd.pgrst.object+json")
	return b
}

// Count will convert the request from selecting content to instead perform only a requets for a count of objects.
// It will perform a HEAD request instead of a full GET. The result from this query will now be a count instead of rows.
func (b *SelectRequestBuilder) Count() *SelectRequestBuilder {
	b.header.Set("Prefer", "count=exact")
	b.isCount = true
	b.httpMethod = "HEAD"
	return b
}
