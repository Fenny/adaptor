// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package adaptor

import (
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/gofiber/fiber"
	"github.com/valyala/fasthttp"
)

func Test_HTTPHandler(t *testing.T) {
	expectedMethod := fiber.MethodPost
	expectedProto := "HTTP/1.1"
	expectedProtoMajor := 1
	expectedProtoMinor := 1
	expectedRequestURI := "/foo/bar?baz=123"
	expectedBody := "body 123 foo bar baz"
	expectedContentLength := len(expectedBody)
	expectedTransferEncoding := "encoding"
	expectedHost := "foobar.com"
	expectedRemoteAddr := "1.2.3.4:6789"
	expectedHeader := map[string]string{
		"Foo-Bar":         "baz",
		"Abc":             "defg",
		"XXX-Remote-Addr": "123.43.4543.345",
	}
	expectedURL, err := url.ParseRequestURI(expectedRequestURI)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	expectedContextKey := "contextKey"
	expectedContextValue := "contextValue"

	callsCount := 0
	nethttpH := func(w http.ResponseWriter, r *http.Request) {
		callsCount++
		if r.Method != expectedMethod {
			t.Fatalf("unexpected method %q. Expecting %q", r.Method, expectedMethod)
		}
		if r.Proto != expectedProto {
			t.Fatalf("unexpected proto %q. Expecting %q", r.Proto, expectedProto)
		}
		if r.ProtoMajor != expectedProtoMajor {
			t.Fatalf("unexpected protoMajor %d. Expecting %d", r.ProtoMajor, expectedProtoMajor)
		}
		if r.ProtoMinor != expectedProtoMinor {
			t.Fatalf("unexpected protoMinor %d. Expecting %d", r.ProtoMinor, expectedProtoMinor)
		}
		if r.RequestURI != expectedRequestURI {
			t.Fatalf("unexpected requestURI %q. Expecting %q", r.RequestURI, expectedRequestURI)
		}
		if r.ContentLength != int64(expectedContentLength) {
			t.Fatalf("unexpected contentLength %d. Expecting %d", r.ContentLength, expectedContentLength)
		}
		if len(r.TransferEncoding) != 1 || r.TransferEncoding[0] != expectedTransferEncoding {
			t.Fatalf("unexpected transferEncoding %q. Expecting %q", r.TransferEncoding, expectedTransferEncoding)
		}
		if r.Host != expectedHost {
			t.Fatalf("unexpected host %q. Expecting %q", r.Host, expectedHost)
		}
		if r.RemoteAddr != expectedRemoteAddr {
			t.Fatalf("unexpected remoteAddr %q. Expecting %q", r.RemoteAddr, expectedRemoteAddr)
		}
		body, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			t.Fatalf("unexpected error when reading request body: %s", err)
		}
		if string(body) != expectedBody {
			t.Fatalf("unexpected body %q. Expecting %q", body, expectedBody)
		}
		if !reflect.DeepEqual(r.URL, expectedURL) {
			t.Fatalf("unexpected URL: %#v. Expecting %#v", r.URL, expectedURL)
		}
		if r.Context().Value(expectedContextKey) != expectedContextValue {
			t.Fatalf("unexpected context value for key %q. Expecting %q", expectedContextKey, expectedContextValue)
		}

		for k, expectedV := range expectedHeader {
			v := r.Header.Get(k)
			if v != expectedV {
				t.Fatalf("unexpected header value %q for key %q. Expecting %q", v, k, expectedV)
			}
		}

		w.Header().Set("Header1", "value1")
		w.Header().Set("Header2", "value2")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "request body is %q", body)
	}
	fiberH := HTTPHandler(http.HandlerFunc(nethttpH))
	fiberH = setFiberContextValueMiddleware(fiberH, expectedContextKey, expectedContextValue)

	var fctx fasthttp.RequestCtx
	var req fasthttp.Request

	req.Header.SetMethod(expectedMethod)
	req.SetRequestURI(expectedRequestURI)
	req.Header.SetHost(expectedHost)
	req.Header.Add(fiber.HeaderTransferEncoding, expectedTransferEncoding)
	req.BodyWriter().Write([]byte(expectedBody)) // nolint:errcheck
	for k, v := range expectedHeader {
		req.Header.Set(k, v)
	}

	remoteAddr, err := net.ResolveTCPAddr("tcp", expectedRemoteAddr)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	fctx.Init(&req, remoteAddr, nil)
	app := fiber.New()
	ctx := app.AcquireCtx(&fctx)
	defer app.ReleaseCtx(ctx)

	fiberH(ctx)

	if callsCount != 1 {
		t.Fatalf("unexpected callsCount: %d. Expecting 1", callsCount)
	}

	resp := &fctx.Response
	if resp.StatusCode() != fiber.StatusBadRequest {
		t.Fatalf("unexpected statusCode: %d. Expecting %d", resp.StatusCode(), fiber.StatusBadRequest)
	}
	if string(resp.Header.Peek("Header1")) != "value1" {
		t.Fatalf("unexpected header value: %q. Expecting %q", resp.Header.Peek("Header1"), "value1")
	}
	if string(resp.Header.Peek("Header2")) != "value2" {
		t.Fatalf("unexpected header value: %q. Expecting %q", resp.Header.Peek("Header2"), "value2")
	}
	expectedResponseBody := fmt.Sprintf("request body is %q", expectedBody)
	if string(resp.Body()) != expectedResponseBody {
		t.Fatalf("unexpected response body %q. Expecting %q", resp.Body(), expectedResponseBody)
	}
}

func Test_FiberHandlerFunc(t *testing.T) {
	expectedMethod := fiber.MethodPost
	expectedRequestURI := "/foo/bar?baz=123"
	expectedBody := "body 123 foo bar baz"
	expectedContentLength := len(expectedBody)
	expectedHost := "foobar.com"
	expectedRemoteAddr := "1.2.3.4:6789"
	expectedHeader := map[string]string{
		"Foo-Bar":         "baz",
		"Abc":             "defg",
		"XXX-Remote-Addr": "123.43.4543.345",
	}
	expectedURL, err := url.ParseRequestURI(expectedRequestURI)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	callsCount := 0
	fiberH := func(c *fiber.Ctx) {
		callsCount++
		if c.Method() != expectedMethod {
			t.Fatalf("unexpected method %q. Expecting %q", c.Method(), expectedMethod)
		}
		if string(c.Fasthttp.RequestURI()) != expectedRequestURI {
			t.Fatalf("unexpected requestURI %q. Expecting %q", string(c.Fasthttp.RequestURI()), expectedRequestURI)
		}
		contentLength := c.Fasthttp.Request.Header.ContentLength()
		if contentLength != expectedContentLength {
			t.Fatalf("unexpected contentLength %d. Expecting %d", contentLength, expectedContentLength)
		}
		if c.Hostname() != expectedHost {
			t.Fatalf("unexpected host %q. Expecting %q", c.Hostname(), expectedHost)
		}
		remoteAddr := c.Fasthttp.RemoteAddr().String()
		if remoteAddr != expectedRemoteAddr {
			t.Fatalf("unexpected remoteAddr %q. Expecting %q", remoteAddr, expectedRemoteAddr)
		}
		body := c.Body()
		if body != expectedBody {
			t.Fatalf("unexpected body %q. Expecting %q", body, expectedBody)
		}
		if c.OriginalURL() != expectedURL.String() {
			t.Fatalf("unexpected URL: %#v. Expecting %#v", c.OriginalURL(), expectedURL)
		}

		for k, expectedV := range expectedHeader {
			v := c.Get(k)
			if v != expectedV {
				t.Fatalf("unexpected header value %q for key %q. Expecting %q", v, k, expectedV)
			}
		}

		c.Set("Header1", "value1")
		c.Set("Header2", "value2")
		c.Status(fiber.StatusBadRequest)
		c.Write(fmt.Sprintf("request body is %q", body))
	}
	nethttpH := FiberHandlerFunc(fiberH)

	var r http.Request

	r.Method = expectedMethod
	r.Body = &netHTTPBody{[]byte(expectedBody)}
	r.RequestURI = expectedRequestURI
	r.ContentLength = int64(expectedContentLength)
	r.Host = expectedHost
	r.RemoteAddr = expectedRemoteAddr

	hdr := make(http.Header)
	for k, v := range expectedHeader {
		hdr.Set(k, v)
	}
	r.Header = hdr

	var w netHTTPResponseWriter
	nethttpH.ServeHTTP(&w, &r)

	if w.StatusCode() != http.StatusBadRequest {
		t.Fatalf("unexpected statusCode: %d. Expecting %d", w.StatusCode(), http.StatusBadRequest)
	}
	if w.Header().Get("Header1") != "value1" {
		t.Fatalf("unexpected header value: %q. Expecting %q", w.Header().Get("Header1"), "value1")
	}
	if w.Header().Get("Header2") != "value2" {
		t.Fatalf("unexpected header value: %q. Expecting %q", w.Header().Get("Header2"), "value2")
	}
	expectedResponseBody := fmt.Sprintf("request body is %q", expectedBody)
	if string(w.body) != expectedResponseBody {
		t.Fatalf("unexpected response body %q. Expecting %q", string(w.body), expectedResponseBody)
	}
}

func setFiberContextValueMiddleware(next func(*fiber.Ctx), key string, value interface{}) func(*fiber.Ctx) {
	return func(c *fiber.Ctx) {
		c.Locals(key, value)
		next(c)
	}
}

type netHTTPBody struct {
	b []byte
}

func (r *netHTTPBody) Read(p []byte) (int, error) {
	if len(r.b) == 0 {
		return 0, io.EOF
	}
	n := copy(p, r.b)
	r.b = r.b[n:]
	return n, nil
}

func (r *netHTTPBody) Close() error {
	r.b = r.b[:0]
	return nil
}

type netHTTPResponseWriter struct {
	statusCode int
	h          http.Header
	body       []byte
}

func (w *netHTTPResponseWriter) StatusCode() int {
	if w.statusCode == 0 {
		return http.StatusOK
	}
	return w.statusCode
}

func (w *netHTTPResponseWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}
	return w.h
}

func (w *netHTTPResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *netHTTPResponseWriter) Write(p []byte) (int, error) {
	w.body = append(w.body, p...)
	return len(p), nil
}
