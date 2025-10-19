import React, { createContext, useContext, useEffect, useState } from 'react';
import { isAuthenticated, logout as authLogout } from './auth';

interface AuthContextType {
  isAuthenticated: boolean;
  logout: () => void;
  loading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [authenticated, setAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check authentication status on mount
    setAuthenticated(isAuthenticated());
    setLoading(false);
  }, []);

  const logout = () => {
    authLogout();
    setAuthenticated(false);
  };

  return (
    <AuthContext.Provider value={{ 
      isAuthenticated: authenticated, 
      logout, 
      loading 
    }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
