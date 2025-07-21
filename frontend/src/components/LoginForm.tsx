import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { 
  Database, 
  Key, 
  AlertCircle, 
  Loader2 
} from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { useToast } from '../hooks/useToast';

interface LoginFormData {
  baseUrl: string;
  apiKey: string;
}

export const LoginForm: React.FC = () => {
  const [isLoading, setIsLoading] = useState(false);
  const [showApiKey, setShowApiKey] = useState(false);
  const { login, state } = useAuth();
  const { success, error } = useToast();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    defaultValues: {
      baseUrl: 'http://localhost:8080',
      apiKey: '',
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    setIsLoading(true);
    try {
      // Remove trailing slash from URL
      const baseUrl = data.baseUrl.replace(/\/+$/, '');
      
      await login({ baseUrl, apiKey: data.apiKey });
      success('Login Successful', 'You have been successfully connected to the backup service.');
    } catch (err) {
      error('Login Failed', err instanceof Error ? err.message : 'Authentication failed. Please check your credentials.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 to-primary-100 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div className="text-center">
          <div className="mx-auto h-16 w-16 bg-primary-600 rounded-full flex items-center justify-center">
            <Database className="h-8 w-8 text-white" />
          </div>
          <h2 className="mt-6 text-3xl font-bold text-gray-900">
            PostgreSQL Backup Manager
          </h2>
          <p className="mt-2 text-sm text-gray-600">
            Enter your service URL and API key to continue
          </p>
        </div>
        
        <form className="mt-8 space-y-6" onSubmit={handleSubmit(onSubmit)}>
          <div className="space-y-4">
            <div>
              <label htmlFor="baseUrl" className="block text-sm font-medium text-gray-700">
                Service URL
              </label>
              <div className="mt-1 relative">
                <input
                  {...register('baseUrl', {
                    required: 'Service URL is required',
                    pattern: {
                      value: /^https?:\/\/.+/,
                      message: 'Please enter a valid URL (http:// or https://)',
                    },
                  })}
                  type="url"
                  className="input-field"
                  placeholder="http://localhost:8080"
                />
                {errors.baseUrl && (
                  <div className="mt-1 flex items-center text-sm text-danger-600">
                    <AlertCircle className="h-4 w-4 mr-1" />
                    {errors.baseUrl.message}
                  </div>
                )}
              </div>
            </div>

            <div>
              <label htmlFor="apiKey" className="block text-sm font-medium text-gray-700">
                API Key
              </label>
              <div className="mt-1 relative">
                <input
                  {...register('apiKey', {
                    required: 'API key is required',
                    minLength: {
                      value: 10,
                      message: 'API key must be at least 10 characters',
                    },
                  })}
                  type={showApiKey ? 'text' : 'password'}
                  className="input-field pr-10"
                  placeholder="Enter your API key"
                />
                <button
                  type="button"
                  className="absolute inset-y-0 right-0 pr-3 flex items-center"
                  onClick={() => setShowApiKey(!showApiKey)}
                >
                  <Key className="h-4 w-4 text-gray-400" />
                </button>
                {errors.apiKey && (
                  <div className="mt-1 flex items-center text-sm text-danger-600">
                    <AlertCircle className="h-4 w-4 mr-1" />
                    {errors.apiKey.message}
                  </div>
                )}
              </div>
            </div>
          </div>

          {(state.error || errors.root) && (
            <div className="bg-danger-50 border border-danger-200 rounded-md p-4">
              <div className="flex">
                <AlertCircle className="h-5 w-5 text-danger-400" />
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-danger-800">
                    Authentication Error
                  </h3>
                  <div className="mt-2 text-sm text-danger-700">
                    {state.error || errors.root?.message}
                  </div>
                </div>
              </div>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={isLoading || state.isLoading}
              className="btn-primary w-full flex justify-center items-center"
            >
              {(isLoading || state.isLoading) ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Connecting...
                </>
              ) : (
                'Connect to Service'
              )}
            </button>
          </div>
        </form>
        
        <div className="text-center">
          <p className="text-xs text-gray-500">
            Make sure your backup service is running and accessible
          </p>
        </div>
      </div>
    </div>
  );
}; 