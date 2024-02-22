// to test db set env DATABASE_DSN before run
package handler

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL         = "localhost:18080"
	fileStoragePath = "/tmp/short-url-db-tests.json"
	databaseDSN     = ""
)

func NewTestConfig() (c *config.Config) {
	c = config.NewConfig()
	c.BaseURL = baseURL
	c.FileStoragePath = fileStoragePath
	c.DatabaseDSN = databaseDSN
	c.WithEnv().CleanParameters()

	var err error
	if c.DatabaseDSN != "" {
		if db, err = sqlx.Open("pgx", c.DatabaseDSN); err != nil {
			log.Fatal(err)
		}
	}
	return
}

var (
	conf = NewTestConfig()
	db   *sqlx.DB
)

func TestHandler_GetShort(t *testing.T) {
	s := service.NewService(repository.NewRepository(repository.Config{StorageFile: conf.FileStoragePath, DB: db}), conf)
	h := NewHandler(s, &conf.Auth).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	// save some values
	testURL1 := "https://practicum.yandex.ru/"
	testURL2 := "https://practicum2.yandex.ru/"
	localURL := "http://" + baseURL + "/"
	ctx := context.TODO()
	testShort1, _ := s.NewShort(ctx, testURL1)
	testShort2, _ := s.NewShort(ctx, testURL2)
	testShort1 = strings.ReplaceAll(testShort1, localURL, "")
	testShort2 = strings.ReplaceAll(testShort2, localURL, "")
	type want struct {
		code            int
		responseContain string
		contentType     string
	}
	type args struct {
		method string
		path   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Get main Route",
			args: args{
				method: http.MethodGet,
				path:   "/",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Get some not exist",
			args: args{
				method: http.MethodGet,
				path:   "/somepage",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Get some not exist 1",
			args: args{
				method: http.MethodGet,
				path:   "/somepage/somepage",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Get some exist",
			args: args{
				method: http.MethodGet,
				path:   "/" + testShort1,
			},
			want: want{
				code:            http.StatusTemporaryRedirect,
				responseContain: testURL1,
				contentType:     "text/html; charset=utf-8",
			},
		},
		{
			name: "Get some exist 2",
			args: args{
				method: http.MethodGet,
				path:   "/" + testShort2,
			},
			want: want{
				code:            http.StatusTemporaryRedirect,
				responseContain: testURL2,
				contentType:     "text/html; charset=utf-8",
			},
		},
		{
			name: "PUT some. Wrong method 2",
			args: args{
				method: http.MethodPut,
				path:   "/" + testShort1,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req, err := http.NewRequest(test.args.method, ts.URL+test.args.path, nil)
			require.NoError(t, err)

			res, err := http.DefaultTransport.RoundTrip(req)

			require.NoError(t, err)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseContain != "" {
				assert.Contains(t, string(resBody), test.want.responseContain)
			}
			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestHandler_MakeShort(t *testing.T) {
	s := service.NewService(repository.NewRepository(repository.Config{StorageFile: conf.FileStoragePath, DB: db}), conf)
	h := NewHandler(s, &conf.Auth).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	// save some values
	testURL := "https://practicum.yandex.ru/?rand_Hash" + helper.NewRandShorter().RandStringBytes().String()
	testURLExist := "https://practicum.yandex.ru/?exist"
	ctx := context.TODO()
	_, _ = s.NewShort(ctx, testURLExist)

	type want struct {
		code            int
		responseContain string
		contentType     string
	}
	type args struct {
		method string
		path   string
		data   string
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Post Main",
			args: args{
				method: http.MethodPost,
				path:   "/",
				data:   testURL,
			},
			want: want{
				code:            http.StatusCreated,
				responseContain: conf.BaseURL,
			},
		},
		{
			name: "Post Main No body",
			args: args{
				method: http.MethodPost,
				path:   "/",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Post Main exist",
			args: args{
				method: http.MethodPost,
				path:   "/",
				data:   testURLExist,
			},
			want: want{
				code:            http.StatusConflict,
				responseContain: conf.BaseURL,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req, err := http.NewRequest(test.args.method, ts.URL+test.args.path, strings.NewReader(test.args.data))
			require.NoError(t, err)

			res, err := http.DefaultClient.Do(req)

			require.NoError(t, err)
			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			func() {
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					require.NoError(t, err)
				}(res.Body)
				resBody, err = io.ReadAll(res.Body)
				require.NoError(t, err)
			}()

			if test.want.responseContain != "" {
				assert.Contains(t, string(resBody), test.want.responseContain)
			}
			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}

func TestHandler_MakeShortJSON(t *testing.T) {
	s := service.NewService(repository.NewRepository(repository.Config{StorageFile: conf.FileStoragePath, DB: db}), conf)
	h := NewHandler(s, &conf.Auth).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	testURL := "https://practicum.yandex.ru/?rand_Hash" + helper.NewRandShorter().RandStringBytes().String()
	testURL1 := "https://practicum.yandex.ru/?rand_Hash" + helper.NewRandShorter().RandStringBytes().String()
	testURL2 := "https://practicum.yandex.ru/?rand_Hash" + helper.NewRandShorter().RandStringBytes().String()
	testURL3 := "https://practicum.yandex.ru/?rand_Hash" + helper.NewRandShorter().RandStringBytes().String()
	testURLExist := "https://practicum.yandex.ru/?exist"
	ctx := context.TODO()
	_, _ = s.NewShort(ctx, testURLExist)

	type want struct {
		code            int
		responseContain string
		contentType     string
		headers         map[string]string
	}
	type args struct {
		method  string
		headers map[string]string
		data    interface{}
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Create new shorten by JSON",
			args: args{
				method: http.MethodPost,
				data: map[string]string{
					"url": testURL,
				},
			},
			want: want{
				code:            http.StatusCreated,
				responseContain: conf.BaseURL,
				contentType:     "application/json; charset=utf-8",
			},
		},
		{
			name: "Create new exist by JSON",
			args: args{
				method: http.MethodPost,
				data: map[string]string{
					"url": testURLExist,
				},
			},
			want: want{
				code:            http.StatusConflict,
				responseContain: conf.BaseURL,
				contentType:     "application/json; charset=utf-8",
			},
		},
		{
			name: "Post No body",
			args: args{
				method: http.MethodPost,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Post wrong json body",
			args: args{
				method: http.MethodPost,
				data: map[string]string{
					"somekey": "https://practicum.yandex.ru/?2",
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Post wrong json body",
			args: args{
				method: http.MethodPost,
				data:   "https://practicum.yandex.ru/?3",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Wrong method (GET)",
			args: args{
				method: http.MethodGet,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Check gzip compress answer",
			args: args{
				method: http.MethodPost,
				data: map[string]string{
					"url": testURL1,
				},
				headers: map[string]string{
					"Accept-Encoding": "gzip",
				},
			},
			want: want{
				code:            http.StatusCreated,
				responseContain: conf.BaseURL,
				contentType:     "application/json; charset=utf-8",
				headers: map[string]string{
					"Content-Encoding": "gzip",
				},
			},
		},
		{
			name: "Check gzip decompress request",
			args: args{
				method: http.MethodPost,
				data: map[string]string{
					"url": testURL2,
				},
				headers: map[string]string{
					"Content-Encoding": "gzip",
				},
			},
			want: want{
				code:            http.StatusCreated,
				responseContain: conf.BaseURL,
				contentType:     "application/json; charset=utf-8",
			},
		},
		{
			name: "Check gzip compress/decompress request",
			args: args{
				method: http.MethodPost,
				data: map[string]string{
					"url": testURL3,
				},
				headers: map[string]string{
					"Content-Encoding": "gzip",
					"Accept-Encoding":  "gzip",
				},
			},
			want: want{
				code:            http.StatusCreated,
				responseContain: conf.BaseURL,
				contentType:     "application/json; charset=utf-8",
				headers: map[string]string{
					"Content-Encoding": "gzip",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			err := json.NewEncoder(b).Encode(test.args.data)
			require.NoError(t, err)
			if len(test.args.headers) > 0 && test.args.headers["Content-Encoding"] == "gzip" {
				compB := new(bytes.Buffer)
				w := gzip.NewWriter(compB)
				_, err = w.Write(b.Bytes())
				b = compB
				require.NoError(t, err)

				err = w.Flush()
				require.NoError(t, err)

				err = w.Close()
				require.NoError(t, err)
			}

			req, err := http.NewRequest(test.args.method, ts.URL+constant.APIRoute+constant.ShortenRoute, b)
			require.NoError(t, err)

			for k, v := range test.args.headers {
				req.Header.Add(k, v)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			resBody, err = io.ReadAll(res.Body)
			require.NoError(t, err)

			for k, v := range test.want.headers {
				assert.True(t, res.Header.Get(k) == v)
			}

			if test.want.responseContain != "" {
				assert.True(t, len(resBody) > 0)
				if len(test.args.headers) > 0 && test.args.headers["Accept-Encoding"] == "gzip" {
					b := bytes.NewBuffer(resBody)
					r, err := gzip.NewReader(b)
					if !errors.Is(err, io.EOF) {
						require.NoError(t, err)
					}
					var resB bytes.Buffer
					_, err = resB.ReadFrom(r)
					require.NoError(t, err)

					resBody = resB.Bytes()
					err = r.Close()
					require.NoError(t, err)
				}
				assert.Contains(t, string(resBody), test.want.responseContain)
			}

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
func TestHandler_MakeShortBatch(t *testing.T) {
	s := service.NewService(repository.NewRepository(repository.Config{StorageFile: conf.FileStoragePath, DB: db}), conf)
	h := NewHandler(s, &conf.Auth).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	type want struct {
		code            int
		responseContain string
		contentType     string
		headers         map[string]string
	}
	type args struct {
		method  string
		headers map[string]string
		data    interface{}
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "Create batch success",
			args: args{
				method: http.MethodPost,
				data: []map[string]string{
					{
						"correlation_id": "1",
						"original_url":   "http://practicum.yandex.ru/?1",
					},
					{
						"correlation_id": "2",
						"original_url":   "http://practicum.yandex.ru/?2",
					},
					{
						"correlation_id": "3",
						"original_url":   "http://practicum.yandex.ru/?3",
					},
				},
			},
			want: want{
				code:            http.StatusCreated,
				responseContain: conf.BaseURL,
				contentType:     "application/json; charset=utf-8",
			},
		},
		{
			name: "Post No body",
			args: args{
				method: http.MethodPost,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Post wrong json body",
			args: args{
				method: http.MethodPost,
				data: map[string]string{
					"somekey": "https://practicum.yandex.ru/?2",
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Post wrong json body",
			args: args{
				method: http.MethodPost,
				data:   "https://practicum.yandex.ru/?3",
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name: "Wrong method (GET)",
			args: args{
				method: http.MethodGet,
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			b := new(bytes.Buffer)
			err := json.NewEncoder(b).Encode(test.args.data)
			require.NoError(t, err)

			req, err := http.NewRequest(test.args.method, ts.URL+constant.APIRoute+constant.ShortenRoute+constant.BatchRoute, b)
			require.NoError(t, err)

			for k, v := range test.args.headers {
				req.Header.Add(k, v)
			}

			res, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer func() {
				err := res.Body.Close()
				require.NoError(t, err)
			}()

			var resBody []byte

			// проверяем код ответа
			require.Equal(t, test.want.code, res.StatusCode)
			resBody, err = io.ReadAll(res.Body)
			require.NoError(t, err)

			for k, v := range test.want.headers {
				assert.True(t, res.Header.Get(k) == v)
			}

			if test.want.responseContain != "" {
				assert.True(t, len(resBody) > 0)
				if len(test.args.headers) > 0 && test.args.headers["Accept-Encoding"] == "gzip" {
					b := bytes.NewBuffer(resBody)
					r, err := gzip.NewReader(b)
					if !errors.Is(err, io.EOF) {
						require.NoError(t, err)
					}
					var resB bytes.Buffer
					_, err = resB.ReadFrom(r)
					require.NoError(t, err)

					resBody = resB.Bytes()
					err = r.Close()
					require.NoError(t, err)
				}
				assert.Contains(t, string(resBody), test.want.responseContain)
			}

			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
		})
	}
}
