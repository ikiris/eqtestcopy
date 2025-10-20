// Get configuration from injected config or fallback to environment variables
function getConfig() {
  // Check if configuration was injected by the backend
  if (typeof window !== 'undefined' && (window as any).__APP_CONFIG__) {
    const injectedConfig = (window as any).__APP_CONFIG__;
    return {
      apiBaseUrl: injectedConfig.apiBaseUrl || '', // Use relative URLs when served by backend
      oidcIssuer: injectedConfig.oidcIssuer || import.meta.env.VITE_OIDC_ISSUER || '',
      clientId: injectedConfig.clientId || import.meta.env.VITE_CLIENT_ID || 'eqtestcopy-spa',
    };
  }
  
  // Fallback to environment variables (only for OIDC config, API uses relative URLs)
  return {
    apiBaseUrl: '', // Empty string means relative URLs
    oidcIssuer: import.meta.env.VITE_OIDC_ISSUER || '',
    clientId: import.meta.env.VITE_CLIENT_ID || 'eqtestcopy-spa',
  };
}

export const config = getConfig();
