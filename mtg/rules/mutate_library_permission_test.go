package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Engine gap #3411. Pinned corpus 6da9ced0…a1b4:
// Parcelbeast (oracle de83b309-ac31-4a57-95e6-ccd3659e6c1f) {2}{G}{U}
// Creature — Elemental Beast, "Mutate {G}{U} …"; Future Sight (oracle
// 93a0e938-b916-47ec-9803-369e885e253c) "Play with the top card of your library
// revealed. You may play lands and cast spells from the top of your library.";
// Bolas's Citadel (oracle 2bd111bb-ce02-414c-b5b7-e0e037d8d96b) "… You may play
// lands and cast spells from the top of your library. If you cast a spell this
// way, pay life equal to its mana value rather than pay its mana cost. …".
// Scryfall rulings retrieved 2026-10-31: Future Sight 2019-06-14 "If the card
// has alternative or additional costs, you may pay those"; Citadel 2019-05-03
// "If you cast a spell for another cost 'rather than pay its mana cost,' you
// can't choose to cast it for any alternative costs"; Parcelbeast's three
// 2020-04-17 rulings cover only its activated ability. Comprehensive Rules
// effective 2026-09-25 (official TXT from magic.wizards.com/en/rules), read:
// 702.140a (mutate is an alternative cost following 601.2b and 601.2f–h; the
// spell targets a non-Human creature with the same owner), 702.140b/c
// (illegal target resolves as a creature spell; legal target merges), 118.9
// and 118.9a (only one alternative cost applies to a spell), 601.2b (no two
// alternative costs), 118.14 (a permission's "mana of any type" rider applies
// only to spells cast that way), 601.2a/601.2h/601.2i and 401.5 (the card
// moves to the stack before payment; a changed top card stays hidden until
// the spell becomes cast). Nothing limits mutate's source zone: Future Sight's
// permission carries the mutate cost; Citadel's life payment is itself an
// alternative cost and can't.

type parcelbeastLibraryFixture struct {
	game   *game.Game
	engine *Engine
	cardID id.ID
	below  id.ID
	target *game.Permanent
}

// newParcelbeastLibraryFixture puts default-registry Parcelbeast on top of
// Player1's library (over another card), an owned non-Human creature, and a
// Forest and an Island for Player1, with Player1 holding main-phase priority.
// permissions names default-registry permanents Player1 controls.
func newParcelbeastLibraryFixture(t *testing.T, permissions ...string) parcelbeastLibraryFixture {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, name := range permissions {
		addCombatPermanent(g, game.Player1, registryCardDef(t, name))
	}
	target := addCombatPermanent(g, game.Player1, mutateTargetCard())
	below := addLibraryTopOgre(g)
	cardID := addCardToLibrary(g, game.Player1, registryCardDef(t, "Parcelbeast"))
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addBasicLandPermanent(g, game.Player1, types.Island)
	g.Players[game.Player1].Life = 20
	giveMainPhasePriority(g, game.Player1)
	return parcelbeastLibraryFixture{game: g, engine: NewEngine(nil), cardID: cardID, below: below, target: target}
}

func (f parcelbeastLibraryFixture) mutate() action.Action {
	return action.CastMutateSpellFromZone(f.cardID, zone.Library, f.target.ObjectID)
}

// assertMutate checks that legal-action enumeration and direct validation agree
// on Player1's mutate cast of the fixture card from the library.
func (f parcelbeastLibraryFixture) assertMutate(t *testing.T, want bool) {
	t.Helper()
	if got := actionsContain(f.engine.legalActions(f.game, game.Player1), f.mutate()); got != want {
		t.Fatalf("legal actions contain the library mutate cast = %v, want %v", got, want)
	}
	if got := canCastMutateSpell(f.game, game.Player1, f.cardID, zone.Library, f.target.ObjectID); got != want {
		t.Fatalf("library mutate validation = %v, want %v", got, want)
	}
}

// assertRejected checks that applying the mutate cast fails and leaves the
// library, life and stack untouched.
func (f parcelbeastLibraryFixture) assertRejected(t *testing.T) {
	t.Helper()
	if f.engine.applyAction(f.game, game.Player1, f.mutate()) {
		t.Fatal("the library mutate cast was applied")
	}
	f.assertLibraryUntouched(t)
}

func (f parcelbeastLibraryFixture) assertLibraryUntouched(t *testing.T) {
	t.Helper()
	library := f.game.Players[game.Player1].Library.All()
	if len(library) != 2 || library[0] != f.cardID || library[1] != f.below {
		t.Fatalf("library = %v, want Parcelbeast still on top of %v", library, f.below)
	}
	if f.game.Stack.Size() != 0 || f.game.Players[game.Player1].Life != 20 {
		t.Fatalf("stack size %d life %d, want an empty stack and 20 life", f.game.Stack.Size(), f.game.Players[game.Player1].Life)
	}
}

// castMutate applies the fixture's mutate cast and checks the stack object,
// the {G}{U} mana payment, the unchanged life total and the spell-cast event.
func (f parcelbeastLibraryFixture) castMutate(t *testing.T) *game.StackObject {
	t.Helper()
	g := f.game
	if !f.engine.applyAction(g, game.Player1, f.mutate()) {
		t.Fatal("Player1 could not mutate-cast Parcelbeast from the top of the library")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.SourceID != f.cardID || !obj.Mutate || obj.MutateTargetID != f.target.ObjectID ||
		obj.SourceZone != zone.Library || obj.Controller != game.Player1 || obj.ManaSpentToCast != 2 {
		t.Fatalf("stack object = %+v, want Player1's {G}{U} mutate cast from the library", obj)
	}
	if len(obj.Targets) != 1 || obj.Targets[0] != game.PermanentTarget(f.target.ObjectID) {
		t.Fatalf("mutate targets = %+v, want the owned non-Human creature", obj.Targets)
	}
	if top, _ := g.Players[game.Player1].Library.Top(); top != f.below || g.Players[game.Player1].Library.Contains(f.cardID) {
		t.Fatalf("library top = %v, want Parcelbeast gone and %v on top", top, f.below)
	}
	if life := g.Players[game.Player1].Life; life != 20 {
		t.Fatalf("life = %d, want 20: the mutate cast pays no life", life)
	}
	assertEvent(t, g.Events, game.EventSpellCast, func(event game.Event) bool {
		return event.CardID == f.cardID && event.FromZone == zone.Library && event.Controller == game.Player1
	})
	return obj
}

// TestMutateFromLibraryUnderFutureSight proves default-registry Future Sight's
// cast-from-top permission allows Parcelbeast's mutate alternative cost
// (CR 702.140a, 118.9a) through enumeration, validation and application, and
// that the cast is rejected without that permission, for a card that is not
// the top card, for an illegal target, or when the mutate cost is unaffordable.
func TestMutateFromLibraryUnderFutureSight(t *testing.T) {
	t.Parallel()
	t.Run("Future Sight", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Future Sight")
		f.assertMutate(t, true)
		f.castMutate(t)
	})
	t.Run("no permission", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t)
		f.assertMutate(t, false)
		f.assertRejected(t)
	})
	t.Run("not the top card", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Future Sight")
		cover := addLibraryTopOgre(f.game)
		f.assertMutate(t, false)
		if f.engine.applyAction(f.game, game.Player1, f.mutate()) {
			t.Fatal("Parcelbeast mutate-cast from below the top of the library")
		}
		if top, _ := f.game.Players[game.Player1].Library.Top(); top != cover || f.game.Stack.Size() != 0 {
			t.Fatal("the rejected cast changed the library or stack")
		}
	})
	t.Run("another player's Future Sight", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t)
		addCombatPermanent(f.game, game.Player2, registryCardDef(t, "Future Sight"))
		f.assertMutate(t, false)
		f.assertRejected(t)
	})
	t.Run("Human target", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Future Sight")
		f.game.CardInstances[f.target.CardInstanceID].Def.Subtypes = []types.Sub{types.Human}
		f.assertMutate(t, false)
		f.assertRejected(t)
	})
	t.Run("creature owned by another player", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Future Sight")
		f.target = addCombatPermanent(f.game, game.Player2, mutateTargetCard())
		f.assertMutate(t, false)
		f.assertRejected(t)
	})
	t.Run("unaffordable mutate cost", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Future Sight")
		for _, permanent := range f.game.Battlefield {
			if permanentHasSubtype(f.game, permanent, types.Island) {
				permanent.Tapped = true
			}
		}
		f.assertMutate(t, false)
		// CR 728: the proposal is undone and the card returns to the top.
		f.assertRejected(t)
	})
}

// TestMutateFromLibraryBolasCitadel proves Bolas's Citadel's permission can't
// carry a mutate cast: its life payment is itself an alternative cost
// (CR 118.9a, 601.2b). With Future Sight also on the battlefield, the mutate
// cast goes through Future Sight's route and pays only the mutate cost.
func TestMutateFromLibraryBolasCitadel(t *testing.T) {
	t.Parallel()
	t.Run("Citadel only", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Bolas's Citadel")
		plain := action.CastSpellFaceFromZone(f.cardID, zone.Library, game.FaceFront, nil, 0, nil)
		if !actionsContain(f.engine.legalActions(f.game, game.Player1), plain) {
			t.Fatal("Citadel does not offer Parcelbeast's ordinary pay-life cast")
		}
		f.assertMutate(t, false)
		f.assertRejected(t)
	})
	for _, order := range [][]string{{"Bolas's Citadel", "Future Sight"}, {"Future Sight", "Bolas's Citadel"}} {
		t.Run(order[0]+" then "+order[1], func(t *testing.T) {
			t.Parallel()
			f := newParcelbeastLibraryFixture(t, order...)
			f.assertMutate(t, true)
			f.castMutate(t)
		})
	}
}

// TestMutateFromLibraryUsesOnlyChosenPermission proves only the chosen
// cast-from-library permission's spending rider and per-turn use apply to a
// mutate cast (CR 118.14).
func TestMutateFromLibraryUsesOnlyChosenPermission(t *testing.T) {
	t.Parallel()
	addOnceAnyMana := func(g *game.Game) *game.Permanent {
		return addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
			Name: "Test Once Any-Mana Seer",
			StaticAbilities: []game.StaticAbility{{
				Text: "Once each turn, you may cast a spell from the top of your library, and mana of any type can be spent to cast it.",
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectCastSpellsFromZone,
					AffectedPlayer: game.PlayerYou,
					CastFromZone:   zone.Library,
					TopCardOnly:    true,
					OncePerTurn:    true,
					SpendAnyMana:   game.ManaSpendAsAnyType,
				}},
			}},
		}})
	}
	tapLands := func(g *game.Game) {
		for _, permanent := range g.Battlefield {
			if permanentHasType(g, permanent, types.Land) {
				permanent.Tapped = true
			}
		}
	}
	t.Run("rider route pays with any mana and spends its use", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Future Sight")
		source := addOnceAnyMana(f.game)
		tapLands(f.game)
		f.game.Players[game.Player1].ManaPool.Add(mana.R, 2)
		f.assertMutate(t, true)
		f.castMutate(t)
		if !f.game.ExilePlayPermissionUsedThisTurn[source.ObjectID] {
			t.Fatal("the mutate cast through the once-per-turn rider did not spend its use")
		}
		next := addCardToLibrary(f.game, game.Player1, registryCardDef(t, "Parcelbeast"))
		f.game.Players[game.Player1].ManaPool.Add(mana.R, 2)
		if canCastMutateSpell(f.game, game.Player1, next, zone.Library, f.target.ObjectID) {
			t.Fatal("a second mutate cast spent red mana through the used once-per-turn permission")
		}
	})
	t.Run("plain route needs mutate colors", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Future Sight")
		tapLands(f.game)
		f.game.Players[game.Player1].ManaPool.Add(mana.R, 2)
		f.assertMutate(t, false)
		f.assertRejected(t)
	})
	t.Run("plain route leaves the once-per-turn use", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Future Sight")
		source := addOnceAnyMana(f.game)
		f.assertMutate(t, true)
		f.castMutate(t)
		if f.game.ExilePlayPermissionUsedThisTurn[source.ObjectID] {
			t.Fatal("the mutate cast spent the once-per-turn use while Future Sight offered the same cost")
		}
	})
}

// TestMutateFromLibraryResolution resolves the library mutate cast: a legal
// target merges (CR 702.140c), and an illegal target leaves an ordinary
// creature spell under the caster's control (CR 702.140b). A cloned game casts
// independently of the original.
func TestMutateFromLibraryResolution(t *testing.T) {
	t.Parallel()
	t.Run("legal target merges", func(t *testing.T) {
		t.Parallel()
		f := newParcelbeastLibraryFixture(t, "Future Sight")
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
		f := newParcelbeastLibraryFixture(t, "Future Sight")
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
		f := newParcelbeastLibraryFixture(t, "Future Sight")
		clone := f.game.Clone()
		cloned := f
		cloned.game = clone
		if p, ok := permanentByObjectID(clone, f.target.ObjectID); ok {
			cloned.target = p
		}
		cloned.assertMutate(t, true)
		cloned.castMutate(t)
		f.assertLibraryUntouched(t)
		f.castMutate(t)
	})
}
