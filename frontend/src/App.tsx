import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { AuthProvider } from './contexts/AuthProvider';
import { ToastProvider } from './components/ToastProvider';
import { ToastContainer } from './components/ToastContainer';
import { useAuth } from './hooks/useAuth';
import { LoginForm } from './components/LoginForm';
import { Layout } from './components/Layout';
import { Dashboard } from './pages/Dashboard';
import { PostgresPage } from './pages/PostgresPage';
import { BackupsPage } from './pages/BackupsPage';
import { LogsPage } from './pages/LogsPage';
import { SettingsPage } from './pages/SettingsPage';

// Create a client
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      gcTime: 1000 * 60 * 10, // 10 minutes (renamed from cacheTime)
      retry: 1,
    },
  },
});

// Protected Route component
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { state } = useAuth();
  
  if (state.isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-2 text-sm text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }
  
  if (!state.user?.isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
};

// Public Route component  
const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { state } = useAuth();
  
  if (state.isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600 mx-auto"></div>
          <p className="mt-2 text-sm text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }
  
  if (state.user?.isAuthenticated) {
    return <Navigate to="/" replace />;
  }
  
  return <>{children}</>;
};

// All pages are now implemented

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <ToastProvider>
          <Router>
            <Routes>
              {/* Public routes */}
              <Route 
                path="/login" 
                element={
                  <PublicRoute>
                    <LoginForm />
                  </PublicRoute>
                } 
              />
              
              {/* Protected routes */}
              <Route 
                path="/" 
                element={
                  <ProtectedRoute>
                    <Layout />
                  </ProtectedRoute>
                }
              >
                <Route index element={<Dashboard />} />
                <Route path="postgres" element={<PostgresPage />} />
                <Route path="backups" element={<BackupsPage />} />
                <Route path="logs" element={<LogsPage />} />
                <Route path="settings" element={<SettingsPage />} />
              </Route>
              
              {/* Catch all route */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
            <ToastContainer />
          </Router>
        </ToastProvider>
      </AuthProvider>
    </QueryClientProvider>
  );
}

export default App;
