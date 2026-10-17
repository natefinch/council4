package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestProductionZoneArrivalsStayInMovementCore(t *testing.T) {
	allowed := map[string]string{
		// Replacement-aware commits and entry transactions own raw placement.
		"zones.go:*":                         "*",
		"stack.go:moveAdventureSpellToExile": "*",
		"stack.go:moveStackCardToZone":       "*",
		"primitive_handlers_zones.go:putResolvedCardsOnBattlefieldCollecting": "*",
		"primitive_handlers_zones.go:putResolvedCardOnBattlefieldValue":       "*",

		// These only reorder cards already in the same library.
		"library_choices.go:revealUntilMatchAndBottomRest":               "Library",
		"library_choices.go:revealUntilChosenBottom":                     "Library",
		"library_choices.go:scryCards":                                   "Library",
		"library_choices.go:digCards":                                    "Library",
		"library_choices.go:routeDigSlotCard":                            "Library",
		"library_choices.go:revealTopPartition":                          "Library",
		"library_choices.go:reorderLibraryTop":                           "Library",
		"keyword_actions.go:searchLibrary":                               "Library",
		"keyword_actions.go:placeFoundCard":                              "Library",
		"hideaway.go:handleHideawayExile":                                "Library",
		"pile_split.go:movePileSplitPile":                                "Library",
		"library_card_placement.go:placeObservedLibraryCard":             "Library",
		"card_program_placement.go:putFoundCardsOnLibraryBottom":         "Library",
		"primitive_handlers_move_permanent.go:placeMovedOnLibraryBottom": "Library",

		// Setup, hidden-state projection and failed-proposal rollback are not moves.
		"mulligan.go:performCommanderMulligan":                           "Library",
		"mulligan.go:bottomCardsFromHand":                                "Library",
		"determinize.go:*":                                               "Hand Library",
		"actions_cast.go:restoreCastSourceCard":                          "*",
		"cast_paid_resolution.go:castPaidSpellFromSource":                "Library",
		"opening_hand_battlefield.go:considerOpeningHandBattlefieldCard": "Hand",
		"stack.go:resolveActivatedAbilityWithChoices":                    "Hand",

		// The payment transaction emits its authoritative arrival through State.
		"payment/apply.go:applyPaymentPlan": "Exile",
	}
	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != "." && (entry.Name() == "testdata" || strings.HasPrefix(entry.Name(), ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		for _, declaration := range file.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			key := filepath.ToSlash(path) + ":" + fn.Name.Name
			zones := allowed[key]
			if zones == "" {
				zones = allowed[filepath.ToSlash(path)+":*"]
			}
			for _, write := range rawZoneArrivals(fn) {
				if zones != "*" && !slices.Contains(strings.Fields(zones), write.zone) {
					t.Errorf("%s: %s directly appends to %s; use shared zone movement, not a card/effect-specific placement",
						fset.Position(write.position), fn.Name.Name, write.zone)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

type rawZoneArrival struct {
	zone     string
	position token.Pos
}

func rawZoneArrivals(fn *ast.FuncDecl) []rawZoneArrival {
	aliases := make(map[string]string)
	for _, parameter := range fn.Type.Params.List {
		kind := parameter.Type
		if pointer, ok := kind.(*ast.StarExpr); ok {
			kind = pointer.X
		}
		if zone, ok := kind.(*ast.SelectorExpr); ok && zone.Sel.Name == "Zone" {
			for _, name := range parameter.Names {
				aliases[name.Name] = "*"
			}
		}
	}
	var writes []rawZoneArrival
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if assignment, ok := node.(*ast.AssignStmt); ok {
			for i, left := range assignment.Lhs {
				name, named := left.(*ast.Ident)
				if !named || i >= len(assignment.Rhs) {
					continue
				}
				aliases[name.Name] = rawZoneReceiver(assignment.Rhs[i], aliases)
			}
		}
		if declaration, ok := node.(*ast.ValueSpec); ok {
			for i, name := range declaration.Names {
				if i < len(declaration.Values) {
					aliases[name.Name] = rawZoneReceiver(declaration.Values[i], aliases)
				}
			}
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		method, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !slices.Contains([]string{"Add", "AddToBottom", "InsertAt"}, method.Sel.Name) {
			return true
		}
		if zone := rawZoneReceiver(method.X, aliases); zone != "" {
			writes = append(writes, rawZoneArrival{zone: zone, position: call.Pos()})
		}
		return true
	})
	return writes
}

func rawZoneReceiver(expression ast.Expr, aliases map[string]string) string {
	if reference, ok := expression.(*ast.UnaryExpr); ok {
		return rawZoneReceiver(reference.X, aliases)
	}
	if field, ok := expression.(*ast.SelectorExpr); ok {
		if slices.Contains([]string{"Hand", "Library", "Graveyard", "Exile", "CommandZone"}, field.Sel.Name) {
			return field.Sel.Name
		}
	}
	if name, ok := expression.(*ast.Ident); ok {
		return aliases[name.Name]
	}
	if call, ok := expression.(*ast.CallExpr); ok {
		name, named := call.Fun.(*ast.Ident)
		if named && name.Name == "destinationZone" && len(call.Args) == 3 {
			if zone, ok := call.Args[2].(*ast.SelectorExpr); ok {
				if slices.Contains([]string{"Hand", "Library", "Graveyard", "Exile", "Command"}, zone.Sel.Name) {
					return zone.Sel.Name
				}
			}
			return "*"
		}
	}
	return ""
}

func TestZoneArrivalGuardRecognizesLocalAliases(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "example.go", `package rules
func bypass(player *Player, g *Game, zoneArg *zone.Zone) {
	player.Hand.Add(1)
	hand := &player.Hand
	hand.Add(2)
	destination, _ := destinationZone(g, 0, zone.Hand)
	destination.Add(3)
	zoneArg.Add(4)
	player.ManaPool.Add(mana.C, 1)
}`, 0)
	if err != nil {
		t.Fatal(err)
	}
	fn, ok := file.Decls[0].(*ast.FuncDecl)
	if !ok {
		t.Fatal("missing fixture function")
	}
	writes := rawZoneArrivals(fn)
	if len(writes) != 4 {
		t.Fatalf("zone writes = %d, want all direct/aliased/parameter writes and no mana write", len(writes))
	}
	for index, write := range writes {
		expected := "Hand"
		if index == 3 {
			expected = "*"
		}
		if write.zone != expected {
			t.Fatalf("recognized zone = %q, want %s", write.zone, expected)
		}
	}
}
