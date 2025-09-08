# FLARE: Lattice-based Private Set Intersection System

FLARE (Fast Lattice-based Authenticated Range Encryption) is a post-quantum secure Private Set Intersection (PSI) system built using lattice-based cryptography and homomorphic encryption techniques. It enables two parties to identify common elements in their datasets without revealing any other sensitive information.

## 🔐 Features

- **Post-Quantum Security**: Built on lattice-based cryptography, resistant to quantum computer attacks
- **Private Set Intersection**: Securely compute intersections between datasets without data leakage
- **Homomorphic Encryption**: Perform computations on encrypted data using the Lattigo library
- **Merkle Tree Accumulators**: Efficient cryptographic data structures for scalable operations
- **Parallel Processing**: Optimized performance using Go's concurrency primitives
- **Database Integration**: SQLite-based storage for cryptographic accumulators and transaction data
- **Flexible CLI**: Command-line interface supporting various encryption modes and parameters

## 🏗️ Architecture

The system consists of several key components:

- **Cryptographic Core** (`pkg/LE/`): Lattice-based encryption scheme implementation
- **Matrix Operations** (`pkg/matrix/`): Efficient polynomial and vector arithmetic
- **Storage Layer** (`internal/storage/`): Database operations and transaction management
- **Crypto Interface** (`internal/crypto/`): High-level cryptographic operations and PSI protocol
- **CLI Tool** (`cmd/Flare/`): User-friendly command-line interface

## 📋 Prerequisites

- Go 1.24.1 or higher
- SQLite3
- Git

## 🚀 Installation

1. Clone the repository:
```bash
git clone https://github.com/SanthoshCheemala/FLARE.git
cd FLARE
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o flare cmd/Flare/main.go
```

## 💻 Usage

### Basic Commands

```bash
# Encrypt transaction data
./flare -encrypt -columns="type,amount" -LIMIT=100

# Decrypt transaction data
./flare -decrypt -columns="type,amount" -LIMIT=100

# Process with merged columns
./flare -encrypt -columns="type,amount" -columns-merge="user_id,timestamp" -LIMIT=50
```

### Command Line Options

- `-encrypt`: Enable encryption mode
- `-decrypt`: Enable decryption mode
- `-columns`: Comma-separated list of columns to process (required)
- `-columns-merge`: Comma-separated list of columns to merge for encryption
- `-LIMIT`: Number of rows to process from the beginning (default: 100)

## 📁 Data Structure

The system expects a SQLite database with financial transaction data. Create your database at `data/transactions.db` with a table named `finanical_transactions` containing the columns you want to process.

Example table structure:
```sql
CREATE TABLE finanical_transactions (
    id INTEGER PRIMARY KEY,
    type TEXT,
    amount REAL,
    user_id TEXT,
    timestamp DATETIME
);
```

## 🔧 Technical Details

### Cryptographic Parameters

The system uses carefully selected parameters for security and efficiency:
- **Ring Dimension (D)**: 256
- **Modulus (Q)**: 180143985094819841
- **Matrix Dimension (N)**: 4
- **Security Level**: Post-quantum secure against lattice attacks

### Performance Optimizations

- **Parallel Encryption/Decryption**: Uses goroutines for concurrent cryptographic operations
- **NTT Transformations**: Efficient polynomial multiplication using Number Theoretic Transform
- **Optimized Matrix Operations**: Custom implementations for lattice-based computations
- **Database Caching**: Efficient storage and retrieval of cryptographic accumulators

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔬 Research Background

This implementation is based on research in lattice-based cryptography and Private Set Intersection protocols. The system provides:

- **Laconic PSI**: Efficient Private Set Intersection with sublinear communication
- **Homomorphic Properties**: Computation on encrypted data without decryption
- **Post-Quantum Security**: Protection against both classical and quantum adversaries

## ⚠️ Security Notice

This is a research implementation. While built on sound cryptographic principles, it should undergo thorough security review before production use. The parameters and implementation choices prioritize educational value and functionality demonstration.


---

**Disclaimer**: This software is provided for research and educational purposes. Users are responsible for ensuring compliance with applicable laws and regulations when using cryptographic software.
