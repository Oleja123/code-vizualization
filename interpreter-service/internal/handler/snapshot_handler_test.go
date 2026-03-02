package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/infrastructure/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCacher is a mock implementation of cache.Cacher for testing
type MockCacher struct {
	getCalled    bool
	setCalled    bool
	getKey       string
	setKey       string
	setValue     cache.CachedInfo
	returnValue  cache.CachedInfo
	returnGetErr error
	returnSetErr error
}

func (m *MockCacher) Get(ctx context.Context, key string) (cache.CachedInfo, error) {
	m.getCalled = true
	m.getKey = key
	return m.returnValue, m.returnGetErr
}

func (m *MockCacher) Set(ctx context.Context, key string, value cache.CachedInfo) error {
	m.setCalled = true
	m.setKey = key
	m.setValue = value
	return m.returnSetErr
}

func (m *MockCacher) Reset() {
	m.getCalled = false
	m.setCalled = false
	m.getKey = ""
	m.setKey = ""
	m.setValue = cache.CachedInfo{}
	m.returnValue = cache.CachedInfo{}
	m.returnGetErr = nil
	m.returnSetErr = nil
}

func TestNewSnapshotHandler_MethodNotAllowed(t *testing.T) {
	h := NewSnapshotHandler("", nil)

	req := httptest.NewRequest(http.MethodGet, "/snapshot", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "method not allowed")
}

func TestNewSnapshotHandler_InvalidBody(t *testing.T) {
	h := NewSnapshotHandler("", nil)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "invalid request body")
}

func TestNewSnapshotHandler_InvalidStep(t *testing.T) {
	h := NewSnapshotHandler("", nil)

	body := SnapshotRequest{Code: "int main(){ return 0; }", Step: -1}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "step must be non-negative")
}

func TestNewSnapshotHandler_Success(t *testing.T) {
	h := NewSnapshotHandler("", nil)

	body := SnapshotRequest{
		Code: `int main() {
	int x = 1;
	return x;
}`,
		Step: 1,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.True(t, resp.Success)
	assert.Equal(t, 1, resp.Step)
	assert.Equal(t, resp.Step, resp.CurrentStep)
	assert.Greater(t, resp.StepsCount, 0)
	require.NotNil(t, resp.Snapshot)
	assert.GreaterOrEqual(t, resp.Snapshot.GetFramesCount(), 1)
}

func TestNewSnapshotHandler_StepOutOfRange(t *testing.T) {
	h := NewSnapshotHandler("", nil)

	body := SnapshotRequest{Code: "int main(){ return 0; }", Step: 100}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)

	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "invalid step index")
}

func TestNewSnapshotHandler_WithCacher_CacheHit(t *testing.T) {
	mockCacher := &MockCacher{}

	// Prepare cached data
	result := 42
	mockCacher.returnValue = cache.CachedInfo{
		Value: []eventdispatcher.Step{
			{StepNumber: 0, Events: nil},
			{StepNumber: 1, Events: nil},
			{StepNumber: 2, Events: nil},
		},
		StepBegin: 0,
		Result:    &result,
		Err:       nil,
	}
	mockCacher.returnGetErr = nil

	h := NewSnapshotHandler("", mockCacher)

	body := SnapshotRequest{
		Code: `int main() { return 42; }`,
		Step: 1,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// Verify that Get was called
	assert.True(t, mockCacher.getCalled, "Cache Get should be called")
	assert.Contains(t, mockCacher.getKey, "code:")

	// Verify that Set was NOT called (cache hit)
	assert.False(t, mockCacher.setCalled, "Cache Set should not be called on cache hit")

	// Verify successful response
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.True(t, resp.Success)
}

func TestNewSnapshotHandler_WithCacher_CacheMiss(t *testing.T) {
	mockCacher := &MockCacher{}

	// Simulate cache miss (empty value)
	mockCacher.returnValue = cache.CachedInfo{}
	mockCacher.returnGetErr = nil

	h := NewSnapshotHandler("", mockCacher)

	body := SnapshotRequest{
		Code: `int main() {
	int x = 1;
	return x;
}`,
		Step: 1,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// Verify that Get was called
	assert.True(t, mockCacher.getCalled, "Cache Get should be called")

	// Verify that Set was called (cache miss, so we store result)
	assert.True(t, mockCacher.setCalled, "Cache Set should be called on cache miss")
	assert.Contains(t, mockCacher.setKey, "code:")
	assert.NotNil(t, mockCacher.setValue.Value, "Cached value should not be nil")
	assert.GreaterOrEqual(t, len(mockCacher.setValue.Value), 0, "Cached steps should be set")

	// Verify successful response
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.True(t, resp.Success)
}

func TestNewSnapshotHandler_WithCacher_GetError(t *testing.T) {
	mockCacher := &MockCacher{}

	// Simulate cache error
	mockCacher.returnValue = cache.CachedInfo{}
	mockCacher.returnGetErr = errors.New("redis connection error")

	h := NewSnapshotHandler("", mockCacher)

	body := SnapshotRequest{
		Code: `int main() {
	int x = 1;
	return x;
}`,
		Step: 1,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// Verify that Get was called
	assert.True(t, mockCacher.getCalled, "Cache Get should be called")

	// Even with cache error, program should still execute
	// and Set should be called
	assert.True(t, mockCacher.setCalled, "Cache Set should be called even after Get error")

	// Verify successful response (cache error doesn't break the handler)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.True(t, resp.Success)
}

func TestNewSnapshotHandler_WithCacher_SetError(t *testing.T) {
	mockCacher := &MockCacher{}

	// Simulate cache miss and Set error
	mockCacher.returnValue = cache.CachedInfo{}
	mockCacher.returnGetErr = nil
	mockCacher.returnSetErr = errors.New("redis write error")

	h := NewSnapshotHandler("", mockCacher)

	body := SnapshotRequest{
		Code: `int main() {
	int x = 1;
	return x;
}`,
		Step: 1,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// Verify that Set was called
	assert.True(t, mockCacher.setCalled, "Cache Set should be called")

	// Verify successful response (cache Set error is ignored)
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.True(t, resp.Success)
}

func TestNewSnapshotHandler_WithCacher_CacheKeyFormat(t *testing.T) {
	mockCacher := &MockCacher{}
	mockCacher.returnValue = cache.CachedInfo{}

	h := NewSnapshotHandler("", mockCacher)

	code := `int main() { return 1; }`
	body := SnapshotRequest{
		Code: code,
		Step: 0,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// Verify cache key format includes code and config parameters
	assert.True(t, mockCacher.getCalled)
	assert.Contains(t, mockCacher.getKey, "code:")
	assert.Contains(t, mockCacher.getKey, code)
	assert.Contains(t, mockCacher.getKey, "max_elements:")
	assert.Contains(t, mockCacher.getKey, "max_steps:")
}

func TestNewSnapshotHandler_NoCacher(t *testing.T) {
	// Test without cacher (nil)
	h := NewSnapshotHandler("", nil)

	body := SnapshotRequest{
		Code: `int main() {
	int x = 42;
	return x;
}`,
		Step: 1,
	}
	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/snapshot", bytes.NewReader(payload))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	// Should work without cache
	assert.Equal(t, http.StatusOK, rr.Code)
	var resp SnapshotResponse
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	assert.True(t, resp.Success)
}
