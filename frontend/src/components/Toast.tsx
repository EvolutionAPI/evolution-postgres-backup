import React, { useEffect, useState } from 'react';
import { 
  CheckCircle, 
  XCircle, 
  AlertTriangle, 
  Info, 
  X 
} from 'lucide-react';
import type { Toast as ToastType } from '../types';

interface ToastProps {
  toast: ToastType;
  onRemove: (id: string) => void;
}

export const Toast: React.FC<ToastProps> = ({ toast, onRemove }) => {
  const [isVisible, setIsVisible] = useState(false);
  const [isLeaving, setIsLeaving] = useState(false);

  useEffect(() => {
    // Trigger entrance animation
    setTimeout(() => setIsVisible(true), 10);
  }, []);

  const handleRemove = () => {
    setIsLeaving(true);
    setTimeout(() => onRemove(toast.id), 300);
  };

  const getIcon = () => {
    switch (toast.type) {
      case 'success':
        return <CheckCircle className="h-5 w-5 text-success-400" />;
      case 'error':
        return <XCircle className="h-5 w-5 text-danger-400" />;
      case 'warning':
        return <AlertTriangle className="h-5 w-5 text-warning-400" />;
      case 'info':
        return <Info className="h-5 w-5 text-primary-400" />;
      default:
        return <Info className="h-5 w-5 text-gray-400" />;
    }
  };

  const getBackgroundColor = () => {
    switch (toast.type) {
      case 'success':
        return 'bg-success-50 border-success-200';
      case 'error':
        return 'bg-danger-50 border-danger-200';
      case 'warning':
        return 'bg-warning-50 border-warning-200';
      case 'info':
        return 'bg-primary-50 border-primary-200';
      default:
        return 'bg-gray-50 border-gray-200';
    }
  };

  const getTitleColor = () => {
    switch (toast.type) {
      case 'success':
        return 'text-success-800';
      case 'error':
        return 'text-danger-800';
      case 'warning':
        return 'text-warning-800';
      case 'info':
        return 'text-primary-800';
      default:
        return 'text-gray-800';
    }
  };

  const getMessageColor = () => {
    switch (toast.type) {
      case 'success':
        return 'text-success-700';
      case 'error':
        return 'text-danger-700';
      case 'warning':
        return 'text-warning-700';
      case 'info':
        return 'text-primary-700';
      default:
        return 'text-gray-700';
    }
  };

  return (
    <div
      className={`
        transform transition-all duration-300 ease-in-out
        ${isVisible && !isLeaving 
          ? 'translate-x-0 opacity-100' 
          : 'translate-x-full opacity-0'
        }
        ${isLeaving ? 'scale-95' : 'scale-100'}
        max-w-sm w-full shadow-lg rounded-lg pointer-events-auto border
        ${getBackgroundColor()}
      `}
    >
      <div className="p-4">
        <div className="flex items-start">
          <div className="flex-shrink-0">
            {getIcon()}
          </div>
          <div className="ml-3 w-0 flex-1">
            <p className={`text-sm font-medium ${getTitleColor()}`}>
              {toast.title}
            </p>
            {toast.message && (
              <p className={`mt-1 text-sm ${getMessageColor()}`}>
                {toast.message}
              </p>
            )}
          </div>
          {toast.dismissible && (
            <div className="ml-4 flex-shrink-0 flex">
              <button
                className={`
                  inline-flex rounded-md p-1.5 focus:outline-none focus:ring-2 focus:ring-offset-2
                  ${toast.type === 'success' ? 'text-success-500 hover:bg-success-100 focus:ring-success-600' : ''}
                  ${toast.type === 'error' ? 'text-danger-500 hover:bg-danger-100 focus:ring-danger-600' : ''}
                  ${toast.type === 'warning' ? 'text-warning-500 hover:bg-warning-100 focus:ring-warning-600' : ''}
                  ${toast.type === 'info' ? 'text-primary-500 hover:bg-primary-100 focus:ring-primary-600' : ''}
                `}
                onClick={handleRemove}
              >
                <span className="sr-only">Dismiss</span>
                <X className="h-4 w-4" />
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}; 