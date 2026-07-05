package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
)

// TestCloakResolvesTopLibraryCardFaceDownAsCloak verifies that resolving a
// Manifest{Cloak: true} effect puts the top card of the controller's library
// onto the battlefield face down as a face-down cloak permanent that is a 2/2
// creature, exactly like manifest but under the cloak face-down kind.
func TestCloakResolvesTopLibraryCardFaceDownAsCloak(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToLibrary(g, game.Player1, manifestNoncreature(cost.Mana{cost.G}))
	addEffectSpellToStack(g, game.Player1, game.Manifest{Cloak: true}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player1].Library.Contains(cardID) {
		t.Fatal("cloaked card remained in library")
	}
	var cloaked *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			cloaked = permanent
			break
		}
	}
	if cloaked == nil {
		t.Fatal("cloaked permanent not found on battlefield")
	}
	if !cloaked.FaceDown || cloaked.FaceDownKind != game.FaceDownCloak {
		t.Fatalf("cloak face-down state = %+v, want face-down cloak", cloaked)
	}
	if effectivePower(g, cloaked) != 2 {
		t.Fatalf("cloak effective power = %d, want 2", effectivePower(g, cloaked))
	}
	if toughness, ok := effectiveToughness(g, cloaked); !ok || toughness != 2 {
		t.Fatalf("cloak effective toughness = %d (ok=%t), want 2", toughness, ok)
	}
}

// TestCloakFaceDownHasWardTwoAndNoShield verifies a face-down cloak permanent
// carries the ward {2} static ability granted to cloaked and disguised
// permanents, and that being cloaked adds no shield counter (unlike turning a
// disguised permanent face up).
func TestCloakFaceDownHasWardTwoAndNoShield(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addFaceDownPermanent(g, game.Player1, manifestCreature(cost.Mana{cost.G}), game.FaceDownCloak)

	abilities := permanentEffectiveAbilities(g, permanent)
	if len(abilities) != 1 {
		t.Fatalf("face-down cloak abilities = %+v, want ward ability", abilities)
	}
	staticBody, ok := abilities[0].(*game.StaticAbility)
	if !ok || !game.BodyHasKeyword(staticBody, game.Ward) {
		t.Fatalf("face-down cloak abilities = %+v, want ward ability", abilities)
	}
	if staticBody.Text != "Ward {2}" {
		t.Fatalf("face-down cloak ward text = %q, want \"Ward {2}\"", staticBody.Text)
	}
	if got := permanent.Counters.Get(counter.Shield); got != 0 {
		t.Fatalf("cloak shield counters = %d, want 0", got)
	}
}

// TestCloakCreatureTurnsFaceUpForManaCostWithoutShield verifies a cloaked
// creature card turns face up for its mana cost — like manifest, not disguise —
// and gains no shield counter on turning up, while a cloaked noncreature card
// cannot be turned face up.
func TestCloakCreatureTurnsFaceUpForManaCostWithoutShield(t *testing.T) {
	t.Run("creature turns up for mana cost without shield", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addFaceDownPermanent(g, game.Player1, manifestCreature(cost.Mana{cost.G}), game.FaceDownCloak)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1

		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("cloak creature turn face-up action failed")
		}
		if permanent.FaceDown {
			t.Fatal("cloak creature remained face-down")
		}
		if permanentEffectiveName(g, permanent) != "Manifest Bear" || effectivePower(g, permanent) != 3 {
			t.Fatalf("face-up cloak characteristics name=%q power=%d, want Manifest Bear/3", permanentEffectiveName(g, permanent), effectivePower(g, permanent))
		}
		if got := permanent.Counters.Get(counter.Shield); got != 0 {
			t.Fatalf("cloak shield counters after turn up = %d, want 0", got)
		}
	})
	t.Run("noncreature cannot turn face up", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addFaceDownPermanent(g, game.Player1, manifestNoncreature(cost.Mana{cost.G}), game.FaceDownCloak)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1

		if engine.canTurnFaceUp(g, game.Player1, permanent.ObjectID) {
			t.Fatal("cloak noncreature was allowed to turn face up")
		}
		if engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("cloak noncreature turn face-up action succeeded")
		}
		if !permanent.FaceDown {
			t.Fatal("cloak noncreature stopped being face-down")
		}
	})
}

// TestCloakedMorphDisguiseCardTurnsUpViaItsOwnCost verifies that a cloaked
// card which also has morph or disguise may be turned face up using that card's
// morph or disguise cost, not just its mana cost (CR 701.56c/d), mirroring the
// manifest rules — and that turning a cloaked permanent face up grants no shield
// counter regardless of the cost paid (unlike a disguised permanent).
func TestCloakedMorphDisguiseCardTurnsUpViaItsOwnCost(t *testing.T) {
	t.Run("cloaked morph creature turns up for morph cost without shield", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addFaceDownPermanent(g, game.Player1, manifestMorphCreature(cost.Mana{cost.W}, cost.Mana{cost.G}), game.FaceDownCloak)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1

		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("cloaked morph creature did not turn face up for its morph cost")
		}
		if permanent.FaceDown || permanentEffectiveName(g, permanent) != "Manifest Morph Bear" {
			t.Fatalf("cloaked morph creature state name=%q faceDown=%t", permanentEffectiveName(g, permanent), permanent.FaceDown)
		}
		if got := permanent.Counters.Get(counter.Shield); got != 0 {
			t.Fatalf("cloak shield counters after morph turn up = %d, want 0", got)
		}
	})
	t.Run("cloaked disguise creature turns up for disguise cost without shield", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		permanent := addFaceDownPermanent(g, game.Player1, disguiseCreature(cost.Mana{cost.G}), game.FaceDownCloak)
		g.Players[game.Player1].ManaPool.Add(mana.G, 1)
		g.Turn.PriorityPlayer = game.Player1

		if !engine.applyAction(g, game.Player1, actionBuild.turnFaceUp(permanent.ObjectID)) {
			t.Fatal("cloaked disguise creature did not turn face up for its disguise cost")
		}
		if permanent.FaceDown || permanentEffectiveName(g, permanent) != "Veiled Guard" {
			t.Fatalf("cloaked disguise creature state name=%q faceDown=%t", permanentEffectiveName(g, permanent), permanent.FaceDown)
		}
		if got := permanent.Counters.Get(counter.Shield); got != 0 {
			t.Fatalf("cloak shield counters after disguise turn up = %d, want 0", got)
		}
	})
}

// TestFaceDownCloakWardCountersSpellWhenCostIsNotPaid verifies the ward {2}
// granted to a face-down cloak permanent counters a spell that targets it when
// the ward cost is not paid, matching the disguise face-down ward behavior.
func TestFaceDownCloakWardCountersSpellWhenCostIsNotPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	warded := addFaceDownPermanent(g, game.Player2, manifestCreature(cost.Mana{cost.W}), game.FaceDownCloak)
	spellID := addCardToHand(g, game.Player1, targetCreatureInstant())
	g.Turn.PriorityPlayer = game.Player1

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, []game.Target{game.PermanentTarget(warded.ObjectID)}, 0, nil)) {
		t.Fatal("targeting spell cast failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("face-down cloak ward trigger was not put on the stack")
	}
	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want cloak ward to counter targeting spell", g.Stack.Size())
	}
}
