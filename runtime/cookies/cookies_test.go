package cookies

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookies_GetWithNilRequest(t *testing.T) {
	// Test defensive programming - should not panic with nil request
	cookies := Cookies{r: nil, w: nil}
	
	cookie, err := cookies.Get("test")
	
	if cookie != nil {
		t.Errorf("Expected nil cookie, got %v", cookie)
	}
	if err != http.ErrNoCookie {
		t.Errorf("Expected http.ErrNoCookie, got %v", err)
	}
}

func TestCookies_AllWithNilRequest(t *testing.T) {
	// Test defensive programming - should not panic with nil request
	cookies := Cookies{r: nil, w: nil}
	
	allCookies := cookies.All()
	
	if allCookies != nil {
		t.Errorf("Expected nil cookies slice, got %v", allCookies)
	}
}

func TestCookies_SetWithNilResponseWriter(t *testing.T) {
	// Test defensive programming - should not panic with nil response writer
	cookies := Cookies{r: nil, w: nil}
	testCookie := &http.Cookie{Name: "test", Value: "value"}
	
	// This should not panic
	cookies.Set(testCookie)
	// No assertion needed - just ensuring no panic
}

func TestNew_WithNilInputs(t *testing.T) {
	// Test constructor validation - should handle nil inputs gracefully
	cookies := New(nil, nil)
	
	// Verify the struct was initialized (zero-value Cookies has nil fields)
	_ = cookies
	
	// Should handle operations gracefully
	cookie, err := cookies.Get("test")
	if cookie != nil || err != http.ErrNoCookie {
		t.Errorf("Expected (nil, http.ErrNoCookie), got (%v, %v)", cookie, err)
	}
}

func TestCookies_Normal_Operations(t *testing.T) {
	// Test normal operations work correctly
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc123"})
	
	w := httptest.NewRecorder()
	cookies := New(req, w)
	
	// Test Get
	cookie, err := cookies.Get("session")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if cookie == nil || cookie.Value != "abc123" {
		t.Errorf("Expected session=abc123, got %v", cookie)
	}
	
	// Test missing cookie
	missing, err := cookies.Get("nonexistent")
	if missing != nil {
		t.Errorf("Expected nil for missing cookie, got %v", missing)
	}
	if err != http.ErrNoCookie {
		t.Errorf("Expected http.ErrNoCookie, got %v", err)
	}
	
	// Test Set
	newCookie := &http.Cookie{Name: "new", Value: "test"}
	cookies.Set(newCookie)
	
	// Verify cookie was set in response
	response := w.Result()
	setCookies := response.Header["Set-Cookie"]
	found := false
	for _, c := range setCookies {
		if c == "new=test" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected cookie 'new=test' to be set, got %v", setCookies)
	}
}