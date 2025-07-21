import React, { useReducer, useEffect } from 'react';
import { apiService } from '../services/api';
import { AuthContext, authReducer, initialState } from './auth-context';
import type { User, ApiConfig } from '../types';

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [state, dispatch] = useReducer(authReducer, initialState);

  // Load saved auth on mount
  useEffect(() => {
    const savedAuth = localStorage.getItem('backup-auth');
    if (savedAuth) {
      try {
        const config: ApiConfig = JSON.parse(savedAuth);
        apiService.initialize(config);
        
        // Verify connection
        apiService.testConnection().then(isValid => {
          if (isValid) {
            const user: User = {
              apiKey: config.apiKey,
              baseUrl: config.baseUrl,
              isAuthenticated: true,
            };
            dispatch({ type: 'LOGIN_SUCCESS', payload: user });
          } else {
            localStorage.removeItem('backup-auth');
          }
        });
      } catch {
        localStorage.removeItem('backup-auth');
      }
    }
  }, []);

  const login = async (config: ApiConfig) => {
    dispatch({ type: 'LOGIN_START' });
    
    try {
      // Initialize API service
      apiService.initialize(config);
      
      // Test connection
      const isConnected = await apiService.testConnection();
      
      if (!isConnected) {
        throw new Error('Failed to connect to the backup service. Please check your URL and API key.');
      }

      const user: User = {
        apiKey: config.apiKey,
        baseUrl: config.baseUrl,
        isAuthenticated: true,
      };

      // Save to localStorage
      localStorage.setItem('backup-auth', JSON.stringify(config));
      
      dispatch({ type: 'LOGIN_SUCCESS', payload: user });
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Authentication failed';
      dispatch({ type: 'LOGIN_ERROR', payload: errorMessage });
      throw error;
    }
  };

  const logout = () => {
    localStorage.removeItem('backup-auth');
    dispatch({ type: 'LOGOUT' });
  };

  const clearError = () => {
    dispatch({ type: 'CLEAR_ERROR' });
  };

  return (
    <AuthContext.Provider 
      value={{ 
        state, 
        login, 
        logout, 
        clearError 
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}; 