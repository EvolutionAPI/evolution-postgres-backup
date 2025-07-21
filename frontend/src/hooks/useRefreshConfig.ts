import { useState, useEffect } from 'react';

interface RefreshConfig {
  instancesInterval: number; // milliseconds
  backupsInterval: number;
  healthInterval: number;
  enabled: boolean;
}

const DEFAULT_CONFIG: RefreshConfig = {
  instancesInterval: 15000, // 15 seconds
  backupsInterval: 5000,    // 5 seconds
  healthInterval: 30000,    // 30 seconds
  enabled: true,
};

export const useRefreshConfig = () => {
  const [config, setConfig] = useState<RefreshConfig>(() => {
    const saved = localStorage.getItem('refresh-config');
    if (saved) {
      try {
        return { ...DEFAULT_CONFIG, ...JSON.parse(saved) };
      } catch {
        return DEFAULT_CONFIG;
      }
    }
    return DEFAULT_CONFIG;
  });

  useEffect(() => {
    localStorage.setItem('refresh-config', JSON.stringify(config));
  }, [config]);

  const updateConfig = (updates: Partial<RefreshConfig>) => {
    setConfig(prev => ({ ...prev, ...updates }));
  };

  const resetToDefaults = () => {
    setConfig(DEFAULT_CONFIG);
  };

  // Helper functions to get intervals (returns false when disabled)
  const getInstancesInterval = () => config.enabled ? config.instancesInterval : false;
  const getBackupsInterval = () => config.enabled ? config.backupsInterval : false;
  const getHealthInterval = () => config.enabled ? config.healthInterval : false;

  return {
    config,
    updateConfig,
    resetToDefaults,
    getInstancesInterval,
    getBackupsInterval,
    getHealthInterval,
  };
}; 