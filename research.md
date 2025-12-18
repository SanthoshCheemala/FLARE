# FLARE: Federated Learning & Anonymous Risk Evaluation
## Research Findings & Technical Insights

### 1. Project Overview
**FLARE** (Federated Learning & Anonymous Risk Evaluation) is a privacy-preserving sanctions screening solution designed to address the conflict between regulatory compliance (AML/KYC) and data privacy (GDPR/CCPA). It enables financial institutions ("Client") to screen their customer lists against a Sanctions Authority's ("Server") blocklists without either party revealing their underlying datasets to the other.

**Core Problem:**
*   **Banks** cannot share raw customer PII with external vendors due to privacy laws.
*   **Sanctions Authorities** (or vendors) often consider their curated lists as proprietary intellectual property and cannot distribute the full raw list to every bank.
*   **Traditional Solution:** Hashing (SHA-256) is vulnerable to dictionary attacks and rainbow tables due to the low entropy of names.

**FLARE Solution:**
FLARE utilizes **Private Set Intersection (PSI)** based on **Laconic Cryptography**. Specifically, it uses the **LE-PSI** (Label-Equipped PSI) protocol powered by the **Lattigo** library (Ring-Learning With Errors cryptography). This ensures that:
1.  The Server learns *nothing* about the Client's customers (not even the set size, ideally).
2.  The Client learns *only* the intersection (customers who are on the sanctions list) and nothing else about the Server's database.
3.  **Laconic Properties:** The server cannot decrypt or reverse-engineer the client's data because the cryptographic scheme allows computing the intersection on compressed (laconic) representations without ever expanding them to the full plaintext.

### 2. System Architecture
The system is architected as a distributed application with two distinct roles:

#### A. FLARE Client (The Bank)
*   **Role:** Data Owner (Customer PII).
*   **Responsibilities:**
    *   Manages local customer datasets.
    *   Initiates screening jobs.
    *   Encrypts customer data using the PSI protocol.
    *   Receives encrypted results and decrypts the intersection.
*   **Tech Stack:** Go (Backend), Next.js (Frontend), SQLite (Local DB).

#### B. FLARE Server (The Sanctions Authority)
*   **Role:** Service Provider (Sanctions List).
*   **Responsibilities:**
    *   Manages global sanctions lists.
    *   Pre-computes cryptographic public parameters (Public Key, Relinearization Keys, etc.) at startup.
    *   Performs the private set intersection computation on encrypted data.
*   **Tech Stack:** Go (Backend), Next.js (Frontend), PostgreSQL (Central DB).

### 3. Technical Implementation Details

#### Cryptographic Core (LE-PSI & Lattigo)
The heart of FLARE is the `psiadapter` package, which wraps the `LE-PSI` library.
*   **Protocol:** Unbalanced PSI (Client set << Server set).
*   **Encryption Scheme:** Ring-LWE based Laconic Cryptography.
*   **Workflow:**
    1.  **Setup:** Server generates global parameters (Polynomial Ring $R_q$, Moduli chain) and pre-processes its sanctions list at startup.
    2.  **Request:** Client requests these public parameters.
    3.  **Encryption:** Client concatenates its inputs -> Hashes them -> Encrypts them into a `ClientCiphertext`.
    4.  **Intersection:** Client sends ciphertexts to Server. Server computes the intersection using Laconic Cryptography.
    5.  **Decryption:** Server returns encrypted matches. Client decrypts them to reveal the IDs of matched records.

#### Unified Frontend with Dynamic Role Adaptation
To streamline development while maintaining logical separation:
*   **Single Codebase:** A unified `flare-ui` repository serves both Client and Server.
*   **Dynamic Adaptation:** The application detects its mode via `NEXT_PUBLIC_APP_MODE` (`CLIENT` or `SERVER`) at build/runtime.
*   **RBAC:**
    *   **Sidebar/Navigation:** Dynamically filters routes (e.g., "Screening" is hidden for Server, "Sanctions Upload" is hidden for Client).
    *   **Authentication:** Strict role enforcement at login (Bank Admin cannot log into Server UI).

### 4. Key Research Findings & Challenges

#### A. Serialization of Complex Cryptographic Structures
**Challenge:** The Lattigo library's structures (`ring.Poly`, `matrix.Matrix`, `LE.LE`) are complex, deeply nested, and often contain unexported fields or pointers that standard Go serialization (`encoding/gob` or `encoding/json`) cannot handle automatically.
**Finding:**
*   **NTT Matrices are Critical:** The encryption performance relies heavily on pre-computed Number Theoretic Transform (NTT) matrices (`A0NTT`, `A1NTT`, `BNTT`).
*   **Issue:** Initial attempts to serialize the `LE` struct failed to include these NTT matrices. While the struct *looked* complete, the encryption algorithm panicked (`segmentation violation`) when trying to access these nil pointers during matrix multiplication.
*   **Solution:** We implemented a custom `SerializedLE` struct and explicit `SerializeParams`/`DeserializeParams` methods. Crucially, we switched to using the official `LE-PSI` library's helper functions (`psi.SerializeParameters`) which are purpose-built to handle these internal states correctly.

#### B. Distributed System Simulation
**Challenge:** Developing a distributed system (Client & Server) on a single local machine.
**Finding:**
*   **Port Conflicts:** Running two Next.js instances simultaneously requires distinct ports (3000 vs 3001) and distinct build directories (`.next-client` vs `.next-server`) to avoid lockfile contention.
*   **Proxying:** The Next.js `rewrites()` must be dynamically configured based on the `APP_MODE` to proxy API requests to the correct backend (Client:8080 or Server:8081).

#### C. Data Normalization & Hashing
**Challenge:** PSI relies on exact string matching. "John Doe" vs "john doe" results in different hashes.
**Finding:**
*   **Strict Normalization:** We implemented a robust normalization pipeline: `Lowercase -> Trim Whitespace -> Concatenate Fields (Name|DOB|Country)`.
*   **Deterministic Hashing:** The hashing algorithm must be identical on both Client and Server. We use SHA-256 truncated to `uint64` for the PSI input.

### 5. Conclusion
FLARE demonstrates that privacy-preserving compliance is practically achievable. By abstracting the complex cryptography behind a user-friendly "Bank" and "Authority" interface, it allows compliance officers to perform their duties without exposing sensitive customer data, satisfying both regulatory and privacy requirements simultaneously.
