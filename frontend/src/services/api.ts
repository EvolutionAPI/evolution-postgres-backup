import axios, { type AxiosInstance } from 'axios';
import type { 
  ApiConfig, 
  PostgresInstance, 
  BackupJob, 
  BackupRequest, 
  RestoreRequest, 
  LogEntry,
  LogFile,
  HealthCheck,
  ApiResponse 
} from '../types';

class ApiService {
  private client: AxiosInstance | null = null;
  private config: ApiConfig | null = null;

  initialize(config: ApiConfig) {
    this.config = config;
    this.client = axios.create({
      baseURL: config.baseUrl,
      headers: {
        'api-key': config.apiKey,
        'Content-Type': 'application/json',
      },
      timeout: 30000,
    });

    // Request interceptor
    this.client.interceptors.request.use(
      (config) => {
        console.log(`API Request: ${config.method?.toUpperCase()} ${config.url}`);
        return config;
      },
      (error) => {
        console.error('API Request Error:', error);
        return Promise.reject(error);
      }
    );

    // Response interceptor
    this.client.interceptors.response.use(
      (response) => {
        console.log(`API Response: ${response.status} ${response.config.url}`);
        return response;
      },
      (error) => {
        console.error('API Response Error:', error.response?.data || error.message);
        return Promise.reject(error);
      }
    );
  }

  private ensureInitialized() {
    if (!this.client) {
      throw new Error('API service not initialized');
    }
    return this.client;
  }

  // Health check
  async checkHealth(): Promise<HealthCheck> {
    const client = this.ensureInitialized();
    const response = await client.get<HealthCheck>('/health');
    return response.data;
  }

  // Test connection
  async testConnection(): Promise<boolean> {
    try {
      await this.checkHealth();
      return true;
    } catch {
      return false;
    }
  }

  // PostgreSQL instances
  async getPostgresInstances(): Promise<PostgresInstance[]> {
    const client = this.ensureInitialized();
    const response = await client.get<ApiResponse<PostgresInstance[]>>('/api/v2/postgres');
    return response.data.data || [];
  }

  async createPostgresInstance(instance: Omit<PostgresInstance, 'id'>): Promise<PostgresInstance> {
    const client = this.ensureInitialized();
    const response = await client.post<ApiResponse<PostgresInstance>>('/api/v2/postgres', instance);
    if (!response.data.data) {
      throw new Error(response.data.error || 'Failed to create instance');
    }
    return response.data.data;
  }

  async updatePostgresInstance(id: string, instance: Partial<PostgresInstance>): Promise<PostgresInstance> {
    const client = this.ensureInitialized();
    const response = await client.put<ApiResponse<PostgresInstance>>(`/api/v2/postgres/${id}`, instance);
    if (!response.data.data) {
      throw new Error(response.data.error || 'Failed to update instance');
    }
    return response.data.data;
  }

  async deletePostgresInstance(id: string): Promise<void> {
    const client = this.ensureInitialized();
    await client.delete(`/api/v2/postgres/${id}`);
  }

  // Backups
  async getBackups(): Promise<BackupJob[]> {
    const client = this.ensureInitialized();
    const response = await client.get<ApiResponse<BackupJob[]>>('/api/v2/backups');
    return response.data.data || [];
  }

  async getBackup(id: string): Promise<BackupJob> {
    const client = this.ensureInitialized();
    const response = await client.get<ApiResponse<BackupJob>>(`/api/v2/backups/${id}`);
    if (!response.data.data) {
      throw new Error(response.data.error || 'Backup not found');
    }
    return response.data.data;
  }

  async createBackup(request: BackupRequest): Promise<BackupJob> {
    const client = this.ensureInitialized();
    const response = await client.post<ApiResponse<BackupJob>>('/api/v2/workers/jobs/backup', request);
    if (!response.data.data) {
      throw new Error(response.data.error || 'Failed to create backup');
    }
    return response.data.data;
  }

  async restoreBackup(request: RestoreRequest): Promise<{ message: string }> {
    const client = this.ensureInitialized();
    const response = await client.post<ApiResponse<{ message: string }>>('/api/v2/workers/jobs/restore', request);
    if (!response.data.data) {
      throw new Error(response.data.error || 'Failed to restore backup');
    }
    return response.data.data;
  }

  // Logs
  async getLogs(params?: { 
    level?: string; 
    job_id?: string; 
    limit?: number; 
    date?: string;
    from?: string; 
    to?: string 
  }): Promise<LogEntry[]> {
    const client = this.ensureInitialized();
    const response = await client.get<ApiResponse<LogEntry[]>>('/api/v2/logs', { params });
    return response.data.data || [];
  }

  async getLogsByBackupId(backupId: string): Promise<LogEntry[]> {
    const client = this.ensureInitialized();
    const response = await client.get<ApiResponse<LogEntry[]>>(`/api/v2/logs/backup/${backupId}`);
    return response.data.data || [];
  }

  async getLogFiles(): Promise<LogFile[]> {
    const client = this.ensureInitialized();
    const response = await client.get<ApiResponse<LogFile[]>>('/api/v2/logs/files');
    return response.data.data || [];
  }

  // Real-time logs via EventSource
  subscribeToLogs(
    onMessage: (log: LogEntry) => void,
    onError?: (error: Event) => void
  ): EventSource | null {
    if (!this.config) {
      throw new Error('API service not initialized');
    }

    // EventSource doesn't support custom headers, so we need to pass auth via URL params
    const url = new URL(`${this.config.baseUrl}/api/v2/logs/stream`);
    url.searchParams.append('api-key', this.config.apiKey);
    
    const eventSource = new EventSource(url.toString());

    eventSource.onmessage = (event) => {
      try {
        const logEntry: LogEntry = JSON.parse(event.data);
        onMessage(logEntry);
      } catch (error) {
        console.error('Failed to parse log entry:', error, 'Data:', event.data);
      }
    };

    eventSource.onerror = (error) => {
      console.error('EventSource error:', error);
      if (onError) {
        onError(error);
      }
    };

    return eventSource;
  }
}

export const apiService = new ApiService(); 