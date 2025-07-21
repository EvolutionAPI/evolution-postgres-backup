export interface ApiConfig {
  baseUrl: string;
  apiKey: string;
}

export interface PostgresInstance {
  id: string;
  name: string;
  host: string;
  port: number;
  username: string;
  password: string;
  databases: string[];
  ssl_mode?: string; // SSL connection mode: 'disable', 'allow', 'prefer', 'require'
  enabled: boolean;
}

export interface BackupJob {
  id: string;
  postgresql_id: string; // Corrigido: era postgres_id
  database_name: string; // Corrigido: era database
  backup_type: 'hourly' | 'daily' | 'weekly' | 'monthly' | 'manual';
  status: 'pending' | 'in_progress' | 'completed' | 'failed';
  created_at: string;
  completed_at?: string;
  file_size?: number;
  s3_key?: string;
  error_message?: string;
  job_id?: string; // Job ID for log correlation
}

export interface BackupRequest {
  postgresql_id: string;
  database_name: string;
  backup_type: 'manual';
}

export interface RestoreRequest {
  postgresql_id: string;
  database_name: string;
  backup_id: string;
}

export interface LogEntry {
  timestamp: string;
  level: 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
  job_id?: string;
  message: string;
}

export interface LogFile {
  name: string;
  date: string;
  size: number;
  modified: string;
  path: string;
}

export interface HealthCheck {
  status: 'ok';
  timestamp: string;
  uptime: string;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
  success?: boolean;
}

export interface Toast {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  title: string;
  message?: string;
  duration?: number;
  dismissible?: boolean;
}

export interface User {
  apiKey: string;
  baseUrl: string;
  isAuthenticated: boolean;
} 