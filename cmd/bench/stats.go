package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"
)

type EndpointStats struct {
	Label           string  `json:"label"`
	Count           int     `json:"count"`
	OK              int     `json:"ok"`
	Errors          int     `json:"errors"`
	Timeouts        int     `json:"timeouts"`
	Min             float64 `json:"min_ms"`
	Mean            float64 `json:"mean_ms"`
	P50             float64 `json:"p50_ms"`
	P95             float64 `json:"p95_ms"`
	P99             float64 `json:"p99_ms"`
	Max             float64 `json:"max_ms"`
	RPS             float64 `json:"rps"`
	TotalDurationMs float64 `json:"total_duration_ms"`
	ServerMin       float64 `json:"server_min_ms"`
	ServerMean      float64 `json:"server_mean_ms"`
	ServerP50       float64 `json:"server_p50_ms"`
	ServerP95       float64 `json:"server_p95_ms"`
	ServerP99       float64 `json:"server_p99_ms"`
	ServerMax       float64 `json:"server_max_ms"`
}

func computeStats(results []callResult, totalDuration time.Duration) map[string]*EndpointStats {
	byEndpoint := make(map[string][]callResult)
	for _, r := range results {
		byEndpoint[r.Endpoint] = append(byEndpoint[r.Endpoint], r)
	}

	stats := make(map[string]*EndpointStats, len(byEndpoint))
	labels := make([]string, 0, len(byEndpoint))
	for label := range byEndpoint {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	for _, label := range labels {
		rs := byEndpoint[label]
		st := &EndpointStats{Label: label, Count: len(rs)}
		durs := make([]time.Duration, 0, len(rs))
		serverDurs := make([]time.Duration, 0, len(rs))

		for _, r := range rs {
			if r.OK {
				st.OK++
				durs = append(durs, r.Duration)
				if r.ServerDur > 0 {
					serverDurs = append(serverDurs, r.ServerDur)
				}
			} else {
				st.Errors++
				if r.ErrorMsg == "timeout" {
					st.Timeouts++
				}
			}
		}

		if len(durs) > 0 {
			sort.Slice(durs, func(i, j int) bool { return durs[i] < durs[j] })
			st.Min = ms(durs[0])
			st.Max = ms(durs[len(durs)-1])
			var sum time.Duration
			for _, d := range durs {
				sum += d
			}
			st.Mean = ms(sum / time.Duration(len(durs)))
			st.P50 = msFloat(percentileDur(durs, 50))
			st.P95 = msFloat(percentileDur(durs, 95))
			st.P99 = msFloat(percentileDur(durs, 99))
		}

		if len(serverDurs) > 0 {
			sort.Slice(serverDurs, func(i, j int) bool { return serverDurs[i] < serverDurs[j] })
			st.ServerMin = ms(serverDurs[0])
			st.ServerMax = ms(serverDurs[len(serverDurs)-1])
			var sum time.Duration
			for _, d := range serverDurs {
				sum += d
			}
			st.ServerMean = ms(sum / time.Duration(len(serverDurs)))
			st.ServerP50 = msFloat(percentileDur(serverDurs, 50))
			st.ServerP95 = msFloat(percentileDur(serverDurs, 95))
			st.ServerP99 = msFloat(percentileDur(serverDurs, 99))
		}

		if totalDuration > 0 {
			st.RPS = float64(st.OK) / totalDuration.Seconds()
		}
		st.TotalDurationMs = ms(totalDuration)

		stats[label] = st
	}
	return stats
}

func percentileDur(durations []time.Duration, p float64) time.Duration {
	n := len(durations)
	if n == 0 {
		return 0
	}
	rank := p / 100 * float64(n-1)
	lower := int(rank)
	upper := lower + 1
	if upper >= n {
		return durations[lower]
	}
	f := rank - float64(lower)
	return durations[lower] + time.Duration(f*float64(durations[upper]-durations[lower]))
}

func ms(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

func msFloat(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

func printTable(stats map[string]*EndpointStats) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "Endpoint\tCount\tOK\tErrors\tTimeouts\tMin(ms)\tP50(ms)\tP95(ms)\tP99(ms)\tMax(ms)\tRPS\tSvP50\tSvP99")
	fmt.Fprintln(w, "--------\t-----\t--\t------\t--------\t-------\t-------\t-------\t-------\t-------\t---\t-----\t-----")
	labels := make([]string, 0, len(stats))
	for label := range stats {
		labels = append(labels, label)
	}
	sort.Strings(labels)
	for _, label := range labels {
		s := stats[label]
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\t%.1f\t%.1f\t%.1f\t%.1f\t%.1f\t%.0f\t%.1f\t%.1f\n",
			label, s.Count, s.OK, s.Errors, s.Timeouts, s.Min, s.P50, s.P95, s.P99, s.Max, s.RPS, s.ServerP50, s.ServerP99)
	}
	w.Flush()
}

func writeResultsJSON(path string, meta interface{}, stats map[string]*EndpointStats) error {
	type output struct {
		Meta  interface{}               `json:"meta"`
		Stats map[string]*EndpointStats `json:"stats"`
	}
	data, err := json.MarshalIndent(output{Meta: meta, Stats: stats}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
