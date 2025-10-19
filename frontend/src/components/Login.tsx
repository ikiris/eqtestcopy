import { initiateLogin } from '../auth/auth';

export function Login() {
  const handleLogin = () => {
    initiateLogin();
  };

  return (
    <div className="login">
      <h1>EQ Test Copy</h1>
      <p>Please log in to access your characters and upload inventory files.</p>
      <button onClick={handleLogin} className="login-button">
        Login with OIDC
      </button>
    </div>
  );
}
