package supabase

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

type Storage struct {
	client *Client
}

// Storage buckets methods

type bucket struct {
	Name string `json:"name"`
}
type bucketResponse struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Owner      string `json:"owner"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}
type bucketMessage struct {
	Message string `json:"message"`
}

type BucketOption struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Public bool   `json:"public"`
}

type storageError struct {
	Err     string `json:"error"`
	Message string `json:"message"`
}

var ErrNotFound = errors.New("file not found")

// CreateBucket creates a new storage bucket
// @param: option:  a bucketOption with the name and id of the bucket you want to create
// @returns: bucket: a response with the details of the bucket of the bucket created
func (s *Storage) CreateBucket(ctx context.Context, option BucketOption) (*bucket, error) {
	reqBody, _ := json.Marshal(option)
	reqURL := fmt.Sprintf("%s/%s/bucket", s.client.BaseURL, StorageEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	injectAuthorizationHeader(req, s.client.apiKey)
	res := bucket{}
	errRes := storageError{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf("%s\n%s", errRes.Err, errRes.Message)
	}

	return &res, nil
}

// GetBucket retrieves a bucket by its id
// @param: id:  the id of the bucket
// @returns: bucketResponse: a response with the details of the bucket
func (s *Storage) GetBucket(ctx context.Context, id string) (*bucketResponse, error) {
	// reqBody, _ := json.Marshal()
	reqURL := fmt.Sprintf("%s/%s/bucket/%s", s.client.BaseURL, StorageEndpoint, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	injectAuthorizationHeader(req, s.client.apiKey)
	res := bucketResponse{}
	errRes := storageError{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf("%s \n %s", errRes.Err, errRes.Message)
	}

	return &res, nil
}

// ListBucket retrieves all buckets ina supabase storage
// @returns: []bucketResponse: a response with the details of all the bucket
func (s *Storage) ListBuckets(ctx context.Context) (*[]bucketResponse, error) {
	// reqBody, _ := json.Marshal()
	reqURL := fmt.Sprintf("%s/%s/bucket/", s.client.BaseURL, StorageEndpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	injectAuthorizationHeader(req, s.client.apiKey)
	res := []bucketResponse{}
	errRes := storageError{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf("%s \n %s", errRes.Err, errRes.Message)
	}

	return &res, nil
}

// EmptyBucket  empties the object of a bucket by id
// @param: id:  the id of the bucket
// @returns bucketMessage: a successful response message or failed
func (s *Storage) EmptyBucket(ctx context.Context, id string) (*bucketMessage, error) {
	// reqBody, _ := json.Marshal()
	reqURL := fmt.Sprintf("%s/%s/bucket/%s/empty", s.client.BaseURL, StorageEndpoint, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	injectAuthorizationHeader(req, s.client.apiKey)
	res := bucketMessage{}
	errRes := storageError{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf("%s \n %s", errRes.Err, errRes.Message)
	}

	return &res, nil
}

// UpdateBucket updates a bucket by its id
// @param: id:  the id of the bucket
// @param: option:  the options to be updated
// @returns bucketMessage: a successful response message or failed
func (s *Storage) UpdateBucket(ctx context.Context, id string, option BucketOption) (*bucketMessage, error) {
	reqBody, _ := json.Marshal(option)
	reqURL := fmt.Sprintf("%s/%s/bucket/%s", s.client.BaseURL, StorageEndpoint, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, reqURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	injectAuthorizationHeader(req, s.client.apiKey)
	res := bucketMessage{}
	errRes := storageError{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf("%s \n %s", errRes.Err, errRes.Message)
	}

	return &res, nil
}

// DeleteBucket deletes a bucket by its id, a bucket can't be deleted except emptied
// @param: id:  the id of the bucket
// @returns bucketMessage: a successful response message or failed
func (s *Storage) DeleteBucket(ctx context.Context, id string) (*bucketResponse, error) {
	// reqBody, _ := json.Marshal()
	reqURL := fmt.Sprintf("%s/%s/bucket/%s", s.client.BaseURL, StorageEndpoint, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	injectAuthorizationHeader(req, s.client.apiKey)
	res := bucketResponse{}
	errRes := storageError{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return nil, fmt.Errorf("%s\n%s", errRes.Err, errRes.Message)
	}

	return &res, nil
}

func (s *Storage) From(bucketId string) *file {
	return &file{BucketId: bucketId, storage: s}
}

// Storage Objects methods

type file struct {
	BucketId string
	storage  *Storage
}

type SortBy struct {
	Column string `json:"column"`
	Order  string `json:"order"`
}

type FileResponse struct {
	Key     string `json:"key"`
	Message string `json:"message"`
}

type FileErrorResponse struct {
	Status     string `json:"statusCode"`
	ShortError string `json:"error"`
	Message    string `json:"message"`
}

func (err *FileErrorResponse) Error() string {
	return err.ShortError + ": " + err.Message
}

type FileSearchOptions struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	SortBy SortBy `json:"sortBy"`
}

type FileObject struct {
	Name           string      `json:"name"`
	BucketId       string      `json:"bucket_id"`
	Owner          string      `json:"owner"`
	Id             string      `json:"id"`
	UpdatedAt      string      `json:"updated_at"`
	CreatedAt      string      `json:"created_at"`
	LastAccessedAt string      `json:"last_accessed_at"`
	Metadata       interface{} `json:"metadata"`
	Buckets        bucket      `json:"buckets"`
}

type ListFileRequest struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	SortBy SortBy `json:"sortBy"`
	Prefix string `json:"prefix"`
}

type SignedUrlResponse struct {
	SignedUrl string `json:"signedURL"`
}

const (
	defaultLimit            = 100
	defaultOffset           = 0
	defaultFileCacheControl = "3600"
	defaultFileContent      = "text/plain;charset=UTF-8"
	defaultFileUpsert       = false
	defaultSortColumn       = "name"
	defaultSortOrder        = "asc"
)

type FileUploadOptions struct {
	CacheControl string
	ContentType  string
	Upsert       bool
}

func (f *file) UploadOrUpdate(path string, data io.Reader, update bool, opts *FileUploadOptions) FileResponse {
	// use default options, then override with whatever is passed in opts
	mergedOpts := FileUploadOptions{
		CacheControl: defaultFileCacheControl,
		ContentType:  defaultFileContent,
		Upsert:       defaultFileUpsert,
	}

	if opts != nil {
		if opts.CacheControl != "" {
			mergedOpts.CacheControl = opts.CacheControl
		}
		if opts.ContentType != "" {
			mergedOpts.ContentType = opts.ContentType
		}
		mergedOpts.Upsert = opts.Upsert
	}

	body := bufio.NewReader(data)
	_path := removeEmptyFolder(f.BucketId + "/" + path)
	client := &http.Client{}

	var (
		method string
		req    *http.Request
		res    *http.Response
		err    error
	)

	if update {
		method = http.MethodPut
	} else {
		method = http.MethodPost
	}

	reqURL := fmt.Sprintf("%s/%s/object/%s", f.storage.client.BaseURL, StorageEndpoint, _path)
	req, err = http.NewRequest(method, reqURL, body)
	if err != nil {
		panic(err)
	}

	injectAuthorizationHeader(req, f.storage.client.apiKey)
	req.Header.Set("cache-control", mergedOpts.CacheControl)
	req.Header.Set("content-type", mergedOpts.ContentType)
	req.Header.Set("x-upsert", strconv.FormatBool(mergedOpts.Upsert))

	res, err = client.Do(req)
	if err != nil {
		panic(err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var response FileResponse
	if err = json.Unmarshal(resBody, &response); err != nil {
		panic(err)
	}

	return response
}

// Update updates a file object in a storage bucket
func (f *file) Update(path string, data io.Reader, opts *FileUploadOptions) FileResponse {
	return f.UploadOrUpdate(path, data, true, opts)
}

// Upload uploads a file object to a storage bucket
func (f *file) Upload(path string, data io.Reader, opts *FileUploadOptions) FileResponse {
	return f.UploadOrUpdate(path, data, false, opts)
}

// Move moves a file object
func (f *file) Move(fromPath string, toPath string) FileResponse {
	_json, _ := json.Marshal(map[string]interface{}{
		"bucketId":      f.BucketId,
		"sourceKey":     fromPath,
		"destintionKey": toPath,
	})

	reqURL := fmt.Sprintf("%s/%s/object/move", f.storage.client.BaseURL, StorageEndpoint)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(_json))
	if err != nil {
		panic(err)
	}

	injectAuthorizationHeader(req, f.storage.client.apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var response FileResponse
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}

	return response
}

// CreatSignedUrl create a signed url for a file object
func (f *file) CreatSignedUrl(filePath string, expiresIn int) SignedUrlResponse {
	_json, _ := json.Marshal(map[string]interface{}{
		"expiresIn": expiresIn,
	})

	reqURL := fmt.Sprintf("%s/%s/object/sign/%s/%s", f.storage.client.BaseURL, StorageEndpoint, f.BucketId, filePath)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(_json))
	if err != nil {
		panic(err)
	}

	injectAuthorizationHeader(req, f.storage.client.apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var response SignedUrlResponse
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}
	response.SignedUrl = f.storage.client.BaseURL + response.SignedUrl

	return response
}

// GetPublicUrl get a public signed url of a file object
func (f *file) GetPublicUrl(filePath string) SignedUrlResponse {
	var response SignedUrlResponse
	response.SignedUrl = fmt.Sprintf("%s/%s/object/public/%s/%s", f.storage.client.BaseURL, StorageEndpoint, f.BucketId, filePath)
	return response
}

// Remove deletes a file object
func (f *file) Remove(filePaths []string) FileResponse {
	_json, _ := json.Marshal(map[string]interface{}{
		"prefixes": filePaths,
	})

	reqURL := fmt.Sprintf("%s/%s/object/%s", f.storage.client.BaseURL, StorageEndpoint, f.BucketId)
	req, err := http.NewRequest(http.MethodDelete, reqURL, bytes.NewBuffer(_json))
	if err != nil {
		panic(err)
	}

	injectAuthorizationHeader(req, f.storage.client.apiKey)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		var response FileResponse
		if err := json.Unmarshal(body, &response); err != nil {
			panic(err)
		}

		return response
	}

	return FileResponse{}
}

// List list all file object
func (f *file) List(queryPath string, options FileSearchOptions) []FileObject {
	if options.Limit == 0 {
		options.Limit = defaultLimit
	}
	if options.Offset == 0 {
		options.Offset = defaultOffset
	}
	if options.SortBy.Order == "" {
		options.SortBy.Order = defaultSortOrder
	}
	if options.SortBy.Column == "" {
		options.SortBy.Column = defaultSortColumn
	}

	_body := ListFileRequest{
		Limit:  options.Limit,
		Offset: options.Offset,
		SortBy: SortBy{
			Column: options.SortBy.Column,
			Order:  options.SortBy.Order,
		},
		Prefix: queryPath,
	}

	_json, _ := json.Marshal(_body)

	reqURL := fmt.Sprintf("%s/%s/object/list/%s", f.storage.client.BaseURL, StorageEndpoint, f.BucketId)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(_json))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		panic(err)
	}

	injectAuthorizationHeader(req, f.storage.client.apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var response []FileObject
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}

	return response
}

// Copy copies a file object
func (f *file) Copy(fromPath, toPath string) FileResponse {
	_json, _ := json.Marshal(map[string]interface{}{
		"bucketId":      f.BucketId,
		"sourceKey":     fromPath,
		"destintionKey": toPath,
	})

	reqURL := fmt.Sprintf("%s/%s/object/copy/%s", f.storage.client.BaseURL, StorageEndpoint, f.BucketId)
	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(_json))
	if err != nil {
		panic(err)
	}

	injectAuthorizationHeader(req, f.storage.client.apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	var response FileResponse
	if err := json.Unmarshal(body, &response); err != nil {
		panic(err)
	}

	return response
}

// Download  retrieves a file object, if it exists, otherwise return file response
func (f *file) Download(filePath string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/%s/object/authenticated/%s/%s", f.storage.client.BaseURL, StorageEndpoint, f.BucketId, filePath)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		panic(err)
	}

	injectAuthorizationHeader(req, f.storage.client.apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	// when not success, supabase will return json insted of file
	if res.StatusCode != 200 {
		var resErr *FileErrorResponse
		if err := json.Unmarshal(body, &resErr); err != nil {
			panic(err)
		}

		if resErr.Status == "404" {
			return nil, ErrNotFound
		}

		return nil, resErr
	}

	return body, nil
}

func removeEmptyFolder(filePath string) string {
	return regexp.MustCompile(`\/\/`).ReplaceAllString(filePath, "/")
}
