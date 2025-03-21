package httpexpect

import (
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestExpect_Constructors(t *testing.T) {
	t.Run("testing.T", func(t *testing.T) {
		_ = Default(&testing.T{}, "")
	})

	t.Run("testing.B", func(t *testing.T) {
		_ = Default(&testing.B{}, "")
	})

	t.Run("testing.TB", func(t *testing.T) {
		_ = Default(testing.TB(&testing.T{}), "")
	})
}

func TestExpect_Requests(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: reporter,
	}

	var reqs [8]*Request

	e := WithConfig(config)

	reqs[0] = e.Request("GET", "/url")
	reqs[1] = e.OPTIONS("/url")
	reqs[2] = e.HEAD("/url")
	reqs[3] = e.GET("/url")
	reqs[4] = e.POST("/url")
	reqs[5] = e.PUT("/url")
	reqs[6] = e.PATCH("/url")
	reqs[7] = e.DELETE("/url")

	assert.Equal(t, "GET", reqs[0].httpReq.Method)
	assert.Equal(t, "OPTIONS", reqs[1].httpReq.Method)
	assert.Equal(t, "HEAD", reqs[2].httpReq.Method)
	assert.Equal(t, "GET", reqs[3].httpReq.Method)
	assert.Equal(t, "POST", reqs[4].httpReq.Method)
	assert.Equal(t, "PUT", reqs[5].httpReq.Method)
	assert.Equal(t, "PATCH", reqs[6].httpReq.Method)
	assert.Equal(t, "DELETE", reqs[7].httpReq.Method)
}

func TestExpect_Builders(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		client := &mockClient{}

		reporter := NewAssertReporter(t)

		config := Config{
			Client:   client,
			Reporter: reporter,
		}

		e := WithConfig(config)

		var reqs1 []*Request

		e1 := e.Builder(func(r *Request) {
			reqs1 = append(reqs1, r)
		})

		var reqs2 []*Request

		e2 := e1.Builder(func(r *Request) {
			reqs2 = append(reqs2, r)
		})

		e.Request("GET", "/url")

		r1 := e1.Request("GET", "/url")
		r2 := e2.Request("GET", "/url")

		assert.Equal(t, 2, len(reqs1))
		assert.Equal(t, 1, len(reqs2))

		assert.Same(t, r1, reqs1[0])
		assert.Same(t, r2, reqs1[1])
		assert.Same(t, r2, reqs2[0])
	})

	t.Run("copying", func(t *testing.T) {
		client := &mockClient{}

		reporter := NewAssertReporter(t)

		config := Config{
			Client:   client,
			Reporter: reporter,
		}

		counter1 := 0
		counter2a := 0
		counter2b := 0

		e0 := WithConfig(config)

		// Simulate the case when many builders are added, and the builders slice
		// have some additioonal capacity. We are going to check that the slice
		// is cloned properly when a new builder is appended.
		for i := 0; i < 10; i++ {
			e0 = e0.Builder(func(r *Request) {})
		}

		e1 := e0.Builder(func(r *Request) {
			counter1++
		})

		e2a := e1.Builder(func(r *Request) {
			counter2a++
		})

		e2b := e1.Builder(func(r *Request) {
			counter2b++
		})

		e0.Request("GET", "/url")
		assert.Equal(t, 0, counter1)
		assert.Equal(t, 0, counter2a)
		assert.Equal(t, 0, counter2b)

		e1.Request("GET", "/url")
		assert.Equal(t, 1, counter1)
		assert.Equal(t, 0, counter2a)
		assert.Equal(t, 0, counter2b)

		e2a.Request("GET", "/url")
		assert.Equal(t, 2, counter1)
		assert.Equal(t, 1, counter2a)
		assert.Equal(t, 0, counter2b)

		e2b.Request("GET", "/url")
		assert.Equal(t, 3, counter1)
		assert.Equal(t, 1, counter2a)
		assert.Equal(t, 1, counter2b)
	})
}

func TestExpect_Matchers(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		client := &mockClient{}

		reporter := NewAssertReporter(t)

		config := Config{
			Client:   client,
			Reporter: reporter,
		}

		e := WithConfig(config)

		var resps1 []*Response

		e1 := e.Matcher(func(r *Response) {
			resps1 = append(resps1, r)
		})

		var resps2 []*Response

		e2 := e1.Matcher(func(r *Response) {
			resps2 = append(resps2, r)
		})

		e.Request("GET", "/url")

		req1 := e1.Request("GET", "/url")
		req2 := e2.Request("GET", "/url")

		assert.Equal(t, 0, len(resps1))
		assert.Equal(t, 0, len(resps2))

		resp1 := req1.Expect()
		resp2 := req2.Expect()

		assert.Equal(t, 2, len(resps1))
		assert.Equal(t, 1, len(resps2))

		assert.Same(t, resp1, resps1[0])
		assert.Same(t, resp2, resps1[1])
		assert.Same(t, resp2, resps2[0])
	})

	t.Run("copying", func(t *testing.T) {
		client := &mockClient{}

		reporter := NewAssertReporter(t)

		config := Config{
			Client:   client,
			Reporter: reporter,
		}

		counter1 := 0
		counter2a := 0
		counter2b := 0

		e0 := WithConfig(config)

		// Simulate the case when many builders are added, and the builders slice
		// have some additioonal capacity. We are going to check that the slice
		// is cloned properly when a new builder is appended.
		for i := 0; i < 10; i++ {
			e0 = e0.Matcher(func(r *Response) {})
		}

		e1 := e0.Matcher(func(r *Response) {
			counter1++
		})

		e2a := e1.Matcher(func(r *Response) {
			counter2a++
		})

		e2b := e1.Matcher(func(r *Response) {
			counter2b++
		})

		e0.Request("GET", "/url").Expect()
		assert.Equal(t, 0, counter1)
		assert.Equal(t, 0, counter2a)
		assert.Equal(t, 0, counter2b)

		e1.Request("GET", "/url").Expect()
		assert.Equal(t, 1, counter1)
		assert.Equal(t, 0, counter2a)
		assert.Equal(t, 0, counter2b)

		e2a.Request("GET", "/url").Expect()
		assert.Equal(t, 2, counter1)
		assert.Equal(t, 1, counter2a)
		assert.Equal(t, 0, counter2b)

		e2b.Request("GET", "/url").Expect()
		assert.Equal(t, 3, counter1)
		assert.Equal(t, 1, counter2a)
		assert.Equal(t, 1, counter2b)
	})
}

func TestExpect_Traverse(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)

	config := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: reporter,
	}

	data := map[string]interface{}{
		"aaa": []interface{}{"bbb", 123, false, nil},
		"bbb": "hello",
		"ccc": 456,
	}

	resp := WithConfig(config).GET("/url").WithJSON(data).Expect()

	m := resp.JSON().Object()

	m.IsEqual(data)

	m.ContainsKey("aaa")
	m.ContainsKey("bbb")
	m.ContainsKey("aaa")

	m.HasValue("aaa", data["aaa"])
	m.HasValue("bbb", data["bbb"])
	m.HasValue("ccc", data["ccc"])

	m.Keys().ConsistsOf("aaa", "bbb", "ccc")
	m.Values().ConsistsOf(data["aaa"], data["bbb"], data["ccc"])

	m.Value("aaa").Array().ConsistsOf("bbb", 123, false, nil)
	m.Value("bbb").String().IsEqual("hello")
	m.Value("ccc").Number().IsEqual(456)

	m.Value("aaa").Array().Value(2).Boolean().IsFalse()
	m.Value("aaa").Array().Value(3).IsNull()
}

func TestExpect_Branches(t *testing.T) {
	client := &mockClient{}

	config := Config{
		BaseURL:  "http://example.com",
		Client:   client,
		Reporter: newMockReporter(t),
	}

	data := map[string]interface{}{
		"foo": []interface{}{"bar", 123, false, nil},
		"bar": "hello",
		"baz": 456,
	}

	req := WithConfig(config).GET("/url").WithJSON(data)
	resp := req.Expect()

	m1 := resp.JSON().Array()  // fail
	m2 := resp.JSON().Object() // ok
	m3 := resp.JSON().Object() // ok

	e1 := m2.Value("foo").Object()                    // fail
	e2 := m2.Value("foo").Array().Value(999).String() // fail
	e3 := m2.Value("foo").Array().Value(0).Number()   // fail
	e4 := m2.Value("foo").Array().Value(0).String()   // ok
	e5 := m2.Value("foo").Array().Value(0).String()   // ok

	e4.IsEqual("qux") // fail
	e5.IsEqual("bar") // ok

	req.chain.assertFlags(t, flagFailedChildren)
	resp.chain.assertFlags(t, flagFailedChildren)

	m1.chain.assertFlags(t, flagFailed)
	m2.chain.assertFlags(t, flagFailedChildren)
	m3.chain.assertFlags(t, 0)

	e1.chain.assertFlags(t, flagFailed)
	e2.chain.assertFlags(t, flagFailed)
	e3.chain.assertFlags(t, flagFailed)
	e4.chain.assertFlags(t, flagFailed)
	e5.chain.assertFlags(t, 0)
}

func TestExpect_Propagation(t *testing.T) {
	t.Run("subsequent operations", func(t *testing.T) {
		ctr := 0
		reporter := newMockReporter(t)
		reporter.reportCb = func() {
			ctr++
		}

		// Failed operation
		value := NewArray(reporter, []interface{}{"foo"})
		value.IsEmpty()
		value.chain.assertFlags(t, flagFailed)
		assert.Equal(t, 1, ctr)

		// Subsequent failed operation won't report failures
		value.IsEmpty()
		value.chain.assertFlags(t, flagFailed)
		assert.Equal(t, 1, ctr)
	})

	t.Run("newly created child", func(t *testing.T) {
		ctr := 0
		reporter := newMockReporter(t)
		reporter.reportCb = func() {
			ctr++
		}

		// Parent's failed operation reports failure
		parent := NewArray(reporter, []interface{}{"foo"})
		parent.IsEmpty()
		parent.chain.assertFlags(t, flagFailed)
		assert.Equal(t, 1, ctr)

		// Child created after parent's failure
		// Child's failed operation won't report failures
		child := parent.Value(0)
		child.IsEqual("bar")
		parent.chain.assertFlags(t, flagFailed|flagFailedChildren)
		child.chain.assertFlags(t, flagFailed)
		assert.Equal(t, 1, ctr)
	})

	t.Run("previously created child", func(t *testing.T) {
		ctr := 0
		reporter := newMockReporter(t)
		reporter.reportCb = func() {
			ctr++
		}

		// Parent and child
		parent := NewArray(reporter, []interface{}{"foo"})
		child := parent.Value(0)

		// Parent's failed operation reports failure
		parent.IsEmpty()
		parent.chain.assertFlags(t, flagFailed)
		child.chain.assertFlags(t, 0)
		assert.Equal(t, 1, ctr)

		// Child was created before parent's failure
		// Child's failed operation will report failures
		child.IsEqual("bar")
		parent.chain.assertFlags(t, flagFailed|flagFailedChildren)
		child.chain.assertFlags(t, flagFailed)
		assert.Equal(t, 2, ctr)
	})

	t.Run("newly created child of parent", func(t *testing.T) {
		ctr := 0
		reporter := newMockReporter(t)
		reporter.reportCb = func() {
			ctr++
		}

		// Parent
		parent := NewArray(reporter, []interface{}{"foo"})

		// Child's failed operation will report failures
		child := parent.Value(0)
		child.IsEqual("bar")
		parent.chain.assertFlags(t, flagFailedChildren)
		child.chain.assertFlags(t, flagFailed)
		assert.Equal(t, 1, ctr)

		// New child created after failure in another child
		// New child's failed operation will report failures
		newChild := parent.Value(0)
		newChild.IsEqual("bar")
		parent.chain.assertFlags(t, flagFailedChildren)
		child.chain.assertFlags(t, flagFailed)
		newChild.chain.assertFlags(t, flagFailed)
		assert.Equal(t, 2, ctr)
	})

	t.Run("previously created child of parent", func(t *testing.T) {
		ctr := 0
		reporter := newMockReporter(t)
		reporter.reportCb = func() {
			ctr++
		}

		// Parent
		parent := NewArray(reporter, []interface{}{"foo"})

		// Children
		child1 := parent.Value(0)
		child2 := parent.Value(0)

		// child1's failed operation will report failures
		child1.IsEqual("bar")
		parent.chain.assertFlags(t, flagFailedChildren)
		child1.chain.assertFlags(t, flagFailed)
		child2.chain.assertFlags(t, 0)
		assert.Equal(t, 1, ctr)

		// child2's failed operation will report failures
		child2.IsEqual("bar")
		parent.chain.assertFlags(t, flagFailedChildren)
		child1.chain.assertFlags(t, flagFailed)
		child2.chain.assertFlags(t, flagFailed)
		assert.Equal(t, 2, ctr)
	})
}

func TestExpect_Inheritance(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		rootReporter := newMockReporter(t)
		req2Reporter := newMockReporter(t)

		e := WithConfig(Config{
			BaseURL:  "http://example.com",
			Client:   &mockClient{},
			Reporter: rootReporter,
		})

		req1 := e.GET("/")
		req2 := e.GET("/")

		req2.WithReporter(req2Reporter)

		// So far OK
		req1.chain.assert(t, success)
		req2.chain.assert(t, success)

		resp1 := req1.Expect()
		resp2 := req2.Expect()

		// So far OK
		resp1.chain.assert(t, success)
		resp2.chain.assert(t, success)

		// Failure on resp1 should be reported to rootReporter,
		// which was inherited from config
		resp1.JSON().Object().Value("foo").chain.assert(t, failure)
		assert.Equal(t, 1, rootReporter.reportCalled)
		assert.Equal(t, 0, req2Reporter.reportCalled)

		// Failure on resp2 should be reported to req2Reporter,
		// which was inherited from req2
		resp2.JSON().Object().Value("foo").chain.assert(t, failure)
		assert.Equal(t, 1, rootReporter.reportCalled)
		assert.Equal(t, 1, req2Reporter.reportCalled)
	})

	t.Run("assertion handler", func(t *testing.T) {
		rootHandler := &mockAssertionHandler{}
		req2Handler := &mockAssertionHandler{}

		e := WithConfig(Config{
			BaseURL:          "http://example.com",
			Client:           &mockClient{},
			AssertionHandler: rootHandler,
		})

		req1 := e.GET("/")
		req2 := e.GET("/")

		req2.WithAssertionHandler(req2Handler)

		// So far OK
		req1.chain.assert(t, success)
		req2.chain.assert(t, success)

		resp1 := req1.Expect()
		resp2 := req2.Expect()

		// So far OK
		resp1.chain.assert(t, success)
		resp2.chain.assert(t, success)

		// Failure on resp1 should be reported to rootReporter,
		// which was inherited from config
		resp1.JSON().Object().Value("foo").chain.assert(t, failure)
		assert.Equal(t, 1, rootHandler.failureCalled)
		assert.Equal(t, 0, req2Handler.failureCalled)

		// Failure on resp2 should be reported to req2Reporter,
		// which was inherited from req2
		resp2.JSON().Object().Value("foo").chain.assert(t, failure)
		assert.Equal(t, 1, rootHandler.failureCalled)
		assert.Equal(t, 1, req2Handler.failureCalled)
	})
}

func TestExpect_RequestFactory(t *testing.T) {
	t.Run("default factory", func(t *testing.T) {
		e := WithConfig(Config{
			BaseURL:  "http://example.com",
			Reporter: NewAssertReporter(t),
		})

		req := e.Request("GET", "/")
		req.chain.assert(t, success)

		assert.NotNil(t, req.httpReq)
	})

	t.Run("custom factory", func(t *testing.T) {
		factory := &mockRequestFactory{}

		e := WithConfig(Config{
			BaseURL:        "http://example.com",
			Reporter:       NewAssertReporter(t),
			RequestFactory: factory,
		})

		req := e.Request("GET", "/")
		req.chain.assert(t, success)

		assert.NotNil(t, factory.lastreq)
		assert.Same(t, req.httpReq, factory.lastreq)
	})

	t.Run("factory failure", func(t *testing.T) {
		factory := &mockRequestFactory{
			fail: true,
		}

		e := WithConfig(Config{
			BaseURL:        "http://example.com",
			Reporter:       newMockReporter(t),
			RequestFactory: factory,
		})

		req := e.Request("GET", "/")
		req.chain.assert(t, failure)

		assert.Nil(t, factory.lastreq)
	})
}

func TestExpect_Panics(t *testing.T) {
	t.Run("nil AssertionHandler, non-nil Reporter", func(t *testing.T) {
		assert.NotPanics(t, func() {
			WithConfig(Config{
				Reporter:         newMockReporter(t),
				AssertionHandler: nil,
			})
		})
	})

	t.Run("non-nil AssertionHandler, nil Reporter", func(t *testing.T) {
		assert.NotPanics(t, func() {
			WithConfig(Config{
				Reporter:         nil,
				AssertionHandler: &mockAssertionHandler{},
			})
		})
	})

	t.Run("nil AssertionHandler, nil Reporter", func(t *testing.T) {
		assert.Panics(t, func() {
			WithConfig(Config{
				Reporter:         nil,
				AssertionHandler: nil,
			})
		})
	})
}

func TestExpect_Config(t *testing.T) {
	t.Run("defaults, non-nil Reporter", func(t *testing.T) {
		config := Config{
			Reporter: newMockReporter(t),
		}

		config = config.withDefaults()

		assert.NotNil(t, config.RequestFactory)
		assert.NotNil(t, config.Client)
		assert.NotNil(t, config.WebsocketDialer)
		assert.NotNil(t, config.AssertionHandler)
		assert.NotNil(t, config.Formatter)
		assert.NotNil(t, config.Reporter)

		assert.NotPanics(t, func() {
			config.validate()
		})
	})

	t.Run("defaults, non-nil AssertionHandler", func(t *testing.T) {
		config := Config{
			AssertionHandler: &mockAssertionHandler{},
		}

		config = config.withDefaults()

		assert.NotNil(t, config.RequestFactory)
		assert.NotNil(t, config.Client)
		assert.NotNil(t, config.WebsocketDialer)
		assert.NotNil(t, config.AssertionHandler)
		assert.Nil(t, config.Formatter)
		assert.Nil(t, config.Reporter)

		assert.NotPanics(t, func() {
			config.validate()
		})
	})

	t.Run("defaults, nil Reporter and AssertionHandler", func(t *testing.T) {
		config := Config{}

		assert.Panics(t, func() {
			config.withDefaults()
		})

		assert.Panics(t, func() {
			config.validate()
		})
	})

	t.Run("validate fields", func(t *testing.T) {
		config := Config{
			Reporter: newMockReporter(t),
		}

		config = config.withDefaults()

		assert.NotPanics(t, func() {
			config.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.RequestFactory = nil
			badConfig.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.Client = nil
			badConfig.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.AssertionHandler = nil
			badConfig.validate()
		})
	})

	t.Run("validate handler", func(t *testing.T) {
		config := Config{
			Reporter: newMockReporter(t),
		}

		config = config.withDefaults()

		assert.NotPanics(t, func() {
			badConfig := config
			badConfig.AssertionHandler = &DefaultAssertionHandler{
				Formatter: &DefaultFormatter{},
				Reporter:  newMockReporter(t),
			}
			badConfig.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.AssertionHandler = &DefaultAssertionHandler{
				Formatter: &DefaultFormatter{},
				Reporter:  nil,
			}
			badConfig.validate()
		})

		assert.Panics(t, func() {
			badConfig := config
			badConfig.AssertionHandler = &DefaultAssertionHandler{
				Formatter: nil,
				Reporter:  newMockReporter(t),
			}
			badConfig.validate()
		})
	})
}

func TestExpect_Adapters(t *testing.T) {
	t.Run("RequestFactoryFunc", func(t *testing.T) {
		called := false
		factory := RequestFactoryFunc(func(
			_ string, _ string, _ io.Reader,
		) (*http.Request, error) {
			called = true
			return nil, nil
		})

		e := WithConfig(Config{
			RequestFactory: factory,
			Reporter:       newMockReporter(t),
		})

		e.Request("GET", "/")

		assert.True(t, called)
	})

	t.Run("ClientFunc", func(t *testing.T) {
		called := false
		client := ClientFunc(func(_ *http.Request) (*http.Response, error) {
			called = true
			return &http.Response{
				Status:     "Test Status",
				StatusCode: 504,
			}, nil
		})

		e := WithConfig(Config{
			Client:   client,
			Reporter: newMockReporter(t),
		})

		req := e.GET("/")
		resp := req.Expect()

		assert.True(t, called)
		assert.Equal(t, resp.httpResp.StatusCode, 504)
		assert.Equal(t, resp.httpResp.Status, "Test Status")
	})

	t.Run("WebsocketDialerFunc", func(t *testing.T) {
		called := false
		dialer := WebsocketDialerFunc(func(
			_ string, _ http.Header,
		) (*websocket.Conn, *http.Response, error) {
			called = true
			return &websocket.Conn{}, &http.Response{}, nil
		})

		e := WithConfig(Config{
			WebsocketDialer: dialer,
			Reporter:        newMockReporter(t),
		})

		e.GET("/path").WithWebsocketUpgrade().Expect().Websocket()

		assert.True(t, called)
	})

	t.Run("ReporterFunc", func(t *testing.T) {
		called := false
		message := ""
		client := ClientFunc(func(r *http.Request) (*http.Response, error) {
			return nil, errors.New("")
		})
		reporter := ReporterFunc(func(_ string, _ ...interface{}) {
			called = true
			message = "test reporter called"
		})

		e := WithConfig(Config{
			Reporter: reporter,
			Client:   client,
		})

		e.GET("/").Expect()

		assert.True(t, called)
		assert.Contains(t, message, "test reporter called")
	})

	t.Run("LoggerFunc", func(t *testing.T) {
		called := false
		message := ""
		logger := LoggerFunc(func(_ string, _ ...interface{}) {
			called = true
			message = "test logger called"
		})

		e := WithConfig(Config{
			Reporter: newMockReporter(t),
			Printers: []Printer{
				NewCompactPrinter(logger),
			},
		})

		e.GET("").Expect()

		assert.True(t, called)
		assert.Contains(t, message, "test logger called")
	})
}

type contextAssertionHandler struct {
	Formatter        Formatter
	Reporter         Reporter
	Logger           Logger
	AssertionContext *AssertionContext
}

// Success implements AssertionHandler.Success.
func (h *contextAssertionHandler) Success(ctx *AssertionContext) {
	if h.Formatter == nil {
		panic("DefaultAssertionHandler.Formatter is nil")
	}
	h.Formatter.FormatSuccess(ctx)
	h.AssertionContext = ctx
}

// Failure implements AssertionHandler.Failure.
func (h *contextAssertionHandler) Failure(
	ctx *AssertionContext, failure *AssertionFailure,
) {
	if h.Formatter == nil {
		panic("DefaultAssertionHandler.Formatter is nil")
	}

	switch failure.Severity {
	case SeverityError:
		if h.Reporter == nil {
			panic("DefaultAssertionHandler.Reporter is nil")
		}

		h.Formatter.FormatFailure(ctx, failure)
		h.AssertionContext = ctx

	case SeverityLog:
		if h.Logger == nil {
			return
		}

		h.Formatter.FormatFailure(ctx, failure)
		h.AssertionContext = ctx
	}
}

func TestExpect_AssertionContextSuccess(t *testing.T) {
	client := &mockClient{}

	reporter := NewAssertReporter(t)
	formatter := &DefaultFormatter{}

	handler := &contextAssertionHandler{Reporter: reporter, Formatter: formatter}

	config := Config{
		BaseURL:          "http://example.com",
		Client:           client,
		Reporter:         reporter,
		Formatter:        formatter,
		AssertionHandler: handler,
	}

	e := WithConfig(config)

	req := e.GET("/test").WithText("test")
	resp := req.Expect()
	body := resp.Body()
	assert.Equal(t, req, handler.AssertionContext.Request)
	assert.Equal(t, body.value, handler.AssertionContext.Response.Body().value)
}

func TestExpect_AssertionContextResponseRequest(t *testing.T) {
	prepare := func(client Client) (*Expect, Config, *contextAssertionHandler) {
		reporter := NewAssertReporter(t)
		formatter := &DefaultFormatter{}

		handler := &contextAssertionHandler{Reporter: reporter, Formatter: formatter}

		config := Config{
			BaseURL:          "http://example.com",
			Client:           client,
			Reporter:         reporter,
			Formatter:        formatter,
			AssertionHandler: handler,
		}
		return WithConfig(config), config, handler
	}

	t.Run("success", func(t *testing.T) {
		e, _, handler := prepare(&mockClient{})
		req := e.GET("/test").WithText("test")
		resp := req.Expect()
		resp.Body()

		req.chain.assertFlags(t, 0)
		resp.chain.assertFlags(t, 0)

		assert.Equal(t, &req, &handler.AssertionContext.Request)
		assert.Equal(t, &resp, &handler.AssertionContext.Response)
	})

	t.Run("fail request", func(t *testing.T) {
		e, _, handler := prepare(&mockClient{err: errors.New("test")})
		req := e.GET("/test").WithText("test")
		resp := req.Expect()
		resp.Body()

		req.chain.assertFlags(t, flagFailed|flagFailedChildren)
		resp.chain.assertFlags(t, flagFailed)

		assert.Equal(t, &req, &handler.AssertionContext.Request)
		assert.Nil(t, handler.AssertionContext.Response)
	})

	t.Run("fail response", func(t *testing.T) {
		client := ClientFunc(func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				Status:     "Test Status",
				StatusCode: 504,
			}, nil
		})
		e, _, handler := prepare(client)
		req := e.GET("/test")
		resp := req.Expect()
		resp.Body()

		req.chain.assertFlags(t, 0)
		resp.chain.assertFlags(t, 0)

		assert.Equal(t, &req, &handler.AssertionContext.Request)
		assert.Equal(t, &resp, &handler.AssertionContext.Response)
	})

	t.Run("fail response children", func(t *testing.T) {
		e, _, handler := prepare(&mockClient{})
		req := e.GET("/test").WithText("{{}")
		resp := req.Expect()
		resp.JSON().Array()

		req.chain.assertFlags(t, flagFailedChildren)
		resp.chain.assertFlags(t, flagFailed|flagFailedChildren)

		assert.Equal(t, &req, &handler.AssertionContext.Request)
		assert.Equal(t, &resp, &handler.AssertionContext.Response)
	})
}
