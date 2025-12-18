module github.com/SanthoshCheemala/FLARE/backend

go 1.24.1

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/lib/pq v1.10.9
	github.com/mattn/go-sqlite3 v1.14.32
	golang.org/x/crypto v0.45.0
)

require (
	github.com/SanthoshCheemala/LE-PSI v0.0.0-00010101000000-000000000000
	github.com/go-chi/chi/v5 v5.2.3
	github.com/gorilla/websocket v1.5.3
	github.com/tuneinsight/lattigo/v3 v3.0.6
)

require golang.org/x/sys v0.38.0 // indirect

replace github.com/SanthoshCheemala/LE-PSI => ../../PSI
