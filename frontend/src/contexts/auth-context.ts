import { createContext } from 'react';
import type { User, ApiConfig } from '../types';

interface AuthState {
  user: User | null;
  isLoading: boolean;
  error: string | null;
}

type AuthAction =
  | { type: 'LOGIN_START' }
  | { type: 'LOGIN_SUCCESS'; payload: User }
  | { type: 'LOGIN_ERROR'; payload: string }
  | { type: 'LOGOUT' }
  | { type: 'CLEAR_ERROR' };

export interface AuthContextType {
  state: AuthState;
  login: (config: ApiConfig) => Promise<void>;
  logout: () => void;
  clearError: () => void;
}

export const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const authReducer = (state: AuthState, action: AuthAction): AuthState => {
  switch (action.type) {
    case 'LOGIN_START':
      return { ...state, isLoading: true, error: null };
    case 'LOGIN_SUCCESS':
      return { 
        user: action.payload, 
        isLoading: false, 
        error: null 
      };
    case 'LOGIN_ERROR':
      return { 
        user: null, 
        isLoading: false, 
        error: action.payload 
      };
    case 'LOGOUT':
      return { 
        user: null, 
        isLoading: false, 
        error: null 
      };
    case 'CLEAR_ERROR':
      return { ...state, error: null };
    default:
      return state;
  }
};

export const initialState: AuthState = {
  user: null,
  isLoading: false,
  error: null,
}; 