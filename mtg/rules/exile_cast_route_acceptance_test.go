package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Overlapping cast-from-exile permissions (#3355): one exiled card covered by
// both a per-card "without paying its mana cost" permission and registered
// Evelyn, the Covetous's once-each-turn paying permission. Validation, legal
// actions (also after cloning) and application share one route set; the
// caster chooses a payable route and only that route's permission is spent.
//
// Source review:
//   - Comprehensive Rules effective 2026-09-25,
//     https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt
//     (sha256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca):
//     601.3 (a player can begin to cast a spell only if a rule or effect
//     allows it), 601.2b (the player announces an alternative cost; two
//     alternative costs can't apply to one spell), 118.9a (only one
//     alternative cost can be applied; "without paying its mana cost" is an
//     alternative cost), 118.14 (mana spending riders apply only to spells
//     cast that way).
//   - Vandalblast, printing 71067239-d46f-4201-a91c-6aa4c41fa8b1 (oracle_id
//     3567c3c8-b3c7-45b7-935b-b1fdbc973720), {R} Sorcery: "Destroy target
//     artifact you don't control. / Overload {4}{R}". Ruling 2024-01-12: "If
//     you are instructed to cast a spell with overload 'without paying its
//     mana cost,' you can't choose to pay its overload cost instead."
//   - Boon Satyr, printing 477a30f3-3cbe-42a5-9fae-fe5885a8cf25 (oracle_id
//     aedcd2fb-813a-4915-b7b9-ac7479863462), {1}{G}{G}: "Flash / Bestow
//     {3}{G}{G} / Enchanted creature gets +4/+2." Ruling 2018-07-13: bestow
//     is an alternative cost and can't be combined with casting the spell
//     "without paying its mana cost."
//   - Evelyn, the Covetous, printing c0dad61f-36cd-46af-82b7-a02e04efd676
//     (oracle_id 4ab6f615-1f04-4ed2-88b2-49faa8356dd0): "... Once each turn,
//     you may play a card from exile with a collection counter on it if it
//     was exiled by an ability you controlled, and you may spend mana as
//     though it were mana of any color to cast it." No rulings.

// overlappingExileRouteScene is Player1 with registered Evelyn and an exiled
// registered card carrying her collection counter and a separate per-card
// free-cast permission.
type overlappingExileRouteScene struct {
	g      *game.Game
	engine *Engine
	evelyn *game.Permanent
	cardID id.ID
	// artifact is an opponent's targetable artifact; warded has hexproof, so
	// only an overloaded Vandalblast can destroy it.
	artifact, warded *game.Permanent
}

func newOverlappingExileRouteScene(t *testing.T, name string, blue int) overlappingExileRouteScene {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	evelyn := groupExileRegistryPermanent(t, g, game.Player1, "Evelyn, the Covetous")
	artifact := addOverloadArtifact(g, game.Player2, false)
	warded := addOverloadArtifact(g, game.Player2, true)
	cardID := collectionExiledCard(g, registryCardDef(t, name))
	freePlayPermission(g, cardID)
	g.Players[game.Player1].ManaPool.Add(mana.U, blue)
	giveMainPhasePriority(g, game.Player1)
	return overlappingExileRouteScene{g: g, engine: NewEngine(nil), evelyn: evelyn, cardID: cardID, artifact: artifact, warded: warded}
}

func (s overlappingExileRouteScene) assertListed(t *testing.T, act action.Action, want bool) {
	t.Helper()
	assertExileCastListed(t, s.g, s.engine, act, want)
	if got := actionsContain(s.engine.legalActions(s.g.Clone(), game.Player1), act); got != want {
		t.Fatalf("cloned game legal actions contain %v = %v, want %v", act, got, want)
	}
}

// spendEvelyn casts another collection-countered card through Evelyn's
// permission alone, spending her once-each-turn use, and resolves it.
func (s overlappingExileRouteScene) spendEvelyn(t *testing.T) {
	t.Helper()
	other := collectionExiledCard(s.g, kickerExileFixture())
	s.g.Players[game.Player1].ManaPool.Add(mana.U, 1)
	if !s.engine.applyAction(s.g, game.Player1, action.CastSpellFaceFromZone(other, zone.Exile, game.FaceFront, nil, 0, nil)) {
		t.Fatal("casting another card through Evelyn rejected")
	}
	s.engine.resolveTopOfStack(s.g, &TurnLog{})
	giveMainPhasePriority(s.g, game.Player1)
	if !s.g.ExilePlayPermissionUsedThisTurn[s.evelyn.ObjectID] {
		t.Fatal("precondition: Evelyn's once-each-turn use is not spent")
	}
}

func (s overlappingExileRouteScene) artifactsDestroyed() int {
	destroyed := 0
	for _, perm := range []*game.Permanent{s.artifact, s.warded} {
		if _, ok := permanentByObjectID(s.g, perm.ObjectID); !ok {
			destroyed++
		}
	}
	return destroyed
}

// TestOverlappingExilePermissionsChooseOneRoute casts registered Vandalblast
// from exile under both a free per-card permission and registered Evelyn.
// The free route is the default and leaves Evelyn's use unspent; choosing
// Evelyn's normal cost pays {R} with blue mana and spends her use; an
// overloaded cast can't use the free-only permission (CR 118.9a), so it is
// listed and applied through Evelyn's route alone and spends her use.
func TestOverlappingExilePermissionsChooseOneRoute(t *testing.T) {
	t.Parallel()
	target := func(s overlappingExileRouteScene) []game.Target {
		return []game.Target{game.PermanentTarget(s.artifact.ObjectID)}
	}

	t.Run("free route is the default and leaves Evelyn unspent", func(t *testing.T) {
		t.Parallel()
		s := newOverlappingExileRouteScene(t, "Vandalblast", 0)
		act := action.CastSpellFaceFromZone(s.cardID, zone.Exile, game.FaceFront, target(s), 0, nil)
		s.assertListed(t, act, true)
		if !s.engine.applyAction(s.g, game.Player1, act) {
			t.Fatal("free cast through the per-card permission rejected")
		}
		if obj, ok := s.g.Stack.Peek(); !ok || obj.SourceID != s.cardID || obj.Overloaded || obj.ManaSpentToCast != 0 {
			t.Fatalf("stack object = %+v, want the single-target free cast", obj)
		}
		if s.g.ExilePlayPermissionUsedThisTurn[s.evelyn.ObjectID] {
			t.Fatal("the free permission's cast spent Evelyn's once-each-turn use")
		}
		s.engine.resolveTopOfStack(s.g, &TurnLog{})
		if got := s.artifactsDestroyed(); got != 1 {
			t.Fatalf("destroyed artifacts = %d, want the one target", got)
		}
	})

	t.Run("choosing Evelyn's normal cost spends her use", func(t *testing.T) {
		t.Parallel()
		s := newOverlappingExileRouteScene(t, "Vandalblast", 1)
		var offered []string
		agents := [game.NumPlayers]PlayerAgent{game.Player1: routeLabelAgent{label: "Normal cost", offered: &offered}}
		act := action.CastSpellFaceFromZone(s.cardID, zone.Exile, game.FaceFront, target(s), 0, nil)
		if !s.engine.applyActionWithChoices(s.g, game.Player1, act, agents, &TurnLog{}) {
			t.Fatal("cast through Evelyn's route rejected")
		}
		if !slices.Contains(offered, "Without paying its mana cost") || !slices.Contains(offered, "Normal cost") {
			t.Fatalf("offered %q, want both permissions' routes", offered)
		}
		if obj, ok := s.g.Stack.Peek(); !ok || obj.SourceID != s.cardID || obj.ManaSpentToCast != 1 {
			t.Fatalf("stack object = %+v, want {R} paid with one blue mana", obj)
		}
		if !s.g.ExilePlayPermissionUsedThisTurn[s.evelyn.ObjectID] {
			t.Fatal("casting through Evelyn's route did not spend her once-each-turn use")
		}
	})

	t.Run("overload casts only through Evelyn", func(t *testing.T) {
		t.Parallel()
		s := newOverlappingExileRouteScene(t, "Vandalblast", 5)
		act := action.CastOverloadedSpellFaceFromZone(s.cardID, zone.Exile, game.FaceFront, nil)
		s.assertListed(t, act, true)
		if !s.engine.applyAction(s.g, game.Player1, act) {
			t.Fatal("overloaded cast through Evelyn's route rejected")
		}
		if obj, ok := s.g.Stack.Peek(); !ok || !obj.Overloaded || obj.ManaSpentToCast != 5 {
			t.Fatalf("stack object = %+v, want overload {4}{R} paid with five blue mana", obj)
		}
		if !s.g.ExilePlayPermissionUsedThisTurn[s.evelyn.ObjectID] {
			t.Fatal("the overloaded cast did not spend Evelyn's once-each-turn use")
		}
		s.engine.resolveTopOfStack(s.g, &TurnLog{})
		if got := s.artifactsDestroyed(); got != 2 {
			t.Fatalf("destroyed artifacts = %d, want both, including the hexproof one", got)
		}
	})

	t.Run("overload is not listed once Evelyn is spent", func(t *testing.T) {
		t.Parallel()
		s := newOverlappingExileRouteScene(t, "Vandalblast", 5)
		s.spendEvelyn(t)
		overload := action.CastOverloadedSpellFaceFromZone(s.cardID, zone.Exile, game.FaceFront, nil)
		s.assertListed(t, overload, false)
		if s.engine.applyAction(s.g, game.Player1, overload) {
			t.Fatal("the free-only permission authorized overload after Evelyn was spent")
		}
		single := action.CastSpellFaceFromZone(s.cardID, zone.Exile, game.FaceFront, target(s), 0, nil)
		s.assertListed(t, single, true)
		if !s.engine.applyAction(s.g, game.Player1, single) {
			t.Fatal("the free permission's single-target cast rejected")
		}
		if obj, ok := s.g.Stack.Peek(); !ok || obj.Overloaded || obj.ManaSpentToCast != 0 {
			t.Fatalf("stack object = %+v, want the free cast with the mana unspent", obj)
		}
	})

	t.Run("overload without mana is neither listed nor applied", func(t *testing.T) {
		t.Parallel()
		s := newOverlappingExileRouteScene(t, "Vandalblast", 4)
		overload := action.CastOverloadedSpellFaceFromZone(s.cardID, zone.Exile, game.FaceFront, nil)
		s.assertListed(t, overload, false)
		if s.engine.applyAction(s.g, game.Player1, overload) {
			t.Fatal("an unpayable overload cast was applied through the free permission")
		}
		if s.g.ExilePlayPermissionUsedThisTurn[s.evelyn.ObjectID] || !s.g.Stack.IsEmpty() {
			t.Fatal("a rejected cast changed the stack or spent Evelyn's use")
		}
	})
}

// TestOverlappingExilePermissionsBestowThroughPayingRoute casts registered
// Boon Satyr from exile under both permissions. Bestow can't combine with the
// free-only permission (CR 118.9a, Boon Satyr ruling), so the bestowed cast is
// listed and applied through Evelyn's route, paying {3}{G}{G} with blue mana
// and spending her use, while the creature cast defaults to the free route.
func TestOverlappingExilePermissionsBestowThroughPayingRoute(t *testing.T) {
	t.Parallel()
	t.Run("bestow", func(t *testing.T) {
		t.Parallel()
		s := newOverlappingExileRouteScene(t, "Boon Satyr", 5)
		act := action.CastBestowSpellFaceFromZone(s.cardID, zone.Exile, game.FaceFront, []game.Target{game.PermanentTarget(s.evelyn.ObjectID)}, 0, nil)
		s.assertListed(t, act, true)
		if !s.engine.applyAction(s.g, game.Player1, act) {
			t.Fatal("bestowed cast through Evelyn's route rejected")
		}
		if obj, ok := s.g.Stack.Peek(); !ok || !obj.Bestowed || obj.ManaSpentToCast != 5 {
			t.Fatalf("stack object = %+v, want bestow paid with five blue mana", obj)
		}
		if !s.g.ExilePlayPermissionUsedThisTurn[s.evelyn.ObjectID] {
			t.Fatal("the bestowed cast did not spend Evelyn's once-each-turn use")
		}
	})
	t.Run("bestow is not listed once Evelyn is spent", func(t *testing.T) {
		t.Parallel()
		s := newOverlappingExileRouteScene(t, "Boon Satyr", 5)
		s.spendEvelyn(t)
		act := action.CastBestowSpellFaceFromZone(s.cardID, zone.Exile, game.FaceFront, []game.Target{game.PermanentTarget(s.evelyn.ObjectID)}, 0, nil)
		s.assertListed(t, act, false)
		if s.engine.applyAction(s.g, game.Player1, act) {
			t.Fatal("the free-only permission authorized bestow")
		}
	})
	t.Run("creature cast defaults to the free route", func(t *testing.T) {
		t.Parallel()
		s := newOverlappingExileRouteScene(t, "Boon Satyr", 5)
		act := action.CastSpellFaceFromZone(s.cardID, zone.Exile, game.FaceFront, nil, 0, nil)
		s.assertListed(t, act, true)
		if !s.engine.applyAction(s.g, game.Player1, act) {
			t.Fatal("creature cast rejected")
		}
		if obj, ok := s.g.Stack.Peek(); !ok || obj.Bestowed || obj.ManaSpentToCast != 0 {
			t.Fatalf("stack object = %+v, want the free creature cast", obj)
		}
		if s.g.ExilePlayPermissionUsedThisTurn[s.evelyn.ObjectID] {
			t.Fatal("the free creature cast spent Evelyn's use")
		}
	})
}
