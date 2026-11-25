package gin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/abmcmanu/sessionx/pkg/session"
	"github.com/gin-gonic/gin"
)

func newRouter(t *testing.T) (*gin.Engine, *session.Manager) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	m, err := session.NewManager(session.DevConfig([]byte("0123456789abcdef0123456789abcdef")))
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.Use(SessionMiddleware(m))
	return r, m
}

func cookieFrom(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == "sessionx" {
			return c
		}
	}
	t.Fatal("session cookie not set")
	return nil
}

func do(r *gin.Engine, path string, c *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if c != nil {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestCounterPersists(t *testing.T) {
	r, _ := newRouter(t)
	r.GET("/", func(c *gin.Context) {
		sess := Get(c)
		n, _ := sess.Data["n"].(float64)
		sess.Data["n"] = n + 1
		c.JSON(http.StatusOK, gin.H{"n": sess.Data["n"]})
	})

	var c *http.Cookie
	var rec *httptest.ResponseRecorder
	for i := 0; i < 3; i++ {
		rec = do(r, "/", c)
		if got := len(rec.Result().Header.Values("Set-Cookie")); got != 1 {
			t.Fatalf("expected 1 Set-Cookie, got %d", got)
		}
		c = cookieFrom(t, rec)
	}
	if rec.Body.String() != `{"n":3}` {
		t.Errorf("unexpected body %s", rec.Body.String())
	}
}

func TestResponseVariantsSaveSession(t *testing.T) {
	r, _ := newRouter(t)
	r.GET("/string", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	r.GET("/status", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	r.GET("/abort", func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) })
	r.GET("/redirect", func(c *gin.Context) { c.Redirect(http.StatusFound, "/") })
	r.GET("/empty", func(c *gin.Context) {})

	for _, path := range []string{"/string", "/status", "/abort", "/redirect", "/empty"} {
		rec := do(r, path, nil)
		if got := len(rec.Result().Header.Values("Set-Cookie")); got != 1 {
			t.Errorf("%s: expected 1 Set-Cookie, got %d", path, got)
		}
	}
}

func TestSessionInRequestContext(t *testing.T) {
	r, _ := newRouter(t)
	r.GET("/", func(c *gin.Context) {
		if session.Get(c.Request) != Get(c) {
			t.Error("request context and gin context hold different sessions")
		}
	})
	do(r, "/", nil)
}

func TestDestroy(t *testing.T) {
	r, m := newRouter(t)
	r.GET("/login", func(c *gin.Context) { Get(c).Data["user"] = "alice" })
	r.GET("/logout", func(c *gin.Context) {
		_ = m.Destroy(c.Writer, c.Request)
		c.String(http.StatusOK, "bye")
	})

	c := cookieFrom(t, do(r, "/login", nil))
	rec := do(r, "/logout", c)
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Errorf("expected a single expiry cookie, got %+v", cookies)
	}
}

func TestGetWithoutMiddleware(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	if Get(c) != nil {
		t.Error("expected nil session")
	}
	c.Set(SessionKey, "wrong type")
	if Get(c) != nil {
		t.Error("expected nil for wrong type")
	}
}
