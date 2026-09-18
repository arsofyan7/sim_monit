package collector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"strings"
	"time"

	"sim_monit/server/models"
)

type HTTPMetricDetails struct {
	DNSLookupMs       float64 `json:"dns_lookup_ms"`
	TCPConnectMs      float64 `json:"tcp_connect_ms"`
	TLSHandshakeMs    float64 `json:"tls_handshake_ms"`
	TTFBMs            float64 `json:"ttfb_ms"`
	ContentDownloadMs float64 `json:"content_download_ms"`
	TotalTimeMs       float64 `json:"total_time_ms"`
	StatusCode        int     `json:"status_code"`
	StatusText        string  `json:"status_text"`
}

func CollectHTTP(target *models.Target) (*models.MetricRaw, models.TargetStatus, error) {
	url := target.Host
	if target.Config != nil && target.Config.URL != "" {
		url = target.Config.URL
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if target.Port == 443 {
			url = "https://" + url
		} else if target.Port == 80 {
			url = "http://" + url
		} else if target.Port > 0 {
			url = fmt.Sprintf("http://%s:%d", url, target.Port)
		} else {
			url = "https://" + url
		}
	}

	method := "GET"
	expectedStatus := 200
	var headers map[string]string

	if target.Config != nil {
		if target.Config.Method != "" {
			method = strings.ToUpper(target.Config.Method)
		}
		if target.Config.ExpectedStatusCode > 0 {
			expectedStatus = target.Config.ExpectedStatusCode
		}
		headers = target.Config.Headers
	}

	var dnsStart, dnsDone time.Time
	var connStart, connDone time.Time
	var tlsStart, tlsDone time.Time
	var ttfbDone time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { dnsDone = time.Now() },
		ConnectStart: func(_, _ string) { connStart = time.Now() },
		ConnectDone: func(_, _ string, _ error) { connDone = time.Now() },
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone:  func(_ tls.ConnectionState, _ error) { tlsDone = time.Now() },
		GotFirstResponseByte: func() { ttfbDone = time.Now() },
	}

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, models.TargetStatusOffline, fmt.Errorf("invalid http request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("User-Agent", "SimMonit/1.0 (Agentless Monitoring Bot)")

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true}, // Allow self-signed test certs
			DisableKeepAlives: true,
		},
	}

	start := time.Now()
	resp, err := client.Do(req)
	totalDuration := time.Since(start)

	if err != nil {
		details, _ := json.Marshal(map[string]any{
			"error": err.Error(),
			"url":   url,
		})
		metric := &models.MetricRaw{
			TargetID:   target.ID,
			Timestamp:  time.Now(),
			LatencyMs:  float64(totalDuration.Milliseconds()),
			RawDetails: string(details),
		}
		return metric, models.TargetStatusOffline, err
	}
	defer resp.Body.Close()

	// Read body to measure content download time (limit to 5MB to prevent DoS from unbounded streams)
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 5*1024*1024))
	endTotal := time.Now()

	// Calculate timing metrics
	dnsLookupMs := 0.0
	if !dnsStart.IsZero() && !dnsDone.IsZero() {
		dnsLookupMs = float64(dnsDone.Sub(dnsStart).Microseconds()) / 1000.0
	}

	tcpConnectMs := 0.0
	if !connStart.IsZero() && !connDone.IsZero() {
		tcpConnectMs = float64(connDone.Sub(connStart).Microseconds()) / 1000.0
	}

	tlsHandshakeMs := 0.0
	if !tlsStart.IsZero() && !tlsDone.IsZero() {
		tlsHandshakeMs = float64(tlsDone.Sub(tlsStart).Microseconds()) / 1000.0
	}

	ttfbMs := 0.0
	if !ttfbDone.IsZero() {
		ttfbMs = float64(ttfbDone.Sub(start).Microseconds()) / 1000.0
	} else {
		ttfbMs = float64(totalDuration.Microseconds()) / 1000.0
	}

	totalMs := float64(endTotal.Sub(start).Microseconds()) / 1000.0
	contentDownloadMs := totalMs - ttfbMs
	if contentDownloadMs < 0 {
		contentDownloadMs = 0
	}

	detailsObj := HTTPMetricDetails{
		DNSLookupMs:       dnsLookupMs,
		TCPConnectMs:      tcpConnectMs,
		TLSHandshakeMs:    tlsHandshakeMs,
		TTFBMs:            ttfbMs,
		ContentDownloadMs: contentDownloadMs,
		TotalTimeMs:       totalMs,
		StatusCode:        resp.StatusCode,
		StatusText:        resp.Status,
	}

	detailsJSON, _ := json.Marshal(detailsObj)

	status := models.TargetStatusOnline
	if expectedStatus > 0 && resp.StatusCode != expectedStatus {
		status = models.TargetStatusOffline
	} else if resp.StatusCode >= 500 {
		status = models.TargetStatusOffline
	}

	metric := &models.MetricRaw{
		TargetID:   target.ID,
		Timestamp:  time.Now(),
		LatencyMs:  totalMs,
		RawDetails: string(detailsJSON),
	}

	return metric, status, nil
}
