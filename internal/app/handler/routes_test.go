package handler

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"
	mocks "github.com/MrSwed/go-musthave-shortener/internal/app/mock/repository"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetShort(t *testing.T) {
	logger := logrus.New()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo)
	h := NewHandler(s, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	// save some values
	testURL1 := "https://practicum.yandex.ru/"
	testURL2 := "https://practicum2.yandex.ru/"

	testShort1 := fmt.Sprint(helper.NewRandShorter().RandStringBytes())
	testShort2 := fmt.Sprint(helper.NewRandShorter().RandStringBytes())

	_ = repo.EXPECT().GetFromShort(testShort1).Return(testURL1, nil).AnyTimes()
	_ = repo.EXPECT().GetFromShort(testShort2).Return(testURL2, nil).AnyTimes()
	_ = repo.EXPECT().GetFromShort(gomock.Any()).Return("", myErr.ErrNotExist).AnyTimes()

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo)
	h := NewHandler(s, logger).
		Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	// save some values
	testURL := "https://practicum.yandex.ru/"
	testShortURL := fmt.Sprintf("%s%s/%s", conf.Scheme, conf.BaseURL, helper.NewRandShorter().RandStringBytes())

	_ = repo.EXPECT().NewShort(testURL).Return(testShortURL, nil).AnyTimes()

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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := mocks.NewMockRepository(ctrl)

	s := service.NewService(repo)
	h := NewHandler(s, logger).Handler()

	ts := httptest.NewServer(h)
	defer ts.Close()

	// save some values
	testURL := "https://practicum.yandex.ru/"

	testShortURL := fmt.Sprintf("%s%s/%s", conf.Scheme, conf.BaseURL, helper.NewRandShorter().RandStringBytes())

	_ = repo.EXPECT().NewShort(testURL).Return(testShortURL, nil).AnyTimes()

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
		{
			name: "Check gzip compress answer",
			args: args{
				method: http.MethodPost,
				data: map[string]string{
					"url": testURL,
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
					"url": testURL,
				},
				headers: map[string]string{
					"Accept-Encoding": "gzip",
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
					"url": testURL,
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

			req, err := http.NewRequest(test.args.method, ts.URL+config.APIRoute+config.ShortenRoute, b)
			require.NoError(t, err)
			defer req.Context()

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
