import { useEffect, useState } from 'react';
import { handleCallback } from '../auth/auth';

export function Callback() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const processCallback = async () => {
      try {
        const success = await handleCallback();
        if (success) {
          // Redirect to main app
          window.location.href = '/';
        } else {
          setError('Authentication failed');
        }
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Authentication error');
      } finally {
        setLoading(false);
      }
    };

    processCallback();
  }, []);

  if (loading) {
    return (
      <div className="callback">
        <h2>Completing login...</h2>
        <p>Please wait while we complete your authentication.</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="callback">
        <h2>Login Failed</h2>
        <p>Error: {error}</p>
        <button onClick={() => window.location.href = '/'}>
          Return to Login
        </button>
      </div>
    );
  }

  return (
    <div className="callback">
      <h2>Login Successful</h2>
      <p>Redirecting to the application...</p>
    </div>
  );
}
