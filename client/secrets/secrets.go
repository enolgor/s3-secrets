package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

type Client interface {
	GetSecret(key string) (string, error)
	GetSecrets(key string) (map[string]interface{}, error)
}

type _client struct {
	req *http.Request
}

var ErrSecretNotFound error = errors.New("secret not found")

func NewClient(rawurl, token string) (Client, error) {
	req, err := http.NewRequest("GET", strings.TrimSuffix(rawurl, "/")+"?token="+token, nil)
	if err != nil {
		return nil, err
	}
	return &_client{req}, nil
}

func (c _client) GetSecret(key string) (string, error) {
	req := c.req.Clone(context.TODO())
	req.URL.Path = "/" + strings.TrimPrefix(key, "/")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", ErrSecretNotFound
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(buf))
	}
	return string(buf), nil
}

func (c _client) GetSecrets(key string) (map[string]interface{}, error) {
	req := c.req.Clone(context.TODO())
	req.URL.Path = "/" + strings.TrimPrefix(key, "/")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrSecretNotFound
	}
	if resp.StatusCode != http.StatusOK {
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(buf))
	}
	secrets := map[string]any{}
	dec := json.NewDecoder(resp.Body)
	return secrets, dec.Decode(&secrets)
}
