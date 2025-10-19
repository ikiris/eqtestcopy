export const config = {
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || 'https://localhost:3000',
  oidcIssuer: import.meta.env.VITE_OIDC_ISSUER || 'https://localhost:8443',
  clientId: import.meta.env.VITE_CLIENT_ID || 'eqtestcopy-spa',
};
