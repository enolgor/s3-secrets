package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	L "github.com/enolgor/go-utils/lambda"
	"github.com/enolgor/go-utils/server"
	"gopkg.in/yaml.v3"
)

var secrets map[string]any = map[string]any{}

var bucket string = os.Getenv("S3_BUCKET")
var file string = os.Getenv("S3_FILE")
var token string = os.Getenv("TOKEN")

func init() {
	if err := readSecrets(bucket, file); err != nil {
		panic(err)
	}
}

func main() {
	mux := server.NewRouter().
		Get("/(.*)", server.Handle(check, handle))
	lambda.Start(L.Handler(mux))
}

func check(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Query().Get("token") != token {
		server.Response(w).Status(http.StatusUnauthorized).WithBody("unauthorized").AsTextPlain()
		return false
	}
	return true
}

func handle(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	secret, ok, err := loadSecrets(path)
	if err != nil {
		server.Response(w).Status(http.StatusInternalServerError).WithBody(err.Error()).AsTextPlain()
		return
	}
	if !ok {
		server.Response(w).Status(http.StatusNotFound).WithBody("secret not found").AsTextPlain()
		return
	}
	switch v := secret.(type) {
	case string:
		server.Response(w).Status(http.StatusOK).WithBody(v).AsTextPlain()
		return
	case map[string]interface{}:
		server.Response(w).Status(http.StatusOK).WithBody(v).AsJson()
		return
	default:
		server.Response(w).Status(http.StatusInternalServerError).WithBody("unknown secret type").AsTextPlain()
		return
	}
}

func readSecrets(bucket, file string) error {
	config, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return err
	}
	client := s3.NewFromConfig(config)
	result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(file),
	})
	if err != nil {
		return err
	}
	return yaml.NewDecoder(result.Body).Decode(secrets)
}

func loadSecrets(path string) (any, bool, error) {
	trimmed := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "/")
	parts := strings.Split(trimmed, "/")
	root := secrets
	for i := range parts {
		value, ok := root[parts[i]]
		if !ok {
			return nil, false, nil
		}
		switch v := value.(type) {
		case string:
			if i == len(parts)-1 {
				return v, true, nil
			}
			return nil, false, nil
		case map[string]interface{}:
			root = v
		default:
			return nil, false, fmt.Errorf("malformed secrets file")
		}
	}
	return root, true, nil
}
