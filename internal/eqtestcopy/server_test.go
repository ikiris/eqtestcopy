package eqtestcopy

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"connectrpc.com/connect"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-cmp/cmp"
	pb "github.com/ikiris/eqtestcopy/proto/eqtestcopy"
	pbconnect "github.com/ikiris/eqtestcopy/proto/eqtestcopy/eqtestcopyconnect"
	"google.golang.org/protobuf/testing/protocmp"
	_ "modernc.org/sqlite"
)

// createMockVerifier creates a mock OIDC verifier that validates our test tokens
func createMockVerifier() *oidc.IDTokenVerifier {
	verifier := &oidc.IDTokenVerifier{}
	return verifier
}

// testAuthInterceptor is a test interceptor that extracts user ID from test tokens
func testAuthInterceptor(next connect.UnaryFunc) connect.UnaryFunc {
	return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		authHeader := req.Header().Get("Authorization")
		if authHeader == "" {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing authorization header"))
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid authorization header format"))
		}

		parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})
		if err != nil || !parsedToken.Valid {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token"))
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token claims"))
		}

		userID := fmt.Sprintf("%v", claims["sub"])
		if userID == "" {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing user ID in token"))
		}

		ctx = context.WithValue(ctx, contextKey("account_id"), userID)
		return next(ctx, req)
	})
}

func TestParseInventoryFile(t *testing.T) {
	t.Parallel()

	server, db := setupTestServer(t)
	defer server.Close()
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close database: %v", err)
		}
	}()

	tests := []struct {
		name           string
		filename       string
		expectedStatus int
		expectedItems  []*pb.InventoryItem
	}{
		// {
		// 	name:           "ValidWithEarrings",
		// 	filename:       "inventory_with_earrings.txt",
		// 	expectedStatus: http.StatusOK,
		// 	expectedItems: []*pb.InventoryItem{
		// 		{SlotId: 1, ItemId: 12345, Charges: 1},   // Ear1 - Silver Earring
		// 		{SlotId: 2, ItemId: 23456, Charges: 1},   // Head - Iron Helm
		// 		{SlotId: 4, ItemId: 34567, Charges: 1},   // Ear2 - Gold Earring
		// 		{SlotId: 5, ItemId: 45678, Charges: 1},   // Neck - Silver Necklace
		// 		{SlotId: 9, ItemId: 56789, Charges: 1},   // Wrist1 - Leather Bracer
		// 		{SlotId: 10, ItemId: 67890, Charges: 1},  // Wrist2 - Iron Bracer
		// 		{SlotId: 15, ItemId: 78901, Charges: 1},  // Finger1 - Silver Ring
		// 		{SlotId: 16, ItemId: 89012, Charges: 1},  // Finger2 - Gold Ring
		// 		{SlotId: 23, ItemId: 99999, Charges: 1},  // General1 - Bag
		// 		{SlotId: 230, ItemId: 11111, Charges: 5}, // General1-Slot1 - Health Potion (23*10 + 1-1 = 230)
		// 		{SlotId: 231, ItemId: 22222, Charges: 3}, // General1-Slot2 - Mana Potion (23*10 + 2-1 = 231)
		// 	},
		// },
		// {
		// 	name:           "ValidAmbiguousSlots",
		// 	filename:       "inventory_ambiguous_slots.txt",
		// 	expectedStatus: http.StatusOK,
		// 	expectedItems: []*pb.InventoryItem{
		// 		{SlotId: 1, ItemId: 11111, Charges: 1},  // Ear - First Earring (first occurrence -> Ear1)
		// 		{SlotId: 4, ItemId: 22222, Charges: 1},  // Ear - Second Earring (second occurrence -> Ear2)
		// 		{SlotId: 9, ItemId: 33333, Charges: 1},  // Wrist - First Bracer (first occurrence -> Wrist1)
		// 		{SlotId: 10, ItemId: 44444, Charges: 1}, // Wrist - Second Bracer (second occurrence -> Wrist2)
		// 		{SlotId: 15, ItemId: 55555, Charges: 1}, // Finger - First Ring (first occurrence -> Finger1)
		// 		{SlotId: 16, ItemId: 66666, Charges: 1}, // Finger - Second Ring (second occurrence -> Finger2)
		// 	},
		// },
		// {
		// 	name:           "InvalidFormat",
		// 	filename:       "inventory_invalid_format.txt",
		// 	expectedStatus: http.StatusBadRequest,
		// 	expectedItems:  nil,
		// },
		{
			name:           "BigXevInventory",
			filename:       "inventory_big_xev.txt",
			expectedStatus: http.StatusOK,
			expectedItems: []*pb.InventoryItem{
				// Equipment slots from database dump
				{SlotId: 1, ItemId: 14702, Charges: 1, Location: "Ear1", SlotName: "Ear 1"},         // Ear1 - Golden Black Sapphire Earring
				{SlotId: 2, ItemId: 1232, Charges: 1, Location: "Head", SlotName: "Head"},           // Head - Blighted Skullcap
				{SlotId: 3, ItemId: 11052, Charges: 1, Location: "Face", SlotName: "Face"},          // Face - Tobrin's Mystical Eyepatch
				{SlotId: 4, ItemId: 14702, Charges: 1, Location: "Ear2", SlotName: "Ear 2"},         // Ear2 - Golden Black Sapphire Earring
				{SlotId: 5, ItemId: 30351, Charges: 1, Location: "Neck", SlotName: "Neck"},          // Neck - Black Sapphire Velium Necklace
				{SlotId: 6, ItemId: 1279, Charges: 1, Location: "Shoulders", SlotName: "Shoulder"},  // Shoulders - Bloodsoaked Raiment
				{SlotId: 7, ItemId: 5723, Charges: 1, Location: "Arms", SlotName: "Arms"},           // Arms - Spirit Wracked Cord
				{SlotId: 8, ItemId: 11603, Charges: 1, Location: "Back", SlotName: "Back"},          // Back - White Dragonscale Cloak
				{SlotId: 9, ItemId: 12804, Charges: 1, Location: "Wrist1", SlotName: "Wrist 1"},     // Wrist1 - Bracelet of Cessation
				{SlotId: 10, ItemId: 25095, Charges: 1, Location: "Wrist2", SlotName: "Wrist 2"},    // Wrist2 - Bone Bracelet of Condemnation
				{SlotId: 11, ItemId: 13169, Charges: 1, Location: "Range", SlotName: "Range"},       // Range - Dagger of Marnek
				{SlotId: 12, ItemId: 31340, Charges: 1, Location: "Hands", SlotName: "Hands"},       // Hands - Shiny Metallic Gloves
				{SlotId: 13, ItemId: 20544, Charges: 1, Location: "Primary", SlotName: "Primary"},   // Primary - Scythe of the Shadowed Soul
				{SlotId: 15, ItemId: 25198, Charges: 1, Location: "Finger1", SlotName: "Ring 1"},    // Finger1 - Ring of Lightning
				{SlotId: 16, ItemId: 5727, Charges: 1, Location: "Finger2", SlotName: "Ring 2"},     // Finger2 - Regal band of Bathezid
				{SlotId: 17, ItemId: 31239, Charges: 1, Location: "Chest", SlotName: "Chest"},       // Chest - Sal`Varae's Robe of Darkness
				{SlotId: 18, ItemId: 31166, Charges: 1, Location: "Legs", SlotName: "Legs"},         // Legs - Rotting Trousers
				{SlotId: 19, ItemId: 29645, Charges: 1, Location: "Feet", SlotName: "Feet"},         // Feet - White Dragonscale Boots
				{SlotId: 20, ItemId: 11666, Charges: 1, Location: "Waist", SlotName: "Waist"},       // Waist - Bone-Clasped Girdle
				{SlotId: 22, ItemId: 17403, Charges: 1, Location: "General1", SlotName: "General1"}, // General1 - Bag of the Tinkerers

				// General inventory slots
				{SlotId: 23, ItemId: 17403, Charges: 1, Location: "General2", SlotName: "General2"}, // General2 - Bag of the Tinkerers
				{SlotId: 24, ItemId: 17403, Charges: 1, Location: "General3", SlotName: "General3"}, // General3 - Bag of the Tinkerers
				{SlotId: 25, ItemId: 17403, Charges: 1, Location: "General4", SlotName: "General4"}, // General4 - Bag of the Tinkerers
				{SlotId: 26, ItemId: 17403, Charges: 1, Location: "General5", SlotName: "General5"}, // General5 - Bag of the Tinkerers
				{SlotId: 27, ItemId: 10895, Charges: 1, Location: "General6", SlotName: "General6"}, // General6 - Fungus Covered Great Stick
				{SlotId: 28, ItemId: 2300, Charges: 1, Location: "General7", SlotName: "General7"},  // General7 - Journeyman's Boots
				{SlotId: 29, ItemId: 14730, Charges: 1, Location: "General8", SlotName: "General8"}, // General8 - Circlet of Shadow

				// General1 bag contents
				{SlotId: 250, ItemId: 1233, Charges: 1, Location: "General1-Slot1", SlotName: "General1 Slot 1"}, // General1-Slot1 - Blighted Robe
				{SlotId: 251, ItemId: 4315, Charges: 1, Location: "General1-Slot2", SlotName: "General1 Slot 2"}, // General1-Slot2 - Obulus Death Shroud
				{SlotId: 252, ItemId: 5728, Charges: 1, Location: "General1-Slot3", SlotName: "General1 Slot 3"}, // General1-Slot3 - Di'zok Signet of Service
				{SlotId: 253, ItemId: 1236, Charges: 1, Location: "General1-Slot4", SlotName: "General1 Slot 4"}, // General1-Slot4 - Blighted Gloves

				// General4 bag contents
				{SlotId: 289, ItemId: 13073, Charges: 20, Location: "General4-Slot10", SlotName: "General4 Slot 10"}, // General4-Slot10 - Bone Chips

				// General5 bag contents
				{SlotId: 290, ItemId: 9991, Charges: 20, Location: "General5-Slot1", SlotName: "General5 Slot 1"},    // General5-Slot1 - Bread Cakes*
				{SlotId: 291, ItemId: 9990, Charges: 20, Location: "General5-Slot2", SlotName: "General5 Slot 2"},    // General5-Slot2 - Skin of Milk
				{SlotId: 293, ItemId: 14727, Charges: 1, Location: "General5-Slot4", SlotName: "General5 Slot 4"},    // General5-Slot4 - Locket of Escape
				{SlotId: 299, ItemId: 10028, Charges: 20, Location: "General5-Slot10", SlotName: "General5 Slot 10"}, // General5-Slot10 - Peridot
			},
		},
		// {
		// 	name:           "BigRealInventory",
		// 	filename:       "inventory_big_real.txt",
		// 	expectedStatus: http.StatusOK,
		// 	expectedItems: []*pb.InventoryItem{
		// 		// Equipment slots
		// 		{SlotId: 1, ItemId: 29859, Charges: 1},  // Ear - Runed Earring of Veracity
		// 		{SlotId: 2, ItemId: 31161, Charges: 1},  // Head - Rotting Crown
		// 		{SlotId: 3, ItemId: 8216, Charges: 1},   // Face - Juzlix's Mask of Torment
		// 		{SlotId: 4, ItemId: 27948, Charges: 1},  // Ear - Lizard Bone Earring
		// 		{SlotId: 5, ItemId: 29459, Charges: 1},  // Neck - Surreptitious Broach
		// 		{SlotId: 6, ItemId: 27889, Charges: 1},  // Shoulders - Terrorclaws Hide
		// 		{SlotId: 7, ItemId: 5723, Charges: 1},   // Arms - Spirit Wracked Cord
		// 		{SlotId: 8, ItemId: 11603, Charges: 1},  // Back - White Dragonscale Cloak
		// 		{SlotId: 9, ItemId: 12804, Charges: 1},  // Wrist - Bracelet of Cessation
		// 		{SlotId: 10, ItemId: 27712, Charges: 1}, // Wrist - Bloody Griffon-Hide Wrist Guard
		// 		{SlotId: 11, ItemId: 5803, Charges: 1},  // Range - A Sandwich of Foul Smelling Herbs
		// 		{SlotId: 12, ItemId: 25026, Charges: 1}, // Hands - Coldain Skin Gloves
		// 		{SlotId: 13, ItemId: 26009, Charges: 1}, // Primary - Zlandicar's Heart
		// 		{SlotId: 14, ItemId: 10219, Charges: 1}, // Secondary - Rokyls Channelling Crystal
		// 		{SlotId: 15, ItemId: 5727, Charges: 1},  // Finger1 - Regal band of Bathezid
		// 		{SlotId: 16, ItemId: 19719, Charges: 1}, // Finger2 - Ring of the Shissar
		// 		{SlotId: 17, ItemId: 1233, Charges: 1},  // Chest - Blighted Robe
		// 		{SlotId: 18, ItemId: 27884, Charges: 1}, // Legs - Dreadfangs Hide
		// 		{SlotId: 19, ItemId: 29454, Charges: 1}, // Feet - Tenuous Dragonscale Slippers
		// 		{SlotId: 20, ItemId: 25858, Charges: 1}, // Waist - Belt of Dwarf Slaying
		// 		{SlotId: 22, ItemId: 8331, Charges: 1},  // Ammo - Turmoil Warts

		// 		// General1 bag and contents
		// 		{SlotId: 23, ItemId: 17404, Charges: 1},   // General1 - Large Soiled Bag
		// 		{SlotId: 230, ItemId: 10050, Charges: 1},  // General1-Slot1 - Sapphire Necklace
		// 		{SlotId: 231, ItemId: 10313, Charges: 1},  // General1-Slot2 - Fishbone Earring
		// 		{SlotId: 232, ItemId: 13073, Charges: 13}, // General1-Slot3 - Bone Chips
		// 		{SlotId: 233, ItemId: 10037, Charges: 2},  // General1-Slot4 - Diamond
		// 		{SlotId: 234, ItemId: 10049, Charges: 1},  // General1-Slot5 - Fire Emerald Ring
		// 		{SlotId: 235, ItemId: 10053, Charges: 1},  // General1-Slot6 - Jacinth
		// 		{SlotId: 236, ItemId: 8331, Charges: 4},   // General1-Slot7 - Turmoil Warts
		// 		{SlotId: 237, ItemId: 13073, Charges: 20}, // General1-Slot8 - Bone Chips
		// 		{SlotId: 239, ItemId: 10028, Charges: 19}, // General1-Slot10 - Peridot

		// 		// General2 bag and contents
		// 		{SlotId: 24, ItemId: 11703, Charges: 1},  // General2 - Box of Abu-Kar
		// 		{SlotId: 240, ItemId: 14742, Charges: 1}, // General2-Slot1 - Hand of the Reaper
		// 		{SlotId: 241, ItemId: 26054, Charges: 1}, // General2-Slot2 - Amulet of the Haven
		// 		{SlotId: 242, ItemId: 27882, Charges: 1}, // General2-Slot3 - Boots of Flowing Slime
		// 		{SlotId: 243, ItemId: 2463, Charges: 1},  // General2-Slot4 - Pegasus Feather Cloak
		// 		{SlotId: 244, ItemId: 30299, Charges: 1}, // General2-Slot5 - Velium Blue Diamond Bracelet
		// 		{SlotId: 245, ItemId: 30299, Charges: 1}, // General2-Slot6 - Velium Blue Diamond Bracelet
		// 		{SlotId: 246, ItemId: 14707, Charges: 1}, // General2-Slot7 - Platinum Diamond Wedding Ring
		// 		{SlotId: 247, ItemId: 14707, Charges: 1}, // General2-Slot8 - Platinum Diamond Wedding Ring
		// 		{SlotId: 248, ItemId: 6351, Charges: 1},  // General2-Slot9 - Fine Steel Morning Star
		// 		{SlotId: 249, ItemId: 5038, Charges: 1},  // General2-Slot10 - Leather Whip

		// 		// General3 bag and contents
		// 		{SlotId: 25, ItemId: 17082, Charges: 1},   // General3 - Box of Nil Space
		// 		{SlotId: 250, ItemId: 2970, Charges: 1},   // General3-Slot1 - Manarock
		// 		{SlotId: 251, ItemId: 9963, Charges: 10},  // General3-Slot2 - Essence Emerald
		// 		{SlotId: 252, ItemId: 1236, Charges: 1},   // General3-Slot3 - Blighted Gloves
		// 		{SlotId: 253, ItemId: 1651, Charges: 1},   // General3-Slot4 - Loam Encrusted Pantaloons
		// 		{SlotId: 254, ItemId: 24890, Charges: 1},  // General3-Slot5 - Holgresh Elder Beads
		// 		{SlotId: 255, ItemId: 27834, Charges: 1},  // General3-Slot6 - Black Pearl
		// 		{SlotId: 256, ItemId: 3426, Charges: 1},   // General3-Slot7 - Black Sapphire
		// 		{SlotId: 257, ItemId: 1348, Charges: 1},   // General3-Slot8 - Black Marble
		// 		{SlotId: 258, ItemId: 24921, Charges: 1},  // General3-Slot9 - Black Pearl
		// 		{SlotId: 259, ItemId: 9962, Charges: 18},  // General3-Slot10 - Some item
		// 		{SlotId: 26, ItemId: 17701, Charges: 1},   // General4 - Some bag
		// 		{SlotId: 260, ItemId: 19726, Charges: 1},  // General4-Slot1 - Some item
		// 		{SlotId: 261, ItemId: 10142, Charges: 1},  // General4-Slot2 - Some item
		// 		{SlotId: 262, ItemId: 31333, Charges: 1},  // General4-Slot3 - Some item
		// 		{SlotId: 263, ItemId: 14706, Charges: 1},  // General4-Slot4 - Some item
		// 		{SlotId: 264, ItemId: 1622, Charges: 1},   // General4-Slot5 - Some item
		// 		{SlotId: 265, ItemId: 11571, Charges: 1},  // General4-Slot6 - Some item
		// 		{SlotId: 267, ItemId: 6353, Charges: 1},   // General4-Slot7 - Some item (skipping 266)
		// 		{SlotId: 268, ItemId: 12863, Charges: 1},  // General4-Slot8 - Some item
		// 		{SlotId: 269, ItemId: 2300, Charges: 1},   // General4-Slot9 - Some item
		// 		{SlotId: 27, ItemId: 17969, Charges: 1},   // General5 - Some bag
		// 		{SlotId: 270, ItemId: 13380, Charges: 1},  // General5-Slot1 - Some item
		// 		{SlotId: 271, ItemId: 30367, Charges: 1},  // General5-Slot2 - Some item
		// 		{SlotId: 272, ItemId: 11917, Charges: 1},  // General5-Slot3 - Some item
		// 		{SlotId: 273, ItemId: 5744, Charges: 1},   // General5-Slot4 - Some item
		// 		{SlotId: 274, ItemId: 30365, Charges: 1},  // General5-Slot5 - Some item
		// 		{SlotId: 275, ItemId: 30365, Charges: 1},  // General5-Slot6 - Some item
		// 		{SlotId: 276, ItemId: 5735, Charges: 1},   // General5-Slot7 - Some item
		// 		{SlotId: 277, ItemId: 1625, Charges: 1},   // General5-Slot8 - Some item
		// 		{SlotId: 279, ItemId: 11052, Charges: 1},  // General5-Slot9 - Some item (skipping 278)
		// 		{SlotId: 28, ItemId: 17969, Charges: 1},   // General6 - Some bag
		// 		{SlotId: 280, ItemId: 20544, Charges: 1},  // General6-Slot1 - Some item
		// 		{SlotId: 281, ItemId: 21810, Charges: 1},  // General6-Slot2 - Some item
		// 		{SlotId: 286, ItemId: 13005, Charges: 11}, // General6-Slot3 - Some item (skipping 282-285)
		// 		{SlotId: 287, ItemId: 13006, Charges: 1},  // General6-Slot4 - Some item
		// 		{SlotId: 288, ItemId: 8194, Charges: 20},  // General6-Slot5 - Some item
		// 		{SlotId: 289, ItemId: 13006, Charges: 20}, // General6-Slot6 - Some item
		// 		{SlotId: 29, ItemId: 14730, Charges: 1},   // General7 - Some bag
		// 		{SlotId: 30, ItemId: 10895, Charges: 1},   // General8 - Some bag
		// 	},
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setupEmptyUser(db, "1"); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			req := createParseInventoryRequest(tt.filename)

			// Create ConnectRPC client
			client := pbconnect.NewEqTestCopyServiceClient(http.DefaultClient, server.URL)

			// Make the request with proper auth header
			connectReq := connect.NewRequest(req)
			connectReq.Header().Set("Authorization", "Bearer "+generateTestToken("1"))

			response, err := client.ParseInventoryFile(context.Background(), connectReq)

			if tt.expectedStatus == http.StatusOK {
				if err != nil {
					t.Fatalf("Expected success but got error: %v", err)
				}

				if len(response.Msg.Errors) > 0 {
					t.Errorf("Parsing errors: %v", response.Msg.Errors)
				}

				// Sort both slices by slot ID for comparison
				sort.Slice(tt.expectedItems, func(i, j int) bool {
					return tt.expectedItems[i].SlotId < tt.expectedItems[j].SlotId
				})
				sort.Slice(response.Msg.Items, func(i, j int) bool {
					return response.Msg.Items[i].SlotId < response.Msg.Items[j].SlotId
				})

				if diff := cmp.Diff(tt.expectedItems, response.Msg.Items, protocmp.Transform()); diff != "" {
					t.Errorf("Items mismatch:\n%s", diff)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error but got success")
				}
			}
		})
	}
}

func TestListCharacters(t *testing.T) {
	t.Parallel()

	server, db := setupTestServer(t)
	defer server.Close()
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close database: %v", err)
		}
	}()

	tests := []struct {
		name           string
		auth           bool
		setup          func(*sql.DB) error
		expectedStatus int
		expectedCount  int
	}{
		{"NoAuth", false, nil, http.StatusUnauthorized, 0},
		{"ValidAuth_Empty", true, func(db *sql.DB) error { return setupEmptyUser(db, "1") }, http.StatusOK, 0},
		{"ValidAuth_WithCharacters", true, func(db *sql.DB) error { return setupUserWithCharacters(db, "1") }, http.StatusOK, 2},
		{"InvalidToken", false, nil, http.StatusUnauthorized, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(db); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			headers := map[string]string{"Content-Type": "application/json"}
			if tt.auth {
				headers["Authorization"] = "Bearer " + generateTestToken("1")
			} else if tt.name == "InvalidToken" {
				headers["Authorization"] = "Bearer invalid-token"
			}

			resp, err := makeRequest(server.URL, "POST", "/eqtestcopy.EqTestCopyService/ListCharacters", headers, &pb.ListCharactersRequest{})
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					t.Fatalf("Failed to close response body: %v", err)
				}
			}()

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if tt.expectedCount >= 0 {
				var response pb.ListCharactersResponse
				if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}
				if len(response.Characters) != tt.expectedCount {
					t.Errorf("Expected %d characters, got %d", tt.expectedCount, len(response.Characters))
				}
			}
		})
	}
}

func TestUpdateInventory(t *testing.T) {
	t.Parallel()

	server, db := setupTestServer(t)
	defer server.Close()
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close database: %v", err)
		}
	}()

	if err := setupUserWithCharacters(db, "1"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	req := &pb.UpdateInventoryRequest{
		CharacterId: 1,
		Inventories: []*pb.InventoryItem{
			{SlotId: 1, ItemId: 99999, Charges: 1},
			{SlotId: 22, ItemId: 12345, Charges: 1},
			{SlotId: 220, ItemId: 11111, Charges: 5},
		},
	}

	resp, err := makeRequest(server.URL, "POST", "/eqtestcopy.EqTestCopyService/UpdateInventory",
		map[string]string{"Content-Type": "application/json", "Authorization": "Bearer " + generateTestToken("1")}, req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response pb.UpdateInventoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if !response.Success {
		t.Errorf("Expected successful update, got: %s", response.Message)
	}
}

// Helper functions
func setupTestServer(t *testing.T) (*httptest.Server, *sql.DB) {
	server, db, err := setupTestServerHelper(t)
	if err != nil {
		t.Fatalf("Failed to setup test server: %v", err)
	}
	return server, db
}

func setupTestServerHelper(t *testing.T) (*httptest.Server, *sql.DB, error) {
	db, err := setupTestDatabase()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to setup test database: %w", err)
	}

	verifier := createMockVerifier()
	server, err := New(db, verifier)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create server: %w", err)
	}

	interceptors := connect.WithInterceptors(connect.UnaryInterceptorFunc(testAuthInterceptor))
	path, serviceHandler := pbconnect.NewEqTestCopyServiceHandler(server, interceptors)

	mux := http.NewServeMux()
	mux.Handle(path, serviceHandler)
	httpServer := httptest.NewServer(mux)

	return httpServer, db, nil
}

func setupTestDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := createTestSchema(db); err != nil {
		return nil, fmt.Errorf("failed to create test schema: %w", err)
	}

	return db, nil
}

func createTestSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		email TEXT,
		revoked BOOLEAN DEFAULT FALSE
	);

	CREATE TABLE IF NOT EXISTS character_data (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		account_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		last_name TEXT DEFAULT '',
		race INTEGER DEFAULT 0,
		class INTEGER DEFAULT 0,
		level INTEGER DEFAULT 0,
		zone_id INTEGER DEFAULT 0,
		x REAL DEFAULT 0,
		y REAL DEFAULT 0,
		z REAL DEFAULT 0,
		cur_hp INTEGER DEFAULT 0,
		mana INTEGER DEFAULT 0,
		str INTEGER DEFAULT 0,
		sta INTEGER DEFAULT 0,
		agi INTEGER DEFAULT 0,
		dex INTEGER DEFAULT 0,
		wis INTEGER DEFAULT 0,
		int INTEGER DEFAULT 0,
		cha INTEGER DEFAULT 0,
		last_login INTEGER DEFAULT 0
	);

	CREATE TABLE IF NOT EXISTS character_inventory (
		id INTEGER NOT NULL DEFAULT 0,
		slotid INTEGER NOT NULL DEFAULT 0,
		itemid INTEGER NULL DEFAULT 0,
		charges INTEGER NULL DEFAULT 0,
		custom_data TEXT NULL DEFAULT NULL,
		serialnumber INTEGER NOT NULL DEFAULT 0,
		initialserial INTEGER NOT NULL DEFAULT 0,
		PRIMARY KEY (id, slotid)
	);

	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL
	);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func generateTestToken(userID string) string {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": 9999999999,
		"iat": 1000000000,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}

func setupEmptyUser(db *sql.DB, userID string) error {
	_, err := db.Exec("INSERT OR REPLACE INTO accounts (id, username, password, email) VALUES (?, ?, ?, ?)",
		userID, "testuser", "hashedpassword", "testuser@example.com")
	return err
}

func setupUserWithCharacters(db *sql.DB, userID string) error {
	if err := setupEmptyUser(db, userID); err != nil {
		return err
	}

	_, err := db.Exec(`
		INSERT OR REPLACE INTO character_data (id, account_id, name, race, class, level, zone_id, cur_hp, mana, str, sta, agi, dex, wis, int, cha, last_login) 
		VALUES 
		(1, ?, 'TestCharacter1', 1, 1, 50, 1, 1000, 500, 100, 100, 100, 100, 100, 100, 100, 1234567890),
		(2, ?, 'TestCharacter2', 2, 2, 30, 2, 800, 400, 80, 80, 80, 80, 80, 80, 80, 1234567890)
	`, userID, userID)
	return err
}

func readTestData(filename string) (string, error) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get test file path")
	}

	testDataDir := filepath.Join(filepath.Dir(testFile), "testdata")
	filePath := filepath.Join(testDataDir, filename)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read test data file %s: %w", filename, err)
	}

	return string(content), nil
}

func createParseInventoryRequest(filename string) *pb.ParseInventoryFileRequest {
	content, err := readTestData(filename)
	if err != nil {
		panic(fmt.Sprintf("Failed to read test data: %v", err))
	}
	return &pb.ParseInventoryFileRequest{FileContent: content}
}

func makeRequest(baseURL, method, path string, headers map[string]string, body interface{}) (*http.Response, error) {
	var bodyReader *strings.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = strings.NewReader(string(bodyBytes))
	}

	req, err := http.NewRequest(method, baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	return client.Do(req)
}
