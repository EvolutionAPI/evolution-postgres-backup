import React, { useState, useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { 
  FileText, 
  Search, 
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
import { useToast } from '../hooks/useToast';
import { format } from 'date-fns';
import type { LogEntry, LogFile } from '../types';

export const LogsPage: React.FC = () => {
  const [isLiveMode, setIsLiveMode] = useState(true);
  const [filter, setFilter] = useState<string>('');
  const [levelFilter, setLevelFilter] = useState<'all' | 'DEBUG' | 'INFO' | 'WARN' | 'ERROR'>('all');
  const [jobIdFilter, setJobIdFilter] = useState<string>('');
  const [selectedDate, setSelectedDate] = useState<string>(new Date().toISOString().split('T')[0]);
  const [liveLogs, setLiveLogs] = useState<LogEntry[]>([]);
  const [eventSource, setEventSource] = useState<EventSource | null>(null);
  const logsEndRef = useRef<HTMLDivElement>(null);
  const { info, success } = useToast();

  const { data: staticLogs = [], isLoading, refetch } = useQuery({
    queryKey: ['logs', levelFilter, jobIdFilter, selectedDate],
    queryFn: () => {
      console.log('🔍 Loading logs for date:', selectedDate);
      return apiService.getLogs({
        level: levelFilter === 'all' ? undefined : levelFilter,
        job_id: jobIdFilter || undefined,
        limit: 100,
        date: selectedDate,
      });
    },
    enabled: true, // Always load historical logs
  });

  // Debug logs and show toast when logs are loaded
  useEffect(() => {
    if (staticLogs.length > 0 && !isLoading) {
      console.log(`✅ Loaded ${staticLogs.length} historical logs for ${selectedDate}`);
      success('Historical Logs Loaded', `Found ${staticLogs.length} log entries from ${selectedDate}`);
    } else if (!isLoading && staticLogs.length === 0) {
      console.log(`❌ No historical logs found for ${selectedDate}`);
      info('No Historical Logs', `No logs found for ${selectedDate}. Check if there were any backup operations on this date.`);
    }
  }, [staticLogs.length, selectedDate, isLoading, success, info]); // Added back success and info

  const { data: logFiles = [] } = useQuery({
    queryKey: ['log-files'],
    queryFn: () => apiService.getLogFiles(),
  });

  // Live logs subscription
  useEffect(() => {
    if (isLiveMode) {
      const source = apiService.subscribeToLogs(
        (logEntry: LogEntry) => {
          setLiveLogs(prev => {
            const newLogs = [...prev, logEntry].slice(-200); // Keep last 200 logs
            return newLogs;
          });
        },
        (error) => {
          console.error('EventSource error:', error);
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
  }, [isLiveMode]);

  // Auto-scroll to bottom in live mode
  useEffect(() => {
    if (isLiveMode && liveLogs.length > 0) {
      logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
  }, [liveLogs, isLiveMode]);

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

  const getLogColor = (level: string) => {
    switch (level) {
      case 'ERROR':
        return 'border-l-danger-500 bg-danger-50';
      case 'WARN':
        return 'border-l-warning-500 bg-warning-50';
      case 'INFO':
        return 'border-l-primary-500 bg-primary-50';
      case 'DEBUG':
        return 'border-l-gray-500 bg-gray-50';
      default:
        return 'border-l-gray-300 bg-white';
    }
  };

  const formatLogMessage = (message: string) => {
    // Highlight job IDs in brackets
    return message.replace(/\[([^\]]+)\]/g, '<span class="font-mono text-primary-600 bg-primary-100 px-1 rounded">[$1]</span>');
  };

  const filteredLiveLogs = liveLogs.filter(log => {
    const matchesLevel = levelFilter === 'all' || log.level === levelFilter;
    const matchesJobId = !jobIdFilter || (log.job_id && log.job_id.includes(jobIdFilter));
    const matchesFilter = !filter || log.message.toLowerCase().includes(filter.toLowerCase());
    return matchesLevel && matchesJobId && matchesFilter;
  });

  const filteredStaticLogs = staticLogs.filter(log => {
    const matchesLevel = levelFilter === 'all' || log.level === levelFilter;
    const matchesJobId = !jobIdFilter || (log.job_id && log.job_id.includes(jobIdFilter));
    const matchesFilter = !filter || log.message.toLowerCase().includes(filter.toLowerCase());
    return matchesLevel && matchesJobId && matchesFilter;
  });

  const currentLogs = isLiveMode ? 
    [...filteredStaticLogs, ...filteredLiveLogs].sort((a, b) => 
      new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    ) : 
    filteredStaticLogs;

  // Auto-scroll to bottom when logs update
  useEffect(() => {
    if (currentLogs.length > 0) {
      setTimeout(() => {
        logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
      }, 100);
    }
  }, [currentLogs.length]);

  // Debug the filtering
  useEffect(() => {
    console.log(`🔧 Logs: ${staticLogs.length} historical + ${liveLogs.length} live = ${currentLogs.length} total (mode: ${isLiveMode ? 'Live' : 'Static'})`);
    
    if (staticLogs.length > 0 && filteredStaticLogs.length === 0) {
      console.log('🚨 All historical logs were filtered out! Check your filters.');
    }
  }, [staticLogs.length, liveLogs.length, currentLogs.length, isLiveMode, filteredStaticLogs.length]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">System Logs</h1>
          <p className="mt-1 text-sm text-gray-500">
            Monitor backup operations and system activities
          </p>
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
                Live Mode
              </>
            ) : (
              <>
                <Play className="h-4 w-4 mr-1" />
                Static Mode
              </>
            )}
          </button>
          
          {!isLiveMode && (
            <button
              onClick={() => refetch()}
              disabled={isLoading}
              className="btn-secondary flex items-center"
            >
              <RefreshCw className={`h-4 w-4 mr-1 ${isLoading ? 'animate-spin' : ''}`} />
              Refresh Logs
            </button>
          )}
          
          <button
            onClick={() => {
              console.log('🐛 Debug Summary:');
              console.log(`- Mode: ${isLiveMode ? 'Live' : 'Static'}`);
              console.log(`- Date: ${selectedDate}`);
              console.log(`- Historical logs: ${staticLogs.length} (filtered: ${filteredStaticLogs.length})`);
              console.log(`- Live logs: ${liveLogs.length} (filtered: ${filteredLiveLogs.length})`);
              console.log(`- Total showing: ${currentLogs.length}`);
              console.log(`- Filters: level=${levelFilter}, jobId="${jobIdFilter}", text="${filter}"`);
              if (staticLogs.length > 0) {
                console.log('- Sample log:', staticLogs[0]);
              }
            }}
            className="btn-secondary flex items-center"
          >
            <Info className="h-4 w-4 mr-1" />
            Debug Info
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="card">
        <div className="grid grid-cols-1 md:grid-cols-5 gap-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Search</label>
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
              <input
                type="text"
                value={filter}
                onChange={(e) => setFilter(e.target.value)}
                className="input-field pl-10"
                placeholder="Search logs..."
              />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Date</label>
            <input
              type="date"
              value={selectedDate}
              onChange={(e) => setSelectedDate(e.target.value)}
              className="input-field"
              disabled={isLiveMode}
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Level</label>
            <select
              value={levelFilter}
              onChange={(e) => setLevelFilter(e.target.value as typeof levelFilter)}
              className="input-field"
            >
              <option value="all">All Levels</option>
              <option value="DEBUG">Debug</option>
              <option value="INFO">Info</option>
              <option value="WARN">Warning</option>
              <option value="ERROR">Error</option>
            </select>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Job ID</label>
            <input
              type="text"
              value={jobIdFilter}
              onChange={(e) => setJobIdFilter(e.target.value)}
              className="input-field"
              placeholder="Filter by job ID..."
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Log Files</label>
            <select className="input-field">
              <option value="">Current Session</option>
              {logFiles.map((file: LogFile) => (
                <option key={file.name} value={file.name}>
                  {file.date} - {file.name} ({(file.size / 1024).toFixed(1)} KB)
                </option>
              ))}
            </select>
          </div>
        </div>

        <div className="mt-4 flex items-center justify-between">
          <div className="flex items-center space-x-4 text-sm text-gray-600">
            <span>Showing {currentLogs.length} log entries</span>
            <span className="text-xs text-gray-500">
              Historical: {filteredStaticLogs.length} | Live: {filteredLiveLogs.length} | Total Raw: {staticLogs.length} | Mode: {isLiveMode ? 'Live' : 'Static'}
            </span>
            {levelFilter !== 'all' && (
              <span className="text-xs bg-primary-100 text-primary-700 px-2 py-1 rounded">
                Level: {levelFilter}
              </span>
            )}
            {jobIdFilter && (
              <span className="text-xs bg-success-100 text-success-700 px-2 py-1 rounded">
                Job: {jobIdFilter}
              </span>
            )}
            {isLiveMode && eventSource && (
              <span className="flex items-center text-success-600">
                <div className="w-2 h-2 bg-success-500 rounded-full animate-pulse mr-1"></div>
                Live streaming
              </span>
            )}
          </div>
          
          <button
            className="btn-secondary flex items-center text-sm"
            onClick={() => {
              const logsText = currentLogs.map(log => 
                `${log.timestamp} [${log.level}] ${log.job_id ? `[${log.job_id}] ` : ''}${log.message}`
              ).join('\n');
              
              const blob = new Blob([logsText], { type: 'text/plain' });
              const url = URL.createObjectURL(blob);
              const a = document.createElement('a');
              a.href = url;
              a.download = `backup-logs-${format(new Date(), 'yyyy-MM-dd-HH-mm')}.txt`;
              a.click();
              URL.revokeObjectURL(url);
            }}
          >
            <Download className="h-4 w-4 mr-1" />
            Export
          </button>
        </div>
      </div>

      {/* Logs Display */}
      <div className="card p-0 overflow-hidden">
        <div className="h-[600px] overflow-y-auto bg-gray-900 text-gray-100 font-mono text-sm">
          {isLoading ? (
            <div className="flex items-center justify-center h-full">
              <Loader2 className="h-6 w-6 animate-spin text-gray-400" />
            </div>
          ) : currentLogs.length === 0 ? (
            <div className="flex items-center justify-center h-full text-gray-400">
              <div className="text-center">
                <FileText className="h-8 w-8 mx-auto mb-2" />
                <p>No logs found</p>
                <p className="text-xs mt-1">Try adjusting your filters or wait for new logs</p>
              </div>
            </div>
          ) : (
            <div className="p-4 space-y-1">
              {currentLogs.map((log, index) => (
                <div
                  key={`${log.timestamp}-${index}`}
                  className={`flex items-start space-x-3 p-2 rounded border-l-4 ${getLogColor(log.level)} bg-opacity-10`}
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
                      {log.job_id && (
                        <span className="bg-primary-900 text-primary-200 px-1.5 py-0.5 rounded text-xs font-medium">
                          {log.job_id}
                        </span>
                      )}
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
    </div>
  );
}; 