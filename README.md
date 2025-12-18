# FLARE: Privacy-Preserving Sanctions Screening

FLARE is a privacy-preserving sanctions screening solution that enables financial institutions to screen customer lists against sanctions databases **without revealing sensitive customer data** to the screening authority, and without the authority exposing the full sanctions list.

## 🔐 How It Works

FLARE uses **Private Set Intersection (PSI)** with Lattice-based cryptography to compute matches between encrypted datasets:

1. **Bank** uploads a customer list and encrypts it locally
2. **Sanctions Authority** holds the encrypted sanctions database
3. **PSI Protocol** finds matches without either party seeing the other's raw data
4. **Results** show only the matching records

> **No raw customer data ever leaves the bank. No full sanctions list is ever shared.**

## 🌟 Key Features

- **Privacy-First**: Zero-knowledge screening using LE-PSI (Lattice-Based Efficient PSI)
- **Dynamic Batch PSI**: Automatic RAM-based batching for large datasets (1000+ records)
- **Dynamic Schema**: Select which columns to match on (name, DOB, country)
- **Real-time Progress**: Live updates during screening via SSE
- **Match Resolution**: View detailed match information with risk levels
- **Dual-Mode UI**: Separate interfaces for Bank (client) and Authority (server)

## 🛠️ Tech Stack

| Layer | Technology |
|-------|------------|
| **Backend** | Go 1.24, Chi Router, SQLite |
| **PSI Library** | [LE-PSI](https://github.com/SanthoshCheemala/LE-PSI), Lattigo |
| **Frontend** | Next.js 15, TypeScript, Tailwind CSS |
| **UI Components** | shadcn/ui, Radix UI, Lucide Icons |

## 🚀 Quick Start

### Prerequisites
- Go 1.24+
- Node.js 20+

### Installation

```bash
# Clone the repository
git clone https://github.com/SanthoshCheemala/FLARE.git
cd FLARE

# Setup Backend
cd backend
go mod download

# Seed the database
cd cmd/seed && go run main.go
cd ../seed_server && go run main.go

# Setup Frontend
cd ../../flare-ui
npm install
```

### Running

**Option 1: Using scripts (recommended)**
```bash
./start_flare.sh    # Starts all services
./stop_flare.sh     # Stops all services
```

**Option 2: Manual**
```bash
# Terminal 1: Server Backend (Sanctions Authority) - Port 8081
cd backend && go run cmd/server/main.go

# Terminal 2: Client Backend (Bank) - Port 8080
cd backend && go run cmd/client/main.go

# Terminal 3: Frontend - Port 3000
cd flare-ui && npm run dev
```

### Access
- **Bank UI**: http://localhost:3000 (Client mode)
- **Authority UI**: http://localhost:3000 (Server mode - set `NEXT_PUBLIC_APP_MODE=server`)

## 📂 Project Structure

```
FLARE/
├── backend/
│   ├── cmd/
│   │   ├── client/      # Bank backend (port 8080)
│   │   ├── server/      # Authority backend (port 8081)
│   │   ├── seed/        # Client database seeder
│   │   └── seed_server/ # Server database seeder
│   ├── internal/
│   │   ├── psiadapter/  # PSI library wrapper (batching, hashing)
│   │   ├── handlers/    # HTTP handlers
│   │   ├── repository/  # Database operations
│   │   └── auth/        # JWT authentication
│   └── data/            # SQLite databases & CSV files
├── flare-ui/
│   ├── src/app/         # Next.js App Router pages
│   └── src/components/  # React components
└── README.md
```

## 🔒 Security Model

| Aspect | Protection |
|--------|------------|
| Customer Data | Never leaves bank; only encrypted hashes sent |
| Sanctions List | Stored encrypted; only matching hashes revealed |
| Communication | TLS encryption in production |
| Authentication | JWT tokens, bcrypt password hashing |

## 📊 Scalability

FLARE includes **Dynamic Batch PSI** for handling large datasets:

- Automatically detects available RAM
- Splits large datasets into optimal batches
- Processes batches sequentially to prevent OOM
- Aggregates results transparently

| Dataset Size | Batching |
|--------------|----------|
| ≤ 500 records | Standard PSI |
| > 500 records | Batch PSI (dynamic) |

## 📄 License

[MIT License](LICENSE)

## 👤 Author

**Santhosh Cheemala**
- GitHub: [@SanthoshCheemala](https://github.com/SanthoshCheemala)
