import { useEffect, useState } from 'react';
import { apiClient } from '../api/client';
import type { CharacterData } from '../generated/proto/eqtestcopy/eqtestcopy_pb';

interface CharacterListProps {
  onCharacterSelect: (character: CharacterData) => void;
}

export function CharacterList({ onCharacterSelect }: CharacterListProps) {
  const [characters, setCharacters] = useState<CharacterData[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadCharacters();
  }, []);

  const loadCharacters = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await apiClient.listCharacters({});
      setCharacters(response.characters);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load characters');
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="character-list">
        <h2>Your Characters</h2>
        <p>Loading characters...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="character-list">
        <h2>Your Characters</h2>
        <div className="error">
          <p>Error: {error}</p>
          <button onClick={loadCharacters}>Retry</button>
        </div>
      </div>
    );
  }

  if (characters.length === 0) {
    return (
      <div className="character-list">
        <h2>Your Characters</h2>
        <p>No characters found.</p>
      </div>
    );
  }

  return (
    <div className="character-list">
      <h2>Your Characters</h2>
      <div className="character-grid">
        {characters.map((character) => (
          <div 
            key={character.id} 
            className="character-card"
            onClick={() => onCharacterSelect(character)}
          >
            <h3>{character.name}</h3>
            <div className="character-info">
              <p><strong>Level:</strong> {character.level}</p>
              <p><strong>Race:</strong> {character.race}</p>
              <p><strong>Class:</strong> {character.class}</p>
              <p><strong>Zone:</strong> {character.zoneId}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
