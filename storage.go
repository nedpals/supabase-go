package supabase

import (
	"bytes"
	"context"
	"encoding/json"
	// "errors"
	"fmt"
	"net/http"
)

type Storage struct {
	client *Client
	// *bucket
	// *file
}

type bucket struct {
	Name string `json:"name"`
}
type bucketResponse struct {
	Id 	 string	`json:"id"`
	Name string `json:"name"`
	Owner string	`json:"owner"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
}
type bucketMessage struct {
	Message 	 string	`json:"message"`
}
// type file struct{}

type bucketOption struct {
	Id  	string 	`json:"id"`
	Name 	string 	`json:"name"`
	Public 	bool 	`json:"public"`
}


type storageError struct {
	Err     string `json:"error"`
	Message string `json:"message"`
}

// CreateBucket creates a new storage bucket
// @param: option:  a bucketOption with the name and id of the bucket you want to create
// @returns: bucket: a response with the details of the bucket of the bucket created
func (s *Storage) CreateBucket(ctx context.Context, option bucketOption) (*bucket, error) {
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
// @param id:  the id of the bucket
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
// @param id:  the id of the bucket
// @param option:  the options to be updated
// @returns bucketMessage: a successful response message or failed 
func (s *Storage) UpdateBucket(ctx context.Context, id string, option bucketOption) (*bucketMessage, error) {
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
// @param id:  the id of the bucket
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

