package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSnapshotHandler_MethodNotAllowed(t *testing.T) {
	h := NewSnapshotHandler()

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
	h := NewSnapshotHandler()

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
	h := NewSnapshotHandler()

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
	h := NewSnapshotHandler()

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
	assert.GreaterOrEqual(t, resp.CurrentStep, resp.Step)
	assert.Greater(t, resp.StepsCount, 0)
	require.NotNil(t, resp.Snapshot)
	assert.GreaterOrEqual(t, resp.Snapshot.GetFramesCount(), 1)
}

func TestNewSnapshotHandler_StepOutOfRange(t *testing.T) {
	h := NewSnapshotHandler()

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
