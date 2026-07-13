package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/c"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// coruscationMagePermanents returns every battlefield permanent named
// "Coruscation Mage" (the cast creature plus any Offspring token copies).
func coruscationMagePermanents(g *game.Game) []*game.Permanent {
	var out []*game.Permanent
	for _, p := range g.Battlefield {
		if def, ok := permanentCardDef(g, p); ok && def.Name == "Coruscation Mage" {
			out = append(out, p)
		}
	}
	return out
}

func offspringCastActionForCard(actions []action.Action, cardID id.ID) bool {
	for _, a := range actions {
		if a.Kind != action.ActionCastSpell {
			continue
		}
		if cast, ok := a.CastSpellPayload(); ok && cast.CardID == cardID && cast.Offspring {
			return true
		}
	}
	return false
}

// TestOffspringNormalCastCreatesNoToken verifies the baseline: casting the real
// Coruscation Mage without paying its Offspring cost resolves a single creature
// and creates no token copy.
func TestOffspringNormalCastCreatesNoToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, c.CoruscationMage())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(normal cast) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || obj.OffspringPaid {
		t.Fatalf("stack object = %#v, want a non-offspring spell", obj)
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})

	if got := len(coruscationMagePermanents(g)); got != 1 {
		t.Fatalf("Coruscation Mage permanents = %d, want 1 (no token on a normal cast)", got)
	}
}

// TestOffspringPaidCastCreatesOneTokenCopy verifies the payoff: paying the
// Offspring cost resolves the creature and its linked ETB trigger creates exactly
// one token copy under the caster's control.
func TestOffspringPaidCastCreatesOneTokenCopy(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, c.CoruscationMage())
	// {1}{R} base + Offspring {2} = four mana.
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastOffspringSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(offspring cast) = false, want true")
	}
	obj, ok := g.Stack.Peek()
	if !ok || !obj.OffspringPaid {
		t.Fatalf("stack object = %#v, want an offspring-paid spell", obj)
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})

	permanents := coruscationMagePermanents(g)
	if got := len(permanents); got != 2 {
		t.Fatalf("Coruscation Mage permanents = %d, want 2 (creature + one token copy)", got)
	}
	var tokens int
	for _, p := range permanents {
		if p.Token {
			tokens++
			if p.Controller != game.Player1 {
				t.Fatalf("token controller = %v, want the caster %v", p.Controller, game.Player1)
			}
		}
	}
	if tokens != 1 {
		t.Fatalf("token copies = %d, want exactly 1", tokens)
	}
}

// TestOffspringTokenIsOneOneCopyWithAbilities verifies the token copy has the
// entering creature's copiable characteristics (types, subtypes, and the noncreature
// -spell-cast trigger and Offspring keyword) but base power and toughness each 1.
func TestOffspringTokenIsOneOneCopyWithAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, c.CoruscationMage())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastOffspringSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(offspring cast) = false, want true")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})

	var token *game.Permanent
	for _, p := range coruscationMagePermanents(g) {
		if p.Token {
			token = p
			break
		}
	}
	if token == nil {
		t.Fatal("no token copy created")
	}
	def, ok := permanentCardDef(g, token)
	if !ok {
		t.Fatal("token copy has no resolvable card definition")
	}
	if got := def.Power.Val.Value; got != 1 {
		t.Fatalf("token base power = %d, want 1", got)
	}
	if got := def.Toughness.Val.Value; got != 1 {
		t.Fatalf("token base toughness = %d, want 1", got)
	}
	if def.DynamicPower.Exists || def.DynamicToughness.Exists {
		t.Fatal("token retained a characteristic-defining P/T; want it cleared to the fixed 1/1")
	}
	if !permanentHasType(g, token, types.Creature) {
		t.Fatal("token copy is not a creature")
	}
	hasOtter := false
	for _, sub := range def.Subtypes {
		if sub == types.Otter {
			hasOtter = true
		}
	}
	if !hasOtter {
		t.Fatalf("token subtypes = %v, want to include Otter (copiable)", def.Subtypes)
	}
	if _, ok := def.OffspringKeyword(); !ok {
		t.Fatal("token copy did not inherit the Offspring keyword static ability")
	}
	if len(def.TriggeredAbilities) == 0 {
		t.Fatal("token copy did not inherit any triggered abilities")
	}
}

// TestOffspringTokenDoesNotRecurse verifies no infinite recursion: the token copy
// carries its own copy of the Offspring ETB trigger, but because the token itself
// was not cast with the offspring cost paid, its intervening-if fails and it makes
// no further token. Exactly one token is produced.
func TestOffspringTokenDoesNotRecurse(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	spellID := addCardToHand(g, game.Player1, c.CoruscationMage())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)

	if !engine.applyAction(g, game.Player1, action.CastOffspringSpell(spellID, nil, 0, nil)) {
		t.Fatal("applyAction(offspring cast) = false, want true")
	}
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})

	if got := len(coruscationMagePermanents(g)); got != 2 {
		t.Fatalf("Coruscation Mage permanents = %d, want 2 (no recursive token from the token copy)", got)
	}
}

// TestOffspringNotOfferedWithoutManaForCost verifies the offspring branch is only
// offered when the additional mana cost is affordable: with exactly the base
// {1}{R} available, the normal cast is offered but the offspring cast is not.
func TestOffspringNotOfferedWithoutManaForCost(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, c.CoruscationMage())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	setMainPhasePriority(g, game.Player1)

	legal := e.Simulator().LegalActions(g, game.Player1)
	if !normalCastActionForCard(legal, spellID) {
		t.Fatal("normal cast not offered for an affordable Offspring creature")
	}
	if offspringCastActionForCard(legal, spellID) {
		t.Fatal("offspring cast offered although the extra {2} cannot be paid")
	}
}

// TestOffspringOffersBothBranchesWhenAffordable verifies that when both the base
// and additional costs are affordable, the player may choose either the normal or
// the offspring cast: Offspring is optional.
func TestOffspringOffersBothBranchesWhenAffordable(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, c.CoruscationMage())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)

	legal := e.Simulator().LegalActions(g, game.Player1)
	if !normalCastActionForCard(legal, spellID) {
		t.Fatal("normal cast not offered for an affordable Offspring creature")
	}
	if !offspringCastActionForCard(legal, spellID) {
		t.Fatal("offspring cast not offered although the extra {2} can be paid")
	}
}

// TestOffspringLegalActionsKeepBranchesDistinct verifies that when both branches
// are affordable the legal action list contains exactly one normal and exactly
// one offspring cast for the card, and that the two are distinct actions: the
// offspring cast must not be satisfiable by matching the plain cast, and vice
// versa. This guards actionsEqual/containsAction against conflating the branches.
func TestOffspringLegalActionsKeepBranchesDistinct(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	spellID := addCardToHand(g, game.Player1, c.CoruscationMage())
	g.Players[game.Player1].ManaPool.Add(mana.R, 1)
	g.Players[game.Player1].ManaPool.Add(mana.C, 3)
	setMainPhasePriority(g, game.Player1)

	legal := e.Simulator().LegalActions(g, game.Player1)

	var normal, offspring int
	for _, a := range legal {
		if a.Kind != action.ActionCastSpell {
			continue
		}
		cast, ok := a.CastSpellPayload()
		if !ok || cast.CardID != spellID {
			continue
		}
		if cast.Offspring {
			offspring++
		} else {
			normal++
		}
	}
	if normal != 1 {
		t.Fatalf("normal Coruscation Mage cast actions = %d, want exactly 1", normal)
	}
	if offspring != 1 {
		t.Fatalf("offspring Coruscation Mage cast actions = %d, want exactly 1", offspring)
	}

	// The synthetic offspring cast must be found in the legal list on its own
	// merits, and the plain cast must not accidentally satisfy it.
	plain := action.CastSpell(spellID, nil, 0, nil)
	paid := action.CastOffspringSpell(spellID, nil, 0, nil)
	if actionsEqual(plain, paid) {
		t.Fatal("actionsEqual() = true for the normal and offspring casts of the same card")
	}
	if !containsAction(legal, plain) {
		t.Fatal("legal action list does not contain the normal cast")
	}
	if !containsAction(legal, paid) {
		t.Fatal("legal action list does not contain the offspring cast")
	}
}

// TestOffspringEnterWithoutCastCreatesNoToken verifies that a Coruscation Mage
// that enters the battlefield without being cast (e.g. reanimated, blinked, or put
// onto the battlefield by an effect) is not offspring-paid and so creates no token
// copy: the payoff is gated on the cast-time additional cost, not merely entering.
func TestOffspringEnterWithoutCastCreatesNoToken(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, c.CoruscationMage())
	// Flush any enter triggers that entering-without-cast might queue.
	engine.putTriggeredAbilitiesOnStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	resolveStackWithTriggers(engine, g, [game.NumPlayers]PlayerAgent{})

	if got := len(coruscationMagePermanents(g)); got != 1 {
		t.Fatalf("Coruscation Mage permanents = %d, want 1 (no token when it enters without an offspring cast)", got)
	}
}

// TestOffspringKeywordAccessorReadsCost verifies the generated Coruscation Mage
// exposes its Offspring keyword with the printed {2} additional cost, the datum
// the payment planner adds to the spell's total when the offspring branch is cast.
func TestOffspringKeywordAccessorReadsCost(t *testing.T) {
	def := c.CoruscationMage()
	offspring, ok := def.OffspringKeyword()
	if !ok {
		t.Fatal("Coruscation Mage does not expose an Offspring keyword")
	}
	if got := len(offspring.Cost); got != 1 {
		t.Fatalf("offspring cost symbols = %d, want 1 (a single generic {2})", got)
	}
	if got := offspring.Cost.ManaValue(); got != 2 {
		t.Fatalf("offspring cost mana value = %d, want 2", got)
	}
}
