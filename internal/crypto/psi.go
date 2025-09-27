package crypto

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/SanthoshCheemala/Crypto/hash"
	"github.com/SanthoshCheemala/FLARE/internal/storage"
	"github.com/SanthoshCheemala/FLARE/pkg/LE"
	"github.com/SanthoshCheemala/FLARE/pkg/matrix"
	"github.com/tuneinsight/lattigo/v3/ring"
)

// cxtx packages the ciphertexts returned by the server-side routine
type cxtx struct {
	c0 []*matrix.Vector
	c1 []*matrix.Vector
	c  *matrix.Vector
	d  *ring.Poly
}

// Laconic_PSI runs client side PSI
func Laconic_PSI(Client_Transaction []storage.Transaction, Server_Transaction []storage.Transaction, Treepath string) (intersection_trans []storage.Transaction, err error) {
	startTime := time.Now()
	cSize := len(Client_Transaction)
	if cSize == 0 {
		return nil, errors.New("client transaction set is empty")
	}

	leParams, err := SetupLEParameters(len(Server_Transaction))
	if err != nil {
		return nil, fmt.Errorf("SetupLEParameters: %w", err)
	}

	db, err := sql.Open("sqlite3", Treepath)
	if err != nil {
		return nil, fmt.Errorf("open tree db: %w", err)
	}
	defer db.Close()

	if err := storage.InitializeTreeDB(db, leParams.Layers); err != nil {
		log.Printf("warning: InitializeTreeDB returned: %v\n", err)
	}

	publicKeys := make([]*matrix.Vector, cSize)
	privateKeys := make([]*matrix.Vector, cSize)
	hashedClient := make([]uint64, cSize)
	mergedClient := make([]string, cSize)

	// --- Encryption timing ---
	encStart := time.Now()
	for i := 0; i < cSize; i++ {
		publicKeys[i], privateKeys[i] = leParams.KeyGen()
		merge := ""
		sortedKeys := getSortedKeys(Client_Transaction[i].Data)
		for _, col := range sortedKeys {
			merge += Client_Transaction[i].Data[col]
		}
		mergedClient[i] = merge

		H := hash.NewSHA256State()
		H.Sha256([]byte(mergedClient[i]))
		raw := binary.BigEndian.Uint64(H.Sum())

		var mask uint64
		bits := uint(leParams.Layers)
		if bits == 0 || bits >= 64 {
			mask = ^uint64(0)
		} else {
			mask = (uint64(1) << bits) - 1
		}
		hashedClient[i] = raw & mask

		LE.Upd(db, hashedClient[i], leParams.Layers, publicKeys[i], leParams)
	}
	encEnd := time.Now()
	encDuration := encEnd.Sub(encStart)

	pp := LE.ReadFromDB(db, 0, 0, leParams).NTT(leParams.R)
	msg := matrix.NewRandomPolyBinary(leParams.R)

	// --- Server encryption timing ---
	serverEncStart := time.Now()
	ciphertexts := Laconic_PSI_server(pp, msg, Server_Transaction, leParams)
	serverEncEnd := time.Now()
	serverEncDuration := serverEncEnd.Sub(serverEncStart)

	// --- Decryption timing ---
	decStart := time.Now()
	witnessesVec1 := make([][]*matrix.Vector, cSize)
	witnessesVec2 := make([][]*matrix.Vector, cSize)
	for i := 0; i < cSize; i++ {
		vec1, vec2 := LE.WitGen(db, leParams, hashedClient[i])
		witnessesVec1[i] = vec1
		witnessesVec2[i] = vec2
	}
	Z := make([]storage.Transaction, 0)
	intersectionMap := make(map[int]bool)

	var totalMaxNoise, totalAvgNoise float64
	var totalMatches, totalErrors int
	var noiseStats []map[string]interface{}
	var errorStats []map[string]interface{}

	for j := range ciphertexts {
		for k := 0; k < cSize; k++ {
			msg2 := LE.Dec(leParams, privateKeys[k], witnessesVec1[k], witnessesVec2[k],
				ciphertexts[j].c0, ciphertexts[j].c1, ciphertexts[j].c, ciphertexts[j].d)

			maxNoise, avgNoise, noiseDistribution := MeasureNoiseLevel(leParams.R, msg, msg2, leParams.Q)
			totalMaxNoise += maxNoise
			totalAvgNoise += avgNoise
			totalMatches++

			matchCount, mismatchCount := 0, 0
			for i := 0; i < leParams.R.N; i++ {
				q14 := leParams.Q / 4
				q34 := (leParams.Q / 4) * 3
				bin := 0
				if msg2.Coeffs[0][i] < q14 || msg2.Coeffs[0][i] > q34 {
					bin = 0
				} else {
					bin = 1
				}
				if bin == int(msg.Coeffs[0][i]) {
					matchCount++
				} else {
					mismatchCount++
				}
			}
			matchPercentage := float64(matchCount) / float64(leParams.R.N)
			errorStats = append(errorStats, map[string]interface{}{
				"ServerIdx": j,
				"ClientIdx": k,
				"Matches":   matchCount,
				"Mismatches": mismatchCount,
				"MatchPct":  matchPercentage,
			})
			if mismatchCount > 0 {
				totalErrors += mismatchCount
			}

			noiseStats = append(noiseStats, map[string]interface{}{
				"ServerIdx": j,
				"ClientIdx": k,
				"MaxNoise":  maxNoise,
				"AvgNoise":  avgNoise,
				"NoiseDist": noiseDistribution,
			})

			if matchPercentage >= 0.95 {
				if _, ok := intersectionMap[k]; !ok {
					Z = append(Z, Client_Transaction[k])
					intersectionMap[k] = true
				}
			}
		}
	}
	decEnd := time.Now()
	decDuration := decEnd.Sub(decStart)
	duration := time.Since(startTime)

	// --- LE parameter analysis ---
	numSlots := 1 << leParams.Layers
	loadFactor := float64(len(Server_Transaction)) / float64(numSlots)
	collisionProb := 1.0 - math.Exp(-math.Pow(float64(len(Server_Transaction)), 2)/(2*float64(numSlots)))
	leAnalysis := map[string]interface{}{
		"Q": leParams.Q,
		"qBits": leParams.QBits,
		"D": leParams.D,
		"N": leParams.N,
		"Layers": leParams.Layers,
		"NumSlots": numSlots,
		"LoadFactor": loadFactor,
		"CollisionProb": collisionProb,
	}

	WriteAdvancedPSIReportHTML(
		"data/psi_report.html",
		noiseStats,
		errorStats,
		totalMatches,
		totalMaxNoise,
		totalAvgNoise,
		totalErrors,
		duration,
		encDuration,
		serverEncDuration,
		decDuration,
		leAnalysis,
	)

	return Z, nil
}

// Laconic_PSI_server constructs encryptions for the server's set using the provided pp and msg.
// Returns a slice of ciphertext bundles (c0, c1, c, d).
func Laconic_PSI_server(pp *matrix.Vector, msg *ring.Poly, server_Transaction []storage.Transaction, le *LE.LE) []cxtx {
	sSize := len(server_Transaction)
	mergedServer := make([]string, sSize)
	for idx, rec := range server_Transaction {
		merge := ""
		sortedKeys := getSortedKeys(rec.Data)
		for _, col := range sortedKeys {
			merge += rec.Data[col]
		}
		mergedServer[idx] = merge
	}

	// Hash server records to indices
	hashed := make([]uint64, sSize)
	for i := 0; i < sSize; i++ {
		H := hash.NewSHA256State()
		H.Sha256([]byte(mergedServer[i]))
		raw := binary.BigEndian.Uint64(H.Sum())

		var mask uint64
		bits := uint(le.Layers)
		if bits == 0 || bits >= 64 {
			mask = ^uint64(0)
		} else {
			mask = (uint64(1) << bits) - 1
		}
		hashed[i] = raw & mask
	}
    fmt.Printf("Server hashes: %v\n", hashed)

	// Build ciphertexts for each hashed server index
	C := make([]cxtx, sSize)
	for i := 0; i < sSize; i++ {
		// r: randomness vectors for each layer
		r := make([]*matrix.Vector, le.Layers+1)
		for j := 0; j < le.Layers+1; j++ {
			r[j] = matrix.NewRandomVec(le.N, le.R, le.PRNG).NTT(le.R)
		}

		e := le.SamplerGaussian.ReadNew()
		e0 := make([]*matrix.Vector, le.Layers+1)
		e1 := make([]*matrix.Vector, le.Layers+1)
		for j := 0; j < le.Layers+1; j++ {
			if j == le.Layers {
				e0[j] = matrix.NewNoiseVec(le.M2, le.R, le.PRNG, le.Sigma, le.Bound).NTT(le.R)
			} else {
				e0[j] = matrix.NewNoiseVec(le.M, le.R, le.PRNG, le.Sigma, le.Bound).NTT(le.R)
			}
			e1[j] = matrix.NewNoiseVec(le.M, le.R, le.PRNG, le.Sigma, le.Bound).NTT(le.R)
		}

		// Enc wraps the encryption; check your LE.Enc signature matches
		c0, c1, cvec, dpoly := LE.Enc(le, pp, hashed[i], msg, r, e0, e1, e)
		C[i].c0 = c0
		C[i].c1 = c1
		C[i].c = cvec
		C[i].d = dpoly
	}

	return C
}

// WriteAdvancedPSIReportHTML generates a beautiful HTML report with advanced stats and LE parameter analysis.
func WriteAdvancedPSIReportHTML(
	filepath string,
	noiseStats []map[string]interface{},
	errorStats []map[string]interface{},
	totalMatches int,
	totalMaxNoise, totalAvgNoise float64,
	totalErrors int,
	duration, encDuration, serverEncDuration, decDuration time.Duration,
	leAnalysis map[string]interface{},
) {
	type NoiseRow struct {
		ServerIdx int
		ClientIdx int
		MaxNoise  float64
		AvgNoise  float64
		NoiseDist map[string]int
	}
	type ErrorRow struct {
		ServerIdx int
		ClientIdx int
		Matches   int
		Mismatches int
		MatchPct  float64
	}
	var noiseRows []NoiseRow
	for _, stat := range noiseStats {
		noiseRows = append(noiseRows, NoiseRow{
			ServerIdx: stat["ServerIdx"].(int),
			ClientIdx: stat["ClientIdx"].(int),
			MaxNoise:  stat["MaxNoise"].(float64),
			AvgNoise:  stat["AvgNoise"].(float64),
			NoiseDist: stat["NoiseDist"].(map[string]int),
		})
	}
	var errorRows []ErrorRow
	for _, stat := range errorStats {
		errorRows = append(errorRows, ErrorRow{
			ServerIdx: stat["ServerIdx"].(int),
			ClientIdx: stat["ClientIdx"].(int),
			Matches:   stat["Matches"].(int),
			Mismatches: stat["Mismatches"].(int),
			MatchPct:  stat["MatchPct"].(float64),
		})
	}
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<title>Advanced PSI Noise & Error Statistics Report</title>
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<style>
		body {
			font-family: 'Segoe UI', Arial, sans-serif;
			background: #f8f9fa;
			margin: 0;
			padding: 0;
		}
		.header {
			background: linear-gradient(90deg, #4f8cff 0%, #6dd5ed 100%);
			color: #fff;
			padding: 2em 2em 1em 2em;
			box-shadow: 0 2px 8px rgba(0,0,0,0.05);
		}
		.container {
			max-width: 1200px;
			margin: 2em auto;
			background: #fff;
			border-radius: 12px;
			box-shadow: 0 2px 16px rgba(0,0,0,0.07);
			padding: 2em;
		}
		h1, h2 {
			font-weight: 700;
			margin-bottom: 0.5em;
		}
		.stats {
			display: flex;
			flex-wrap: wrap;
			gap: 2em;
			margin-bottom: 2em;
		}
		.stat-card {
			background: #e3f2fd;
			border-radius: 8px;
			padding: 1em 2em;
			box-shadow: 0 1px 4px rgba(0,0,0,0.04);
			min-width: 220px;
			flex: 1;
		}
		.stat-title {
			font-size: 1.1em;
			color: #1565c0;
			margin-bottom: 0.3em;
		}
		.stat-value {
			font-size: 1.5em;
			font-weight: 600;
			color: #0d47a1;
		}
		table {
			border-collapse: collapse;
			width: 100%;
			margin-bottom: 2em;
			font-size: 0.98em;
		}
		th, td {
			border: 1px solid #bdbdbd;
			padding: 8px 6px;
			text-align: center;
		}
		th {
			background: #e3f2fd;
			color: #1565c0;
		}
		tr:nth-child(even) {
			background: #f5f5f5;
		}
		.noise-bar {
			display: inline-block;
			height: 12px;
			border-radius: 4px;
			background: linear-gradient(90deg, #4caf50 0%, #ffeb3b 50%, #f44336 100%);
		}
		.le-param-table {
			margin-bottom: 2em;
			width: 60%;
			border: 1px solid #bdbdbd;
			border-radius: 8px;
			overflow: hidden;
		}
		.le-param-table th {
			background: #1565c0;
			color: #fff;
		}
		.le-param-table td {
			background: #e3f2fd;
			color: #0d47a1;
		}
		@media (max-width: 900px) {
			.container { padding: 1em; }
			.le-param-table { width: 100%; }
		}
	</style>
</head>
<body>
	<div class="header">
		<h1>Advanced PSI Noise & Error Statistics Report</h1>
		<p>Beautifully generated after each run. Analyze performance, correctness, and LE parameter impact.</p>
	</div>
	<div class="container">
		<h2>Summary Statistics</h2>
		<div class="stats">
			<div class="stat-card">
				<div class="stat-title">Total Execution Time</div>
				<div class="stat-value">{{.Duration}}</div>
			</div>
			<div class="stat-card">
				<div class="stat-title">Encryption Time (Client)</div>
				<div class="stat-value">{{.EncDuration}}</div>
			</div>
			<div class="stat-card">
				<div class="stat-title">Encryption Time (Server)</div>
				<div class="stat-value">{{.ServerEncDuration}}</div>
			</div>
			<div class="stat-card">
				<div class="stat-title">Decryption Time</div>
				<div class="stat-value">{{.DecDuration}}</div>
			</div>
			<div class="stat-card">
				<div class="stat-title">Total Matches</div>
				<div class="stat-value">{{.TotalMatches}}</div>
			</div>
			<div class="stat-card">
				<div class="stat-title">Total Errors</div>
				<div class="stat-value">{{.TotalErrors}}</div>
			</div>
			<div class="stat-card">
				<div class="stat-title">Avg Max Noise</div>
				<div class="stat-value">{{printf "%.6f" .AvgMaxNoise}}</div>
			</div>
			<div class="stat-card">
				<div class="stat-title">Avg Mean Noise</div>
				<div class="stat-value">{{printf "%.6f" .AvgMeanNoise}}</div>
			</div>
		</div>
		<h2>LE Parameter Analysis</h2>
		<table class="le-param-table">
			<tr><th>Parameter</th><th>Value</th></tr>
			<tr><td>Q (Modulus)</td><td>{{.LEAnalysis.Q}}</td></tr>
			<tr><td>qBits</td><td>{{.LEAnalysis.qBits}}</td></tr>
			<tr><td>D (Ring Dimension)</td><td>{{.LEAnalysis.D}}</td></tr>
			<tr><td>N (Matrix Dimension)</td><td>{{.LEAnalysis.N}}</td></tr>
			<tr><td>Layers</td><td>{{.LEAnalysis.Layers}}</td></tr>
			<tr><td>Num Slots</td><td>{{.LEAnalysis.NumSlots}}</td></tr>
			<tr><td>Load Factor</td><td>{{printf "%.6f" .LEAnalysis.LoadFactor}}</td></tr>
			<tr><td>Collision Probability</td><td>{{printf "%.6e" .LEAnalysis.CollisionProb}}</td></tr>
		</table>
		<h2>Error Statistics (Decryption Correctness)</h2>
		<table>
			<tr>
				<th>Server Index</th>
				<th>Client Index</th>
				<th>Matches</th>
				<th>Mismatches</th>
				<th>Match %</th>
			</tr>
			{{range .ErrorRows}}
			<tr>
				<td>{{.ServerIdx}}</td>
				<td>{{.ClientIdx}}</td>
				<td>{{.Matches}}</td>
				<td>{{.Mismatches}}</td>
				<td>{{printf "%.2f" (mul100 .MatchPct)}}%</td>
			</tr>
			{{end}}
		</table>
		<h2>Noise Statistics</h2>
		<table>
			<tr>
				<th>Server Index</th>
				<th>Client Index</th>
				<th>Max Noise</th>
				<th>Avg Noise</th>
				<th>Noise Distribution</th>
			</tr>
			{{range .NoiseRows}}
			<tr>
				<td>{{.ServerIdx}}</td>
				<td>{{.ClientIdx}}</td>
				<td>{{printf "%.6f" .MaxNoise}}</td>
				<td>{{printf "%.6f" .AvgNoise}}</td>
				<td>
					{{range $cat, $count := .NoiseDist}}
						{{$cat}}: {{$count}}<br>
						<span class="noise-bar" style="width:{{mul $count 1.5}}px"></span>
					{{end}}
				</td>
			</tr>
			{{end}}
		</table>
		<p style="color:#888;">Report generated at: {{.Timestamp}}</p>
	</div>
</body>
</html>
`
	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"mul100": func(a float64) float64 { return a * 100 },
	}
	data := struct {
		NoiseRows        []NoiseRow
		ErrorRows        []ErrorRow
		TotalMatches     int
		TotalErrors      int
		AvgMaxNoise      float64
		AvgMeanNoise     float64
		Duration         string
		EncDuration      string
		ServerEncDuration string
		DecDuration      string
		LEAnalysis       map[string]interface{}
		Timestamp        string
	}{
		NoiseRows: noiseRows,
		ErrorRows: errorRows,
		TotalMatches: totalMatches,
		TotalErrors: totalErrors,
		AvgMaxNoise: totalMaxNoise / float64(totalMatches),
		AvgMeanNoise: totalAvgNoise / float64(totalMatches),
		Duration: duration.String(),
		EncDuration: encDuration.String(),
		ServerEncDuration: serverEncDuration.String(),
		DecDuration: decDuration.String(),
		LEAnalysis: leAnalysis,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	f, err := os.Create(filepath)
	if err != nil {
		fmt.Printf("Error creating HTML report: %v\n", err)
		return
	}
	defer f.Close()
	t := template.Must(template.New("report").Funcs(funcMap).Parse(tmpl))
	if err := t.Execute(f, data); err != nil {
		fmt.Printf("Error writing HTML report: %v\n", err)
	}
}