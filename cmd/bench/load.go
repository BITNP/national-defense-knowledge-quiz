package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type startResponse struct {
	Title    string            `json:"title"`
	Intro    string            `json:"intro"`
	LogID    uint              `json:"log_id"`
	Problems []responseProblem `json:"problems"`
}

type responseProblem struct {
	Type string   `json:"type"`
	Text string   `json:"text"`
	Data []string `json:"data"`
}

type callResult struct {
	Endpoint   string
	Duration   time.Duration
	OK         bool
	StatusCode int
	ErrorMsg   string
	ServerDur  time.Duration
}

func runLoad(cfg Config) ([]callResult, time.Duration, error) {
	tr := &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  true,
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   cfg.Timeout,
	}

	var uidCounter int64
	uidGen := func() string {
		n := atomic.AddInt64(&uidCounter, 1)
		return fmt.Sprintf("bench-%d-%d-%d", cfg.Concurrency, n, time.Now().UnixNano())
	}

	tasks := make(chan string, cfg.Requests)
	for i := 0; i < cfg.Requests; i++ {
		tasks <- uidGen()
	}
	close(tasks)

	var mu sync.Mutex
	var results []callResult
	var wg sync.WaitGroup

	start := time.Now()
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for uid := range tasks {
				switch cfg.Scenario {
				case "info":
					r := doInfo(client, cfg.URL, cfg.ExamID)
					mu.Lock()
					results = append(results, r)
					mu.Unlock()
				case "start":
					r := doStart(client, cfg.URL, cfg.ExamID, uid)
					mu.Lock()
					results = append(results, r)
					mu.Unlock()
				case "workflow":
					rs := doAll(client, cfg.URL, cfg.ExamID, uid)
					mu.Lock()
					results = append(results, rs...)
					mu.Unlock()
				case "start_submit":
					rs := doStartSubmit(client, cfg.URL, cfg.ExamID, uid)
					mu.Lock()
					results = append(results, rs...)
					mu.Unlock()
				default:
					mu.Lock()
					results = append(results, callResult{Endpoint: cfg.Scenario, OK: false, ErrorMsg: "unknown scenario"})
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)
	return results, elapsed, nil
}

func doInfo(client *http.Client, baseURL string, examID uint) callResult {
	url := fmt.Sprintf("%s/exam/info/?exam=%d", baseURL, examID)
	t0 := time.Now()
	resp, err := client.Get(url)
	dur := time.Since(t0)
	if err != nil {
		return callResult{Endpoint: "GET /exam/info/", Duration: dur, OK: false, ErrorMsg: err.Error()}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	return callResult{
		Endpoint:   "GET /exam/info/",
		Duration:   dur,
		OK:         ok,
		StatusCode: resp.StatusCode,
		ServerDur:  parseServerDuration(resp.Header),
	}
}

func doStart(client *http.Client, baseURL string, examID uint, uid string) callResult {
	body := map[string]interface{}{
		"exam":       examID,
		"student_id": uid,
		"name":       uid,
	}
	jsonBody, _ := json.Marshal(body)
	url := baseURL + "/exam/start/"

	t0 := time.Now()
	resp, err := client.Post(url, "application/json", bytes.NewReader(jsonBody))
	dur := time.Since(t0)
	if err != nil {
		return callResult{Endpoint: "POST /exam/start/", Duration: dur, OK: false, ErrorMsg: err.Error()}
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	return callResult{
		Endpoint:   "POST /exam/start/",
		Duration:   dur,
		OK:         ok,
		StatusCode: resp.StatusCode,
		ServerDur:  parseServerDuration(resp.Header),
	}
}

func doStartSubmit(client *http.Client, baseURL string, examID uint, uid string) []callResult {
	body := map[string]interface{}{
		"exam":       examID,
		"student_id": uid,
		"name":       uid,
	}
	jsonBody, _ := json.Marshal(body)
	startURL := baseURL + "/exam/start/"

	t0 := time.Now()
	resp, err := client.Post(startURL, "application/json", bytes.NewReader(jsonBody))
	startDur := time.Since(t0)

	var results []callResult

	if err != nil {
		results = append(results, callResult{Endpoint: "POST /exam/start/", Duration: startDur, OK: false, ErrorMsg: err.Error()})
		results = append(results, callResult{Endpoint: "POST /exam/submit/", Duration: 0, OK: false, ErrorMsg: "start failed"})
		return results
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	startOK := resp.StatusCode >= 200 && resp.StatusCode < 300
	results = append(results, callResult{
		Endpoint:   "POST /exam/start/",
		Duration:   startDur,
		OK:         startOK,
		StatusCode: resp.StatusCode,
		ServerDur:  parseServerDuration(resp.Header),
	})
	if !startOK {
		results = append(results, callResult{Endpoint: "POST /exam/submit/", Duration: 0, OK: false, ErrorMsg: "start non-2xx", StatusCode: resp.StatusCode})
		return results
	}

	var startResp startResponse
	if err := json.Unmarshal(respBody, &startResp); err != nil || startResp.LogID == 0 || len(startResp.Problems) == 0 {
		results = append(results, callResult{Endpoint: "POST /exam/submit/", Duration: 0, OK: false, ErrorMsg: "invalid start response"})
		return results
	}

	answers := make([]string, len(startResp.Problems))
	for i, p := range startResp.Problems {
		if len(p.Data) > 0 {
			answers[i] = p.Data[0]
		}
	}
	answersJSON, _ := json.Marshal(answers)

	submitBody := map[string]interface{}{
		"log_id":     startResp.LogID,
		"student_id": uid,
		"name":       uid,
		"answers":    string(answersJSON),
	}
	submitJSON, _ := json.Marshal(submitBody)
	submitURL := baseURL + "/exam/submit/"

	t0 = time.Now()
	subResp, err := client.Post(submitURL, "application/json", bytes.NewReader(submitJSON))
	submitDur := time.Since(t0)
	if err != nil {
		results = append(results, callResult{Endpoint: "POST /exam/submit/", Duration: submitDur, OK: false, ErrorMsg: err.Error()})
		return results
	}
	defer subResp.Body.Close()
	_, _ = io.Copy(io.Discard, subResp.Body)

	results = append(results, callResult{
		Endpoint:   "POST /exam/submit/",
		Duration:   submitDur,
		OK:         subResp.StatusCode >= 200 && subResp.StatusCode < 300,
		StatusCode: subResp.StatusCode,
		ServerDur:  parseServerDuration(subResp.Header),
	})
	return results
}

func doAll(client *http.Client, baseURL string, examID uint, uid string) []callResult {
	infoRes := doInfo(client, baseURL, examID)
	if !infoRes.OK {
		return []callResult{
			infoRes,
			{Endpoint: "POST /exam/start/", Duration: 0, OK: false, ErrorMsg: "info failed"},
			{Endpoint: "POST /exam/submit/", Duration: 0, OK: false, ErrorMsg: "info failed"},
		}
	}
	startRes := doStart(client, baseURL, examID, uid)
	if !startRes.OK {
		return []callResult{infoRes, startRes,
			{Endpoint: "POST /exam/submit/", Duration: 0, OK: false, ErrorMsg: "start failed"},
		}
	}
	submitRes := doStartSubmit(client, baseURL, examID, uid)
	return append([]callResult{infoRes, startRes}, submitRes...)
}

func parseServerDuration(h http.Header) time.Duration {
	s := h.Get("X-Request-Duration")
	if s == "" {
		return 0
	}
	s = strings.TrimSuffix(s, "ms")
	ms, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return time.Duration(ms) * time.Millisecond
}
