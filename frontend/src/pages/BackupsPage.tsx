import React, { useState, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { 
  Archive, 
  Plus, 
  Download, 
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2,
  Database,
  Calendar,
  HardDrive,
  RotateCcw,
  FileText
} from 'lucide-react';
import { apiService } from '../services/api';
import { useRefreshConfig } from '../hooks/useRefreshConfig';
import { useToast } from '../hooks/useToast';
import { formatDistanceToNow, format } from 'date-fns';
import { BackupLogsModal } from '../components/BackupLogsModal';
import type { BackupJob, BackupRequest, RestoreRequest } from '../types';

interface BackupFormData {
  postgres_id: string;
  database: string;
}

interface RestoreFormData {
  postgres_id: string;
  database: string;
}

export const BackupsPage: React.FC = () => {
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [isRestoreModalOpen, setIsRestoreModalOpen] = useState(false);
  const [isLogsModalOpen, setIsLogsModalOpen] = useState(false);
  const [selectedBackup, setSelectedBackup] = useState<BackupJob | null>(null);
  const [logsBackup, setLogsBackup] = useState<BackupJob | null>(null);
  const [filter, setFilter] = useState<'all' | 'completed' | 'failed' | 'in_progress'>('all');
  const [selectedInstanceId, setSelectedInstanceId] = useState<string>('');
  const queryClient = useQueryClient();
  const { getBackupsInterval, getInstancesInterval } = useRefreshConfig();
  const { success, error } = useToast();

  const { data: backups = [], isLoading: backupsLoading } = useQuery({
    queryKey: ['backups'],
    queryFn: () => apiService.getBackups(),
    refetchInterval: getBackupsInterval(),
    refetchIntervalInBackground: true,
  });

  const { data: instances = [] } = useQuery({
    queryKey: ['postgres-instances'],
    queryFn: () => apiService.getPostgresInstances(),
    refetchInterval: getInstancesInterval(),
    refetchIntervalInBackground: true,
  });

  const createBackupMutation = useMutation({
    mutationFn: (data: BackupRequest) => apiService.createBackup(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['backups'] });
      setIsCreateModalOpen(false);
      createBackupForm.reset();
      success('Backup Created', 'Your backup has been created successfully and is now processing.');
    },
    onError: (err) => {
      error('Backup Failed', err instanceof Error ? err.message : 'Failed to create backup. Please try again.');
    },
  });

  const restoreBackupMutation = useMutation({
    mutationFn: (data: RestoreRequest) => apiService.restoreBackup(data),
    onSuccess: () => {
      setIsRestoreModalOpen(false);
      setSelectedBackup(null);
      restoreForm.reset();
      success('Backup Restored', 'Your backup has been restored successfully.');
    },
    onError: (err) => {
      error('Restore Failed', err instanceof Error ? err.message : 'Failed to restore backup. Please try again.');
    },
  });

  const createBackupForm = useForm<BackupFormData>();
  const restoreForm = useForm<RestoreFormData>();

  const enabledInstances = instances.filter(instance => instance.enabled);

  // Clean up state when modal closes
  useEffect(() => {
    if (!isCreateModalOpen) {
      setSelectedInstanceId('');
      createBackupForm.reset();
    }
  }, [isCreateModalOpen, createBackupForm]);

  const onCreateBackup = (data: BackupFormData) => {
    createBackupMutation.mutate({
      postgresql_id: data.postgres_id,
      database_name: data.database,
      backup_type: 'manual',
    });
  };

  const onRestoreBackup = (data: RestoreFormData) => {
    if (!selectedBackup) return;
    
    restoreBackupMutation.mutate({
      postgresql_id: data.postgres_id,
      database_name: data.database,
      backup_id: selectedBackup.id,
    });
  };

  const handleRestore = (backup: BackupJob) => {
    setSelectedBackup(backup);
    restoreForm.setValue('postgres_id', backup.postgresql_id);
    restoreForm.setValue('database', backup.database_name);
    setIsRestoreModalOpen(true);
  };

  const handleViewLogs = (backup: BackupJob) => {
    setLogsBackup(backup);
    setIsLogsModalOpen(true);
  };

  const handleCloseLogsModal = () => {
    setIsLogsModalOpen(false);
    setLogsBackup(null);
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-5 w-5 text-success-500" />;
      case 'failed':
        return <XCircle className="h-5 w-5 text-danger-500" />;
      case 'in_progress':
        return <Loader2 className="h-5 w-5 text-warning-500 animate-spin" />;
      default:
        return <Clock className="h-5 w-5 text-gray-400" />;
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'text-success-700 bg-success-50 border-success-200';
      case 'failed':
        return 'text-danger-700 bg-danger-50 border-danger-200';
      case 'in_progress':
        return 'text-warning-700 bg-warning-50 border-warning-200';
      default:
        return 'text-gray-700 bg-gray-50 border-gray-200';
    }
  };

  const formatFileSize = (bytes?: number) => {
    if (!bytes) return 'Unknown';
    const units = ['B', 'KB', 'MB', 'GB', 'TB'];
    let size = bytes;
    let unitIndex = 0;
    
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }
    
    return `${size.toFixed(1)} ${units[unitIndex]}`;
  };

  const filteredBackups = backups
    .filter(backup => filter === 'all' || backup.status === filter)
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()); // Order by date desc

  const getInstanceName = (postgresId: string) => {
    const instance = instances.find(inst => inst.id === postgresId);
    return instance?.name || `Unknown Instance (${postgresId})`;
  };

  if (backupsLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <Loader2 className="h-8 w-8 animate-spin text-primary-600" />
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Backups</h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage your database backups and restore points
          </p>
        </div>
        <button
          onClick={() => setIsCreateModalOpen(true)}
          disabled={enabledInstances.length === 0}
          className="btn-primary flex items-center disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <Plus className="h-4 w-4 mr-2" />
          Create Backup
        </button>
      </div>

      {/* Filters */}
      <div className="flex items-center space-x-4">
        <span className="text-sm font-medium text-gray-700">Filter by status:</span>
        <div className="flex space-x-2">
          {[
            { key: 'all', label: 'All' },
            { key: 'completed', label: 'Completed' },
            { key: 'in_progress', label: 'In Progress' },
            { key: 'failed', label: 'Failed' },
          ].map(({ key, label }) => (
                         <button
               key={key}
               onClick={() => setFilter(key as typeof filter)}
               className={`px-3 py-1 rounded-full text-sm font-medium transition-colors ${
                filter === key
                  ? 'bg-primary-100 text-primary-800 border border-primary-200'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      {/* Backups List */}
      {filteredBackups.length === 0 ? (
        <div className="text-center py-12">
          <Archive className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-lg font-medium text-gray-900">
            {filter === 'all' ? 'No backups found' : `No ${filter.replace('_', ' ')} backups`}
          </h3>
          <p className="mt-1 text-gray-500">
            {enabledInstances.length === 0 
              ? 'Add and enable a PostgreSQL instance first to create backups.'
              : 'Create your first backup to get started.'
            }
          </p>
          {enabledInstances.length > 0 && (
            <button
              onClick={() => setIsCreateModalOpen(true)}
              className="btn-primary mt-4"
            >
              <Plus className="h-4 w-4 mr-2" />
              Create Backup
            </button>
          )}
        </div>
      ) : (
        <div className="space-y-4">
          {filteredBackups.map((backup) => (
            <div key={backup.id} className="card">
              <div className="flex items-start justify-between">
                <div className="flex items-start space-x-4">
                  <div className="flex-shrink-0">
                    {getStatusIcon(backup.status)}
                  </div>
                  
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2">
                      <h3 className="text-lg font-medium text-gray-900 truncate">
                        {backup.database_name}
                      </h3>
                      <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${getStatusColor(backup.status)}`}>
                        {backup.status.replace('_', ' ')}
                      </span>
                    </div>
                    
                    <div className="mt-1 flex items-center space-x-4 text-sm text-gray-500">
                      <div className="flex items-center">
                        <Database className="h-4 w-4 mr-1" />
                        {getInstanceName(backup.postgresql_id)}
                      </div>
                      <div className="flex items-center">
                        <Calendar className="h-4 w-4 mr-1" />
                        {format(new Date(backup.created_at), 'MMM d, yyyy HH:mm')}
                      </div>
                      {backup.file_size && (
                        <div className="flex items-center">
                          <HardDrive className="h-4 w-4 mr-1" />
                          {formatFileSize(backup.file_size)}
                        </div>
                      )}
                    </div>

                    <div className="mt-2 text-sm text-gray-600">
                      <span className="inline-flex items-center px-2 py-1 rounded text-xs font-medium bg-gray-100 text-gray-800">
                        {backup.backup_type}
                      </span>
                      <span className="ml-2 text-gray-500">
                        {formatDistanceToNow(new Date(backup.created_at), { addSuffix: true })}
                      </span>
                      {backup.completed_at && (
                        <span className="ml-2 text-gray-500">
                          • Completed {formatDistanceToNow(new Date(backup.completed_at), { addSuffix: true })}
                        </span>
                      )}
                    </div>

                    {backup.error_message && (
                      <div className="mt-2 p-2 bg-danger-50 border border-danger-200 rounded text-sm text-danger-700">
                        <AlertCircle className="h-4 w-4 inline mr-1" />
                        {backup.error_message}
                      </div>
                    )}
                  </div>
                </div>

                <div className="flex items-center space-x-2">
                  <button
                    onClick={() => handleViewLogs(backup)}
                    className="p-2 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded"
                    title="View Logs"
                  >
                    <FileText className="h-4 w-4" />
                  </button>

                  {backup.status === 'completed' && (
                    <button
                      onClick={() => handleRestore(backup)}
                      className="p-2 text-gray-400 hover:text-primary-600 hover:bg-primary-50 rounded"
                      title="Restore Backup"
                    >
                      <RotateCcw className="h-4 w-4" />
                    </button>
                  )}
                  
                  {backup.s3_key && backup.status === 'completed' && (
                    <button
                      className="p-2 text-gray-400 hover:text-success-600 hover:bg-success-50 rounded"
                      title="Download Backup"
                      onClick={() => {
                        // Note: This would require implementing a download endpoint
                        alert('Download functionality would be implemented with a download endpoint');
                      }}
                    >
                      <Download className="h-4 w-4" />
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create Backup Modal */}
      {isCreateModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" onClick={() => setIsCreateModalOpen(false)} />
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <form onSubmit={createBackupForm.handleSubmit(onCreateBackup)}>
                <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                  <div className="flex items-center mb-4">
                    <Archive className="h-6 w-6 text-primary-600 mr-2" />
                    <h3 className="text-lg font-medium text-gray-900">Create Manual Backup</h3>
                  </div>

                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700">PostgreSQL Instance</label>
                      <select
                        {...createBackupForm.register('postgres_id', { required: 'Instance is required' })}
                        className="input-field mt-1"
                        onChange={(e) => {
                          const value = e.target.value;
                          createBackupForm.setValue('postgres_id', value); // Ensure form value is set
                          
                          const selectedInstance = enabledInstances.find(inst => inst.id === value);
                          if (selectedInstance) {
                            setSelectedInstanceId(value);
                            if (selectedInstance.databases.length === 1) {
                              createBackupForm.setValue('database', selectedInstance.databases[0]);
                            } else {
                              createBackupForm.setValue('database', '');
                            }
                          } else {
                            setSelectedInstanceId('');
                            createBackupForm.setValue('database', '');
                          }
                        }}
                      >
                        <option value="">Select instance...</option>
                        {enabledInstances.map((instance) => (
                          <option key={instance.id} value={instance.id}>
                            {instance.name} ({instance.host}:{instance.port})
                          </option>
                        ))}
                      </select>
                      {createBackupForm.formState.errors.postgres_id && (
                        <div className="mt-1 flex items-center text-sm text-danger-600">
                          <AlertCircle className="h-4 w-4 mr-1" />
                          {createBackupForm.formState.errors.postgres_id.message}
                        </div>
                      )}
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">Database</label>
                      <select
                        {...createBackupForm.register('database', { required: 'Database is required' })}
                        className="input-field mt-1"
                      >
                        <option value="">Select database...</option>
                        {selectedInstanceId && 
                          enabledInstances
                            .find(inst => inst.id === selectedInstanceId)
                            ?.databases.map((db) => (
                              <option key={db} value={db}>{db}</option>
                            ))
                        }
                      </select>
                      {selectedInstanceId && (
                        <p className="mt-1 text-xs text-gray-500">
                          Available databases: {
                            enabledInstances
                              .find(inst => inst.id === selectedInstanceId)
                              ?.databases.length || 0
                          } found
                        </p>
                      )}
                      {!selectedInstanceId && (
                        <p className="mt-1 text-xs text-gray-500">
                          Please select a PostgreSQL instance first
                        </p>
                      )}
                      {createBackupForm.formState.errors.database && (
                        <div className="mt-1 flex items-center text-sm text-danger-600">
                          <AlertCircle className="h-4 w-4 mr-1" />
                          {createBackupForm.formState.errors.database.message}
                        </div>
                      )}
                    </div>
                  </div>
                </div>

                <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                  <button
                    type="submit"
                    disabled={createBackupMutation.isPending}
                    className="btn-primary w-full sm:w-auto sm:ml-3"
                  >
                    {createBackupMutation.isPending && (
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    )}
                    Create Backup
                  </button>
                  <button
                    type="button"
                    onClick={() => setIsCreateModalOpen(false)}
                    className="btn-secondary w-full sm:w-auto mt-3 sm:mt-0"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Restore Modal */}
      {isRestoreModalOpen && selectedBackup && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" onClick={() => setIsRestoreModalOpen(false)} />
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <form onSubmit={restoreForm.handleSubmit(onRestoreBackup)}>
                <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                  <div className="flex items-center mb-4">
                    <RotateCcw className="h-6 w-6 text-warning-600 mr-2" />
                    <h3 className="text-lg font-medium text-gray-900">Restore Backup</h3>
                  </div>

                  <div className="mb-4 p-4 bg-yellow-50 border border-yellow-200 rounded">
                    <div className="flex">
                      <AlertCircle className="h-5 w-5 text-yellow-400" />
                      <div className="ml-3">
                        <h3 className="text-sm font-medium text-yellow-800">Warning</h3>
                        <div className="mt-2 text-sm text-yellow-700">
                          This will restore the backup and may overwrite existing data. Please ensure you have a recent backup before proceeding.
                        </div>
                      </div>
                    </div>
                  </div>

                  <div className="space-y-4">
                    <div className="bg-gray-50 p-3 rounded">
                      <h4 className="text-sm font-medium text-gray-900">Backup Details</h4>
                      <div className="mt-1 text-sm text-gray-600">
                        <p>Database: <span className="font-medium">{selectedBackup.database_name}</span></p>
                        <p>Created: <span className="font-medium">{format(new Date(selectedBackup.created_at), 'MMM d, yyyy HH:mm')}</span></p>
                        {selectedBackup.file_size && (
                          <p>Size: <span className="font-medium">{formatFileSize(selectedBackup.file_size)}</span></p>
                        )}
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">Restore to PostgreSQL Instance</label>
                      <select
                        {...restoreForm.register('postgres_id', { required: 'Instance is required' })}
                        className="input-field mt-1"
                      >
                        <option value="">Select instance...</option>
                        {enabledInstances.map((instance) => (
                          <option key={instance.id} value={instance.id}>
                            {instance.name} ({instance.host}:{instance.port})
                          </option>
                        ))}
                      </select>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">Restore to Database</label>
                      <input
                        {...restoreForm.register('database', { required: 'Database is required' })}
                        className="input-field mt-1"
                        placeholder="database_name"
                      />
                      <p className="mt-1 text-xs text-gray-500">
                        Database will be created if it doesn't exist
                      </p>
                    </div>
                  </div>
                </div>

                <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                  <button
                    type="submit"
                    disabled={restoreBackupMutation.isPending}
                    className="btn-danger w-full sm:w-auto sm:ml-3"
                  >
                    {restoreBackupMutation.isPending && (
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    )}
                    Restore Backup
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      setIsRestoreModalOpen(false);
                      setSelectedBackup(null);
                    }}
                    className="btn-secondary w-full sm:w-auto mt-3 sm:mt-0"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          </div>
        </div>
      )}

      {/* Backup Logs Modal */}
      {logsBackup && (
        <BackupLogsModal
          backup={logsBackup}
          isOpen={isLogsModalOpen}
          onClose={handleCloseLogsModal}
        />
      )}
    </div>
  );
}; 