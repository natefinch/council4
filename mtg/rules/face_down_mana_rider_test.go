package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Source review: issue #3375. CR 2026-09-25 (verified copy of
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt,
// effective September 25, 2026): 702.37c and 702.168b make a card cast with
// morph or disguise a 2/2 face-down creature spell paid for with {3} under the
// alternative-cost rules, which "enters the battlefield with the same
// characteristics the spell had"; 708.2 says those are its only
// characteristics. 707.2 says a copy acquires copiable values and casting
// choices, while "other effects ... status, counters, and stickers are not
// copied", and 707.10 that a copy of a spell isn't cast. Generator Servant
// 5ea15262-c00f-4865-bb6a-37ccb2ef62bd ("If any of that mana is spent on a
// creature spell, it gains haste until end of turn."; 2018-12-07 rulings: each
// such spell "(and resulting creature)" gains haste, including when the mana
// pays only an alternative or additional cost) and Arena of Glory
// dd148edc-9e43-41aa-bb50-f912115d3e72 (same rider; 2024-06-07 rulings: mana
// spent on any part of a creature spell's cost, including an alternative cost,
// gives the spell and resulting creature haste) supply the rider, modelled here
// directly as a ManaSpendCastCreatureSpell SpellGainsKeywords unit because no
// registered card declares it yet. Default-registry cards: Zoetic Cavern
// 37f10035-bf05-460d-9390-433caa2570f4 (Land, Morph {2}), Birchlore Rangers
// (Morph {G}), Aveline de Grandpré (Disguise {B}{G}) and Muldrotha, the
// Gravetide e4625704-1d52-44e4-804f-2f45644d76ac (cast a permanent spell of
// each permanent type from your graveyard).

func addHasteRiderUnit(g *game.Game, unit mana.Color) {
	player := g.Players[game.Player1]
	player.ManaRiders = append(player.ManaRiders, game.ManaRiderInstance{
		Unit:       mana.Unit{Color: unit},
		Controller: game.Player1,
		Rider: game.ManaSpendRider{
			Condition:          game.ManaSpendCastCreatureSpell,
			SpellGainsKeywords: []game.Keyword{game.Haste},
		},
	})
}

func assertHasteRiderTagged(t *testing.T, g *game.Game, obj *game.StackObject, want bool) {
	t.Helper()
	if got := slices.Contains(obj.GainsKeywordsUntilEndOfTurn, game.Haste); got != want {
		t.Fatalf("stack object haste rider = %v, want %v", got, want)
	}
	if want && len(g.Players[game.Player1].ManaRiders) != 0 {
		t.Fatalf("riders remaining = %d, want the spent haste unit consumed", len(g.Players[game.Player1].ManaRiders))
	}
}

func resolveFaceDownPermanent(t *testing.T, g *game.Game, engine *Engine, cardID id.ID) *game.Permanent {
	t.Helper()
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent, ok := findPermanentByCardID(g, cardID)
	if !ok || !permanent.FaceDown {
		t.Fatalf("face-down spell resolved to %+v (found %v), want a face-down permanent", permanent, ok)
	}
	return permanent
}

// A face-down spell cast from hand with haste-tagged mana spent on its {3}
// alternative cost resolves as a face-down creature with haste until cleanup.
func TestFaceDownHandCastKeepsManaSpendHasteRider(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name  string
		color mana.Color
		kind  game.FaceDownKind
	}{
		{"Zoetic Cavern", mana.C, game.FaceDownMorph},
		{"Birchlore Rangers", mana.G, game.FaceDownMorph},
		{"Aveline de Grandpré", mana.G, game.FaceDownDisguise},
	} {
		g, engine, cardID := faceDownCostGame(t, tc.name, tc.color, 3)
		addHasteRiderUnit(g, tc.color)
		obj := castFaceDownPaying(t, engine, g, cardID, tc.kind, 3)
		assertHasteRiderTagged(t, g, obj, true)
		permanent := resolveFaceDownPermanent(t, g, engine, cardID)
		if !hasKeyword(g, permanent, game.Haste) {
			t.Fatalf("%s: face-down permanent paid with haste-tagged mana lacks haste", tc.name)
		}
		if got := effectivePower(g, permanent); got != 2 {
			t.Fatalf("%s: face-down permanent power = %d, want 2", tc.name, got)
		}
		expireCleanupDurations(g)
		if hasKeyword(g, permanent, game.Haste) {
			t.Fatalf("%s: rider haste persisted past cleanup", tc.name)
		}
	}
}

// Untagged mana grants nothing: a face-down cast paid with ordinary mana
// resolves without haste.
func TestFaceDownHandCastWithUntaggedManaHasNoHaste(t *testing.T) {
	t.Parallel()
	g, engine, cardID := faceDownCostGame(t, "Zoetic Cavern", mana.C, 3)
	obj := castFaceDownPaying(t, engine, g, cardID, game.FaceDownMorph, 3)
	assertHasteRiderTagged(t, g, obj, false)
	if permanent := resolveFaceDownPermanent(t, g, engine, cardID); hasKeyword(g, permanent, game.Haste) {
		t.Fatal("face-down permanent cast with untagged mana gained haste")
	}
}

// Muldrotha's graveyard permission admits Zoetic Cavern cast face down as a
// creature spell; haste-tagged mana spent on it gives the face-down permanent
// haste, while an untagged graveyard cast does not.
func TestFaceDownGraveyardCastKeepsManaSpendHasteRider(t *testing.T) {
	for _, tagged := range []bool{true, false} {
		g, e, registry := lifetimeGame(t)
		source := addCombatPermanent(g, game.Player1, lifetimeDefinition(t, registry, muldrothaID))
		cavern := addCardToGraveyard(g, game.Player1, productionChoiceCard(t, "Zoetic Cavern"))
		g.Players[game.Player1].ManaPool.Add(mana.C, 3)
		if tagged {
			addHasteRiderUnit(g, mana.C)
		}
		cast := action.CastFaceDownFromZone(cavern, zone.Graveyard, game.FaceFront, game.FaceDownMorph).
			WithPlayPermission(muldrothaPermission(source, types.Creature))
		if !e.applyAction(g, game.Player1, cast) {
			t.Fatal("face-down Zoetic Cavern graveyard cast failed")
		}
		obj, ok := g.Stack.Peek()
		if !ok || !obj.FaceDown || obj.SourceZone != zone.Graveyard {
			t.Fatalf("stack top = %+v, want face-down spell cast from graveyard", obj)
		}
		assertHasteRiderTagged(t, g, obj, tagged)
		permanent := resolveFaceDownPermanent(t, g, e, cavern)
		if got := hasKeyword(g, permanent, game.Haste); got != tagged {
			t.Fatalf("tagged=%v: graveyard face-down permanent haste = %v", tagged, got)
		}
	}
}

// Turning the creature face up keeps the same permanent (CR 708.8), so the
// rider's until-end-of-turn haste still applies to it.
func TestFaceDownRiderHasteSurvivesTurningFaceUp(t *testing.T) {
	t.Parallel()
	g, engine, cardID := faceDownCostGame(t, "Birchlore Rangers", mana.G, 3)
	addHasteRiderUnit(g, mana.G)
	castFaceDownPaying(t, engine, g, cardID, game.FaceDownMorph, 3)
	permanent := resolveFaceDownPermanent(t, g, engine, cardID)
	g.Players[game.Player1].ManaPool.Add(mana.G, 1)
	if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
		t.Fatal("turning Birchlore Rangers face up failed")
	}
	if permanent.FaceDown || !hasKeyword(g, permanent, game.Haste) {
		t.Fatalf("face-up Birchlore Rangers FaceDown=%v haste=%v, want face up with haste", permanent.FaceDown, hasKeyword(g, permanent, game.Haste))
	}
}

// The rider recorded on the stack survives game cloning (replay/search), and
// the clone's face-down permanent gains haste independently of the original.
func TestFaceDownRiderSurvivesClone(t *testing.T) {
	t.Parallel()
	g, engine, cardID := faceDownCostGame(t, "Zoetic Cavern", mana.C, 3)
	addHasteRiderUnit(g, mana.C)
	castFaceDownPaying(t, engine, g, cardID, game.FaceDownMorph, 3)
	clone := g.Clone()
	cloneObj, ok := clone.Stack.Peek()
	if !ok || !slices.Contains(cloneObj.GainsKeywordsUntilEndOfTurn, game.Haste) {
		t.Fatalf("cloned stack top = %+v, want the haste rider", cloneObj)
	}
	if permanent := resolveFaceDownPermanent(t, clone, engine, cardID); !hasKeyword(clone, permanent, game.Haste) {
		t.Fatal("cloned face-down spell resolved without haste")
	}
	if _, ok := findPermanentByCardID(g, cardID); ok {
		t.Fatal("resolving the clone changed the original game")
	}
	if permanent := resolveFaceDownPermanent(t, g, engine, cardID); !hasKeyword(g, permanent, game.Haste) {
		t.Fatal("original face-down spell resolved without haste")
	}
}

// A copy of a haste-tagged spell is not cast and spent no mana, and the rider
// is an effect rather than a copiable value (CR 707.2, 707.10): the copy's token
// has no haste while the original still gains it. Covers a face-down original
// and an ordinary creature spell.
func TestCopyOfHasteTaggedSpellDoesNotGainHaste(t *testing.T) {
	t.Parallel()
	t.Run("face down", func(t *testing.T) {
		t.Parallel()
		g, engine, cardID := faceDownCostGame(t, "Zoetic Cavern", mana.C, 3)
		addHasteRiderUnit(g, mana.C)
		obj := castFaceDownPaying(t, engine, g, cardID, game.FaceDownMorph, 3)
		token := resolveCopyOfTaggedSpell(t, g, engine, obj)
		if hasKeyword(g, token, game.Haste) {
			t.Fatal("copy of a haste-tagged face-down spell gained haste")
		}
		if permanent := resolveFaceDownPermanent(t, g, engine, cardID); !hasKeyword(g, permanent, game.Haste) {
			t.Fatal("original face-down spell lost its haste rider")
		}
	})
	t.Run("face up", func(t *testing.T) {
		t.Parallel()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		g.Players[game.Player1].ManaPool.Add(mana.R, 1)
		addHasteRiderUnit(g, mana.R)
		goblin := creatureSpellDef("Goblin", types.Goblin)
		goblin.ManaCost = opt.Val(cost.Mana{cost.R})
		goblin.Power = opt.Val(game.PT{Value: 1})
		goblin.Toughness = opt.Val(game.PT{Value: 1})
		spellID := addCardToHand(g, game.Player1, goblin)
		g.Turn.Phase = game.PhasePrecombatMain
		g.Turn.Step = game.StepNone
		if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
			t.Fatal("cast creature with haste-tagged mana failed")
		}
		obj, _ := g.Stack.Peek()
		assertHasteRiderTagged(t, g, obj, true)
		token := resolveCopyOfTaggedSpell(t, g, engine, obj)
		if hasKeyword(g, token, game.Haste) {
			t.Fatal("copy of a haste-tagged creature spell gained haste")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		permanent, ok := findPermanentByCardID(g, spellID)
		if !ok || !hasKeyword(g, permanent, game.Haste) {
			t.Fatal("original creature spell lost its haste rider")
		}
	})
}

// resolveCopyOfTaggedSpell copies obj with an ordinary copy effect, checks the
// copy carries no rider, and resolves the copy to its token permanent.
func resolveCopyOfTaggedSpell(t *testing.T, g *game.Game, engine *Engine, obj *game.StackObject) *game.Permanent {
	t.Helper()
	addEffectSpellToStack(g, game.Player1, game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(obj.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	copyObj, ok := g.Stack.Peek()
	if !ok || !copyObj.Copy || copyObj.ID == obj.ID {
		t.Fatalf("stack top = %+v, want the spell copy", copyObj)
	}
	if len(copyObj.GainsKeywordsUntilEndOfTurn) != 0 {
		t.Fatalf("spell copy carries rider keywords %v", copyObj.GainsKeywordsUntilEndOfTurn)
	}
	before := len(g.Battlefield)
	engine.resolveTopOfStack(g, &TurnLog{})
	if len(g.Battlefield) != before+1 {
		t.Fatalf("battlefield size = %d, want the copy's token added to %d", len(g.Battlefield), before)
	}
	token := g.Battlefield[len(g.Battlefield)-1]
	if !permanentHasType(g, token, types.Creature) {
		t.Fatalf("copy token %+v is not a creature", token)
	}
	return token
}
