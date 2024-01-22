package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetShort(t *testing.T) {
	conf := config.NewConfig()
	logger := logrus.New()
	h := NewHandler(
		service.NewService(
			repository.NewRepository(conf.FileStoragePath), conf), logger).
		InitRoutes()

	ts := httptest.NewServer(h.r)
	defer ts.Close()

	// save some values
	testURL1 := "https://practicum.yandex.ru/"
	testURL2 := "https://practicum2.yandex.ru/"
	localURL := "http://localhost:8080/"

	testShort1, _ := h.s.NewShort(testURL1)
	testShort2, _ := h.s.NewShort(testURL2)
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
			name: "PUT some exist. Wrong method 2",
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

			defer req.Context()
			res, err := http.DefaultTransport.RoundTrip(req)
			//res, err := http.DefaultClient.Do(req)

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
	conf := config.NewConfig()
	logger := logrus.New()
	h := NewHandler(
		service.NewService(
			repository.NewRepository(conf.FileStoragePath), conf), logger).
		InitRoutes()

	ts := httptest.NewServer(h.r)
	defer ts.Close()

	// save some values
	testURL := "https://practicum.yandex.ru/"

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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			req, err := http.NewRequest(test.args.method, ts.URL+test.args.path, strings.NewReader(test.args.data))
			require.NoError(t, err)

			defer req.Context()

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
	conf := config.NewConfig()
	logger := logrus.New()
	h := NewHandler(
		service.NewService(
			repository.NewRepository(conf.FileStoragePath), conf), logger).
		InitRoutes()

	ts := httptest.NewServer(h.r)
	defer ts.Close()

	// save some values
	testURL := "https://practicum.yandex.ru/"

	type want struct {
		code            int
		responseContain string
		contentType     string
	}
	type args struct {
		method string
		data   interface{}
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
				data: map[string]interface{}{
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
				data: map[string]interface{}{
					"somekey": testURL,
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
				data:   testURL,
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

			req, err := http.NewRequest(test.args.method, ts.URL+config.APIRoute+config.ShortenRoute, b)
			require.NoError(t, err)

			defer req.Context()

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
