# FLARE: Privacy-Preserving Sanctions Screening

FLARE is a cutting-edge privacy-preserving sanctions screening solution that leverages **Private Set Intersection (PSI)** and **Homomorphic Encryption (HE)** to allow financial institutions to screen customer lists against sanctions lists without revealing sensitive customer data to the screening provider, and without the provider revealing the full sanctions list to the institution.

![FLARE Dashboard](https://placehold.co/1200x600/1e293b/ffffff?text=FLARE+Dashboard+Preview)

## 🌟 Key Features

- **Privacy-First Architecture**: Uses Lattice-based cryptography (Lattigo) and LE-PSI to ensure data privacy for both parties.
- **Secure Screening**: Perform intersection checks between encrypted customer datasets and sanctions lists.
- **Modern Dashboard**: A comprehensive Next.js-based UI for managing screenings, customers, and results.
- **Role-Based Access Control (RBAC)**: Granular permissions for Admins and standard Users.
- **Real-time Progress**: WebSocket/SSE integration for tracking long-running screening jobs.
- **Audit Logging**: Complete trail of all screening activities for compliance.

## 🛠️ Tech Stack

### Backend
- **Language**: Go 1.24
- **Framework**: Chi Router
- **Cryptography**: [Lattigo](https://github.com/tuneinsight/lattigo) (Homomorphic Encryption), [LE-PSI](https://github.com/SanthoshCheemala/LE-PSI)
- **Database**: PostgreSQL
- **Authentication**: JWT & Bcrypt

### Frontend
- **Framework**: Next.js 16 (App Router)
- **Styling**: Tailwind CSS v4
- **Components**: shadcn/ui, Radix UI
- **Icons**: Lucide React
- **Language**: TypeScript

## 🚀 Getting Started

### Prerequisites
- Go 1.24+
- Node.js 20+
- Docker & Docker Compose (for PostgreSQL)

### Installation

1.  **Clone the repository**
    ```bash
    git clone https://github.com/SanthoshCheemala/FLARE.git
    cd FLARE
    ```

2.  **Setup Backend**
    ```bash
    cd backend
    cp .env.example .env
    # Update .env with your database credentials
    go mod download
    ```

3.  **Setup Frontend**
    ```bash
    cd ../flare-ui
    npm install
    ```

4.  **Start Database**
    ```bash
    # From project root or backend directory
    docker-compose up -d postgres
    ```

### Running the Application

1.  **Start Backend Server**
    ```bash
    cd backend
    go run cmd/api/main.go
    ```

2.  **Start Frontend Client**
    ```bash
    cd flare-ui
    npm run dev
    ```

3.  **Access the UI**
    Open [http://localhost:3000](http://localhost:3000) in your browser.

## 📂 Project Structure

```
FLARE/
├── backend/                # Go Backend Service
│   ├── cmd/                # Entry points
│   ├── internal/           # Core business logic (PSI, Auth, Jobs)
│   ├── migrations/         # SQL database migrations
│   └── ...
├── flare-ui/               # Next.js Frontend Application
│   ├── src/app/            # App Router pages
│   ├── src/components/     # UI Components
│   └── ...
└── ...
```

## 🔒 Security

FLARE is designed with security as a priority.
- **Zero-Knowledge Screening**: The server never sees the raw customer data.
- **Encrypted Communication**: All data in transit is encrypted.
- **Secure Storage**: Sensitive data is hashed or encrypted at rest.


## 📄 License

[MIT License](LICENSE)
