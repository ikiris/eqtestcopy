import { config } from '../config';

// PKCE utilities
function generateCodeVerifier(): string {
  const array = new Uint8Array(32);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

async function generateCodeChallenge(verifier: string): Promise<string> {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return btoa(String.fromCharCode(...new Uint8Array(digest)))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

// OAuth2/OIDC flow
export async function initiateLogin(): Promise<void> {
  const codeVerifier = generateCodeVerifier();
  const codeChallenge = await generateCodeChallenge(codeVerifier);
  
  // Store code verifier for later use
  sessionStorage.setItem('code_verifier', codeVerifier);
  
  // Build authorization URL
  const authUrl = new URL('/auth', config.oidcIssuer);
  authUrl.searchParams.set('response_type', 'code');
  authUrl.searchParams.set('client_id', config.clientId);
  authUrl.searchParams.set('redirect_uri', window.location.origin + '/callback');
  authUrl.searchParams.set('scope', 'openid profile email');
  authUrl.searchParams.set('code_challenge', codeChallenge);
  authUrl.searchParams.set('code_challenge_method', 'S256');
  authUrl.searchParams.set('state', generateState());
  
  // Redirect to authorization server
  window.location.href = authUrl.toString();
}

function generateState(): string {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  return btoa(String.fromCharCode(...array))
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '');
}

export async function handleCallback(): Promise<boolean> {
  console.log('Starting callback handling...');
  const urlParams = new URLSearchParams(window.location.search);
  const code = urlParams.get('code');
  // const state = urlParams.get('state'); // TODO: Validate state parameter
  const error = urlParams.get('error');
  
  console.log('URL params:', { code: code ? 'present' : 'missing', error });
  
  if (error) {
    console.error('OAuth error:', error);
    return false;
  }
  
  if (!code) {
    console.error('No authorization code received');
    return false;
  }
  
  const codeVerifier = sessionStorage.getItem('code_verifier');
  console.log('Code verifier:', codeVerifier ? 'present' : 'missing');
  if (!codeVerifier) {
    console.error('No code verifier found');
    return false;
  }
  
  try {
    const tokenUrl = `${config.oidcIssuer}/token`;
    const tokenBody = new URLSearchParams({
      grant_type: 'authorization_code',
      client_id: config.clientId,
      code,
      redirect_uri: window.location.origin + '/callback',
      code_verifier: codeVerifier,
    });
    
    console.log('Making token exchange request to:', tokenUrl);
    console.log('Request body:', tokenBody.toString());
    
    // Exchange code for tokens
    const tokenResponse = await fetch(tokenUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: tokenBody,
    });
    
    console.log('Token response status:', tokenResponse.status, tokenResponse.statusText);
    
    if (!tokenResponse.ok) {
      const errorText = await tokenResponse.text();
      console.error('Token exchange failed:', errorText);
      return false;
    }
    
    const tokens = await tokenResponse.json();
    
    // Store ID token (needed for API authentication)
    sessionStorage.setItem('access_token', tokens.id_token);
    
    // Clean up
    sessionStorage.removeItem('code_verifier');
    
    return true;
  } catch (error) {
    console.error('Token exchange error:', error);
    return false;
  }
}

export function logout(): void {
  sessionStorage.removeItem('access_token');
  window.location.href = '/';
}

export function isAuthenticated(): boolean {
  return !!sessionStorage.getItem('access_token');
}
