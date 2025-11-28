# Sanctions Screening Frontend - Requirements & Features

**Project:** LE-PSI Sanctions Screening System  
**Version:** 1.0  
**Last Updated:** November 15, 2025

---

## Overview

A web-based frontend application for privacy-preserving sanctions screening using LE-PSI (Lattice-based Private Set Intersection). The system enables financial institutions to screen customer lists against sanctions databases (OFAC, UN, EU) without revealing sensitive customer information.

---

## Core Features

### 1. Dashboard

**Purpose:** Central hub providing overview of screening operations and quick access to key functions.

**Features:**
- **Summary Cards**
  - Total customers screened (last 30 days)
  - Active matches requiring review
  - Last screening timestamp
  - System status (healthy/processing/error)

- **Recent Activity Timeline**
  - Last 10 screening operations with status
  - Quick view of match counts per screening
  - Direct links to results

- **Quick Actions**
  - "Run New Screening" button
  - "Upload Customer List" button
  - "View Pending Matches" button

- **Statistics Charts**
  - Screening frequency (last 7 days/30 days)
  - Match trends over time
  - Processing time metrics

**User Roles:** All users

---

### 2. Customer List Management

**Purpose:** Upload, manage, and validate customer data for screening.

**Features:**
- **Upload Interface**
  - Drag-and-drop file upload
  - Support formats: CSV, Excel (.xlsx), JSON
  - Sample template download
  - Bulk import (up to 5,000 records)

- **Data Validation**
  - Required fields check: Name, Date of Birth, Country, Customer ID
  - Optional fields: Address, Nationality, Passport Number, Tax ID
  - Format validation (DOB: YYYY-MM-DD, Country: ISO codes)
  - Duplicate detection
  - Real-time validation feedback

- **Preview & Review**
  - Table view of uploaded data (first 100 rows)
  - Edit individual records before submission
  - Column mapping (if headers don't match)
  - Error highlighting with fix suggestions

- **Saved Lists**
  - View previously uploaded customer lists
  - Metadata: upload date, uploader, record count
  - Reuse lists for new screenings
  - Delete old lists (with confirmation)

- **Manual Entry**
  - Form to add individual customers
  - Auto-save drafts
  - Batch entry mode (multiple records at once)

**Data Structure:**
```json
{
  "customerId": "C12345",
  "name": "John Smith",
  "dateOfBirth": "1980-01-15",
  "country": "US",
  "address": "123 Main St, New York, NY",
  "nationality": "US",
  "passportNumber": "P123456789",
  "taxId": "123-45-6789"
}
```

**User Roles:** Admin, Compliance Officer

---

### 3. Sanctions List Management

**Purpose:** Maintain up-to-date sanctions databases from multiple sources.

**Features:**
- **Active Lists Overview**
  - List of active sanctions databases
  - Source: OFAC, UN, EU, HMT, custom
  - Record count per list
  - Last updated timestamp
  - Status: active/inactive/updating

- **Upload/Update**
  - Upload new sanctions list (CSV, Excel, JSON)
  - Auto-update from official sources (API integration)
  - Schedule automatic updates (daily/weekly)
  - Version control (keep history of updates)

- **List Details**
  - View all entries in a sanctions list
  - Search and filter by name, country, entity type
  - Export subset for offline review
  - Compare versions (show what changed)

- **Data Fields**
  - Name (individual/entity)
  - Aliases/AKAs
  - Date of Birth (for individuals)
  - Country/Nationality
  - Sanction program (e.g., OFAC SDN, UN Taliban)
  - Added date
  - Remarks/Notes

- **Multiple List Support**
  - Enable/disable specific lists for screening
  - Merge multiple lists for single screening
  - Priority ranking (if same entity in multiple lists)

**Data Structure:**
```json
{
  "sanctionId": "OFAC-12345",
  "name": "Ivan Petrov",
  "aliases": ["I. Petrov", "Petrov Ivan"],
  "dateOfBirth": "1970-05-10",
  "country": "RU",
  "nationality": "Russian",
  "sanctionProgram": "OFAC SDN",
  "addedDate": "2022-03-01",
  "remarks": "Designated for involvement in..."
}
```

**User Roles:** Admin

---

### 4. Run Screening

**Purpose:** Execute privacy-preserving PSI screening between customer and sanctions lists.

**Features:**
- **Screening Configuration**
  - Select customer list (dropdown of uploaded lists)
  - Select sanctions list(s) (multi-select checkboxes)
  - Screening name/description (optional)
  - Advanced options:
    - Fuzzy matching threshold (0-100%)
    - Date of birth tolerance (±years)
    - Country exact match vs. related countries

- **Execution**
  - "Start Screening" button
  - Background processing with job ID
  - Real-time progress indicator
    - Stage 1: Preparing customer data
    - Stage 2: Server initialization (adaptive threading logs)
    - Stage 3: Client encryption
    - Stage 4: Intersection detection
  - Estimated time remaining (based on dataset size)
  - Cancel option (if needed)

- **Scheduling**
  - Schedule recurring screenings (daily, weekly, monthly)
  - Set time for execution (e.g., 2 AM daily)
  - Email notification on completion
  - Auto-export results to specified location

- **Performance Metrics Display**
  - Number of workers used (adaptive threading)
  - Memory usage (estimated RAM)
  - Execution time per phase
  - Throughput (records/second)

**User Roles:** Admin, Compliance Officer

---

### 5. Screening Results

**Purpose:** Display matches found during screening and enable investigation workflow.

**Features:**
- **Results Overview**
  - Screening metadata: date/time, lists used, total records screened
  - Match summary: total matches, high/medium/low confidence
  - Status: pending review, reviewed, escalated, false positive
  - Export options: CSV, Excel, PDF report

- **Matches Table**
  - Columns: Customer Name, Customer ID, Matched Sanctions Entity, Match Score, Status, Actions
  - Sorting by any column
  - Filtering:
    - By status (pending, reviewed, escalated, false positive)
    - By confidence score (>90%, 70-90%, <70%)
    - By sanctions source (OFAC, UN, EU)
    - By date range
  - Pagination (25/50/100 per page)

- **Search & Filter**
  - Quick search by customer name or ID
  - Advanced filters sidebar
  - Save filter presets

- **Bulk Actions**
  - Select multiple matches (checkboxes)
  - Bulk mark as reviewed/false positive
  - Bulk export selected matches
  - Bulk assign to investigator

- **Match Score Indicator**
  - Color-coded badges:
    - Red: 90-100% (high confidence)
    - Yellow: 70-89% (medium confidence)
    - Gray: <70% (low confidence)
  - Tooltip explaining match criteria

**User Roles:** All users (view), Compliance Officer (edit)

---

### 6. Match Details & Investigation

**Purpose:** Deep dive into individual matches for compliance review.

**Features:**
- **Side-by-Side Comparison**
  - Left panel: Customer information
  - Right panel: Sanctions entity information
  - Highlighted matching fields (name, DOB, country)

- **Match Score Breakdown**
  - Name similarity: 95%
  - DOB match: exact/within tolerance
  - Country match: exact/related
  - Overall confidence: 92%

- **Investigation Workflow**
  - Status dropdown: Pending → Under Review → Reviewed → Escalated → False Positive
  - Assign to investigator (dropdown of users)
  - Priority flag (high/medium/low)
  - Due date (SLA tracking)

- **Notes & Comments**
  - Add investigation notes (rich text editor)
  - Timestamp and author for each note
  - Attach supporting documents (PDFs, images)
  - Internal comments vs. compliance audit trail

- **Actions**
  - Mark as false positive (with reason)
  - Escalate to senior compliance (with reason)
  - Generate compliance report (PDF)
  - Block customer account (if integrated with core banking)
  - File SAR (Suspicious Activity Report) - integration

- **Audit Trail**
  - Full history of status changes
  - Who reviewed, when, what action taken
  - Exportable for regulatory compliance

**User Roles:** Compliance Officer, Admin

---

### 7. Audit Log & History

**Purpose:** Complete audit trail of all screening activities for compliance.

**Features:**
- **Screening History**
  - List of all past screenings (chronological)
  - Columns: Date, Initiated By, Customer List, Sanctions List, Matches Found, Status
  - Filter by date range, user, status
  - Search by screening ID or name

- **Detailed Logs**
  - Click on screening to view full details
  - Input parameters (lists used, options)
  - Execution logs (timestamps, phases, errors)
  - Performance metrics (workers, memory, time)
  - Output: number of matches, where results stored

- **User Activity Log**
  - Who logged in when
  - What actions taken (upload, screening, review, export)
  - Failed login attempts
  - Data access logs (GDPR compliance)

- **System Events**
  - Sanctions list updates
  - Scheduled screening executions
  - System errors/warnings
  - Database backups

- **Export Logs**
  - Download full audit log (CSV, JSON)
  - Filter before export
  - Encrypted export for sensitive data

- **Retention Policy Display**
  - How long logs are kept (e.g., 7 years for compliance)
  - Archive old logs
  - Auto-delete after retention period

**User Roles:** Admin (full access), Compliance Officer (screening logs only)

---

### 8. User Management

**Purpose:** Manage user access and permissions (multi-user support).

**Features:**
- **User List**
  - Table: Name, Email, Role, Status (active/inactive), Last Login
  - Search and filter users
  - Add new user button

- **Add/Edit User**
  - Form fields: Name, Email, Password (hashed), Role
  - Role options:
    - **Admin**: Full access (manage users, lists, screenings, settings)
    - **Compliance Officer**: Run screenings, review matches, view logs
    - **Viewer**: Read-only access to results and logs
  - Email verification (send activation link)
  - Two-factor authentication (2FA) option

- **Role Permissions Matrix**
  - Visual table showing what each role can do
  - Granular permissions (e.g., view vs. edit vs. delete)

- **User Activity**
  - View individual user's activity log
  - Last actions, screenings run, matches reviewed
  - Session management (force logout)

- **Deactivate/Delete**
  - Soft delete (deactivate) vs. hard delete
  - Transfer ownership of screenings/matches before deletion
  - Audit trail of deletions

**User Roles:** Admin only

---

### 9. Settings & Configuration

**Purpose:** System-wide configuration and customization.

**Features:**
- **General Settings**
  - Organization name and logo
  - Time zone
  - Date/time format preferences
  - Language (if multi-language support)

- **Screening Parameters**
  - Default fuzzy matching threshold (e.g., 85%)
  - Default DOB tolerance (e.g., ±2 years)
  - Country matching rules (exact vs. related)
  - Maximum customer list size (default: 5,000)

- **Performance Settings**
  - Adaptive threading configuration:
    - Available RAM (default: 117 GB)
    - CPU cores (default: 48)
    - Safety margin (default: 15%)
  - Database path for tree storage
  - Verbose logging toggle (PSI_VERBOSE)

- **Notifications**
  - Email notifications for:
    - Screening completion
    - New matches found
    - High-priority matches (>90% confidence)
    - System errors
  - Webhook/API notifications (for integrations)
  - SMTP configuration (email server settings)

- **Data Retention**
  - How long to keep screening results (e.g., 7 years)
  - Auto-archive old screenings
  - Purge policy for customer/sanctions lists

- **API Keys & Integrations**
  - Generate API keys for external access
  - Integrate with sanctions list providers (OFAC API, etc.)
  - Webhook URLs for third-party integrations
  - Core banking system integration (optional)

- **Security**
  - Password policy (length, complexity, expiration)
  - Session timeout (e.g., 30 minutes)
  - IP whitelist (restrict access by IP)
  - Enable/disable 2FA for all users

- **Backup & Recovery**
  - Schedule automatic database backups
  - Backup location (local/cloud)
  - Restore from backup option
  - Export full system configuration

**User Roles:** Admin only

---

## Additional Features

### 10. Reporting & Analytics

**Purpose:** Generate compliance reports and business intelligence.

**Features:**
- **Predefined Reports**
  - Monthly screening summary (PDF/Excel)
  - Matches by sanctions source (OFAC, UN, EU)
  - False positive rate over time
  - Investigator performance (matches reviewed, avg time)

- **Custom Reports**
  - Report builder: select fields, filters, date range
  - Save report templates
  - Schedule automated report generation

- **Dashboards**
  - Executive dashboard (high-level metrics)
  - Compliance dashboard (detailed screening stats)
  - Operational dashboard (system performance)

- **Export Options**
  - PDF (for regulatory submission)
  - Excel (for further analysis)
  - CSV (for data import)
  - JSON (for API access)

**User Roles:** Admin, Compliance Officer

---

### 11. Help & Documentation

**Purpose:** In-app guidance and support.

**Features:**
- **User Guide**
  - Step-by-step tutorials
  - Video walkthroughs
  - FAQ section
  - Glossary of terms

- **Contextual Help**
  - Tooltips on hover (explain fields)
  - "?" icons next to complex features
  - Inline documentation

- **Support**
  - Contact support form
  - Live chat (if available)
  - Submit bug reports
  - Feature request submission

**User Roles:** All users

---

## Technical Specifications

### Frontend Tech Stack Recommendations

**Framework:**
- React.js (with TypeScript)
- Next.js (for SSR and SEO)
- Vue.js (alternative)

**UI Library:**
- Material-UI (MUI)
- Ant Design
- Tailwind CSS + Headless UI

**State Management:**
- Redux Toolkit (for complex state)
- React Query (for API data)
- Zustand (lightweight alternative)

**Charts/Visualization:**
- Chart.js
- Recharts
- D3.js (advanced)

**File Upload:**
- react-dropzone
- Uppy

**Forms:**
- React Hook Form
- Formik

**Tables:**
- TanStack Table (formerly React Table)
- ag-Grid (enterprise features)

---

### Backend Integration

**API Endpoints Required:**

```
POST   /api/customers/upload          - Upload customer list
GET    /api/customers                 - Get saved customer lists
DELETE /api/customers/:id             - Delete customer list

POST   /api/sanctions/upload          - Upload sanctions list
GET    /api/sanctions                 - Get sanctions lists
PUT    /api/sanctions/:id             - Update sanctions list

POST   /api/screening/start           - Start new screening
GET    /api/screening/:id/status      - Get screening progress
GET    /api/screening/:id/results     - Get screening results
GET    /api/screening/history         - Get screening history

GET    /api/matches                   - Get all matches
GET    /api/matches/:id               - Get match details
PUT    /api/matches/:id               - Update match status/notes
POST   /api/matches/:id/escalate      - Escalate match

GET    /api/audit-log                 - Get audit logs
GET    /api/users                     - Get users (admin)
POST   /api/users                     - Create user (admin)
PUT    /api/users/:id                 - Update user (admin)

GET    /api/settings                  - Get settings
PUT    /api/settings                  - Update settings

POST   /api/reports/generate          - Generate report
GET    /api/reports/:id               - Download report
```

**WebSocket/SSE:**
- Real-time screening progress updates
- Notification delivery

---

### Security Requirements

**Authentication:**
- JWT tokens (short-lived access token + refresh token)
- OAuth 2.0 (optional, for SSO)
- 2FA using TOTP (Google Authenticator)

**Authorization:**
- Role-based access control (RBAC)
- Route guards on frontend
- API endpoint permissions

**Data Security:**
- HTTPS/TLS for all communications
- Data encryption at rest (database)
- Data encryption in transit
- Secure password hashing (bcrypt/Argon2)
- XSS protection
- CSRF tokens
- Rate limiting on APIs

**Compliance:**
- GDPR compliance (data privacy)
- CCPA compliance (for US)
- SOC 2 Type II (for SaaS)
- Audit logging for all data access

---

### Performance Requirements

**Response Times:**
- Page load: <2 seconds
- API calls: <500ms (excluding screening execution)
- Search/filter: <300ms
- File upload: Progress indicator for >5 seconds

**Scalability:**
- Support 100+ concurrent users
- Handle 5,000 customer records upload
- Handle 10,000+ sanctions entries
- Store 1 year of screening history (estimated 50GB)

**Browser Support:**
- Chrome (last 2 versions)
- Firefox (last 2 versions)
- Safari (last 2 versions)
- Edge (last 2 versions)

**Responsive Design:**
- Desktop (1920×1080, 1366×768)
- Tablet (iPad, 1024×768)
- Mobile (375×667) - view-only mode

---

## User Flows

### Flow 1: First-Time Screening

1. User logs in → Dashboard
2. "Upload Customer List" → Upload CSV → Validate → Save
3. "Manage Sanctions List" → Upload OFAC list → Save
4. "Run Screening" → Select lists → Configure → Start
5. Wait for completion (progress bar)
6. View results → See matches
7. Click on match → Investigate → Mark status
8. Export report for compliance

### Flow 2: Scheduled Daily Screening

1. Admin sets up scheduled screening (daily at 2 AM)
2. System auto-runs screening with latest customer list
3. Email notification sent if matches found
4. Compliance officer logs in → Reviews new matches
5. Investigates each match → Updates status
6. High-confidence matches escalated to senior compliance
7. Monthly report auto-generated for regulators

### Flow 3: Sanctions List Update

1. Admin logs in → Sanctions List Management
2. "Update OFAC List" → Upload new CSV
3. System compares new vs. old → Shows changes
4. Confirm update
5. System triggers re-screening of recent customers (optional)
6. Notification sent to compliance team

---

## Wireframe Suggestions

### Dashboard (High-Level)
```
+----------------------------------------------------------+
| [Logo] Sanctions Screening System          [User Menu ▼] |
+----------------------------------------------------------+
| Dashboard | Customers | Sanctions | Run Screening | ... |
+----------------------------------------------------------+
|                                                          |
| Summary Cards:                                           |
| +-------------+ +-------------+ +-------------+          |
| | 5,234       | | 12          | | 2 hours ago |          |
| | Customers   | | Matches     | | Last Scan   |          |
| | Screened    | | Pending     | |             |          |
| +-------------+ +-------------+ +-------------+          |
|                                                          |
| Recent Screenings:                                       |
| +----------------------------------------------------+   |
| | Date       | Customer List | Matches | Status     |   |
| |------------|---------------|---------|------------|   |
| | Nov 15 2AM | Q4-2025       | 3       | Complete   |   |
| | Nov 14 2AM | Q4-2025       | 2       | Complete   |   |
| +----------------------------------------------------+   |
|                                                          |
| [Chart: Screening Frequency]  [Chart: Match Trends]     |
|                                                          |
+----------------------------------------------------------+
```

### Screening Results
```
+----------------------------------------------------------+
| Screening Results: Q4-2025 Customers vs OFAC             |
| Date: Nov 15, 2025 2:00 AM | Matches: 3 | Export [CSV ▼]|
+----------------------------------------------------------+
| Filter: [Status ▼] [Confidence ▼] [Search...        🔍] |
+----------------------------------------------------------+
| ☑ | Customer Name | Customer ID | Matched Entity | Score | Status      | Actions     |
|---|---------------|-------------|----------------|-------|-------------|-------------|
| ☐ | John Smith    | C12345      | Ivan Petrov    | 92%   | Pending     | [View][Edit]|
| ☐ | Maria Garcia  | C67890      | M. Garcia      | 88%   | Under Review| [View][Edit]|
| ☐ | Alex Johnson  | C11223      | A. Johnsson    | 75%   | False Pos.  | [View][Edit]|
+----------------------------------------------------------+
| Bulk Actions: [Mark Reviewed] [Export Selected]         |
+----------------------------------------------------------+
```

---

## Success Metrics

**User Adoption:**
- 90%+ of compliance officers use system weekly
- <2 hours average training time for new users

**Efficiency:**
- 50% reduction in manual screening time
- 70% faster match investigation

**Accuracy:**
- <5% false positive rate
- 100% detection of true matches (no false negatives)

**Compliance:**
- 100% audit trail coverage
- <24 hours match resolution time (SLA)
- Zero data breaches

---

## Future Enhancements (Phase 2)

1. **AI-Powered Fuzzy Matching**
   - Machine learning for better name matching
   - OCR for document verification

2. **Mobile App**
   - iOS/Android app for on-the-go review
   - Push notifications for urgent matches

3. **Advanced Analytics**
   - Predictive risk scoring
   - Network analysis (find hidden relationships)

4. **Multi-Tenancy**
   - Support multiple organizations on same platform
   - Data isolation and white-labeling

5. **Blockchain Integration**
   - Immutable audit trail on blockchain
   - Decentralized sanctions list verification

6. **API Marketplace**
   - Third-party integrations (Chainalysis, Refinitiv)
   - Plugin architecture for custom extensions

---

## Conclusion

This frontend will provide a comprehensive, user-friendly interface for privacy-preserving sanctions screening. The design prioritizes compliance, auditability, and operational efficiency while leveraging the LE-PSI library's post-quantum security advantages.

**Next Steps:**
1. Create detailed wireframes/mockups
2. Build API backend (Go/Node.js)
3. Develop frontend components (React)
4. Integration testing with LE-PSI library
5. Security audit and penetration testing
6. Pilot deployment with test users
7. Production rollout

---

**Document Version:** 1.0  
**Author:** Development Team  
**Review Date:** November 15, 2025
