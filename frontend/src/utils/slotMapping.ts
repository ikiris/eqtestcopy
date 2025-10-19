// Slot mapping for TAKP inventory format to EQ slot IDs
// Based on the example inventory file structure
//
// Note: Some slots like "Ear", "Wrist", and "Fingers" can appear multiple times
// in TAKP inventory files. The backend handles this by assigning sequential
// slot IDs based on the order they appear (e.g., first "Ear" = slot 0, second "Ear" = slot 1)

export const SLOT_MAP: Record<string, number> = {
  // Equipment slots (0-22) - using proper TAKP slot names
  // EQ Slot ID Reference:
  // 0 = Charm, 1 = Ear 1, 2 = Head, 3 = Face, 4 = Ear 2, 5 = Neck, 6 = Shoulder,
  // 7 = Arms, 8 = Back, 9 = Bracer 1, 10 = Bracer 2, 11 = Range, 12 = Hands,
  // 13 = Primary, 14 = Secondary, 15 = Ring 1, 16 = Ring 2, 17 = Chest, 18 = Legs,
  // 19 = Feet, 20 = Waist, 21 = Powersource, 22 = Ammo
  'Charm': 0,      // Charm slot
  'Ear 1': 1,      // First earring slot
  'Head': 2,
  'Face': 3,
  'Ear 2': 4,      // Second earring slot
  'Neck': 5,
  'Shoulder': 6,   // Note: EQ uses "Shoulder", TAKP might use "Shoulders"
  'Arms': 7,
  'Back': 8,
  'Bracer 1': 9,   // First bracer slot
  'Bracer 2': 10,  // Second bracer slot
  'Range': 11,
  'Hands': 12,
  'Primary': 13,
  'Secondary': 14,
  'Ring 1': 15,    // First ring slot
  'Ring 2': 16,    // Second ring slot
  'Chest': 17,
  'Legs': 18,
  'Feet': 19,
  'Waist': 20,
  'Powersource': 21, // Powersource slot
  'Ammo': 22,      // Ammo slot
  // Legacy mappings for backward compatibility
  'Ear': 1,        // Maps to Ear 1 for compatibility
  'Wrist': 9,      // Maps to Bracer 1 for compatibility
  'Wrist2': 10,    // Maps to Bracer 2 for compatibility
  'Fingers': 15,   // Maps to Ring 1 for compatibility
  'Fingers2': 16,  // Maps to Ring 2 for compatibility
  'Shoulders': 6,  // Maps to Shoulder for compatibility
  'Held': 0,       // Maps to Charm for compatibility
  'General-Coin': 21, // Maps to Powersource for compatibility

  // General inventory slots (23-30) - starts after Ammo slot (22)
  'General1': 23,
  'General2': 24,
  'General3': 25,
  'General4': 26,
  'General5': 27,
  'General6': 28,
  'General7': 29,
  'General8': 30,

  // Bank slots (31-60) - starts after general inventory
  'Bank1': 31,
  'Bank2': 32,
  'Bank3': 33,
  'Bank4': 34,
  'Bank5': 35,
  'Bank6': 36,
  'Bank7': 37,
  'Bank8': 38,
  'Bank9': 39,
  'Bank10': 40,
  'Bank11': 41,
  'Bank12': 42,
  'Bank13': 43,
  'Bank14': 44,
  'Bank15': 45,
  'Bank16': 46,
  'Bank17': 47,
  'Bank18': 48,
  'Bank19': 49,
  'Bank20': 50,
  'Bank21': 51,
  'Bank22': 52,
  'Bank23': 53,
  'Bank24': 54,
  'Bank25': 55,
  'Bank26': 56,
  'Bank27': 57,
  'Bank28': 58,
  'Bank29': 59,
  'Bank30': 60,

  // Shared bank slots (61-90) - starts after bank slots
  'SharedBank1': 61,
  'SharedBank2': 62,
  'SharedBank3': 63,
  'SharedBank4': 64,
  'SharedBank5': 65,
  'SharedBank6': 66,
  'SharedBank7': 67,
  'SharedBank8': 68,
  'SharedBank9': 69,
  'SharedBank10': 70,
  'SharedBank11': 71,
  'SharedBank12': 72,
  'SharedBank13': 73,
  'SharedBank14': 74,
  'SharedBank15': 75,
  'SharedBank16': 76,
  'SharedBank17': 77,
  'SharedBank18': 78,
  'SharedBank19': 79,
  'SharedBank20': 80,
  'SharedBank21': 81,
  'SharedBank22': 82,
  'SharedBank23': 83,
  'SharedBank24': 84,
  'SharedBank25': 85,
  'SharedBank26': 86,
  'SharedBank27': 87,
  'SharedBank28': 88,
  'SharedBank29': 89,
  'SharedBank30': 90,

  // Bank coin slot
  'Bank-Coin': 91,
};

/**
 * Maps TAKP location string to EQ slot ID
 * Handles nested slots like "General1-Slot2"
 */
export function mapLocationToSlotId(location: string): number | null {
  // Handle nested slots (e.g., "General1-Slot2")
  if (location.includes('-')) {
    const [parentSlot, subSlot] = location.split('-', 2);
    const parentSlotId = SLOT_MAP[parentSlot];
    
    if (parentSlotId === undefined) {
      return null;
    }

    // For nested slots, we need to calculate the sub-slot ID
    // EQMacEmu uses the formula: (parent_slot_id * 10) + (sub_slot_number - 1)
    if (subSlot.startsWith('Slot')) {
      const slotNumber = parseInt(subSlot.replace('Slot', ''), 10);
      if (isNaN(slotNumber) || slotNumber < 1 || slotNumber > 10) {
        return null; // Invalid slot number
      }
      // EQMacEmu bag slot calculation: (parent_slot_id * 10) + (sub_slot_number - 1)
      return (parentSlotId * 10) + (slotNumber - 1);
    }
  }

  // Direct slot mapping
  return SLOT_MAP[location] ?? null;
}

/**
 * Checks if a location should be included in character inventory updates
 * Excludes bank and shared bank slots
 */
export function isCharacterInventorySlot(location: string): boolean {
  // Only include equipment slots and general inventory slots
  // Exclude bank slots, shared bank slots, and coin slots
  return !location.startsWith('Bank') && 
         !location.startsWith('SharedBank') && 
         !location.includes('-Coin');
}

/**
 * Gets all valid character inventory slot names
 */
export function getCharacterInventorySlots(): string[] {
  return Object.keys(SLOT_MAP).filter(isCharacterInventorySlot);
}
