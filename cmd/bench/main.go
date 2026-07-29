package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type Config struct {
	URL             string
	Scenario        string
	Concurrency     int
	Requests        int
	Timeout         time.Duration
	ExamID          uint
	DBType          string
	DBURL           string
	SkipDBReset     bool
	Profile         bool
	ProfileDuration time.Duration
	PprofURL        string
	OutputDir       string
	Verbose         bool
}

func main() {
	cfg := parseFlags()
	log.Printf("Benchmark config: scenario=%s concurrency=%d requests=%d", cfg.Scenario, cfg.Concurrency, cfg.Requests)

	// DB reset
	if !cfg.SkipDBReset {
		log.Printf("Resetting session tables (type=%s url=%s)...", cfg.DBType, cfg.DBURL)
		if err := resetDB(cfg.DBType, cfg.DBURL); err != nil {
			log.Fatalf("DB reset failed: %v", err)
		}
		log.Println("DB reset done.")
	}

	// Output directory
	if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// CPU profile goroutine
	var profileErr error
	var profileBody []byte
	profileDone := make(chan struct{})
	if cfg.Profile {
		go func() {
			defer close(profileDone)
			pURL := cfg.PprofURL
			if pURL == "" {
				pURL = cfg.URL + "/debug/pprof/profile"
			}
			profileURL := fmt.Sprintf("%s?seconds=%.0f", pURL, cfg.ProfileDuration.Seconds())
			log.Printf("Collecting CPU profile from %s ...", profileURL)
			client := &http.Client{Timeout: cfg.ProfileDuration + 10*time.Second}
			resp, err := client.Get(profileURL)
			if err != nil {
				profileErr = fmt.Errorf("profile fetch failed: %w", err)
				return
			}
			defer resp.Body.Close()
			profileBody, err = io.ReadAll(resp.Body)
			if err != nil {
				profileErr = fmt.Errorf("profile read failed: %w", err)
			}
		}()
	}

	// Run load test
	log.Println("Starting load test...")
	results, totalDuration, err := runLoad(cfg)
	if err != nil {
		log.Fatalf("Load test failed: %v", err)
	}
	log.Printf("Load test completed in %.2fs", totalDuration.Seconds())

	// Wait for profile and print top functions
	if cfg.Profile {
		log.Println("Waiting for CPU profile to finish...")
		<-profileDone
		if profileErr != nil {
			log.Printf("WARNING: CPU profile error: %v", profileErr)
		} else if len(profileBody) > 0 {
			profilePath := filepath.Join(cfg.OutputDir, "cpu.pb.gz")
			if err := os.WriteFile(profilePath, profileBody, 0644); err != nil {
				log.Printf("WARNING: failed to save profile: %v", err)
			} else {
				log.Printf("CPU profile saved to %s", profilePath)
				if goPath, err := exec.LookPath("go"); err == nil {
					log.Println("Top functions (cumulative):")
					cmd := exec.Command(goPath, "tool", "pprof", "-top", "-cum", "-n", "25", profilePath)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					_ = cmd.Run()
					log.Println("Top functions (flat):")
					cmd = exec.Command(goPath, "tool", "pprof", "-top", "-n", "25", profilePath)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					_ = cmd.Run()
				} else {
					log.Println("go tool pprof not found; profile saved for later inspection")
				}
			}
		}
	}

	// Compute and display stats
	stats := computeStats(results, totalDuration)
	fmt.Println()
	printTable(stats)
	fmt.Println()

	// Write results JSON
	meta := map[string]interface{}{
		"url":               cfg.URL,
		"scenario":          cfg.Scenario,
		"concurrency":       cfg.Concurrency,
		"requests":          cfg.Requests,
		"timeout":           cfg.Timeout.String(),
		"exam_id":           cfg.ExamID,
		"db_type":           cfg.DBType,
		"db_url":            cfg.DBURL,
		"skip_db_reset":     cfg.SkipDBReset,
		"profile":           cfg.Profile,
		"profile_duration":  cfg.ProfileDuration.String(),
		"total_duration_ms": ms(totalDuration),
	}
	jsonPath := filepath.Join(cfg.OutputDir, "results.json")
	if err := writeResultsJSON(jsonPath, meta, stats); err != nil {
		log.Printf("WARNING: failed to write results JSON: %v", err)
	} else {
		log.Printf("Results saved to %s", jsonPath)
	}
}

func parseFlags() Config {
	var cfg Config
	flag.StringVar(&cfg.URL, "url", "http://localhost:9002", "Backend base URL")
	flag.StringVar(&cfg.Scenario, "scenario", "workflow", "Scenario: workflow (all APIs), start_submit, info, start")
	flag.IntVar(&cfg.Concurrency, "concurrency", 100, "Number of concurrent workers")
	flag.IntVar(&cfg.Requests, "requests", 1000, "Total iterations")
	flag.DurationVar(&cfg.Timeout, "timeout", 30*time.Second, "Per-request timeout")
	flag.UintVar(&cfg.ExamID, "exam", 1, "Exam ID")
	flag.StringVar(&cfg.DBType, "db-type", "sqlite", "Database type: sqlite or postgres")
	flag.StringVar(&cfg.DBURL, "db-url", "./dev.db", "Database DSN")
	flag.BoolVar(&cfg.SkipDBReset, "skip-db-reset", false, "Skip DB reset")
	flag.BoolVar(&cfg.Profile, "profile", false, "Collect CPU profile")
	flag.DurationVar(&cfg.ProfileDuration, "profile-duration", 10*time.Second, "CPU profile duration")
	flag.StringVar(&cfg.PprofURL, "pprof-url", "", "pprof profile URL (default: <url>/debug/pprof/profile)")
	flag.StringVar(&cfg.OutputDir, "output", "bench_results", "Output directory")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "Verbose logging")
	flag.Parse()
	switch cfg.Scenario {
	case "workflow", "start_submit", "info", "start":
	default:
		log.Fatalf("Unknown scenario %q; valid values: workflow, start_submit, info, start", cfg.Scenario)
	}
	return cfg
}
