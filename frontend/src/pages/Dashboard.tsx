import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { 
  Database, 
  Archive, 
  CheckCircle, 
  XCircle, 
  Clock,
  Activity,
  AlertTriangle
} from 'lucide-react';
import { apiService } from '../services/api';
import { useRefreshConfig } from '../hooks/useRefreshConfig';
import { formatDistanceToNow } from 'date-fns';

export const Dashboard: React.FC = () => {
  const { getInstancesInterval, getBackupsInterval, getHealthInterval } = useRefreshConfig();

  const { data: postgresInstances = [] } = useQuery({
    queryKey: ['postgres-instances'],
    queryFn: () => apiService.getPostgresInstances(),
    refetchInterval: getInstancesInterval(),
    refetchIntervalInBackground: true, // Keep refreshing even when tab is not active
  });

  const { data: backups = [] } = useQuery({
    queryKey: ['backups'],
    queryFn: () => apiService.getBackups(),
    refetchInterval: getBackupsInterval(),
    refetchIntervalInBackground: true,
  });

  const { data: health } = useQuery({
    queryKey: ['health'],
    queryFn: () => apiService.checkHealth(),
    refetchInterval: getHealthInterval(),
    refetchIntervalInBackground: true,
  });

  // Calculate metrics
  const enabledInstances = postgresInstances.filter(instance => instance.enabled);
  const totalDatabases = postgresInstances.reduce((acc, instance) => acc + instance.databases.length, 0);
  
  const recentBackups = backups
    .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
    .slice(0, 5);

  const backupStats = {
    total: backups.length,
    completed: backups.filter(b => b.status === 'completed').length,
    failed: backups.filter(b => b.status === 'failed').length,
    inProgress: backups.filter(b => b.status === 'in_progress').length,
  };

  const cards = [
    {
      title: 'PostgreSQL Instances',
      value: postgresInstances.length,
      subtitle: `${enabledInstances.length} enabled`,
      icon: Database,
      color: 'primary',
    },
    {
      title: 'Total Databases',
      value: totalDatabases,
      subtitle: 'Across all instances',
      icon: Database,
      color: 'success',
    },
    {
      title: 'Total Backups',
      value: backupStats.total,
      subtitle: `${backupStats.completed} completed`,
      icon: Archive,
      color: 'primary',
    },
    {
      title: 'Failed Backups',
      value: backupStats.failed,
      subtitle: 'Need attention',
      icon: AlertTriangle,
      color: 'danger',
    },
  ];

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-success-500" />;
      case 'failed':
        return <XCircle className="h-4 w-4 text-danger-500" />;
      case 'in_progress':
        return <Clock className="h-4 w-4 text-warning-500" />;
      default:
        return <Clock className="h-4 w-4 text-gray-400" />;
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

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="mt-1 text-sm text-gray-500">
            Overview of your PostgreSQL backup system
          </p>
        </div>
        {health && (
          <div className="flex items-center space-x-2 px-3 py-2 bg-success-50 border border-success-200 rounded-lg">
            <Activity className="h-4 w-4 text-success-500" />
            <span className="text-sm font-medium text-success-700">
              System Healthy
            </span>
            <span className="text-xs text-success-600">
              (Uptime: {health.uptime})
            </span>
          </div>
        )}
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {cards.map((card) => (
          <div key={card.title} className="card">
            <div className="flex items-center">
              <div className={`p-2 rounded-lg bg-${card.color}-100`}>
                <card.icon className={`h-6 w-6 text-${card.color}-600`} />
              </div>
              <div className="ml-4">
                <p className="text-sm font-medium text-gray-600">{card.title}</p>
                <p className="text-2xl font-bold text-gray-900">{card.value}</p>
                <p className="text-xs text-gray-500">{card.subtitle}</p>
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Backups */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-medium text-gray-900">Recent Backups</h3>
            <Archive className="h-5 w-5 text-gray-400" />
          </div>
          
          {recentBackups.length === 0 ? (
            <p className="text-sm text-gray-500 text-center py-8">
              No backups found. Create your first backup to get started.
            </p>
          ) : (
            <div className="space-y-3">
              {recentBackups.map((backup) => (
                <div
                  key={backup.id}
                  className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                >
                  <div className="flex items-center space-x-3">
                    {getStatusIcon(backup.status)}
                    <div>
                      <p className="text-sm font-medium text-gray-900">
                        {backup.database_name}
                      </p>
                      <p className="text-xs text-gray-500">
                        {backup.backup_type} • {formatDistanceToNow(new Date(backup.created_at), { addSuffix: true })}
                      </p>
                    </div>
                  </div>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${getStatusColor(backup.status)}`}>
                    {backup.status.replace('_', ' ')}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* PostgreSQL Instances Status */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-medium text-gray-900">PostgreSQL Instances</h3>
            <Database className="h-5 w-5 text-gray-400" />
          </div>
          
          {postgresInstances.length === 0 ? (
            <p className="text-sm text-gray-500 text-center py-8">
              No PostgreSQL instances configured. Add your first instance to get started.
            </p>
          ) : (
            <div className="space-y-3">
              {postgresInstances.map((instance) => (
                <div
                  key={instance.id}
                  className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                >
                  <div className="flex items-center space-x-3">
                    <div className={`h-3 w-3 rounded-full ${instance.enabled ? 'bg-success-400' : 'bg-gray-300'}`} />
                    <div>
                      <p className="text-sm font-medium text-gray-900">
                        {instance.name}
                      </p>
                      <p className="text-xs text-gray-500">
                        {instance.host}:{instance.port} • {instance.databases.length} database{instance.databases.length !== 1 ? 's' : ''}
                      </p>
                    </div>
                  </div>
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border ${
                    instance.enabled 
                      ? 'text-success-700 bg-success-50 border-success-200'
                      : 'text-gray-700 bg-gray-50 border-gray-200'
                  }`}>
                    {instance.enabled ? 'Enabled' : 'Disabled'}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Quick Actions */}
      <div className="card">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Quick Actions</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <button className="btn-primary text-left p-4 h-auto">
            <Database className="h-6 w-6 mb-2" />
            <div>
              <p className="font-medium">Add PostgreSQL Instance</p>
              <p className="text-sm opacity-90">Configure a new database connection</p>
            </div>
          </button>
          
          <button className="btn-secondary text-left p-4 h-auto">
            <Archive className="h-6 w-6 mb-2" />
            <div>
              <p className="font-medium">Create Manual Backup</p>
              <p className="text-sm opacity-90">Run an immediate backup job</p>
            </div>
          </button>
          
          <button className="btn-secondary text-left p-4 h-auto">
            <Activity className="h-6 w-6 mb-2" />
            <div>
              <p className="font-medium">View System Logs</p>
              <p className="text-sm opacity-90">Monitor backup activities</p>
            </div>
          </button>
        </div>
      </div>
    </div>
  );
};