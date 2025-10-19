import { create } from '@bufbuild/protobuf';
import { InventoryItemSchema, type InventoryItem } from '../generated/proto/eqtestcopy/eqtestcopy_pb';

export interface TAKPInventoryRow {
  location: string;
  name: string;
  id: number;
  count: number;
  slots: number;
}

export interface ParsedInventory {
  items: InventoryItem[];
  errors: string[];
  warnings: string[];
}

/**
 * Parses a TAKP inventory file content
 * @param content The tab-delimited content of the inventory file
 * @returns Parsed inventory with items and any errors/warnings
 */
export function parseTAKPInventory(content: string): ParsedInventory {
  const lines = content.trim().split('\n');
  const items: InventoryItem[] = [];
  const errors: string[] = [];
  const warnings: string[] = [];

  // Skip header row
  const dataLines = lines.slice(1);

  for (let i = 0; i < dataLines.length; i++) {
    const line = dataLines[i].trim();
    if (!line) continue;

    try {
      const parsedRow = parseInventoryRow(line);
      
      // Skip empty slots
      if (parsedRow.id === 0 || parsedRow.count === 0) {
        continue;
      }

      // Create inventory item - backend will handle location parsing and slot ID calculation
      const item = create(InventoryItemSchema, {
        slotId: 0, // Backend will calculate this from location
        itemId: parsedRow.id,
        charges: parsedRow.count,
        location: parsedRow.location, // Include location for backend slot mapping
      });

      items.push(item);

    } catch (error) {
      errors.push(`Failed to parse line ${i + 2}: ${error}`);
    }
  }

  return {
    items,
    errors,
    warnings,
  };
}

/**
 * Parses a single tab-delimited row from the inventory file
 */
function parseInventoryRow(line: string): TAKPInventoryRow {
  const parts = line.split('\t');
  
  if (parts.length < 5) {
    throw new Error(`Invalid row format: expected 5 columns, got ${parts.length}`);
  }

  const location = parts[0].trim();
  const name = parts[1].trim();
  const id = parseInt(parts[2].trim(), 10);
  const count = parseInt(parts[3].trim(), 10);
  const slots = parseInt(parts[4].trim(), 10);

  if (isNaN(id) || isNaN(count) || isNaN(slots)) {
    throw new Error(`Invalid numeric values: id=${parts[2]}, count=${parts[3]}, slots=${parts[4]}`);
  }

  return {
    location,
    name,
    id,
    count,
    slots,
  };
}

/**
 * Validates that the file content looks like a TAKP inventory file
 */
export function validateTAKPInventoryFormat(content: string): { isValid: boolean; error?: string } {
  const lines = content.trim().split('\n');
  
  if (lines.length < 2) {
    return { isValid: false, error: 'File must have at least a header and one data row' };
  }

  const header = lines[0].trim();
  const expectedHeader = 'Location\tName\tID\tCount\tSlots';
  
  if (header !== expectedHeader) {
    return { 
      isValid: false, 
      error: `Invalid header format. Expected: "${expectedHeader}", Got: "${header}"` 
    };
  }

  return { isValid: true };
}

/**
 * Converts parsed inventory items to a preview format for display
 */
export function createInventoryPreview(items: InventoryItem[], itemNames?: Map<number, string>): string {
  if (items.length === 0) {
    return 'No items to upload';
  }

  // Show all items, not just the first 10
  return items
    .map(item => {
      const itemName = itemNames?.get(item.itemId) || `Item ${item.itemId}`;
      return `Slot ${item.slotId}: ${itemName} (${item.charges} charges)`;
    })
    .join('\n');
}
