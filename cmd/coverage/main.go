package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"

	"golang.org/x/tools/cover"
)

type FileCoverage struct {
	File         string  `json:"file"`
	TotalStmts   int     `json:"total_stmts"`
	CoveredStmts int     `json:"covered_stmts"`
	CoveragePct  float64 `json:"coverage_pct"`
}

func main() {
	var profilePath string
	var format string
	flag.StringVar(&profilePath, "coverprofile", "coverage.out", "path to coverage profile")
	flag.StringVar(&format, "format", "json", "output format: json|csv")
	flag.Parse()

	profiles, err := cover.ParseProfiles(profilePath)
	if err != nil {
		log.Fatalf("failed to parse cover profile: %v", err)
	}

	total := map[string]int{}
	covered := map[string]int{}

	for _, p := range profiles {
		for _, b := range p.Blocks {
			total[p.FileName] += b.NumStmt
			if b.Count > 0 {
				covered[p.FileName] += b.NumStmt
			}
		}
	}

	var rows []FileCoverage
	for f, t := range total {
		c := covered[f]
		pct := 0.0
		if t > 0 {
			pct = float64(c) / float64(t) * 100.0
		}
		rows = append(rows, FileCoverage{
			File:         f,
			TotalStmts:   t,
			CoveredStmts: c,
			CoveragePct:  pct,
		})
	}

	// stabile Ausgabe sortiert nach Dateiname
	sort.Slice(rows, func(i, j int) bool { return rows[i].File < rows[j].File })

	switch format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rows); err != nil {
			log.Fatalf("failed to write JSON: %v", err)
		}
	case "csv":
		w := csv.NewWriter(os.Stdout)
		defer w.Flush()
		_ = w.Write([]string{"file", "total_stmts", "covered_stmts", "coverage_pct"})
		for _, r := range rows {
			_ = w.Write([]string{
				r.File,
				fmt.Sprintf("%d", r.TotalStmts),
				fmt.Sprintf("%d", r.CoveredStmts),
				fmt.Sprintf("%.2f", r.CoveragePct),
			})
		}
	default:
		log.Fatalf("unsupported format: %q (use json|csv)", format)
	}
}
