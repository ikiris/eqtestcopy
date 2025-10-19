import { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './auth/AuthContext';
import { CharacterList } from './components/CharacterList';
import { InventoryUpload } from './components/InventoryUpload';
import { Login } from './components/Login';
import { Callback } from './components/Callback';
import type { CharacterData } from './generated/proto/eqtestcopy/eqtestcopy_pb';
import './App.css';

function AppContent() {
  const { isAuthenticated, logout, loading } = useAuth();
  const [selectedCharacter, setSelectedCharacter] = useState<CharacterData | null>(null);

  if (loading) {
    return <div className="app">Loading...</div>;
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>EQ Test Copy</h1>
        {isAuthenticated && (
          <button onClick={logout} className="logout-button">
            Logout
          </button>
        )}
      </header>

      <main className="app-main">
        {!isAuthenticated ? (
          <Login />
        ) : (
          <div className="main-content">
            <div className="content-grid">
              <div className="left-panel">
                <CharacterList onCharacterSelect={setSelectedCharacter} />
              </div>
              <div className="right-panel">
                <InventoryUpload 
                  selectedCharacter={selectedCharacter} 
                  onUploadComplete={() => {
                    // Optionally refresh character data or show success message
                    console.log('Upload completed');
                  }}
                />
              </div>
            </div>
          </div>
        )}
      </main>
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          <Route path="/callback" element={<Callback />} />
          <Route path="/" element={<AppContent />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;
