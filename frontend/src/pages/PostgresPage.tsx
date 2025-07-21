import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { 
  Database, 
  Plus, 
  Edit, 
  Trash2, 
  Eye, 
  EyeOff,
  CheckCircle,
  XCircle,
  AlertCircle,
  Loader2
} from 'lucide-react';
import { apiService } from '../services/api';
import { useRefreshConfig } from '../hooks/useRefreshConfig';
import { useToast } from '../hooks/useToast';
import type { PostgresInstance } from '../types';

interface PostgresFormData {
  name: string;
  host: string;
  port: number;
  username: string;
  password: string;
  databases: string;
  ssl_mode: string;
  enabled: boolean;
}

export const PostgresPage: React.FC = () => {
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingInstance, setEditingInstance] = useState<PostgresInstance | null>(null);
  const [showPassword, setShowPassword] = useState(false);
  const queryClient = useQueryClient();
  const { getInstancesInterval } = useRefreshConfig();
  const { success, error } = useToast();

  const { data: instances = [], isLoading } = useQuery({
    queryKey: ['postgres-instances'],
    queryFn: () => apiService.getPostgresInstances(),
    refetchInterval: getInstancesInterval(),
    refetchIntervalInBackground: true, // Keep refreshing even when tab is not active
  });

  const createMutation = useMutation({
    mutationFn: (data: Omit<PostgresInstance, 'id'>) => apiService.createPostgresInstance(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['postgres-instances'] });
      setIsModalOpen(false);
      reset();
      success('Instance Created', 'PostgreSQL instance has been created successfully.');
    },
    onError: (err) => {
      error('Creation Failed', err instanceof Error ? err.message : 'Failed to create PostgreSQL instance.');
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<PostgresInstance> }) => 
      apiService.updatePostgresInstance(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['postgres-instances'] });
      setIsModalOpen(false);
      setEditingInstance(null);
      reset();
      success('Instance Updated', 'PostgreSQL instance has been updated successfully.');
    },
    onError: (err) => {
      error('Update Failed', err instanceof Error ? err.message : 'Failed to update PostgreSQL instance.');
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => apiService.deletePostgresInstance(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['postgres-instances'] });
      success('Instance Deleted', 'PostgreSQL instance has been deleted successfully.');
    },
    onError: (err) => {
      error('Deletion Failed', err instanceof Error ? err.message : 'Failed to delete PostgreSQL instance.');
    },
  });

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
    setValue,
  } = useForm<PostgresFormData>({
    defaultValues: {
      enabled: true,
      port: 5432,
      ssl_mode: 'prefer',
    },
  });

  const onSubmit = (data: PostgresFormData) => {
    const instanceData = {
      ...data,
      databases: data.databases.split(',').map(db => db.trim()).filter(Boolean),
    };

    if (editingInstance) {
      updateMutation.mutate({ id: editingInstance.id, data: instanceData });
    } else {
      createMutation.mutate(instanceData);
    }
  };

  const handleEdit = (instance: PostgresInstance) => {
    setEditingInstance(instance);
    setValue('name', instance.name);
    setValue('host', instance.host);
    setValue('port', instance.port);
    setValue('username', instance.username);
    setValue('password', instance.password);
    setValue('databases', (instance.databases || []).join(', '));
    setValue('ssl_mode', instance.ssl_mode || 'prefer');
    setValue('enabled', instance.enabled);
    setIsModalOpen(true);
  };

  const handleDelete = (id: string) => {
    if (confirm('Are you sure you want to delete this PostgreSQL instance?')) {
      deleteMutation.mutate(id);
    }
  };

  const toggleEnabled = (instance: PostgresInstance) => {
    updateMutation.mutate({
      id: instance.id,
      data: { enabled: !instance.enabled }
    });
    
    // Show immediate feedback since this is a quick toggle
    if (instance.enabled) {
      success('Instance Disabled', `${instance.name} has been disabled.`);
    } else {
      success('Instance Enabled', `${instance.name} has been enabled.`);
    }
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setEditingInstance(null);
    reset();
    setShowPassword(false);
  };

  if (isLoading) {
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
          <h1 className="text-2xl font-bold text-gray-900">PostgreSQL Instances</h1>
          <p className="mt-1 text-sm text-gray-500">
            Manage your PostgreSQL database connections
          </p>
        </div>
        <button
          onClick={() => setIsModalOpen(true)}
          className="btn-primary flex items-center"
        >
          <Plus className="h-4 w-4 mr-2" />
          Add Instance
        </button>
      </div>

      {/* Instances Grid */}
      {instances.length === 0 ? (
        <div className="text-center py-12">
          <Database className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-lg font-medium text-gray-900">No PostgreSQL instances</h3>
          <p className="mt-1 text-gray-500">
            Get started by adding your first PostgreSQL instance.
          </p>
          <button
            onClick={() => setIsModalOpen(true)}
            className="btn-primary mt-4"
          >
            <Plus className="h-4 w-4 mr-2" />
            Add Instance
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
          {instances.map((instance) => (
            <div key={instance.id} className="card">
              <div className="flex items-start justify-between">
                <div className="flex items-center">
                  <div className={`p-2 rounded-lg ${instance.enabled ? 'bg-success-100' : 'bg-gray-100'}`}>
                    <Database className={`h-6 w-6 ${instance.enabled ? 'text-success-600' : 'text-gray-400'}`} />
                  </div>
                  <div className="ml-3">
                    <h3 className="text-lg font-medium text-gray-900">{instance.name}</h3>
                    <p className="text-sm text-gray-500">{instance.host}:{instance.port}</p>
                  </div>
                </div>
                <div className="flex items-center space-x-2">
                  <button
                    onClick={() => toggleEnabled(instance)}
                    className={`p-1 rounded ${instance.enabled ? 'text-success-600 hover:bg-success-50' : 'text-gray-400 hover:bg-gray-50'}`}
                    title={instance.enabled ? 'Disable' : 'Enable'}
                  >
                    {instance.enabled ? <CheckCircle className="h-4 w-4" /> : <XCircle className="h-4 w-4" />}
                  </button>
                  <button
                    onClick={() => handleEdit(instance)}
                    className="p-1 text-gray-400 hover:text-primary-600 hover:bg-primary-50 rounded"
                    title="Edit"
                  >
                    <Edit className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => handleDelete(instance.id)}
                    className="p-1 text-gray-400 hover:text-danger-600 hover:bg-danger-50 rounded"
                    title="Delete"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>

              <div className="mt-4">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-gray-500">Username:</span>
                  <span className="font-medium">{instance.username}</span>
                </div>
                <div className="flex items-center justify-between text-sm mt-2">
                  <span className="text-gray-500">SSL Mode:</span>
                  <span className="font-medium capitalize">{instance.ssl_mode || 'prefer'}</span>
                </div>
                <div className="flex items-center justify-between text-sm mt-2">
                  <span className="text-gray-500">Databases:</span>
                  <span className="font-medium">{instance.databases?.length || 0}</span>
                </div>
                <div className="mt-3">
                  <div className="flex flex-wrap gap-1">
                    {(instance.databases || []).map((db) => (
                      <span
                        key={db}
                        className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-primary-100 text-primary-800"
                      >
                        {db}
                      </span>
                    ))}
                    {(!instance.databases || instance.databases.length === 0) && (
                      <span className="text-xs text-gray-500">No databases configured</span>
                    )}
                  </div>
                </div>
              </div>

              <div className="mt-4 pt-4 border-t border-gray-200">
                <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                  instance.enabled 
                    ? 'text-success-700 bg-success-50 border border-success-200'
                    : 'text-gray-700 bg-gray-50 border border-gray-200'
                }`}>
                  {instance.enabled ? 'Enabled' : 'Disabled'}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 overflow-y-auto">
          <div className="flex items-center justify-center min-h-screen px-4 pt-4 pb-20 text-center sm:block sm:p-0">
            <div className="fixed inset-0 transition-opacity bg-gray-500 bg-opacity-75" onClick={handleCloseModal} />
            
            <div className="inline-block align-bottom bg-white rounded-lg text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg sm:w-full">
              <form onSubmit={handleSubmit(onSubmit)}>
                <div className="bg-white px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
                  <div className="flex items-center mb-4">
                    <Database className="h-6 w-6 text-primary-600 mr-2" />
                    <h3 className="text-lg font-medium text-gray-900">
                      {editingInstance ? 'Edit PostgreSQL Instance' : 'Add PostgreSQL Instance'}
                    </h3>
                  </div>

                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700">Name</label>
                      <input
                        {...register('name', { required: 'Name is required' })}
                        className="input-field mt-1"
                        placeholder="Production Database"
                      />
                      {errors.name && (
                        <div className="mt-1 flex items-center text-sm text-danger-600">
                          <AlertCircle className="h-4 w-4 mr-1" />
                          {errors.name.message}
                        </div>
                      )}
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
                      <div className="sm:col-span-2">
                        <label className="block text-sm font-medium text-gray-700">Host</label>
                        <input
                          {...register('host', { required: 'Host is required' })}
                          className="input-field mt-1"
                          placeholder="localhost"
                        />
                        {errors.host && (
                          <div className="mt-1 flex items-center text-sm text-danger-600">
                            <AlertCircle className="h-4 w-4 mr-1" />
                            {errors.host.message}
                          </div>
                        )}
                      </div>

                      <div>
                        <label className="block text-sm font-medium text-gray-700">Port</label>
                        <input
                          {...register('port', { 
                            required: 'Port is required',
                            valueAsNumber: true,
                            min: { value: 1, message: 'Port must be greater than 0' },
                            max: { value: 65535, message: 'Port must be less than 65536' }
                          })}
                          type="number"
                          className="input-field mt-1"
                          placeholder="5432"
                        />
                        {errors.port && (
                          <div className="mt-1 flex items-center text-sm text-danger-600">
                            <AlertCircle className="h-4 w-4 mr-1" />
                            {errors.port.message}
                          </div>
                        )}
                      </div>
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">Username</label>
                      <input
                        {...register('username', { required: 'Username is required' })}
                        className="input-field mt-1"
                        placeholder="postgres"
                      />
                      {errors.username && (
                        <div className="mt-1 flex items-center text-sm text-danger-600">
                          <AlertCircle className="h-4 w-4 mr-1" />
                          {errors.username.message}
                        </div>
                      )}
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">Password</label>
                      <div className="mt-1 relative">
                        <input
                          {...register('password', { required: 'Password is required' })}
                          type={showPassword ? 'text' : 'password'}
                          className="input-field pr-10"
                          placeholder="••••••••"
                        />
                        <button
                          type="button"
                          className="absolute inset-y-0 right-0 pr-3 flex items-center"
                          onClick={() => setShowPassword(!showPassword)}
                        >
                          {showPassword ? (
                            <EyeOff className="h-4 w-4 text-gray-400" />
                          ) : (
                            <Eye className="h-4 w-4 text-gray-400" />
                          )}
                        </button>
                      </div>
                      {errors.password && (
                        <div className="mt-1 flex items-center text-sm text-danger-600">
                          <AlertCircle className="h-4 w-4 mr-1" />
                          {errors.password.message}
                        </div>
                      )}
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">Databases</label>
                      <input
                        {...register('databases', { required: 'At least one database is required' })}
                        className="input-field mt-1"
                        placeholder="database1, database2, database3"
                      />
                      <p className="mt-1 text-xs text-gray-500">
                        Comma-separated list of database names
                      </p>
                      {errors.databases && (
                        <div className="mt-1 flex items-center text-sm text-danger-600">
                          <AlertCircle className="h-4 w-4 mr-1" />
                          {errors.databases.message}
                        </div>
                      )}
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700">SSL Mode</label>
                      <select
                        {...register('ssl_mode', { required: 'SSL Mode is required' })}
                        className="input-field mt-1"
                      >
                        <option value="disable">Disable</option>
                        <option value="allow">Allow</option>
                        <option value="prefer">Prefer (Recommended)</option>
                        <option value="require">Require</option>
                      </select>
                      <p className="mt-1 text-xs text-gray-500">
                        SSL connection mode for PostgreSQL
                      </p>
                      {errors.ssl_mode && (
                        <div className="mt-1 flex items-center text-sm text-danger-600">
                          <AlertCircle className="h-4 w-4 mr-1" />
                          {errors.ssl_mode.message}
                        </div>
                      )}
                    </div>

                    <div className="flex items-center">
                      <input
                        {...register('enabled')}
                        type="checkbox"
                        className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                      />
                      <label className="ml-2 block text-sm text-gray-700">
                        Enable this instance
                      </label>
                    </div>
                  </div>
                </div>

                <div className="bg-gray-50 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse">
                  <button
                    type="submit"
                    disabled={createMutation.isPending || updateMutation.isPending}
                    className="btn-primary w-full sm:w-auto sm:ml-3"
                  >
                    {(createMutation.isPending || updateMutation.isPending) && (
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                    )}
                    {editingInstance ? 'Update' : 'Create'}
                  </button>
                  <button
                    type="button"
                    onClick={handleCloseModal}
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
    </div>
  );
}; 