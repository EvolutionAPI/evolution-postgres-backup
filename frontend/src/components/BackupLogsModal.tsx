import React, { useState, useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { 
  X,
  FileText, 
  Download, 
  Pause, 
  Play,
  AlertCircle,
  Info,
  AlertTriangle,
  XCircle,
  Loader2,
  RefreshCw
} from 'lucide-react';
import { apiService } from '../services/api';
import { format } from 'date-fns';
import type { LogEntry, BackupJob } from '../types';

interface BackupLogsModalProps {
  backup: BackupJob;
  isOpen: boolean;
  onClose: () => void;
}

export const BackupLogsModal: React.FC<BackupLogsModalProps> = ({ backup, isOpen, onClose }) => {
  const [isLiveMode, setIsLiveMode] = useState(true);
  const [liveLogs, setLiveLogs] = useState<LogEntry[]>([]);
  const [eventSource, setEventSource] = useState<EventSource | null>(null);
  const logsEndRef = useRef<HTMLDivElement>(null);

  // Use job_id from backup if available, otherwise extract from backup ID (fallback for old backups)
  const shortJobId = backup.job_id || backup.id.split('-')[0];

  const { data: staticLogs = [], isLoading, refetch } = useQuery({
    queryKey: ['backup-logs', backup.id],
    queryFn: () => apiService.getLogsByBackupId(backup.id),
    enabled: isOpen, // Always load when modal is open
  });

  // Live logs subscription
  useEffect(() => {
    if (isLiveMode && isOpen) {
      const source = apiService.subscribeToLogs(
        (logEntry: LogEntry) => {
          // Only show logs for this specific backup using short job ID
          if (logEntry.job_id === shortJobId) {
            setLiveLogs(prev => {
              const newLogs = [...prev, logEntry].slice(-200); // Keep last 200 logs
              return newLogs;
            });
          }
        },
        (error) => {
          console.error('EventSource error in backup modal:', error);
          setIsLiveMode(false);
        }
      );
      
      if (source) {
        setEventSource(source);
      }

      return () => {
        if (source) {
          source.close();
        }
      };
    } else {
      if (eventSource) {
        eventSource.close();
        setEventSource(null);
      }
    }
  }, [isLiveMode, isOpen, shortJobId]);

  // Auto-scroll to bottom when logs update
  useEffect(() => {
    if (staticLogs.length > 0 || liveLogs.length > 0) {
      setTimeout(() => {
        logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
      }, 100);
    }
  }, [staticLogs.length, liveLogs.length]);

  // Reset logs when modal opens/closes
  useEffect(() => {
    if (isOpen) {
      setLiveLogs([]);
    } else {
      if (eventSource) {
        eventSource.close();
        setEventSource(null);
      }
    }
  }, [isOpen]);
  
  // Handle ESC key to close modal
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEsc);
      return () => document.removeEventListener('keydown', handleEsc);
    }
  }, [isOpen, onClose]);

  const toggleLiveMode = () => {
    setIsLiveMode(!isLiveMode);
    if (!isLiveMode) {
      setLiveLogs([]);
    }
  };

  const getLogIcon = (level: string) => {
    switch (level) {
      case 'ERROR':
        return <XCircle className="h-4 w-4 text-danger-500" />;
      case 'WARN':
        return <AlertTriangle className="h-4 w-4 text-warning-500" />;
      case 'INFO':
        return <Info className="h-4 w-4 text-primary-500" />;
      case 'DEBUG':
        return <AlertCircle className="h-4 w-4 text-gray-500" />;
      default:
        return <FileText className="h-4 w-4 text-gray-500" />;
    }
  };

  const formatLogMessage = (message: string) => {
    // Highlight job IDs in brackets
    return message.replace(/\[([^\]]+)\]/g, '<span class="font-mono text-primary-600 bg-primary-100 px-1 rounded">[$1]</span>');
  };

  const currentLogs = isLiveMode ? 
    [...(Array.isArray(staticLogs) ? staticLogs : []), ...liveLogs]
      .filter(log => log.job_id === shortJobId) // Use short job ID for filtering
      .sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()) : 
    (Array.isArray(staticLogs) ? staticLogs : []);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 overflow-hidden">
      <div className="absolute inset-0 bg-gray-500 bg-opacity-75" onClick={onClose} />
      
      <div className="absolute inset-4 bg-white rounded-lg shadow-xl flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-gray-200">
          <div className="flex items-center space-x-3">
            <FileText className="h-6 w-6 text-primary-600" />
            <div>
              <h2 className="text-lg font-medium text-gray-900">
                Backup Logs: {backup.database_name}
              </h2>
              <p className="text-sm text-gray-500">
                Job ID: {backup.id} • {format(new Date(backup.created_at), 'MMM d, yyyy HH:mm')}
              </p>
            </div>
          </div>
          
          <div className="flex items-center space-x-3">
            <button
              onClick={toggleLiveMode}
              className={`flex items-center px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                isLiveMode 
                  ? 'bg-success-100 text-success-700 border border-success-200'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {isLiveMode ? (
                <>
                  <Pause className="h-4 w-4 mr-1" />
                  Live
                </>
              ) : (
                <>
                  <Play className="h-4 w-4 mr-1" />
                  Static
                </>
              )}
            </button>
            
            {!isLiveMode && (
              <button
                onClick={() => refetch()}
                disabled={isLoading}
                className="btn-secondary flex items-center text-sm"
              >
                <RefreshCw className={`h-4 w-4 mr-1 ${isLoading ? 'animate-spin' : ''}`} />
                Refresh
              </button>
            )}
            
            <button
              onClick={() => {
                const logsText = currentLogs.map(log => 
                  `${log.timestamp} [${log.level}] ${log.message}`
                ).join('\n');
                
                const blob = new Blob([logsText], { type: 'text/plain' });
                const url = URL.createObjectURL(blob);
                const a = document.createElement('a');
                a.href = url;
                a.download = `backup-${backup.id}-logs-${format(new Date(), 'yyyy-MM-dd-HH-mm')}.txt`;
                a.click();
                URL.revokeObjectURL(url);
              }}
              className="btn-secondary flex items-center text-sm"
            >
              <Download className="h-4 w-4 mr-1" />
              Export
            </button>
            
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 p-1"
            >
              <X className="h-6 w-6" />
            </button>
          </div>
        </div>

        {/* Status and Info */}
        <div className="p-4 bg-gray-50 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4 text-sm">
              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                backup.status === 'completed' ? 'text-success-700 bg-success-50 border-success-200' :
                backup.status === 'failed' ? 'text-danger-700 bg-danger-50 border-danger-200' :
                backup.status === 'in_progress' ? 'text-warning-700 bg-warning-50 border-warning-200' :
                'text-gray-700 bg-gray-50 border-gray-200'
              }`}>
                {backup.status.replace('_', ' ')}
              </span>
              
              <span className="text-gray-600">
                Showing {currentLogs.length} log entries
                {isLiveMode && (
                  <span className="text-xs text-gray-500 ml-2">
                    (Historical: {staticLogs.length} + Live: {liveLogs.filter(log => log.job_id === shortJobId).length})
                  </span>
                )}
              </span>
              
              {isLiveMode && eventSource && (
                <span className="flex items-center text-success-600">
                  <div className="w-2 h-2 bg-success-500 rounded-full animate-pulse mr-1"></div>
                  Live streaming
                </span>
              )}
            </div>
            
            <div className="text-sm text-gray-500">
              {backup.backup_type} backup • {backup.postgresql_id}
            </div>
          </div>
        </div>

        {/* Logs Display */}
        <div className="flex-1 overflow-hidden">
          <div className="h-full overflow-y-auto bg-gray-900 text-gray-100 font-mono text-sm">
            {isLoading ? (
              <div className="flex items-center justify-center h-full">
                <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
              </div>
            ) : currentLogs.length === 0 ? (
              <div className="flex items-center justify-center h-full text-gray-400">
                <div className="text-center">
                  <FileText className="h-12 w-12 mx-auto mb-4" />
                  <p className="text-lg">No logs found for this backup</p>
                  <p className="text-sm mt-2">
                    {isLiveMode 
                      ? 'Waiting for new logs...' 
                      : 'Try switching to live mode or refresh'
                    }
                  </p>
                </div>
              </div>
            ) : (
              <div className="p-4 space-y-1">
                {currentLogs.map((log, index) => (
                  <div
                    key={`${log.timestamp}-${index}`}
                    className="flex items-start space-x-3 p-2 rounded border-l-4 border-l-primary-500 bg-gray-800 bg-opacity-50"
                  >
                    <div className="flex-shrink-0 mt-0.5">
                      {getLogIcon(log.level)}
                    </div>
                    
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-2 text-xs text-gray-400 mb-1">
                        <span>{format(new Date(log.timestamp), 'HH:mm:ss.SSS')}</span>
                        <span className={`px-1.5 py-0.5 rounded text-xs font-medium ${
                          log.level === 'ERROR' ? 'bg-danger-900 text-danger-200' :
                          log.level === 'WARN' ? 'bg-warning-900 text-warning-200' :
                          log.level === 'INFO' ? 'bg-primary-900 text-primary-200' :
                          'bg-gray-800 text-gray-300'
                        }`}>
                          {log.level}
                        </span>
                        <span className="bg-primary-900 text-primary-200 px-1.5 py-0.5 rounded text-xs font-medium">
                          {backup.id}
                        </span>
                      </div>
                      <div 
                        className="text-gray-100 break-words"
                        dangerouslySetInnerHTML={{ __html: formatLogMessage(log.message) }}
                      />
                    </div>
                  </div>
                ))}
                <div ref={logsEndRef} />
              </div>
            )}
          </div>
        </div>

        {/* Footer */}
        <div className="p-4 bg-gray-50 border-t border-gray-200">
          <div className="flex items-center justify-between text-xs text-gray-500">
            <div>
              Use the live mode to see real-time logs as the backup progresses
            </div>
            <div>
              Press ESC to close or click outside
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}; 