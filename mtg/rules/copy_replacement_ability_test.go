package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: issue #3389. Pinned corpus Oracle text:
// Torbran, Thane of Red Fell (ELD 147, 79f591cd-d277-4ba5-b1bf-1c09cac9cb8a)
// "If a red source you control would deal damage to an opponent or a permanent
// an opponent controls, it deals that much damage plus 2 instead." Its three
// 2019-10-04 rulings keep the extra damage on the original source.
// Mirage Mirror (e516bc14-749c-4da6-b5d3-cb5c16316e0c) "{2}: This artifact
// becomes a copy of target artifact, creature, enchantment, or land until end
// of turn."; 2017-07-14 rulings: it copies copiable values and each resolution
// overwrites what it is copying. Clone (5e648262-3b9b-4c58-8e29-48356e3cb064)
// "You may have this creature enter as a copy of any creature on the
// battlefield."; 2025-10-02 rulings: it copies exactly the printed values.
// Lightning Bolt deals 3 damage to any target; Grizzly Bears is vanilla.
// CR 2026-09-25 (verified official copy): 707.2 copiable values, 613.2a copy
// effects in layer 1, 707.9a copy exceptions become copiable values, 611.3a-b
// static abilities apply only while the object has them, 613.1f ability loss,
// 708.2 face-down permanents have no abilities, 614.1 replacement effects.

func copyReplacementGame(t *testing.T) (*game.Game, *Engine, *cards.Registry) {
	t.Helper()
	registry := cards.NewDefaultRegistry()
	return game.NewGame([game.NumPlayers]game.PlayerConfig{}), NewEngineWithPrograms(nil, registry.Programs()), registry
}

func copyReplacementCard(t *testing.T, registry *cards.Registry, name string) *game.CardDef {
	t.Helper()
	def := registry.Lookup(name)
	if def == nil {
		t.Fatalf("production %s is missing", name)
	}
	return def
}

// boltOpponent casts and resolves a real Lightning Bolt at Player2 and returns
// the damage dealt.
func boltOpponent(t *testing.T, g *game.Game, e *Engine, registry *cards.Registry) int {
	t.Helper()
	boltID := addCardToHand(g, game.Player1, copyReplacementCard(t, registry, "Lightning Bolt"))
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	setMainPhasePriority(g, game.Player1)
	if !e.applyAction(g, game.Player1, action.CastSpell(boltID, []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)) {
		t.Fatal("Lightning Bolt cast failed")
	}
	before := g.Players[game.Player2].Life
	e.resolveTopOfStack(g, &TurnLog{})
	return before - g.Players[game.Player2].Life
}

func activateMirageMirror(t *testing.T, g *game.Game, e *Engine, mirror, target *game.Permanent) {
	t.Helper()
	putMirageMirrorActivation(t, g, e, mirror, target)
	e.resolveTopOfStack(g, &TurnLog{})
	if got, want := permanentEffectiveName(g, mirror), permanentEffectiveName(g, target); got != want {
		t.Fatalf("Mirage Mirror is %q, want copy of %q", got, want)
	}
}

func putMirageMirrorActivation(t *testing.T, g *game.Game, e *Engine, mirror, target *game.Permanent) {
	t.Helper()
	g.Players[game.Player1].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)
	if !e.applyAction(g, game.Player1, action.ActivateAbility(
		mirror.ObjectID, 0, []game.Target{game.PermanentTarget(target.ObjectID)}, 0,
	)) {
		t.Fatal("Mirage Mirror activation failed")
	}
}

func sourceReplacementOrigins(g *game.Game, source *game.Permanent) []id.ID {
	var origins []id.ID
	for i := range g.ReplacementEffects {
		replacement := &g.ReplacementEffects[i]
		if replacement.FromSourceAbility && replacement.SourceObjectID == source.ObjectID {
			origins = append(origins, replacement.SourceAbilityOrigin)
		}
	}
	return origins
}

func TestCopyEffectSuspendsAndRestoresOriginalReplacementAbility(t *testing.T) {
	g, e, registry := copyReplacementGame(t)
	torbran := addReplacementPermanent(t, g, game.Player1, copyReplacementCard(t, registry, "Torbran, Thane of Red Fell"))
	bears := addCombatPermanent(g, game.Player2, copyReplacementCard(t, registry, "Grizzly Bears"))
	if got := boltOpponent(t, g, e, registry); got != 5 {
		t.Fatalf("Torbran's bolt dealt %d, want 5", got)
	}

	// While a copy effect makes Torbran a Grizzly Bears, the copy's copiable
	// values supply its abilities, so its printed replacement does not apply.
	becomeStaticCopy(t, g, torbran, bears, true)
	if got := boltOpponent(t, g, e, registry); got != 3 {
		t.Fatalf("bolt while Torbran copies Grizzly Bears dealt %d, want 3", got)
	}
	branch := g.Clone()

	// When the copy effect ends, the printed ability functions again.
	expireCleanupDurations(g)
	if permanentEffectiveName(g, torbran) != "Torbran, Thane of Red Fell" {
		t.Fatal("copy effect did not end at cleanup")
	}
	if got := boltOpponent(t, g, e, registry); got != 5 {
		t.Fatalf("bolt after the copy ended dealt %d, want 5", got)
	}
	if got := boltOpponent(t, branch, e, registry); got != 3 {
		t.Fatalf("cloned game shared copy expiry: bolt dealt %d, want 3", got)
	}
}

func TestMirageMirrorCopyInstallsAndExpiresCopiedReplacementAbility(t *testing.T) {
	g, e, registry := copyReplacementGame(t)
	torbran := addReplacementPermanent(t, g, game.Player1, copyReplacementCard(t, registry, "Torbran, Thane of Red Fell"))
	mirror := addReplacementPermanent(t, g, game.Player1, copyReplacementCard(t, registry, "Mirage Mirror"))
	activateMirageMirror(t, g, e, mirror, torbran)

	// Both Torbran and the Mirror-as-Torbran add 2 to the same red source.
	if got := boltOpponent(t, g, e, registry); got != 7 {
		t.Fatalf("bolt with Torbran and Mirror-Torbran dealt %d, want 7", got)
	}
	// The copied ability belongs to the Mirror, independent of the original.
	movePermanentToZone(g, torbran, zone.Graveyard)
	if got := boltOpponent(t, g, e, registry); got != 5 {
		t.Fatalf("bolt with only Mirror-Torbran dealt %d, want 5", got)
	}

	expireCleanupDurations(g)
	if got := boltOpponent(t, g, e, registry); got != 3 {
		t.Fatalf("bolt after Mirror's copy expired dealt %d, want 3", got)
	}
	if origins := sourceReplacementOrigins(g, mirror); len(origins) != 0 {
		t.Fatalf("expired copy left Mirror replacements with origins %v", origins)
	}
}

func TestOverwritingCopySuspendsEarlierCopiedReplacementAbility(t *testing.T) {
	g, e, registry := copyReplacementGame(t)
	torbran := addReplacementPermanent(t, g, game.Player1, copyReplacementCard(t, registry, "Torbran, Thane of Red Fell"))
	bears := addCombatPermanent(g, game.Player2, copyReplacementCard(t, registry, "Grizzly Bears"))
	mirror := addReplacementPermanent(t, g, game.Player1, copyReplacementCard(t, registry, "Mirage Mirror"))
	// Activated in response to itself: the Torbran copy resolves first, then
	// the Grizzly Bears copy overwrites it (Mirage Mirror 2017-07-14 ruling).
	putMirageMirrorActivation(t, g, e, mirror, bears)
	putMirageMirrorActivation(t, g, e, mirror, torbran)
	e.resolveTopOfStack(g, &TurnLog{})
	if permanentEffectiveName(g, mirror) != "Torbran, Thane of Red Fell" {
		t.Fatal("first resolution did not make Mirror a copy of Torbran")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if permanentEffectiveName(g, mirror) != "Grizzly Bears" {
		t.Fatal("second resolution did not overwrite Mirror's copy with Grizzly Bears")
	}
	if origins := sourceReplacementOrigins(g, mirror); len(origins) != 1 || origins[0] == permanentCopiableAbilityOrigin(g, mirror) {
		t.Fatalf("Mirror replacement origins = %v, want only the overwritten Torbran copy", origins)
	}

	if got := boltOpponent(t, g, e, registry); got != 5 {
		t.Fatalf("bolt after Mirror overwrote its Torbran copy dealt %d, want 5", got)
	}
}

func TestCloneEnteringAsCopyGainsCopiedReplacementAbility(t *testing.T) {
	g, e, registry := copyReplacementGame(t)
	torbran := addReplacementPermanent(t, g, game.Player1, copyReplacementCard(t, registry, "Torbran, Thane of Red Fell"))
	for range 4 {
		addBasicLandPermanent(g, game.Player1, types.Island)
	}
	cloneID := addCardToHand(g, game.Player1, copyReplacementCard(t, registry, "Clone"))
	setMainPhasePriority(g, game.Player1)
	if !e.applyAction(g, game.Player1, action.CastSpell(cloneID, nil, 0, nil)) {
		t.Fatal("Clone cast failed")
	}
	agents := simPassPolicies()
	agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{1}}}
	e.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	clone := permanentByCardID(g, cloneID)
	if clone == nil || permanentEffectiveName(g, clone) != "Torbran, Thane of Red Fell" {
		t.Fatal("Clone did not enter as a copy of Torbran")
	}
	// Remove the original so only the copied ability can add damage.
	movePermanentToZone(g, torbran, zone.Graveyard)
	if got := boltOpponent(t, g, e, registry); got != 5 {
		t.Fatalf("bolt with Clone-Torbran dealt %d, want 5", got)
	}
	if origins := sourceReplacementOrigins(g, clone); len(origins) != 1 || origins[0] != permanentCopiableAbilityOrigin(g, clone) || origins[0] == 0 {
		t.Fatalf("Clone replacement origins = %v, want one from its copy effect", origins)
	}
}

func TestAbilityLossAndFaceDownSuppressSourceReplacementAbilities(t *testing.T) {
	t.Run("copied ability lost", func(t *testing.T) {
		g, e, registry := copyReplacementGame(t)
		torbran := addReplacementPermanent(t, g, game.Player1, copyReplacementCard(t, registry, "Torbran, Thane of Red Fell"))
		mirror := addReplacementPermanent(t, g, game.Player1, copyReplacementCard(t, registry, "Mirage Mirror"))
		activateMirageMirror(t, g, e, mirror, torbran)
		movePermanentToZone(g, torbran, zone.Graveyard)
		g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
			ID: g.IDGen.Next(), Timestamp: 1 << 30, AffectedObjectID: mirror.ObjectID,
			Layer: game.LayerAbility, RemoveAllAbilities: true,
		})
		if got := boltOpponent(t, g, e, registry); got != 3 {
			t.Fatalf("bolt after Mirror-Torbran lost all abilities dealt %d, want 3", got)
		}
	})
	t.Run("face down", func(t *testing.T) {
		g, e, registry := copyReplacementGame(t)
		torbran := addReplacementPermanent(t, g, game.Player1, copyReplacementCard(t, registry, "Torbran, Thane of Red Fell"))
		torbran.FaceDown = true
		torbran.FaceDownKind = game.FaceDownMorph
		if got := boltOpponent(t, g, e, registry); got != 3 {
			t.Fatalf("bolt with face-down Torbran dealt %d, want 3", got)
		}
		torbran.FaceDown = false
		if got := boltOpponent(t, g, e, registry); got != 5 {
			t.Fatalf("bolt after Torbran turned face up dealt %d, want 5", got)
		}
	})
}
