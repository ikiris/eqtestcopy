import { useState, useRef } from 'react';
import { apiClient } from '../api/client';
import type { CharacterData, InventoryItem } from '../generated/proto/eqtestcopy/eqtestcopy_pb';

interface InventoryUploadProps {
  selectedCharacter: CharacterData | null;
  onUploadComplete: () => void;
}

export function InventoryUpload({ selectedCharacter, onUploadComplete }: InventoryUploadProps) {
  const [file, setFile] = useState<File | null>(null);
  const [parsedInventory, setParsedInventory] = useState<InventoryItem[]>([]);
  const [itemNames, setItemNames] = useState<Map<number, string>>(new Map());
  const [errors, setErrors] = useState<string[]>([]);
  const [warnings, setWarnings] = useState<string[]>([]);
  const [uploading, setUploading] = useState(false);
  const [uploadSuccess, setUploadSuccess] = useState<string | null>(null);
  const [parsing, setParsing] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileSelect = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = event.target.files?.[0];
    if (!selectedFile) return;

    setFile(selectedFile);
    setParsedInventory([]);
    setErrors([]);
    setWarnings([]);
    setUploadSuccess(null);

    // Read file content and send to backend for parsing
    const reader = new FileReader();
    reader.onload = async (e) => {
      const content = e.target?.result as string;
      await parseInventoryFile(content);
    };
    reader.readAsText(selectedFile);
  };


  const parseInventoryFile = async (content: string) => {
    setParsing(true);
    setErrors([]);
    setWarnings([]);
    
    try {
      // Send raw file content to backend for secure parsing
      const response = await apiClient.parseInventoryFile({
        fileContent: content,
      });

      if (response.success) {
        // Convert item names map to Map
        const itemNamesMap = new Map<number, string>();
        Object.entries(response.itemNames).forEach(([id, name]) => {
          itemNamesMap.set(parseInt(id), name);
        });

        setParsedInventory(response.items);
        setItemNames(itemNamesMap);
        setErrors(response.errors);
        setWarnings(response.warnings);
      } else {
        setErrors([response.message]);
      }
    } catch (err) {
      setErrors([err instanceof Error ? err.message : 'Failed to parse inventory file']);
    } finally {
      setParsing(false);
    }
  };

  const handleUpload = async () => {
    if (!selectedCharacter || parsedInventory.length === 0) {
      setErrors(['Please select a character and parse a valid inventory file first']);
      return;
    }

    setUploading(true);
    setErrors([]);
    setUploadSuccess(null);

    try {
      const response = await apiClient.updateInventory({
        characterId: selectedCharacter.id,
        inventories: parsedInventory,
      });

      if (response.success) {
        setUploadSuccess(response.message);
        onUploadComplete();
      } else {
        setErrors([response.message]);
      }
    } catch (err) {
      setErrors([err instanceof Error ? err.message : 'Upload failed']);
    } finally {
      setUploading(false);
    }
  };

  const clearFile = () => {
    setFile(null);
    setParsedInventory([]);
    setItemNames(new Map());
    setErrors([]);
    setWarnings([]);
    setUploadSuccess(null);
    if (fileInputRef.current) {
      fileInputRef.current.value = '';
    }
  };

  if (!selectedCharacter) {
    return (
      <div className="inventory-upload">
        <h2>Inventory Upload</h2>
        <p>Please select a character first to upload inventory.</p>
      </div>
    );
  }

  return (
    <div className="inventory-upload">
      <h2>Upload Inventory for {selectedCharacter.name}</h2>
      
      <div className="upload-section">
        <div className="file-input">
          <label htmlFor="inventory-file">Select TAKP Inventory File:</label>
          <input
            ref={fileInputRef}
            id="inventory-file"
            type="file"
            accept=".txt"
            onChange={handleFileSelect}
            disabled={uploading}
          />
          {file && (
            <button onClick={clearFile} disabled={uploading}>
              Clear File
            </button>
          )}
        </div>

        {file && (
          <div className="file-info">
            <p><strong>File:</strong> {file.name}</p>
          </div>
        )}

        {parsing && (
          <div className="parsing-status">
            <h3>Parsing inventory file...</h3>
            <p>Please wait while we securely parse your inventory file on the server.</p>
          </div>
        )}

        {parsedInventory.length > 0 && !parsing && (
          <div className="inventory-preview">
            <h3>Parsed Inventory ({parsedInventory.length} items)</h3>
            <div className="inventory-items">
              {parsedInventory.map((item, index) => {
                const itemName = itemNames.get(item.itemId) || `Item ${item.itemId}`;
                return (
                  <div key={index} className="inventory-item">
                    <span className="slot">{item.slotName || `Slot ${item.slotId}`}</span>
                    <span className="item-name">{itemName}</span>
                    <span className="charges">({item.charges} charges)</span>
                    {item.location && <span className="location">from {item.location}</span>}
                  </div>
                );
              })}
            </div>
            
            <button 
              onClick={handleUpload} 
              disabled={uploading || errors.length > 0}
              className="upload-button"
            >
              {uploading ? 'Uploading...' : 'Upload Inventory'}
            </button>
          </div>
        )}

        {errors.length > 0 && (
          <div className="errors">
            <h3>Errors:</h3>
            <ul>
              {errors.map((error, index) => (
                <li key={index} className="error">{error}</li>
              ))}
            </ul>
          </div>
        )}

        {warnings.length > 0 && (
          <div className="warnings">
            <h3>Warnings:</h3>
            <ul>
              {warnings.map((warning, index) => (
                <li key={index} className="warning">{warning}</li>
              ))}
            </ul>
          </div>
        )}

        {uploadSuccess && (
          <div className="success">
            <h3>Success!</h3>
            <p>{uploadSuccess}</p>
          </div>
        )}
      </div>
    </div>
  );
}
