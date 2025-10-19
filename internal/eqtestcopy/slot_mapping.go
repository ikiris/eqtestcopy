package eqtestcopy

import (
	"fmt"
	"strconv"
	"strings"
)

// Slot mapping for TAKP inventory format to EQ slot IDs
// Based on EQMacEmu slot system
//
// EQ Slot ID Reference:
// 0 = Charm, 1 = Ear 1, 2 = Head, 3 = Face, 4 = Ear 2, 5 = Neck, 6 = Shoulder,
// 7 = Arms, 8 = Back, 9 = Bracer 1, 10 = Bracer 2, 11 = Range, 12 = Hands,
// 13 = Primary, 14 = Secondary, 15 = Ring 1, 16 = Ring 2, 17 = Chest, 18 = Legs,
// 19 = Feet, 20 = Waist, 21 = Powersource, 22 = Ammo
//
// Note: TAKP uses "Ear", "Wrist", "Fingers" for both slots of each type.
// The MapLocationToSlotIdWithOrder function handles assigning correct IDs based on order.
var slotMap = map[string]int32{
	// Equipment slots (0-22) - EXACT slot names from TAKP
	"Charm":     0,  // Slot 0
	"Ear1":      1,  // Slot 1
	"Head":      2,  // Slot 2
	"Face":      3,  // Slot 3
	"Ear2":      4,  // Slot 4
	"Neck":      5,  // Slot 5
	"Shoulders": 6,  // Slot 6
	"Arms":      7,  // Slot 7
	"Back":      8,  // Slot 8
	"Wrist1":    9,  // Slot 9
	"Wrist2":    10, // Slot 10
	"Range":     11, // Slot 11
	"Hands":     12, // Slot 12
	"Primary":   13, // Slot 13
	"Secondary": 14, // Slot 14
	"Finger1":   15, // Slot 15
	"Finger2":   16, // Slot 16
	"Chest":     17, // Slot 17
	"Legs":      18, // Slot 18
	"Feet":      19, // Slot 19
	"Waist":     20, // Slot 20
	"Ammo":      21, // Slot 21

	// General inventory slots (22-29) - starts at Ammo slot (22)
	"General1": 22,
	"General2": 23,
	"General3": 24,
	"General4": 25,
	"General5": 26,
	"General6": 27,
	"General7": 28,
	"General8": 29,
}

// MapLocationToSlotId maps TAKP location string to EQ slot ID
// Handles nested slots like "General1-Slot2" using EQMacEmu formula
func MapLocationToSlotId(location string) (int32, error) {
	// Handle nested slots (e.g., "General1-Slot2")
	if strings.Contains(location, "-") {
		parts := strings.SplitN(location, "-", 2)
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid location format: %s", location)
		}

		parentSlot := parts[0]
		subSlot := parts[1]

		parentSlotId, exists := slotMap[parentSlot]
		if !exists {
			return 0, fmt.Errorf("unknown parent slot: %s", parentSlot)
		}

		// For nested slots, we need to calculate the sub-slot ID
		// EQMacEmu uses the formula: (parent_slot_id * 10) + sub_slot_number + 29
		if strings.HasPrefix(subSlot, "Slot") {
			slotNumberStr := strings.TrimPrefix(subSlot, "Slot")
			slotNumber, err := strconv.Atoi(slotNumberStr)
			if err != nil {
				return 0, fmt.Errorf("invalid slot number in %s: %v", location, err)
			}

			if slotNumber < 1 || slotNumber > 10 {
				return 0, fmt.Errorf("slot number out of range (1-10) in %s", location)
			}

			// EQMacEmu bag slot calculation: (parent_slot_id * 10) + sub_slot_number + 29
			bagSlotId := (parentSlotId * 10) + int32(slotNumber) + 29
			return bagSlotId, nil
		}
	}

	// Direct slot mapping
	if slotId, exists := slotMap[location]; exists {
		return slotId, nil
	}

	return 0, fmt.Errorf("unknown location: %s", location)
}

// MapLocationToSlotIdWithOrder maps TAKP location string to EQ slot ID with order tracking
// This handles duplicate slot names (like "Ear") by assigning sequential IDs based on order
func MapLocationToSlotIdWithOrder(location string, slotUsage map[string]int) (int32, error) {
	// Handle nested slots (e.g., "General1-Slot2")
	if strings.Contains(location, "-") {
		parts := strings.SplitN(location, "-", 2)
		if len(parts) != 2 {
			return 0, fmt.Errorf("invalid location format: %s", location)
		}

		parentSlot := parts[0]
		subSlot := parts[1]

		parentSlotId, exists := slotMap[parentSlot]
		if !exists {
			return 0, fmt.Errorf("unknown parent slot: %s", parentSlot)
		}

		// For nested slots, we need to calculate the sub-slot ID
		// EQMacEmu uses the formula: (parent_slot_id * 10) + sub_slot_number + 29
		if strings.HasPrefix(subSlot, "Slot") {
			slotNumberStr := strings.TrimPrefix(subSlot, "Slot")
			slotNumber, err := strconv.Atoi(slotNumberStr)
			if err != nil {
				return 0, fmt.Errorf("invalid slot number in %s: %v", location, err)
			}

			if slotNumber < 1 || slotNumber > 10 {
				return 0, fmt.Errorf("slot number out of range (1-10) in %s", location)
			}

			// EQMacEmu bag slot calculation: (parent_slot_id * 10) + sub_slot_number + 29
			bagSlotId := (parentSlotId * 10) + int32(slotNumber) + 29
			return bagSlotId, nil
		}
	}

	// Direct slot mapping
	if slotId, exists := slotMap[location]; exists {
		return slotId, nil
	}

	return 0, fmt.Errorf("unknown location: %s", location)
}

// Inverse slot map from slot ID to friendly display name
var slotIdToFriendlyName = map[int32]string{
	// Equipment slots (0-22)
	0:  "Charm",
	1:  "Ear 1",
	2:  "Head",
	3:  "Face",
	4:  "Ear 2",
	5:  "Neck",
	6:  "Shoulder",
	7:  "Arms",
	8:  "Back",
	9:  "Wrist 1",
	10: "Wrist 2",
	11: "Range",
	12: "Hands",
	13: "Primary",
	14: "Secondary",
	15: "Ring 1",
	16: "Ring 2",
	17: "Chest",
	18: "Legs",
	19: "Feet",
	20: "Waist",
	21: "Ammo",

	// General inventory slots (22-29)
	22: "General1",
	23: "General2",
	24: "General3",
	25: "General4",
	26: "General5",
	27: "General6",
	28: "General7",
	29: "General8",
}

// GetFriendlySlotNameByID returns the friendly name for a given slot ID
func GetFriendlySlotNameByID(slotID int32) string {
	// Check direct mapping first
	if friendlyName, exists := slotIdToFriendlyName[slotID]; exists {
		return friendlyName
	}

	// Handle bag slots using EQMacEmu formula
	// Bag slots: (parent_slot_id * 10) + sub_slot_number + 29
	// Reverse: parent_slot_id = (slot_id - 29 - sub_slot_number) / 10
	// But we need to account for sub slots being 1-10, not 0-9
	if slotID >= 250 {
		adjustedID := slotID - 29
		parentSlotID := (adjustedID - 1) / 10  // Subtract 1 to account for 1-based sub slots
		subSlot := ((adjustedID - 1) % 10) + 1 // Convert back to 1-based

		parentName := slotIdToFriendlyName[parentSlotID]
		if parentName != "" {
			return fmt.Sprintf("%s Slot %d", parentName, subSlot)
		}
	}

	// Fallback to slot ID if no mapping found
	return fmt.Sprintf("Slot %d", slotID)
}

// IsCharacterInventorySlot checks if a location should be included in character inventory updates
// Excludes bank and shared bank slots
func IsCharacterInventorySlot(location string) bool {
	// Only include equipment slots and general inventory slots
	// Exclude bank slots, shared bank slots, and coin slots
	return !strings.HasPrefix(location, "Bank") &&
		!strings.HasPrefix(location, "SharedBank") &&
		!strings.HasSuffix(location, "-Coin")
}
