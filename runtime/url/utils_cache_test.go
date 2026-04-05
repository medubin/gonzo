package url

// Internal tests for the field metadata cache (getFieldMeta).
// Uses package url (not url_test) to access unexported symbols.

import (
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type cachedURL struct {
	ID   string `url:"id"`
	Name string `url:"name"`
}

type cachedJSON struct {
	Page  *int `json:"page,omitempty"`
	Limit *int `json:"limit,omitempty"`
}

func TestGetFieldMeta_ReturnsCorrectMetadata(t *testing.T) {
	meta := getFieldMeta(reflect.TypeOf(cachedURL{}), "url")
	require.Len(t, meta, 2)
	assert.Equal(t, "id", meta[0].tag)
	assert.Equal(t, 0, meta[0].index)
	assert.Equal(t, "name", meta[1].tag)
	assert.Equal(t, 1, meta[1].index)
}

func TestGetFieldMeta_OmitemptySuffixStripped(t *testing.T) {
	meta := getFieldMeta(reflect.TypeOf(cachedJSON{}), "json")
	require.Len(t, meta, 2)
	assert.Equal(t, "page", meta[0].tag)
	assert.Equal(t, "limit", meta[1].tag)
}

func TestGetFieldMeta_CachesResult(t *testing.T) {
	// Two calls for the same type must return identical slices (same pointer).
	typ := reflect.TypeOf(cachedURL{})
	first := getFieldMeta(typ, "url")
	second := getFieldMeta(typ, "url")
	// reflect.SliceHeader comparison via pointer to first element
	if len(first) > 0 && len(second) > 0 {
		assert.Same(t, &first[0], &second[0], "expected cache hit to return same backing array")
	}
}

func TestGetFieldMeta_DifferentTagNames_IndependentCaches(t *testing.T) {
	type dual struct {
		A string `url:"a" json:"alpha"`
	}
	urlMeta := getFieldMeta(reflect.TypeOf(dual{}), "url")
	jsonMeta := getFieldMeta(reflect.TypeOf(dual{}), "json")
	require.Len(t, urlMeta, 1)
	require.Len(t, jsonMeta, 1)
	assert.Equal(t, "a", urlMeta[0].tag)
	assert.Equal(t, "alpha", jsonMeta[0].tag)
}

func TestGetFieldMeta_NoTaggedFields_ReturnsEmpty(t *testing.T) {
	type noTags struct {
		X string
		Y int
	}
	meta := getFieldMeta(reflect.TypeOf(noTags{}), "url")
	assert.Empty(t, meta)
}

func TestGetFieldMeta_Concurrent(t *testing.T) {
	// Many goroutines requesting metadata for the same type concurrently.
	// The -race detector will catch any data race.
	type target struct {
		ID string `url:"id"`
	}
	typ := reflect.TypeOf(target{})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			meta := getFieldMeta(typ, "url")
			if len(meta) != 1 || meta[0].tag != "id" {
				t.Errorf("unexpected meta: %v", meta)
			}
		}()
	}
	wg.Wait()
}
