package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Parcelbeast (oracle de83b309-ac31-4a57-95e6-ccd3659e6c1f),
// pinned corpus 6da9ced0…a1b4: {2}{G}{U} Creature — Elemental Beast, "Mutate
// {G}{U} (If you cast this spell for its mutate cost, put it over or under
// target non-Human creature you own. …)". Scryfall rulings (2020-04-17) cover
// only its activated ability. Comprehensive Rules effective 2026-09-25 (local
// copy of the official TXT): 702.140a (mutate is an alternative cost; the
// mutating spell targets a non-Human creature with the same owner as the
// spell, following 601.2b and 601.2f–h), 702.140b (illegal target: resolves
// as a creature spell), 702.140c (legal target: merges), 118.9a/601.2b (only
// one alternative cost or alternative casting method applies, so a permission
// that casts without paying the mana cost cannot carry a mutate cast) and
// 715.3d (the Adventure exile permission belongs to the Adventure's
// controller). Nothing in these rules limits mutate to a cast from hand: any
// effect that lets the owner cast the creature from exile allows its mutate
// alternative cost.

// parcelbeastExileFixture puts default-registry Parcelbeast into owner's exile
// with a non-Human creature target owned by owner, a Forest and an Island for
// Player1, and Player1 holding main-phase priority.
type parcelbeastExileFixture struct {
	game   *game.Game
	engine *Engine
	cardID id.ID
	target *game.Permanent
}

func newParcelbeastExileFixture(t *testing.T, owner game.PlayerID) parcelbeastExileFixture {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	target := addCombatPermanent(g, owner, mutateTargetCard())
	cardID := addCardToExile(g, owner, registryCardDef(t, "Parcelbeast"))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Island)
	giveMainPhasePriority(g, game.Player1)
	return parcelbeastExileFixture{game: g, engine: NewEngine(nil), cardID: cardID, target: target}
}

func (f parcelbeastExileFixture) mutate() action.Action {
	return action.CastMutateSpellFromZone(f.cardID, zone.Exile, f.target.ObjectID)
}

// assertMutate checks that legal-action enumeration and direct validation agree
// on Player1's mutate cast of the fixture card from exile.
func (f parcelbeastExileFixture) assertMutate(t *testing.T, want bool) {
	t.Helper()
	if got := actionsContain(f.engine.legalActions(f.game, game.Player1), f.mutate()); got != want {
		t.Fatalf("legal actions contain the exile mutate cast = %v, want %v", got, want)
	}
	if got := canCastMutateSpell(f.game, game.Player1, f.cardID, zone.Exile, f.target.ObjectID); got != want {
		t.Fatalf("exile mutate validation = %v, want %v", got, want)
	}
}

// castMutate applies the fixture's mutate cast and checks the resulting stack
// object, payment and spell-cast event.
func (f parcelbeastExileFixture) castMutate(t *testing.T) {
	t.Helper()
	g := f.game
	if !f.engine.applyAction(g, game.Player1, f.mutate()) {
		t.Fatal("Player1 could not mutate-cast Parcelbeast from exile")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != f.cardID || !obj.Mutate || obj.MutateTargetID != f.target.ObjectID ||
		obj.SourceZone != zone.Exile || obj.Controller != game.Player1 || obj.ManaSpentToCast != 2 {
		t.Fatalf("stack object = %+v, want Player1's {G}{U} mutate cast from exile", obj)
	}
	for _, player := range g.Players {
		if player.Exile.Contains(f.cardID) {
			t.Fatal("mutate-cast Parcelbeast is still in exile")
		}
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == f.cardID && event.FromZone == zone.Exile && event.Controller == game.Player1
	})
}

func perCardExileEffect(g *game.Game, kind game.RuleEffectKind, controller game.PlayerID, cardID id.ID) game.RuleEffect {
	return game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           kind,
		Controller:     controller,
		AffectedPlayer: game.PlayerYou,
		CastFromZone:   zone.Exile,
		AffectedCardID: cardID,
	}
}

// TestMutateFromExileUnderOrdinaryPermissions proves that an ordinary
// per-card or group permission to cast or play a card from exile authorizes
// its mutate alternative cost (CR 702.140a, 118.9a) for a card no Adventure
// ever exiled, through enumeration, validation and application, and that a
// permission that does not reach the card or cannot carry an alternative cost
// does not.
func TestMutateFromExileUnderOrdinaryPermissions(t *testing.T) {
	t.Parallel()
	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastExileFixture(t, game.Player1)
		f.assertMutate(t, false)
		if f.engine.applyAction(f.game, game.Player1, f.mutate()) {
			t.Fatal("Parcelbeast mutate-cast from exile with no permission")
		}
		if !f.game.Players[game.Player1].Exile.Contains(f.cardID) || f.game.Stack.Size() != 0 {
			t.Fatal("rejected mutate cast moved Parcelbeast out of exile")
		}
	})
	for name, kind := range map[string]game.RuleEffectKind{"cast": game.RuleEffectCastFromZone, "play": game.RuleEffectPlayFromZone} {
		t.Run("per-card "+name+" permission", func(t *testing.T) {
			t.Parallel()
			f := newParcelbeastExileFixture(t, game.Player1)
			f.game.RuleEffects = append(f.game.RuleEffects, perCardExileEffect(f.game, kind, game.Player1, f.cardID))
			f.assertMutate(t, true)
			f.castMutate(t)
		})
	}
	t.Run("group cast-spells permission", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastExileFixture(t, game.Player1)
		f.game.RuleEffects = append(f.game.RuleEffects, castSpellsFromExileEffect(f.game, game.Player1, types.Creature))
		f.assertMutate(t, true)
		f.castMutate(t)
	})
	t.Run("group permission for other spell types", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastExileFixture(t, game.Player1)
		f.game.RuleEffects = append(f.game.RuleEffects, castSpellsFromExileEffect(f.game, game.Player1, types.Instant, types.Sorcery))
		f.assertMutate(t, false)
	})
	t.Run("per-card permission for another card", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastExileFixture(t, game.Player1)
		other := addCardToExile(f.game, game.Player1, registryCardDef(t, "Parcelbeast"))
		f.game.RuleEffects = append(f.game.RuleEffects, perCardExileEffect(f.game, game.RuleEffectCastFromZone, game.Player1, other))
		f.assertMutate(t, false)
	})
	t.Run("another player's per-card permission", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastExileFixture(t, game.Player1)
		f.game.RuleEffects = append(f.game.RuleEffects, perCardExileEffect(f.game, game.RuleEffectCastFromZone, game.Player3, f.cardID))
		f.assertMutate(t, false)
		if f.engine.applyAction(f.game, game.Player1, f.mutate()) {
			t.Fatal("Player1 mutate-cast through Player3's permission")
		}
	})
	t.Run("unaffordable mutate cost", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastExileFixture(t, game.Player1)
		f.game.RuleEffects = append(f.game.RuleEffects, perCardExileEffect(f.game, game.RuleEffectCastFromZone, game.Player1, f.cardID))
		for _, permanent := range f.game.Battlefield {
			if permanentHasSubtype(f.game, permanent, types.Island) {
				permanent.Tapped = true
			}
		}
		f.assertMutate(t, false)
		if f.engine.applyAction(f.game, game.Player1, f.mutate()) {
			t.Fatal("Parcelbeast mutate-cast for {G}{U} with only a Forest untapped")
		}
		if !f.game.Players[game.Player1].Exile.Contains(f.cardID) || f.game.Stack.Size() != 0 {
			t.Fatal("unpaid mutate cast did not return Parcelbeast to exile")
		}
	})
	t.Run("Human target", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastExileFixture(t, game.Player1)
		f.game.RuleEffects = append(f.game.RuleEffects, perCardExileEffect(f.game, game.RuleEffectCastFromZone, game.Player1, f.cardID))
		f.game.CardInstances[f.target.CardInstanceID].Def.Subtypes = []types.Sub{types.Human}
		f.assertMutate(t, false)
	})
	// CR 118.9a, 601.2b: a permission to cast the card without paying its mana
	// cost is itself an alternative cost, so it cannot carry mutate.
	t.Run("free permission only", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastExileFixture(t, game.Player1)
		free := perCardExileEffect(f.game, game.RuleEffectPlayFromZone, game.Player1, f.cardID)
		free.WithoutPayingManaCost = true
		f.game.RuleEffects = append(f.game.RuleEffects, free)
		plain := action.CastSpellFaceFromZone(f.cardID, zone.Exile, game.FaceFront, nil, 0, nil)
		if !actionsContain(f.engine.legalActions(f.game, game.Player1), plain) {
			t.Fatal("the free permission does not offer Parcelbeast's ordinary free cast")
		}
		f.assertMutate(t, false)
		if f.engine.applyAction(f.game, game.Player1, f.mutate()) {
			t.Fatal("Parcelbeast mutate-cast through a without-paying-its-mana-cost permission")
		}
	})
	t.Run("free and paying permissions", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastExileFixture(t, game.Player1)
		free := perCardExileEffect(f.game, game.RuleEffectPlayFromZone, game.Player1, f.cardID)
		free.WithoutPayingManaCost = true
		f.game.RuleEffects = append(f.game.RuleEffects, free, castSpellsFromExileEffect(f.game, game.Player1))
		f.assertMutate(t, true)
		f.castMutate(t)
	})
}

// TestMutateFromExileSpendsOncePerTurnPermission proves a mutate cast through a
// once-per-turn group permission spends that permission's use.
func TestMutateFromExileSpendsOncePerTurnPermission(t *testing.T) {
	t.Parallel()
	f := newParcelbeastExileFixture(t, game.Player1)
	source := addCombatPermanent(f.game, game.Player1, mutateTargetCard())
	f.game.CardInstances[source.CardInstanceID].Def.Types = []types.Card{types.Artifact}
	permission := castSpellsFromExileEffect(f.game, game.Player1)
	permission.OncePerTurn = true
	permission.SourceObjectID = source.ObjectID
	permission.SourceCardID = source.CardInstanceID
	f.game.RuleEffects = append(f.game.RuleEffects, permission)
	f.assertMutate(t, true)
	f.castMutate(t)
	if !f.game.ExilePlayPermissionUsedThisTurn[source.ObjectID] {
		t.Fatal("the mutate cast did not spend the once-per-turn permission")
	}
	other := addCardToExile(f.game, game.Player1, registryCardDef(t, "Parcelbeast"))
	if canCastMutateSpell(f.game, game.Player1, other, zone.Exile, f.target.ObjectID) {
		t.Fatal("a second card mutate-cast through the spent once-per-turn permission")
	}
}

// TestMutateFromAnotherPlayersExileTargetsSpellOwnersCreature casts an
// opponent's exiled Parcelbeast through Player1's per-card permission: the
// mutating spell must target a creature with the same owner as the spell
// (CR 702.140a), not one its caster owns.
func TestMutateFromAnotherPlayersExileTargetsSpellOwnersCreature(t *testing.T) {
	t.Parallel()
	f := newParcelbeastExileFixture(t, game.Player2)
	ownTarget := addCombatPermanent(f.game, game.Player1, mutateTargetCard())
	f.game.RuleEffects = append(f.game.RuleEffects, perCardExileEffect(f.game, game.RuleEffectCastFromZone, game.Player1, f.cardID))
	if canCastMutateSpell(f.game, game.Player1, f.cardID, zone.Exile, ownTarget.ObjectID) ||
		actionsContain(f.engine.legalActions(f.game, game.Player1), action.CastMutateSpellFromZone(f.cardID, zone.Exile, ownTarget.ObjectID)) {
		t.Fatal("Player2's Parcelbeast could target a creature Player1 owns")
	}
	f.assertMutate(t, true)
	f.castMutate(t)
}

// TestMutateFromExileResolution resolves the exile mutate cast: a legal target
// merges (CR 702.140c), and an illegal target leaves an ordinary creature
// spell that enters under the caster's control (CR 702.140b). A cloned game
// casts independently of the original.
func TestMutateFromExileResolution(t *testing.T) {
	t.Parallel()
	setup := func(t *testing.T) parcelbeastExileFixture {
		t.Helper()
		f := newParcelbeastExileFixture(t, game.Player1)
		f.game.RuleEffects = append(f.game.RuleEffects, perCardExileEffect(f.game, game.RuleEffectCastFromZone, game.Player1, f.cardID))
		return f
	}
	t.Run("legal target merges", func(t *testing.T) {
		t.Parallel()
		f := setup(t)
		f.castMutate(t)
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
		f.engine.resolveTopOfStackWithChoices(f.game, agents, &TurnLog{})
		merged, ok := permanentByObjectID(f.game, f.target.ObjectID)
		if !ok || merged.CardInstanceID != f.cardID || len(merged.MergedCards) != 1 {
			t.Fatalf("mutated permanent = %+v, want Parcelbeast over the target", merged)
		}
		if got := permanentEffectiveName(f.game, merged); got != "Parcelbeast" {
			t.Fatalf("merged permanent name = %q, want Parcelbeast", got)
		}
		assertEvent(t, f.game.Events, game.EventPermanentMutated, func(event game.Event) bool {
			return event.PermanentID == f.target.ObjectID && event.CardID == f.cardID
		})
	})
	t.Run("illegal target resolves as creature", func(t *testing.T) {
		t.Parallel()
		f := setup(t)
		f.castMutate(t)
		if !movePermanentToZone(f.game, f.target, zone.Graveyard) {
			t.Fatal("failed to remove the mutate target")
		}
		f.engine.resolveTopOfStack(f.game, &TurnLog{})
		parcelbeast := permanentWithCardID(f.game, f.cardID)
		if parcelbeast == nil || len(parcelbeast.MergedCards) != 0 || parcelbeast.Controller != game.Player1 {
			t.Fatalf("resolved Parcelbeast = %+v, want an ordinary creature under Player1's control", parcelbeast)
		}
	})
	t.Run("clone", func(t *testing.T) {
		t.Parallel()
		f := setup(t)
		clone := f.game.Clone()
		cloned := parcelbeastExileFixture{game: clone, engine: f.engine, cardID: f.cardID, target: f.target}
		if p, ok := permanentByObjectID(clone, f.target.ObjectID); ok {
			cloned.target = p
		}
		cloned.assertMutate(t, true)
		cloned.castMutate(t)
		if !f.game.Players[game.Player1].Exile.Contains(f.cardID) || f.game.Stack.Size() != 0 {
			t.Fatal("casting in the clone changed the original game")
		}
		f.castMutate(t)
	})
}
