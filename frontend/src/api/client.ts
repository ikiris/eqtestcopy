import { createConnectTransport } from '@connectrpc/connect-web';
import { createClient } from '@connectrpc/connect';
import { EqTestCopyService } from '../generated/proto/eqtestcopy/eqtestcopy_pb';
import { config } from '../config';

// Create transport with auth interceptor
const transport = createConnectTransport({
  baseUrl: config.apiBaseUrl,
  interceptors: [
    (next) => async (req) => {
      // Add authorization header if token exists
      const token = getAuthToken();
      if (token) {
        req.header.set('authorization', `Bearer ${token}`);
      }
      return next(req);
    },
  ],
});

// Create the client exactly like the documentation example
export const apiClient = createClient(EqTestCopyService, transport);

// Simple token storage (in a real app, you'd use proper auth state management)
function getAuthToken(): string | null {
  return sessionStorage.getItem('access_token');
}

export function setAuthToken(token: string): void {
  sessionStorage.setItem('access_token', token);
}

export function clearAuthToken(): void {
  sessionStorage.removeItem('access_token');
}
