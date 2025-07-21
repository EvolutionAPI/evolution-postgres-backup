import React, { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import {
  Settings,
  Save,
  RefreshCw,
  Database,
  Cloud,
  Shield,
  Bell,
  AlertCircle,
  CheckCircle,
  Info,
} from "lucide-react";
import { useAuth } from "../hooks/useAuth";
import { useRefreshConfig } from "../hooks/useRefreshConfig";
import { useToast } from "../hooks/useToast";
import { apiService } from "../services/api";

export const SettingsPage: React.FC = () => {
  const { state: authState } = useAuth();
  const [saveStatus, setSaveStatus] = useState<
    "idle" | "saving" | "saved" | "error"
  >("idle");
  const {
    config,
    updateConfig,
    getInstancesInterval,
    getBackupsInterval,
    getHealthInterval,
  } = useRefreshConfig();
  const { success, error, info } = useToast();

  const { data: health } = useQuery({
    queryKey: ["health"],
    queryFn: () => apiService.checkHealth(),
    refetchInterval: getHealthInterval(),
    refetchIntervalInBackground: true,
  });

  const { data: instances = [] } = useQuery({
    queryKey: ["postgres-instances"],
    queryFn: () => apiService.getPostgresInstances(),
    refetchInterval: getInstancesInterval(),
    refetchIntervalInBackground: true,
  });

  const { data: backups = [] } = useQuery({
    queryKey: ["backups"],
    queryFn: () => apiService.getBackups(),
    refetchInterval: getBackupsInterval(),
    refetchIntervalInBackground: true,
  });

  const handleSave = async () => {
    setSaveStatus("saving");
    info('Saving Settings', 'Your settings are being saved...');

    // Simulate saving settings
    await new Promise((resolve) => setTimeout(resolve, 1000));

    setSaveStatus("saved");
    success('Settings Saved', 'Your auto-refresh settings have been saved successfully.');
    setTimeout(() => setSaveStatus("idle"), 2000);
  };

  const testConnection = async () => {
    try {
      info('Testing Connection', 'Checking connection to backup service...');
      const isConnected = await apiService.testConnection();
      if (isConnected) {
        success('Connection Successful', 'Successfully connected to the backup service.');
      } else {
        error('Connection Failed', 'Failed to connect to the backup service. Please check your configuration.');
      }
    } catch (err) {
      error('Connection Error', err instanceof Error ? err.message : 'An unexpected error occurred while testing the connection.');
    }
  };

  const stats = {
    totalInstances: instances.length,
    enabledInstances: instances.filter((i) => i.enabled).length,
    totalBackups: backups.length,
    successfulBackups: backups.filter((b) => b.status === "completed").length,
    failedBackups: backups.filter((b) => b.status === "failed").length,
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
          <p className="mt-1 text-sm text-gray-500">
            Configure your backup system settings and preferences
          </p>
        </div>
        <button
          onClick={handleSave}
          disabled={saveStatus === "saving"}
          className="btn-primary flex items-center"
        >
          {saveStatus === "saving" ? (
            <RefreshCw className="h-4 w-4 mr-2 animate-spin" />
          ) : saveStatus === "saved" ? (
            <CheckCircle className="h-4 w-4 mr-2" />
          ) : (
            <Save className="h-4 w-4 mr-2" />
          )}
          {saveStatus === "saving"
            ? "Saving..."
            : saveStatus === "saved"
            ? "Saved!"
            : "Save Changes"}
        </button>
      </div>

      {/* Connection Status */}
      <div className="card">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-medium text-gray-900 flex items-center">
            <Shield className="h-5 w-5 mr-2 text-primary-600" />
            Connection Status
          </h2>
          <button onClick={testConnection} className="btn-secondary text-sm">
            Test Connection
          </button>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 bg-gray-50 rounded-lg">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-700">
                Service URL
              </span>
              <span className="text-sm text-gray-600">
                {authState.user?.baseUrl}
              </span>
            </div>
          </div>

          <div className="p-4 bg-gray-50 rounded-lg">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-gray-700">API Key</span>
              <span className="text-sm text-gray-600 font-mono">
                {authState.user?.apiKey
                  ? "••••••••••••" + authState.user.apiKey.slice(-4)
                  : "Not set"}
              </span>
            </div>
          </div>

          {health && (
            <div className="p-4 bg-success-50 rounded-lg">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-success-700">
                  System Status
                </span>
                <span className="text-sm text-success-600">Healthy</span>
              </div>
              <div className="mt-1 text-xs text-success-600">
                Uptime: {health.uptime}
              </div>
            </div>
          )}

          <div className="p-4 bg-primary-50 rounded-lg">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium text-primary-700">
                Last Updated
              </span>
              <span className="text-sm text-primary-600">
                {new Date().toLocaleString()}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* System Statistics */}
      <div className="card">
        <h2 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
          <Database className="h-5 w-5 mr-2 text-primary-600" />
          System Statistics
        </h2>

        <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
          <div className="text-center p-4 bg-gray-50 rounded-lg">
            <div className="text-2xl font-bold text-gray-900">
              {stats.totalInstances}
            </div>
            <div className="text-sm text-gray-600">Total Instances</div>
          </div>

          <div className="text-center p-4 bg-success-50 rounded-lg">
            <div className="text-2xl font-bold text-success-700">
              {stats.enabledInstances}
            </div>
            <div className="text-sm text-success-600">Enabled Instances</div>
          </div>

          <div className="text-center p-4 bg-primary-50 rounded-lg">
            <div className="text-2xl font-bold text-primary-700">
              {stats.totalBackups}
            </div>
            <div className="text-sm text-primary-600">Total Backups</div>
          </div>

          <div className="text-center p-4 bg-success-50 rounded-lg">
            <div className="text-2xl font-bold text-success-700">
              {stats.successfulBackups}
            </div>
            <div className="text-sm text-success-600">Successful</div>
          </div>

          <div className="text-center p-4 bg-danger-50 rounded-lg">
            <div className="text-2xl font-bold text-danger-700">
              {stats.failedBackups}
            </div>
            <div className="text-sm text-danger-600">Failed</div>
          </div>
        </div>
      </div>

      {/* Application Settings */}
      <div className="card">
        <h2 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
          <Settings className="h-5 w-5 mr-2 text-primary-600" />
          Application Settings
        </h2>

        <div className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Auto-refresh Enabled
              </label>
              <div className="flex items-center">
                <input
                  type="checkbox"
                  checked={config.enabled}
                  onChange={(e) => updateConfig({ enabled: e.target.checked })}
                  className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                />
                <label className="ml-2 block text-sm text-gray-700">
                  Enable automatic data refresh
                </label>
              </div>
              <p className="mt-1 text-xs text-gray-500">
                Automatically refresh data without manual page reload
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                PostgreSQL Instances Refresh
              </label>
              <select 
                value={config.instancesInterval / 1000} 
                onChange={(e) => updateConfig({ instancesInterval: parseInt(e.target.value) * 1000 })}
                className="input-field"
                disabled={!config.enabled}
              >
                <option value="5">Every 5 seconds</option>
                <option value="10">Every 10 seconds</option>
                <option value="15">Every 15 seconds</option>
                <option value="30">Every 30 seconds</option>
                <option value="60">Every minute</option>
              </select>
              <p className="mt-1 text-xs text-gray-500">
                How often to refresh PostgreSQL instance status
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Backups Refresh
              </label>
              <select 
                value={config.backupsInterval / 1000} 
                onChange={(e) => updateConfig({ backupsInterval: parseInt(e.target.value) * 1000 })}
                className="input-field"
                disabled={!config.enabled}
              >
                <option value="3">Every 3 seconds</option>
                <option value="5">Every 5 seconds</option>
                <option value="10">Every 10 seconds</option>
                <option value="15">Every 15 seconds</option>
                <option value="30">Every 30 seconds</option>
              </select>
              <p className="mt-1 text-xs text-gray-500">
                How often to refresh backup status (more frequent for real-time updates)
              </p>
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                System Health Refresh
              </label>
              <select 
                value={config.healthInterval / 1000} 
                onChange={(e) => updateConfig({ healthInterval: parseInt(e.target.value) * 1000 })}
                className="input-field"
                disabled={!config.enabled}
              >
                <option value="15">Every 15 seconds</option>
                <option value="30">Every 30 seconds</option>
                <option value="60">Every minute</option>
                <option value="120">Every 2 minutes</option>
                <option value="300">Every 5 minutes</option>
              </select>
              <p className="mt-1 text-xs text-gray-500">
                How often to check system health status
              </p>
            </div>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <div className="flex items-start">
              <Info className="h-5 w-5 text-blue-500 mt-0.5 mr-3" />
              <div>
                <h4 className="text-sm font-medium text-blue-900">Auto-refresh Information</h4>
                <p className="text-sm text-blue-700 mt-1">
                  Auto-refresh automatically updates data in real-time without requiring manual page refresh. 
                  This helps you monitor backup progress and system status as they happen. Lower intervals provide 
                  more real-time updates but use more resources.
                </p>
                <p className="text-xs text-blue-600 mt-2">
                  Current status: {config.enabled ? '✅ Enabled' : '❌ Disabled'}
                </p>
              </div>
            </div>
          </div>

          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-3 flex items-center">
              <Bell className="h-4 w-4 mr-1" />
              Notifications
            </h3>
            <div className="space-y-3">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  defaultChecked
                  className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                />
                <span className="ml-2 text-sm text-gray-700">
                  Show backup completion notifications
                </span>
              </label>

              <label className="flex items-center">
                <input
                  type="checkbox"
                  defaultChecked
                  className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                />
                <span className="ml-2 text-sm text-gray-700">
                  Alert on backup failures
                </span>
              </label>

              <label className="flex items-center">
                <input
                  type="checkbox"
                  className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                />
                <span className="ml-2 text-sm text-gray-700">
                  Daily summary emails
                </span>
              </label>
            </div>
          </div>
        </div>
      </div>

      {/* Storage Settings */}
      <div className="card">
        <h2 className="text-lg font-medium text-gray-900 mb-4 flex items-center">
          <Cloud className="h-5 w-5 mr-2 text-primary-600" />
          Storage Settings
        </h2>

        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex">
            <Info className="h-5 w-5 text-blue-400" />
            <div className="ml-3">
              <h3 className="text-sm font-medium text-blue-800">
                Storage Configuration
              </h3>
              <div className="mt-2 text-sm text-blue-700">
                Storage settings are configured on the server side through
                environment variables. Contact your system administrator to
                modify S3 bucket, credentials, or retention policies.
              </div>
            </div>
          </div>
        </div>

        <div className="mt-4 grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="p-4 bg-gray-50 rounded-lg">
            <div className="text-sm font-medium text-gray-700">
              Retention Policy
            </div>
            <div className="mt-1 text-xs text-gray-600">
              <div>• Hourly: 24 hours</div>
              <div>• Daily: 30 days</div>
              <div>• Weekly: 8 weeks</div>
              <div>• Monthly: 12 months</div>
            </div>
          </div>

          <div className="p-4 bg-gray-50 rounded-lg">
            <div className="text-sm font-medium text-gray-700">
              Backup Types
            </div>
            <div className="mt-1 text-xs text-gray-600">
              <div>• Manual backups</div>
              <div>• Scheduled backups</div>
              <div>• Full database dumps</div>
              <div>• Multiple database support</div>
            </div>
          </div>
        </div>
      </div>

      {/* Danger Zone */}
      <div className="card border-danger-200">
        <h2 className="text-lg font-medium text-danger-900 mb-4 flex items-center">
          <AlertCircle className="h-5 w-5 mr-2 text-danger-600" />
          Danger Zone
        </h2>

        <div className="space-y-4">
          <div className="flex items-center justify-between p-4 bg-danger-50 rounded-lg border border-danger-200">
            <div>
              <h3 className="text-sm font-medium text-danger-900">
                Clear Application Data
              </h3>
              <p className="text-sm text-danger-700">
                Remove all stored connection settings and cached data from this
                browser.
              </p>
            </div>
            <button
              onClick={() => {
                if (
                  confirm(
                    "Are you sure? This will log you out and clear all local data."
                  )
                ) {
                  localStorage.clear();
                  window.location.reload();
                }
              }}
              className="btn-danger text-sm"
            >
              Clear Data
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
