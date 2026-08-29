// Copyright 2022 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resty

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/weijinwen/jupiter/pkg/core/metric"
)

func counterValue(t *testing.T, c prometheus.Counter) float64 {
	t.Helper()
	var m dto.Metric
	require.NoError(t, c.Write(&m))
	return m.GetCounter().GetValue()
}

func TestOnAfterResponse_MetricLabelUsesConfigName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultConfig()
	config.Addr = server.URL
	config.Name = "resty.metric-name-case"

	client := config.MustBuild()

	counter := metric.ClientHandleCounter.WithLabelValues(metric.TypeHTTP, config.Name, http.MethodGet, client.HostURL, "200 OK")
	before := counterValue(t, counter)

	_, err := client.R().Get("/ping")
	require.NoError(t, err)

	after := counterValue(t, counter)
	assert.Equal(t, before+1, after, "counter labelled with config.Name should increase by 1")
}

func TestOnAfterResponse_AccessLog_LogsWhenEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	core, logs := observer.New(zapcore.DebugLevel)

	config := DefaultConfig()
	config.Addr = server.URL
	config.EnableAccessLog = true
	config.logger = zap.New(core)

	client := config.MustBuild()
	_, err := client.R().Get("/ping")
	require.NoError(t, err)

	entries := logs.FilterMessage("access").All()
	require.Len(t, entries, 1)
	ctx := entries[0].ContextMap()
	assert.Equal(t, "GET", ctx["meth"])
	assert.Equal(t, int32(http.StatusOK), ctx["code"])
}

func TestOnAfterResponse_AccessLog_NoLogWhenDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	core, logs := observer.New(zapcore.DebugLevel)

	config := DefaultConfig()
	config.Addr = server.URL
	config.logger = zap.New(core)

	client := config.MustBuild()
	_, err := client.R().Get("/ping")
	require.NoError(t, err)

	entries := logs.FilterMessage("access").All()
	assert.Empty(t, entries)
}
