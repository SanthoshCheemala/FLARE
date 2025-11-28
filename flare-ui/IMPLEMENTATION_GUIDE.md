# FLARE UI - Privacy-Preserving Sanctions Screening Interface

## 🚀 Quick Start

The frontend is now fully implemented and running!

### Access the Application

1. **Start the development server** (if not already running):
   ```bash
   cd flare-ui
   npm run dev
   ```

2. **Open your browser**:
   - Local: http://localhost:3000
   - Network: http://172.16.203.161:3000

3. **Navigate automatically to Dashboard** - The root path `/` redirects to `/dashboard`

## 📁 Project Structure

```
flare-ui/
├── src/
│   ├── app/
│   │   ├── (dashboard)/           # Dashboard layout group
│   │   │   ├── layout.tsx         # Shared layout with sidebar
│   │   │   ├── dashboard/         # Main dashboard page
│   │   │   ├── screening/         # Screening execution page (COMPLEX)
│   │   │   ├── customers/         # Customer management
│   │   │   ├── sanctions/         # Sanctions lists
│   │   │   ├── results/           # Results viewer
│   │   │   ├── audit/             # Audit logs
│   │   │   └── settings/          # Settings
│   │   ├── layout.tsx             # Root layout
│   │   ├── page.tsx               # Redirects to /dashboard
│   │   └── globals.css            # Global styles
│   ├── components/
│   │   ├── ui/                    # Reusable UI components
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   └── progress.tsx
│   │   └── shared/
│   │       └── sidebar.tsx        # Navigation sidebar
│   └── lib/
│       └── utils.ts               # Utility functions
└── package.json
```

## 🎯 Implemented Features

### ✅ 1. Dashboard (`/dashboard`)
- **Summary Cards**: Total screened, active matches, last screening, system status
- **Recent Activity Timeline**: Shows recent screening jobs with match counts
- **Quick Actions**: Buttons for common tasks
- **Responsive Design**: Works on desktop and mobile

### ✅ 2. Run Screening Page (`/screening`) - MOST COMPLEX
This is the flagship feature demonstrating the LE-PSI workflow:

**Configuration Panel:**
- Customer list selection dropdown
- Sanctions database checkboxes (OFAC, UN, EU)
- Fuzzy match threshold slider
- Start screening button with loading states

**Execution Status Panel:**
- Real-time progress bar
- Visual pipeline stages:
  1. **Prep** - File preparation
  2. **Init** - Server initialization with adaptive threading
  3. **Encrypt** - PSI-Lattice encryption
  4. **PSI Match** - Private set intersection
- Terminal-style log output with timestamped messages
- Real-time metrics: Throughput, Memory Usage, CPU Load
- Completion indicator with "View Results" button

**Features:**
- Simulated multi-stage execution (8-second demo)
- Stage-based UI updates with colored indicators
- Disabled controls during execution
- Professional terminal-style logging

### ✅ 3. Sidebar Navigation
- Persistent sidebar with all routes
- Active route highlighting
- User profile section
- Icons from lucide-react

### ✅ 4. Placeholder Pages
- Customers (`/customers`)
- Sanctions Lists (`/sanctions`)
- Results (`/results`)
- Audit Log (`/audit`)
- Settings (`/settings`)

All placeholders are ready for implementation with proper headers and "Coming soon" indicators.

## 🎨 Design System

### Colors
- Primary: Slate 900 (dark backgrounds, buttons)
- Accent: Blue (active states)
- Success: Green (completion states)
- Warning: Amber (pending reviews)
- Destructive: Red (errors, critical matches)

### Components
All components use **shadcn/ui** patterns:
- `Button`: Multiple variants (default, outline, secondary, ghost, link)
- `Card`: Structured containers with headers and content
- `Progress`: Animated progress bars
- Custom utility: `cn()` for class merging

## 🔧 Technologies

- **Framework**: Next.js 16.0.1 with App Router
- **Styling**: Tailwind CSS v4
- **Icons**: lucide-react
- **TypeScript**: Full type safety
- **Components**: shadcn/ui patterns

## 📊 Data Flow (Currently Mocked)

The screening page uses `setTimeout` to simulate backend calls:
```typescript
// Example from screening page
startScreening() => {
  1. Set status to 'preparing' (5% progress)
  2. After 1.5s: 'init_server' (15% progress)
  3. After 3.5s: 'encrypting' (45% progress)
  4. After 6s: 'intersecting' (80% progress)
  5. After 8s: 'complete' (100% progress)
}
```

## 🔌 Future Backend Integration

To connect to your Go backend:

1. **Create API route handlers** in `src/app/api/`:
   ```typescript
   // src/app/api/screening/route.ts
   export async function POST(request: Request) {
     const body = await request.json()
     // Call Go backend
     const response = await fetch('http://localhost:8080/api/screen', {
       method: 'POST',
       headers: { 'Content-Type': 'application/json' },
       body: JSON.stringify(body)
     })
     return Response.json(await response.json())
   }
   ```

2. **Update the screening page** to use real API:
   ```typescript
   const response = await fetch('/api/screening', {
     method: 'POST',
     body: JSON.stringify({ customerList, sanctions })
   })
   ```

3. **WebSocket for real-time updates** (optional):
   ```typescript
   const ws = new WebSocket('ws://localhost:8080/ws/screening')
   ws.onmessage = (event) => {
     const data = JSON.parse(event.data)
     setProgress(data.progress)
     setLogs(prev => [...prev, data.log])
   }
   ```

## 🚀 Next Steps

### High Priority
1. **Connect to Backend**: Replace mock data with real API calls
2. **Results Page**: Display screening results with match details
3. **Customer Upload**: Implement CSV/Excel file upload
4. **Authentication**: Add login/logout functionality

### Medium Priority
1. **Sanctions Management**: CRUD operations for sanctions lists
2. **Audit Log**: Real-time activity tracking
3. **Settings**: System configuration panel
4. **Export Results**: Download screening results as PDF/CSV

### Nice to Have
1. **Dark Mode**: Theme toggle
2. **Notifications**: Real-time alerts for matches
3. **Charts**: Visualize screening trends
4. **Bulk Operations**: Process multiple lists simultaneously

## 🐛 Known Issues / Limitations

1. **Mock Data**: All data is currently simulated
2. **No Persistence**: Data resets on page refresh
3. **No Error Handling**: Production needs proper error boundaries
4. **No Loading States**: Some pages need skeleton loaders
5. **No Validation**: Form inputs need validation logic

## 📝 Development Tips

### Add a New Page
1. Create file in `src/app/(dashboard)/newpage/page.tsx`
2. Add route to sidebar in `src/components/shared/sidebar.tsx`
3. Use the dashboard layout automatically

### Add a New Component
1. Create in `src/components/ui/component-name.tsx`
2. Follow shadcn/ui patterns
3. Export from component file

### Customize Styling
- Edit theme colors in `src/app/globals.css`
- Extend Tailwind in `tailwind.config.ts` (if needed)

## 🎉 Success!

Your FLARE UI is now running with:
- ✅ Professional dashboard layout
- ✅ Working navigation
- ✅ Interactive screening workflow
- ✅ Real-time progress visualization
- ✅ Responsive design
- ✅ Type-safe TypeScript
- ✅ Modern UI components

**Ready to connect to your Go backend!**
