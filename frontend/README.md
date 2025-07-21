# PostgreSQL Backup Manager - Frontend

A modern React frontend for managing PostgreSQL backups with a clean, professional interface built with TypeScript, Tailwind CSS, and React Query.

## Features

- 🔐 **Secure Authentication** - Login with API URL and key
- 📊 **Real-time Dashboard** - Monitor backup status and system health
- 🗄️ **PostgreSQL Management** - Add and configure database instances
- 📦 **Backup Operations** - Create, view, and restore backups
- 📋 **Live Logs** - Real-time log streaming with filtering
- 📱 **Responsive Design** - Works perfectly on desktop and mobile
- 🎨 **Modern UI/UX** - Clean, professional interface with Tailwind CSS

## Tech Stack

- **React 19** with TypeScript
- **Vite** for fast development and building
- **Tailwind CSS** for styling
- **React Router** for navigation
- **React Query (TanStack Query)** for data fetching and caching
- **React Hook Form** for form management
- **Lucide React** for icons
- **Axios** for HTTP requests
- **date-fns** for date formatting

## Getting Started

### Prerequisites

- Node.js 18+ 
- npm or yarn
- PostgreSQL Backup Service running (backend)

### Installation

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

The application will be available at `http://localhost:3000` and will automatically open in your browser.

### Building for Production

```bash
npm run build
```

The built files will be in the `dist` directory.

### Preview Production Build

```bash
npm run preview
```

## Configuration

The frontend connects to your PostgreSQL Backup Service backend. On first use:

1. Navigate to the login page
2. Enter your service URL (e.g., `http://localhost:8080`)
3. Enter your API key
4. Click "Connect to Service"

The connection details are saved locally and the app will automatically reconnect on subsequent visits.

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

## Project Structure

```
src/
├── components/          # Reusable UI components
│   ├── Layout.tsx      # Main app layout with sidebar
│   └── LoginForm.tsx   # Authentication form
├── contexts/           # React contexts
│   └── AuthContext.tsx # Authentication state management
├── pages/             # Route components
│   └── Dashboard.tsx  # Main dashboard page
├── services/          # API and external services
│   └── api.ts         # Backend API service
├── types/             # TypeScript type definitions
│   └── index.ts       # Shared interfaces and types
├── App.tsx            # Main app component
└── main.tsx           # App entry point
```

## Features Overview

### Dashboard
- System health monitoring
- Quick stats (instances, databases, backups)
- Recent backup activity
- Quick action buttons

### Authentication
- Secure API key-based authentication
- Connection testing before login
- Persistent session storage
- Automatic reconnection

### Responsive Design
- Mobile-first approach
- Collapsible sidebar on mobile
- Touch-friendly interactions
- Optimized for all screen sizes

## Development Notes

### State Management
- Uses React Query for server state
- React Context for authentication state
- Local component state with useState/useReducer

### Styling
- Tailwind CSS with custom color palette
- Consistent spacing and typography
- Dark mode ready (can be enabled)
- Component-based utility classes

### Type Safety
- Full TypeScript coverage
- Strict type checking
- API response typing
- Form validation typing

## Contributing

1. Follow the existing code style
2. Use TypeScript for all new files
3. Add proper type definitions
4. Test responsive design
5. Update this README if needed

## License

This project is part of the PostgreSQL Backup Manager system.
