import { IDENTITY_API } from './api';

export const validateToken = async (token) => {
  try {
    const response = await fetch(`${IDENTITY_API}/validate-token`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    if (response.ok) {
      const userData = await response.json();
      return { success: true, user: userData };
    } else {
      return { success: false, error: 'Invalid token' };
    }
  } catch (error) {
    console.error('Token validation failed:', error);
    return { success: false, error: error.message };
  }
};

export const saveAuthData = (user, token) => {
  localStorage.setItem('user', JSON.stringify(user));
  localStorage.setItem('jwt_token', token);
};

export const clearAuthData = () => {
  localStorage.removeItem('user');
  localStorage.removeItem('jwt_token');
};

export const getSavedAuth = () => {
  const savedUser = localStorage.getItem('user');
  const savedToken = localStorage.getItem('jwt_token');
  
  if (savedUser && savedToken) {
    try {
      return { user: JSON.parse(savedUser), token: savedToken };
    } catch (e) {
      clearAuthData();
      return null;
    }
  }
  return null;
};
