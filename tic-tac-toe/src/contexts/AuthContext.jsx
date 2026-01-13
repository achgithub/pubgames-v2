import React, { createContext, useContext, useState, useEffect } from 'react';
import { validateToken, saveAuthData, clearAuthData, getSavedAuth } from '../services/authApi';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const initAuth = async () => {
      // Check for token in URL
      const params = new URLSearchParams(window.location.search);
      const token = params.get('token');
      
      if (token) {
        const result = await validateToken(token);
        if (result.success) {
          setUser(result.user);
          saveAuthData(result.user, token);
        }
        window.history.replaceState({}, '', window.location.pathname);
        setLoading(false);
        return;
      }
      
      // Check for saved auth
      const savedAuth = getSavedAuth();
      if (savedAuth) {
        setUser(savedAuth.user);
      }
      
      setLoading(false);
    };

    initAuth();
  }, []);

  const logout = () => {
    setUser(null);
    clearAuthData();
  };

  return (
    <AuthContext.Provider value={{ user, loading, logout }}>
      {children}
    </AuthContext.Provider>
  );
};
