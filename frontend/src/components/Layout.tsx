import React, { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { 
  Database, 
  Archive, 
  Settings, 
  FileText, 
  LogOut, 
  Menu, 
  X,
  Home 
} from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { clsx } from 'clsx';

const navigation = [
  { name: 'Dashboard', href: '/', icon: Home },
  { name: 'PostgreSQL Instances', href: '/postgres', icon: Database },
  { name: 'Backups', href: '/backups', icon: Archive },
  { name: 'Logs', href: '/logs', icon: FileText },
  { name: 'Settings', href: '/settings', icon: Settings },
];

export const Layout: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const { state, logout } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Mobile sidebar */}
      <div className={clsx(
        'fixed inset-0 z-50 lg:hidden',
        sidebarOpen ? 'block' : 'hidden'
      )}>
        <div className="fixed inset-0 bg-gray-600 bg-opacity-75" />
        <div className="fixed inset-y-0 left-0 flex w-64 flex-col bg-white">
          <div className="flex h-16 shrink-0 items-center justify-between px-4 border-b border-gray-200">
            <div className="flex items-center">
              <Database className="h-8 w-8 text-primary-600" />
              <span className="ml-2 text-lg font-semibold text-gray-900">
                Backup Manager
              </span>
            </div>
            <button
              onClick={() => setSidebarOpen(false)}
              className="text-gray-400 hover:text-gray-600"
            >
              <X className="h-6 w-6" />
            </button>
          </div>
          <nav className="flex-1 space-y-1 px-2 py-4">
            {navigation.map((item) => {
              const isActive = location.pathname === item.href;
              return (
                <button
                  key={item.name}
                  onClick={() => {
                    navigate(item.href);
                    setSidebarOpen(false);
                  }}
                  className={clsx(
                    'group flex w-full items-center rounded-md px-3 py-2 text-sm font-medium',
                    isActive
                      ? 'bg-primary-100 text-primary-900'
                      : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                  )}
                >
                  <item.icon
                    className={clsx(
                      'mr-3 h-5 w-5 flex-shrink-0',
                      isActive ? 'text-primary-500' : 'text-gray-400 group-hover:text-gray-500'
                    )}
                  />
                  {item.name}
                </button>
              );
            })}
          </nav>
          <div className="border-t border-gray-200 p-4">
            <div className="flex items-center mb-3">
              <div className="h-8 w-8 bg-primary-600 rounded-full flex items-center justify-center">
                <span className="text-xs font-medium text-white">
                  {state.user?.baseUrl.charAt(8) || 'U'}
                </span>
              </div>
              <div className="ml-3">
                <p className="text-sm font-medium text-gray-900">Connected</p>
                <p className="text-xs text-gray-500 truncate">
                  {state.user?.baseUrl}
                </p>
              </div>
            </div>
            <button
              onClick={handleLogout}
              className="btn-secondary w-full flex items-center justify-center"
            >
              <LogOut className="h-4 w-4 mr-2" />
              Logout
            </button>
          </div>
        </div>
      </div>

      {/* Desktop sidebar */}
      <div className="hidden lg:fixed lg:inset-y-0 lg:z-50 lg:flex lg:w-64 lg:flex-col">
        <div className="flex grow flex-col gap-y-5 overflow-y-auto bg-white border-r border-gray-200">
          <div className="flex h-16 shrink-0 items-center px-4 border-b border-gray-200">
            <Database className="h-8 w-8 text-primary-600" />
            <span className="ml-2 text-lg font-semibold text-gray-900">
              Backup Manager
            </span>
          </div>
          <nav className="flex flex-1 flex-col px-4">
            <ul className="flex flex-1 flex-col gap-y-7">
              <li>
                <ul className="space-y-1">
                  {navigation.map((item) => {
                    const isActive = location.pathname === item.href;
                    return (
                      <li key={item.name}>
                        <button
                          onClick={() => navigate(item.href)}
                          className={clsx(
                            'group flex w-full items-center rounded-md px-3 py-2 text-sm font-medium',
                            isActive
                              ? 'bg-primary-100 text-primary-900'
                              : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                          )}
                        >
                          <item.icon
                            className={clsx(
                              'mr-3 h-5 w-5 flex-shrink-0',
                              isActive ? 'text-primary-500' : 'text-gray-400 group-hover:text-gray-500'
                            )}
                          />
                          {item.name}
                        </button>
                      </li>
                    );
                  })}
                </ul>
              </li>
              <li className="mt-auto">
                <div className="border-t border-gray-200 pt-4">
                  <div className="flex items-center mb-3">
                    <div className="h-8 w-8 bg-primary-600 rounded-full flex items-center justify-center">
                      <span className="text-xs font-medium text-white">
                        {state.user?.baseUrl.charAt(8) || 'U'}
                      </span>
                    </div>
                    <div className="ml-3">
                      <p className="text-sm font-medium text-gray-900">Connected</p>
                      <p className="text-xs text-gray-500 truncate">
                        {state.user?.baseUrl}
                      </p>
                    </div>
                  </div>
                  <button
                    onClick={handleLogout}
                    className="btn-secondary w-full flex items-center justify-center"
                  >
                    <LogOut className="h-4 w-4 mr-2" />
                    Logout
                  </button>
                </div>
              </li>
            </ul>
          </nav>
        </div>
      </div>

      {/* Main content */}
      <div className="lg:pl-64">
        {/* Top bar for mobile */}
        <div className="sticky top-0 z-40 flex h-16 shrink-0 items-center gap-x-4 border-b border-gray-200 bg-white px-4 shadow-sm sm:gap-x-6 sm:px-6 lg:hidden">
          <button
            onClick={() => setSidebarOpen(true)}
            className="text-gray-700"
          >
            <Menu className="h-6 w-6" />
          </button>
          <div className="flex flex-1 items-center justify-between">
            <h1 className="text-lg font-semibold text-gray-900">
              Backup Manager
            </h1>
            <button
              onClick={handleLogout}
              className="text-gray-400 hover:text-gray-600"
            >
              <LogOut className="h-5 w-5" />
            </button>
          </div>
        </div>

        {/* Page content */}
        <main className="p-4 sm:p-6 lg:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}; 