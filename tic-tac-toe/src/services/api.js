import axios from 'axios';

const getHostname = () => window.location.hostname;

export const API_BASE = `http://${getHostname()}:30041/api`;
export const IDENTITY_URL = `http://${getHostname()}:30000`;
export const IDENTITY_API = `http://${getHostname()}:3001/api`;

// Setup axios interceptor to add JWT token
axios.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('jwt_token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export default axios;
