// Get configuration from injected config or fallback to environment variables
function getConfig() {
  // Check if configuration was injected by the backend
  if (typeof window !== 'undefined' && (window as any).__APP_CONFIG__) {
    const injectedConfig = (window as any).__APP_CONFIG__;
    return {
      apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'https://localhost:3000',
      oidcIssuer: injectedConfig.oidcIssuer || 'https://localhost:8443',
      clientId: injectedConfig.clientId || 'eqtestcopy-spa',
    };
  }
  
  // Fallback to environment variables for development
  return {
    apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'https://localhost:3000',
    oidcIssuer: import.meta.env.VITE_OIDC_ISSUER || 'https://localhost:8443',
    clientId: import.meta.env.VITE_CLIENT_ID || 'eqtestcopy-spa',
  };
}

export const config = getConfig();
