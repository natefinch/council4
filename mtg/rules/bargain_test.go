package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// bargainDrawSpell is a {0} instant with Bargain whose only effect ("draw a
// card") is gated on the "if this spell was bargained" condition. The zero mana
// cost isolates the Bargain additional cost from mana so the tests need no lands,
// and the gated draw makes the bargained cast branch observable at resolution.
func bargainDrawSpell(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		ManaCost: opt.Val(cost.Mana{cost.O(0)}),
		Types:    []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{
			game.BargainStaticBody,
		},
		SpellAbility: opt.Val(game.Mode{
			Sequence: []game.Instruction{
				{
					Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
					Condition: opt.Val(game.EffectCondition{
						Condition: opt.Val(game.Condition{SpellWasBargained: true}),
					}),
				},
			},
		}.Ability()),
	}}
}

// artifactCreaturePermanent, enchantmentPermanent, plainCreaturePermanent, and
// landPermanent are the sacrifice-union probes: the first three each match the
// Bargain cost (artifact, enchantment, or token), the last two do not.
func artifactCreaturePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Ornithopter",
		Types:     []types.Card{types.Artifact, types.Creature},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

func enchantmentPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Wild Growth",
		Types: []types.Card{types.Enchantment},
	}})
}

func plainCreaturePermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      "Grizzly Bears",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

func landPermanent(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})
}

func bargainedCastActionForCard(actions []action.Action, cardID id.ID) bool {
	for _, a := range actions {
		if a.Kind != action.ActionCastSpell {
			continue
		}
		if cast, ok := a.CastSpellPayload(); ok && cast.CardID == cardID && cast.Bargained {
			return true
		}
	}
	return false
}

func normalCastActionForCard(actions []action.Action, cardID id.ID) bool {
	for _, a := range actions {
		if a.Kind != action.ActionCastSpell {
			continue
		}
		if cast, ok := a.CastSpellPayload(); ok && cast.CardID == cardID && !cast.Bargained {
			return true
		}
	}
	return false
}

// TestBargainOffersBothBranchesWithEligibleSacrifice verifies that a Bargain
// spell whose controller has an eligible object to sacrifice is offered both the
// normal (unbargained) and bargained cast actions: Bargain is optional, so the
// player may choose either branch.
func TestBargainOffersBothBranchesWithEligibleSacrifice(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, bargainDrawSpell("Bargain Draw"))
	enchantmentPermanent(g, game.Player1)
	setMainPhasePriority(g, game.Player1)

	legal := e.Simulator().LegalActions(g, game.Player1)
	if !normalCastActionForCard(legal, spellID) {
		t.Fatal("normal (unbargained) cast not offered for a Bargain spell")
	}
	if !bargainedCastActionForCard(legal, spellID) {
		t.Fatal("bargained cast not offered although an eligible object is available to sacrifice")
	}
}

// TestBargainNotOfferedWithoutEligibleObject verifies that when the controller
// has no artifact, enchantment, or token to sacrifice, only the normal cast is
// offered: the bargained branch requires the additional cost to be payable.
func TestBargainNotOfferedWithoutEligibleObject(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, bargainDrawSpell("Bargain Draw"))
	plainCreaturePermanent(g, game.Player1)
	landPermanent(g, game.Player1)
	setMainPhasePriority(g, game.Player1)

	legal := e.Simulator().LegalActions(g, game.Player1)
	if !normalCastActionForCard(legal, spellID) {
		t.Fatal("normal cast not offered for a Bargain spell with no eligible sacrifice")
	}
	if bargainedCastActionForCard(legal, spellID) {
		t.Fatal("bargained cast offered although no artifact, enchantment, or token is available")
	}
}

// TestBargainSacrificeUnion verifies the exact "artifact, enchantment, or token"
// union: an artifact creature, a nontoken enchantment, and a token each enable
// the bargained branch, while a plain nontoken creature and a land do not.
func TestBargainSacrificeUnion(t *testing.T) {
	cases := []struct {
		name    string
		add     func(*game.Game, game.PlayerID)
		bargain bool
	}{
		{"artifact creature", func(g *game.Game, p game.PlayerID) { artifactCreaturePermanent(g, p) }, true},
		{"nontoken enchantment", func(g *game.Game, p game.PlayerID) { enchantmentPermanent(g, p) }, true},
		{"token creature", func(g *game.Game, p game.PlayerID) { addTokenCreaturePermanent(g, p, "Servo") }, true},
		{"plain creature", func(g *game.Game, p game.PlayerID) { plainCreaturePermanent(g, p) }, false},
		{"land", func(g *game.Game, p game.PlayerID) { landPermanent(g, p) }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := newSimEngine()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			spellID := addCardToHand(g, game.Player1, bargainDrawSpell("Bargain Draw"))
			tc.add(g, game.Player1)
			setMainPhasePriority(g, game.Player1)

			legal := e.Simulator().LegalActions(g, game.Player1)
			if got := bargainedCastActionForCard(legal, spellID); got != tc.bargain {
				t.Fatalf("bargained offered = %v, want %v for %s", got, tc.bargain, tc.name)
			}
		})
	}
}

// TestBargainedCastSacrificesExactlyOneObject verifies the Bargain additional
// cost sacrifices exactly one object: with two eligible permanents on the
// battlefield, a bargained cast removes exactly one and records a single
// sacrificed-as-cost id (no duplicate sacrifices).
func TestBargainedCastSacrificesExactlyOneObject(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, bargainDrawSpell("Bargain Draw"))
	enchantmentPermanent(g, game.Player1)
	enchantmentPermanent(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	setMainPhasePriority(g, game.Player1)

	before := len(g.Battlefield)
	if !e.applyAction(g, game.Player1, action.CastBargainedSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast bargained) = false, want true")
	}
	if got := before - len(g.Battlefield); got != 1 {
		t.Fatalf("permanents sacrificed = %d, want exactly 1", got)
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("no spell on the stack after casting bargained")
	}
	if !obj.Bargained {
		t.Fatal("stack object does not record Bargained = true")
	}
	if got := len(obj.SacrificedAsCostIDs); got != 1 {
		t.Fatalf("SacrificedAsCostIDs = %d, want exactly 1", got)
	}
}

// TestBargainedBranchPersistsAfterSacrificeGone verifies that the bargained
// state lives on the stack object, not on the sacrificed object: once the
// Bargain cost is paid the object is gone from the battlefield, yet the
// resolving spell still counts as bargained and its gated draw resolves.
func TestBargainedBranchPersistsAfterSacrificeGone(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, bargainDrawSpell("Bargain Draw"))
	sacrificed := enchantmentPermanent(g, game.Player1)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	setMainPhasePriority(g, game.Player1)

	if !e.applyAction(g, game.Player1, action.CastBargainedSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(cast bargained) = false, want true")
	}
	if _, ok := g.PermanentByID(sacrificed.ObjectID); ok {
		t.Fatal("sacrificed object is still on the battlefield after paying the Bargain cost")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want 1 (bargained gated draw must resolve although the sacrificed object is gone)", got)
	}
}

// TestUnbargainedCastSkipsGatedEffect is the negative control: casting the same
// spell without bargaining leaves the "if this spell was bargained" draw
// unresolved, so the base spell does nothing.
func TestUnbargainedCastSkipsGatedEffect(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, bargainDrawSpell("Bargain Draw"))
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	setMainPhasePriority(g, game.Player1)

	if !e.applyAction(g, game.Player1, action.CastSpellFromZone(spellID, zone.Hand, nil, 0, nil)) {
		t.Fatal("applyAction(cast unbargained) = false, want true")
	}
	e.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want 0 (unbargained spell must not draw)", got)
	}
}

// TestBargainedEffectGatingByStackState exercises the "if this spell was
// bargained" resolution condition directly: the gated draw runs only when the
// resolving stack object recorded a bargained cast, and never for a copy (copies
// are never bargained, CR 707.10c / 702.166).
func TestBargainedEffectGatingByStackState(t *testing.T) {
	gatedDraw := func() *game.Instruction {
		return &game.Instruction{
			Primitive: game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
			Condition: opt.Val(game.EffectCondition{
				Condition: opt.Val(game.Condition{SpellWasBargained: true}),
			}),
		}
	}

	resolve := func(bargained, isCopy bool) int {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		obj := &game.StackObject{
			Kind:       game.StackSpell,
			Controller: game.Player1,
			Bargained:  bargained,
			Copy:       isCopy,
		}
		engine.resolveInstructionWithChoices(g, obj, gatedDraw(), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
		return g.Players[game.Player1].Hand.Size()
	}

	if got := resolve(true, false); got != 1 {
		t.Fatalf("bargained draw: hand size = %d, want 1", got)
	}
	if got := resolve(false, false); got != 0 {
		t.Fatalf("unbargained draw: hand size = %d, want 0", got)
	}
	if got := resolve(true, true); got != 0 {
		t.Fatalf("copied bargained spell: hand size = %d, want 0 (copies are never bargained)", got)
	}
}
