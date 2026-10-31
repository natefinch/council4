package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Adventure exile permission lets its beneficiary play a land front face
// (#3354, bounded Engine slice of #170).
//
// Source review:
//   - Lindblum, Industrial Regency // Mage Siege, oracle_id
//     4cc014f3-05e0-442e-9dee-03eab1aa65a3, revision
//     548dd152-f0b6-4e8f-9afc-a4ec1671b648 (pinned corpus): "Land — Town //
//     Instant — Adventure"; Lindblum "This land enters tapped. {T}: Add {R}.";
//     Mage Siege {2}{R} "Create a 0/1 black Wizard creature token ...".
//   - Scryfall rulings (2025-06-06): "If a spell is cast as an Adventure, its
//     controller exiles it ... For as long as it remains exiled, that player may
//     play it using its primary characteristics."; "You must still follow any
//     timing restrictions and permissions for the card you play from exile. In
//     the case of any of the five lands in this release, you'll be able to play
//     it only during your main phase while the stack is empty and only if you
//     have an available land play remaining."; exiling it any other way gives no
//     permission.
//   - Comprehensive Rules effective 2026-09-25 (sha256
//     8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     715.3d ("For as long as that card remains exiled, that player may play
//     it."), 715.4 (only normal characteristics outside the stack), 116.2a and
//     305.1 (a land play is a special action on your own main phase with an
//     empty stack), 305.2a/305.2b (land-play count), 305.3 (not on another
//     player's turn) and 400.7 (a zone change makes a new object).

func lindblumRegistry(t *testing.T) (*cards.Registry, *game.CardDef) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	def := registry.Lookup("Lindblum, Industrial Regency")
	if def == nil {
		t.Fatal("registered Lindblum, Industrial Regency // Mage Siege not found")
	}
	return registry, def
}

// resolvedMageSiege has Player1 cast registered Mage Siege from hand on their
// main phase, optionally lets thief gain control of it, and resolves it into
// Player1's exile.
func resolvedMageSiege(t *testing.T, thief game.PlayerID) (*game.Game, *Engine, id.ID) {
	t.Helper()
	registry, def := lindblumRegistry(t)
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngineWithPrograms(nil, registry.Programs())
	cardID := addCardToHand(g, game.Player1, def)
	for range 3 {
		addBasicLandPermanent(g, game.Player1, types.Mountain)
	}
	giveMainPhasePriority(g, game.Player1)
	if !engine.applyAction(g, game.Player1, action.CastSpellFace(cardID, game.FaceAlternate, nil, 0, nil)) {
		t.Fatal("Player1 could not cast Mage Siege")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != cardID || obj.Face != game.FaceAlternate {
		t.Fatalf("stack object = %+v, want Mage Siege", obj)
	}
	if thief != game.Player1 {
		stealStackSpell(t, g, engine, obj.ID, thief)
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Exile.Contains(cardID) || g.AdventureCards[cardID].Player != thief {
		t.Fatalf("Mage Siege exile = %v permission = %+v, want %v's permission in Player1's exile",
			g.Players[game.Player1].Exile.Contains(cardID), g.AdventureCards[cardID], thief)
	}
	if wizards := countPermanentsNamed(g, thief, "Wizard"); wizards != 1 {
		t.Fatalf("%v Wizard tokens = %d, want Mage Siege's one", thief, wizards)
	}
	g.Turn.LandsPlayedThisTurn = 0
	return g, engine, cardID
}

func countPermanentsNamed(g *game.Game, controller game.PlayerID, name string) int {
	count := 0
	for _, permanent := range g.Battlefield {
		if permanent.Controller == controller && permanentEffectiveName(g, permanent) == name {
			count++
		}
	}
	return count
}

func lindblumFromExile(cardID id.ID, face game.FaceIndex) action.Action {
	return action.PlayLandFaceFromZone(cardID, zone.Exile, face)
}

// assertAdventureLandPlayable checks both legal-action enumeration and direct
// apply for player playing Lindblum's land front from exile, on a clone so the
// direct apply does not alter g. It also checks the Adventure face is never
// offered as a land.
func assertAdventureLandPlayable(t *testing.T, g *game.Game, engine *Engine, player game.PlayerID, cardID id.ID, want bool) {
	t.Helper()
	actions := engine.legalActions(g, player)
	if got := actionsContain(actions, lindblumFromExile(cardID, game.FaceFront)); got != want {
		t.Fatalf("legal actions for %v contain Lindblum play from exile = %v, want %v", player, got, want)
	}
	if actionsContain(actions, lindblumFromExile(cardID, game.FaceAlternate)) {
		t.Fatalf("legal actions for %v offer Mage Siege as a land", player)
	}
	clone := g.Clone()
	if got := engine.applyAction(clone, player, lindblumFromExile(cardID, game.FaceFront)); got != want {
		t.Fatalf("direct Lindblum play from exile by %v = %v, want %v", player, got, want)
	}
}

func assertLindblumPlayed(t *testing.T, g *game.Game, cardID id.ID, controller game.PlayerID) {
	t.Helper()
	if g.Players[game.Player1].Exile.Contains(cardID) {
		t.Fatal("played Lindblum remained in its owner's exile")
	}
	permanent := permanentForCard(g, cardID)
	if permanent == nil || permanent.Owner != game.Player1 || effectiveController(g, permanent) != controller ||
		permanentEffectiveName(g, permanent) != "Lindblum, Industrial Regency" || !permanent.Tapped {
		t.Fatalf("permanent = %+v, want tapped Player1-owned Lindblum controlled by %v", permanent, controller)
	}
	if g.Turn.LandsPlayedThisTurn != 1 {
		t.Fatalf("lands played = %d, want the play to use the land drop", g.Turn.LandsPlayedThisTurn)
	}
}

func TestAdventureLandFrontPlayableByResolvingController(t *testing.T) {
	t.Run("owner-controlled Mage Siege", func(t *testing.T) {
		g, engine, cardID := resolvedMageSiege(t, game.Player1)
		mountain := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Mountain", Types: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic},
			Subtypes: []types.Sub{types.Mountain},
		}})
		giveMainPhasePriority(g, game.Player2)
		assertAdventureLandPlayable(t, g, engine, game.Player2, cardID, false)

		// Another player's turn (CR 305.3).
		giveMainPhasePriority(g, game.Player2)
		g.Turn.PriorityPlayer = game.Player1
		assertAdventureLandPlayable(t, g, engine, game.Player1, cardID, false)

		// A non-empty stack (CR 116.2a, 305.1).
		giveMainPhasePriority(g, game.Player1)
		addEffectSpellToStack(g, game.Player2, game.GainLife{Amount: game.Fixed(1), Player: game.ControllerReference()}, nil)
		assertAdventureLandPlayable(t, g, engine, game.Player1, cardID, false)
		g.Stack.Pop()

		// No land play remaining (CR 305.2b).
		g.Turn.LandsPlayedThisTurn = 1
		assertAdventureLandPlayable(t, g, engine, game.Player1, cardID, false)
		g.Turn.LandsPlayedThisTurn = 0

		if !g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatal("Lindblum left exile before being played")
		}
		assertAdventureLandPlayable(t, g, engine, game.Player1, cardID, true)
		if !engine.applyAction(g, game.Player1, lindblumFromExile(cardID, game.FaceFront)) {
			t.Fatal("owner could not play Lindblum from exile")
		}
		assertLindblumPlayed(t, g, cardID, game.Player1)
		if actionsContain(engine.legalActions(g, game.Player1), action.PlayLand(mountain)) ||
			engine.applyAction(g.Clone(), game.Player1, action.PlayLand(mountain)) {
			t.Fatal("second land drop was allowed after playing Lindblum from exile")
		}
	})
	t.Run("stolen Mage Siege", func(t *testing.T) {
		g, engine, cardID := resolvedMageSiege(t, game.Player2)
		giveMainPhasePriority(g, game.Player1)
		assertAdventureLandPlayable(t, g, engine, game.Player1, cardID, false)
		giveMainPhasePriority(g, game.Player2)
		assertAdventureLandPlayable(t, g, engine, game.Player2, cardID, true)
		if !engine.applyAction(g, game.Player2, lindblumFromExile(cardID, game.FaceFront)) {
			t.Fatal("Player2 could not play Lindblum from Player1's exile")
		}
		assertLindblumPlayed(t, g, cardID, game.Player2)
	})
}

func TestAdventureLandPermissionExpiresWithExileIncarnation(t *testing.T) {
	for _, route := range []string{"exile to exile", "exile to graveyard and back"} {
		t.Run(route, func(t *testing.T) {
			g, engine, cardID := resolvedMageSiege(t, game.Player2)
			stale := g.AdventureCards[cardID]
			if route == "exile to exile" {
				if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Exile) {
					t.Fatal("exile-to-exile move failed")
				}
			} else if !moveCardBetweenZones(g, game.Player1, cardID, zone.Exile, zone.Graveyard) ||
				!moveCardBetweenZones(g, game.Player1, cardID, zone.Graveyard, zone.Exile) {
				t.Fatal("exile round trip failed")
			}
			if !g.Players[game.Player1].Exile.Contains(cardID) {
				t.Fatal("same physical card is not back in exile")
			}
			for _, player := range []game.PlayerID{game.Player1, game.Player2} {
				giveMainPhasePriority(g, player)
				assertAdventureLandPlayable(t, g, engine, player, cardID, false)
			}
			// A stale record cannot revive the permission for the new object.
			g.AdventureCards[cardID] = stale
			giveMainPhasePriority(g, game.Player2)
			assertAdventureLandPlayable(t, g, engine, game.Player2, cardID, false)
		})
	}
}

// TestAdventureLandPlayLeavesOncePerTurnPermissionUnspent overlaps the
// Adventure permission with a once-per-turn exile land-play permission: the
// unlimited Adventure permission authorizes the play, so the once-per-turn use
// stays available.
func TestAdventureLandPlayLeavesOncePerTurnPermissionUnspent(t *testing.T) {
	g, engine, cardID := resolvedMageSiege(t, game.Player1)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Exile Land Permission", Types: []types.Card{types.Artifact},
	}})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectPlayLandsFromZone,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerYou,
		CastFromZone:   zone.Exile,
		PermanentTypes: []types.Card{types.Land},
		OncePerTurn:    true,
		SourceObjectID: source.ObjectID,
	})
	giveMainPhasePriority(g, game.Player1)
	if !canPlayLandFromZoneByRuleEffect(g, game.Player1, cardID, zone.Exile) {
		t.Fatal("once-per-turn exile land permission does not cover Lindblum")
	}
	assertAdventureLandPlayable(t, g, engine, game.Player1, cardID, true)
	if !engine.applyAction(g, game.Player1, lindblumFromExile(cardID, game.FaceFront)) {
		t.Fatal("owner could not play Lindblum from exile")
	}
	assertLindblumPlayed(t, g, cardID, game.Player1)
	if g.ExilePlayPermissionUsedThisTurn[source.ObjectID] {
		t.Fatal("Adventure land play spent the once-per-turn exile permission")
	}
}
