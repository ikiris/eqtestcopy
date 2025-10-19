package eqtestcopy

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"connectrpc.com/connect"

	"github.com/ikiris/eqmaclib/eqdb"
	pb "github.com/ikiris/eqtestcopy/proto/eqtestcopy"
)

// TAKP slot name translation map
var takpSlotNameMap = map[string]string{
	"Bracer":  "Wrist",
	"Fingers": "Finger",
}

// Ambiguous slots that need order-based numbering
var ambiguousSlots = map[string]struct{}{
	"Ear":    struct{}{},
	"Wrist":  struct{}{},
	"Finger": struct{}{},
}

// GetCharacter retrieves a character by ID
func (s *server) GetCharacter(ctx context.Context, req *connect.Request[pb.GetCharacterRequest]) (*connect.Response[pb.CharacterData], error) {
	// Get character from database

	accountIDStr, ok := ctx.Value(contextKey("account_id")).(string)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("account ID not found in context"))
	}

	// Convert string to int32
	var accountID int32
	if _, err := fmt.Sscanf(accountIDStr, "%d", &accountID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid account ID format: %v", err))
	}

	char, inventory, err := s.db.GetCharacter(ctx, accountID, req.Msg.CharacterId)
	if err != nil {
		slog.Error("failed to get character", "error", err, "character_id", req.Msg.CharacterId)

		return nil, connect.NewError(connect.CodeNotFound, fmt.Errorf("character not found"))
	}

	// Convert to protobuf format
	response := &pb.CharacterData{
		Id:     char.ID,
		Name:   char.Name,
		Race:   char.Race,
		Class:  char.Class,
		Level:  char.Level,
		ZoneId: char.ZoneID,
		Stats: &pb.Stat{
			Hp:    char.Stats.HP,
			Mana:  char.Stats.Mana,
			Str:   char.Stats.Str,
			Sta:   char.Stats.Sta,
			Agi:   char.Stats.Agi,
			Dex:   char.Stats.Dex,
			Wis:   char.Stats.Wis,
			Intel: char.Stats.Int,
			Cha:   char.Stats.Cha,
		},
	}

	// Convert inventory items
	for _, item := range inventory {
		response.Inventories = append(response.Inventories, &pb.InventoryItem{
			SlotId:   item.SlotID,
			ItemId:   item.ItemID,
			Charges:  item.Charges,
			Location: item.Location,
			SlotName: GetFriendlySlotNameByID(item.SlotID),
		})
	}

	return connect.NewResponse(response), nil
}

// ListCharacters retrieves characters for a user with pagination
func (s *server) ListCharacters(ctx context.Context, req *connect.Request[pb.ListCharactersRequest]) (*connect.Response[pb.ListCharactersResponse], error) {
	// Get user ID from context
	accountIDStr, ok := ctx.Value(contextKey("account_id")).(string)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("user ID not found in context"))
	}

	// Convert string to int32
	var accountID int32
	if _, err := fmt.Sscanf(accountIDStr, "%d", &accountID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid user ID format: %v", err))
	}

	// Query database
	listReq := eqdb.ListCharactersRequest{
		AccountID: accountID,
	}

	characters, err := s.db.ListCharacters(ctx, listReq)
	if err != nil {
		slog.Error("failed to list characters", "error", err, "account_id", accountID)

		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to list characters"))
	}

	// Convert to protobuf format
	var pbCharacters []*pb.CharacterData
	for _, char := range characters {
		pbChar := &pb.CharacterData{
			Id:     char.ID,
			Name:   char.Name,
			Race:   char.Race,
			Class:  char.Class,
			Level:  char.Level,
			ZoneId: char.ZoneID,
		}
		pbCharacters = append(pbCharacters, pbChar)
	}

	return connect.NewResponse(&pb.ListCharactersResponse{
		Characters: pbCharacters,
	}), nil
}

// UpdateInventory updates a character's inventory
func (s *server) UpdateInventory(ctx context.Context, req *connect.Request[pb.UpdateInventoryRequest]) (*connect.Response[pb.UpdateInventoryResponse], error) {
	// Get account ID from context
	accountIDStr, ok := ctx.Value(contextKey("account_id")).(string)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("account ID not found in context"))
	}

	// Convert string to int32
	var accountID int32
	if _, err := fmt.Sscanf(accountIDStr, "%d", &accountID); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("invalid account ID format: %v", err))
	}

	// Convert protobuf inventory to database format
	var inventory eqdb.Inventory
	slotUsage := make(map[string]int) // Track slot usage for duplicate slot names

	for _, item := range req.Msg.Inventories {
		// Calculate slot ID from location string (security: backend controls slot calculation)
		var slotID int32
		if item.Location != "" {
			// Use location string to calculate slot ID with order tracking
			calculatedSlotID, err := MapLocationToSlotIdWithOrder(item.Location, slotUsage)
			if err != nil {
				slog.Error("Failed to map location to slot ID", "error", err, "location", item.Location)
				return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid location: %s", item.Location))
			}
			slotID = calculatedSlotID
		} else {
			// Fallback to provided slot ID (for backward compatibility)
			slotID = item.SlotId
		}

		// Validate that this is a character inventory slot (not bank/shared bank)
		if !IsCharacterInventorySlot(item.Location) {
			slog.Warn("Skipping non-character inventory slot", "location", item.Location)
			continue
		}

		inventory = append(inventory, eqdb.InventoryItem{
			SlotID:   slotID,
			ItemID:   item.ItemId,
			Charges:  item.Charges,
			Location: item.Location,
		})
	}

	// Update inventory in database
	err := s.db.UpdateInventory(ctx, req.Msg.CharacterId, accountID, inventory)
	if err != nil {
		slog.Error("failed to update inventory", "error", err, "character_id", req.Msg.CharacterId, "account_id", accountID)

		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to update inventory: %v", err))
	}

	return connect.NewResponse(&pb.UpdateInventoryResponse{
		Success: true,
		Message: "Inventory updated successfully",
	}), nil
}

// GetItemNames retrieves item names for the given item IDs
func (s *server) GetItemNames(ctx context.Context, req *connect.Request[pb.GetItemNamesRequest]) (*connect.Response[pb.GetItemNamesResponse], error) {
	// Get item names from database
	itemNames, err := s.db.GetItemNames(ctx, req.Msg.ItemIds)
	if err != nil {
		slog.Error("Failed to get item names", "error", err, "item_ids", req.Msg.ItemIds)
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to get item names: %v", err))
	}

	return connect.NewResponse(&pb.GetItemNamesResponse{
		ItemNames: itemNames,
	}), nil
}

// ParseInventoryFile parses a TAKP inventory file and returns structured data for approval
func (s *server) ParseInventoryFile(ctx context.Context, req *connect.Request[pb.ParseInventoryFileRequest]) (*connect.Response[pb.ParseInventoryFileResponse], error) {
	content := req.Msg.FileContent
	if content == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("file content cannot be empty"))
	}

	// Parse the inventory file content
	items, errors, warnings, err := s.parseTAKPInventoryContent(content)
	if err != nil {
		slog.Error("Failed to parse inventory file", "error", err)
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("failed to parse inventory file: %v", err))
	}

	// Get item names for display
	var itemIds []int32
	for _, item := range items {
		itemIds = append(itemIds, item.ItemID)
	}

	itemNames, err := s.db.GetItemNames(ctx, itemIds)
	if err != nil {
		slog.Warn("Failed to get item names", "error", err)
		// Don't fail the request, just continue without item names
		itemNames = make(map[int32]string)
	}

	// Convert to protobuf format
	var pbItems []*pb.InventoryItem
	for _, item := range items {
		pbItems = append(pbItems, &pb.InventoryItem{
			SlotId:   item.SlotID,
			ItemId:   item.ItemID,
			Charges:  item.Charges,
			Location: item.Location,
			SlotName: GetFriendlySlotNameByID(item.SlotID),
		})
	}

	response := &pb.ParseInventoryFileResponse{
		Success:   len(errors) == 0,
		Message:   fmt.Sprintf("Parsed %d items", len(pbItems)),
		Items:     pbItems,
		Errors:    errors,
		Warnings:  warnings,
		ItemNames: itemNames,
	}

	return connect.NewResponse(response), nil
}

// parseTAKPInventoryContent parses TAKP inventory file content and returns structured data
func (s *server) parseTAKPInventoryContent(content string) ([]eqdb.InventoryItem, []string, []string, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		return nil, nil, nil, fmt.Errorf("file must have at least a header and one data row")
	}

	// Validate header
	header := strings.TrimSpace(lines[0])
	expectedHeader := "Location\tName\tID\tCount\tSlots"
	if header != expectedHeader {
		return nil, nil, nil, fmt.Errorf("invalid header format. Expected: %q, Got: %q", expectedHeader, header)
	}

	var items []eqdb.InventoryItem
	var errors []string
	var warnings []string

	// Track slot usage for handling duplicate slot names (like "Ear")
	slotUsage := make(map[string]int)

	// Parse data lines
	for i, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 5 {
			errors = append(errors, fmt.Sprintf("Line %d: Invalid row format: expected 5 columns, got %d", i+2, len(parts)))
			continue
		}

		location := strings.TrimSpace(parts[0])

		// Translate TAKP slot names to our names
		if translatedName, exists := takpSlotNameMap[location]; exists {
			location = translatedName
		}

		if _, exists := ambiguousSlots[location]; exists {
			// Check usage count for this ambiguous slot
			slotUsage[location]++
			location = location + strconv.Itoa(slotUsage[location])
		}

		_ = strings.TrimSpace(parts[1]) // name - not used but keep for completeness
		idStr := strings.TrimSpace(parts[2])
		countStr := strings.TrimSpace(parts[3])
		slotsStr := strings.TrimSpace(parts[4])

		// Parse numeric values
		id, err := strconv.ParseInt(idStr, 10, 32)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Line %d: Invalid item ID: %s", i+2, idStr))
			continue
		}

		count, err := strconv.ParseInt(countStr, 10, 32)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Line %d: Invalid count: %s", i+2, countStr))
			continue
		}

		_, err = strconv.ParseInt(slotsStr, 10, 32)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Line %d: Invalid slots: %s", i+2, slotsStr))
			continue
		}

		// Skip empty slots
		if id == 0 || count == 0 {
			continue
		}

		// Validate that this is a character inventory slot (not bank/shared bank) first
		if !IsCharacterInventorySlot(location) {
			warnings = append(warnings, fmt.Sprintf("Line %d: Skipping non-character inventory slot: %s", i+2, location))
			continue
		}

		// Calculate slot ID from location with order tracking
		slotID, err := MapLocationToSlotIdWithOrder(location, slotUsage)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Line %d: Invalid location %q: %v", i+2, location, err))
			continue
		}

		items = append(items, eqdb.InventoryItem{
			SlotID:   slotID,
			ItemID:   int32(id),
			Charges:  int32(count),
			Location: location,
		})
	}

	return items, errors, warnings, nil
}
