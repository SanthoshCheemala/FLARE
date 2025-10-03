# FLARE 2.0: Advanced Lattice-based Private Set Intersection System

FLARE (Fast Lattice-based Authenticated Range Encryption) 2.0 is a next-generation post-quantum secure Private Set Intersection (PSI) system built using cutting-edge lattice-based cryptography. It enables two parties to securely identify common elements in their datasets without revealing any sensitive information, featuring comprehensive analytics and modular architecture.

## 🚀 What's New in Version 2.0

- **🏗️ Modular PSI Architecture**: Separated client and server logic for better maintainability
- **📊 Advanced Analytics Engine**: Comprehensive noise analysis, performance profiling, and security assessment
- **🔧 Build Tag System**: Production and analytics builds with different optimization levels
- **📈 Real-time Performance Monitoring**: Detailed timing, throughput, and quality metrics
- **🛡️ Enhanced Security Assessment**: Post-quantum security evaluation and recommendations
- **📋 Multi-format Reporting**: HTML dashboards, JSON statistics, and optimization reports

## 🔐 Core Features

### Security & Cryptography
- **Post-Quantum Security**: Built on lattice-based cryptography, resistant to quantum computer attacks
- **Laconic PSI Protocol**: Efficient Private Set Intersection with sublinear communication complexity
- **Homomorphic Operations**: Perform computations on encrypted data using advanced polynomial arithmetic
- **Configurable Security Levels**: Support for ring dimensions 256, 512, 1024, and 2048 bits

### Performance & Analytics
- **Comprehensive Noise Analysis**: Real-time monitoring of cryptographic noise levels and distribution
- **Performance Profiling**: Detailed timing analysis, throughput measurement, and efficiency scoring
- **Quality Metrics**: Advanced correctness checking with match percentage analysis
- **Memory Optimization**: Efficient polynomial operations with NTT (Number Theoretic Transform)

### Architecture & Modularity
- **Clean Separation**: Distinct client and server implementations in modular packages
- **Flexible Storage**: SQLite-based cryptographic accumulator storage with tree structures
- **Parallel Processing**: Optimized Go concurrency for enhanced performance
- **Extensible Design**: Easy to extend with new PSI variants and optimization techniques

## 🏗️ Project Structure

```
FLARE/
├── cmd/Flare/                    # Application entry points
│   ├── main.go                   # Production build (fast execution)
│   └── main_analytics.go         # Analytics build (comprehensive reporting)
├── internal/                     # Internal packages
│   ├── crypto/                   # Cryptographic operations
│   │   ├── PSI/                  # Modular PSI implementation
│   │   │   ├── common.go         # Shared utilities and types
│   │   │   ├── client.go         # Client-side PSI logic
│   │   │   └── server.go         # Server-side PSI logic
│   │   ├── psi.go                # Main PSI interface
│   │   ├── psi_analytics.go      # Analytics-enabled PSI (build tag)
│   │   ├── helpers.go            # Cryptographic utilities
│   │   └── parameters.go         # System parameter management
│   └── storage/                  # Database operations
├── pkg/                          # Public packages
│   ├── LE/                       # Lattice Encryption core
│   └── matrix/                   # Polynomial and matrix operations
├── utils/                        # Utility functions
│   └── report_generation.go     # Advanced analytics reporting
├── data/                         # Data storage
└── results/                      # Generated reports and analytics
```

## 📋 Prerequisites

- **Go**: Version 1.24.1 or higher
- **SQLite3**: For cryptographic accumulator storage
- **Git**: For version control and dependency management

## 🚀 Installation

1. **Clone the repository:**
```bash
git clone https://github.com/SanthoshCheemala/FLARE.git
cd FLARE
```

2. **Install dependencies:**
```bash
go mod tidy
```

3. **Build the application:**

**Production Build (Optimized for Performance):**
```bash
go build -o flare cmd/Flare/main.go
```

**Analytics Build (Comprehensive Reporting):**
```bash
go build -tags analytics -o flare-analytics cmd/Flare/main_analytics.go
```

## 💻 Usage

### Production Mode (Fast Execution)

```bash
# Basic PSI with default parameters
./flare -columns="type,amount" -LIMIT=100

# Process specific columns
./flare -columns="user_id,timestamp,amount" -LIMIT=50
```

### Analytics Mode (Comprehensive Analysis)

```bash
# Full analytics with default settings
./flare-analytics -columns="type,amount" -LIMIT=50

# Advanced analytics with custom parameters
./flare-analytics \
  -columns="type,amount" \
  -columns-merge="user_id,timestamp" \
  -LIMIT=100 \
  -ring-dimension=512 \
  -output-dir="analysis_results" \
  -advanced-analytics=true \
  -verbose=true

# Custom security analysis
./flare-analytics \
  -columns="sensitive_data" \
  -ring-dimension=1024 \
  -report-format="both" \
  -LIMIT=200
```

### Command Line Options

#### Common Options
- `-columns`: Comma-separated list of columns to process (required)
- `-LIMIT`: Number of rows to process (default: production=2, analytics=50)

#### Analytics-Specific Options
- `-columns-merge`: Additional columns to merge for enhanced security
- `-ring-dimension`: Lattice ring dimension - 256, 512, 1024, 2048 (default: 256)
- `-output-dir`: Directory for generated reports (default: "data")
- `-advanced-analytics`: Enable comprehensive analytics (default: true)
- `-report-format`: Output format - "html", "json", or "both" (default: "html")
- `-verbose`: Enable detailed logging (default: false)

## 📁 Data Structure

The system expects a SQLite database at `data/transactions.db` with financial transaction data:

```sql
CREATE TABLE finanical_transactions (
    id INTEGER PRIMARY KEY,
    type TEXT,
    amount REAL,
    user_id TEXT,
    timestamp DATETIME,
    category TEXT,
    description TEXT
);
```

## � Analytics & Reporting

### Generated Reports

**Analytics mode generates comprehensive reports:**

1. **📋 Advanced HTML Dashboard** (`flare_psi_advanced_report.html`)
   - Real-time noise analysis with interactive charts
   - Performance metrics and timing breakdowns
   - Security assessment and recommendations
   - Quality scoring and efficiency analysis

2. **📈 JSON Statistics** (`flare_psi_statistics.json`)
   - Detailed numerical data for further analysis
   - Noise distribution patterns
   - Error analysis and timing metrics
   - Cryptographic parameter effectiveness

3. **⚡ Performance Profile** (`performance_profile.json`)
   - Operation timing distributions
   - Throughput analysis and bottleneck identification
   - Memory usage patterns
   - Optimization recommendations

4. **🛡️ Security Assessment** (`security_assessment.json`)
   - Post-quantum security evaluation
   - Vulnerability risk analysis
   - Parameter strength assessment
   - Compliance scoring

5. **🔧 Optimization Report** (`optimization_recommendations.json`)
   - Performance improvement opportunities
   - Parameter tuning suggestions
   - ROI analysis for optimizations
   - Implementation roadmap

### Key Metrics Tracked

- **Noise Analysis**: Maximum/average noise levels, distribution patterns
- **Performance**: Throughput, latency, efficiency scores
- **Quality**: Match percentages, correctness validation
- **Security**: Parameter strength, post-quantum readiness
- **Stability**: System reliability and error patterns

## 🔧 Technical Implementation

### Cryptographic Core

**Lattice Encryption (LE) Parameters:**
```go
type LE struct {
    Q      uint64  // Modulus (180143985094819841)
    D      int     // Ring dimension (256/512/1024/2048)
    N      int     // Matrix dimension (4)
    Layers int     // Tree layers (auto-calculated)
    Sigma  float64 // Gaussian noise parameter
    // ... additional cryptographic parameters
}
```

**PSI Architecture:**
```go
// Modular PSI package structure
package psi

// Ciphertext structure for secure communication
type Cxtx struct {
    C0 []*matrix.Vector  // First ciphertext component
    C1 []*matrix.Vector  // Second ciphertext component
    C  *matrix.Vector    // Combined ciphertext
    D  *ring.Poly        // Polynomial component
}

// Client-side PSI implementation
func Client(clientTx, serverTx []Transaction, treePath string) ([]Transaction, error)

// Server-side PSI implementation  
func Server(pp *Vector, msg *Poly, serverTx []Transaction, le *LE) []Cxtx
```

### Performance Optimizations

- **🔄 Parallel Cryptographic Operations**: Concurrent key generation and encryption
- **⚡ NTT Transformations**: Efficient polynomial multiplication algorithms
- **💾 Optimized Memory Management**: Smart allocation patterns for large datasets
- **🏗️ Modular Architecture**: Clean separation enabling independent optimization
- **📊 Real-time Monitoring**: Performance tracking with minimal overhead

### Security Features

- **🛡️ Post-Quantum Resistance**: Based on Learning With Errors (LWE) problem
- **🔒 Configurable Security Levels**: Adjustable ring dimensions for different threat models
- **📈 Noise Management**: Advanced noise analysis and threshold monitoring
- **🔍 Correctness Validation**: Probabilistic verification with configurable thresholds
- **📋 Security Assessment**: Automated evaluation of cryptographic strength

## 🤝 Contributing

We welcome contributions! Please see our contribution guidelines:

1. **Fork the repository** and create a feature branch
2. **Follow Go best practices** and maintain code quality
3. **Add tests** for new functionality
4. **Update documentation** including this README
5. **Submit a pull request** with clear description

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🔬 Research Background

FLARE 2.0 implements state-of-the-art research in:

- **Laconic Private Set Intersection**: Efficient PSI with sublinear communication
- **Lattice-based Cryptography**: Post-quantum secure cryptographic primitives  
- **Homomorphic Encryption**: Computation on encrypted data without decryption
- **Cryptographic Accumulators**: Efficient membership proofs using tree structures
- **Performance Analytics**: Real-time cryptographic system monitoring

### Academic References

This implementation builds upon research in:
- Lattice-based cryptography and the LWE problem
- Private Set Intersection protocols and optimizations
- Homomorphic encryption schemes and applications
- Post-quantum cryptographic security analysis

## ⚠️ Security Notice

**Important**: This is a research implementation designed for educational and experimental purposes. While built on sound cryptographic principles:

- **🔍 Security Review Required**: Thorough security audit recommended before production use
- **📊 Parameter Validation**: Default parameters optimized for demonstration, not production security
- **🛡️ Threat Model**: Designed for semi-honest adversaries in research contexts
- **⚖️ Compliance**: Users responsible for regulatory compliance in their jurisdiction

## 📞 Support & Community

- **🐛 Issues**: Report bugs via GitHub Issues
- **💡 Feature Requests**: Suggest improvements via GitHub Discussions  
- **📚 Documentation**: Additional docs in `/doc` directory
- **🔬 Research**: Contact maintainers for academic collaboration

---

**⚡ FLARE 2.0 - Advancing the frontiers of post-quantum Private Set Intersection**

*Built with ❤️ for the cryptographic research community*
