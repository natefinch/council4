package rules

import (
	"testing"

	cards "github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

// These tests exercise the real, compiler-generated Six end to end through the
// engine (not a hand-crafted mirror), proving its attack trigger — "Whenever Six
// attacks, mill three cards. You may put a land card from among them into your
// hand." — wires from combat detection through the mill and the linked,
// milled-only, at-most-one, optional land return. The land pool is restricted to
// exactly the cards this resolution milled by object identity, so a same-named
// graveyard land or another Six's milled card is never offered.

func sixAttackContent() game.AbilityContent {
	return cards.Six().TriggeredAbilities[0].Content
}

// sixLandChoicePrimitive extracts Six's real "put a land card from among them
// into your hand" ChooseFromZone so a test can resolve just that step against a
// hand-published linked set (for interposing an intervening zone change).
func sixLandChoicePrimitive(t *testing.T) game.ChooseFromZone {
	t.Helper()
	seq := sixAttackContent().Modes[0].Sequence
	if len(seq) != 2 {
		t.Fatalf("Six attack sequence length = %d, want 2 (mill then choose)", len(seq))
	}
	choose, ok := seq[1].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("Six attack sequence[1] primitive = %T, want game.ChooseFromZone", seq[1].Primitive)
	}
	return choose
}

func sixLandCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Land}}}
}

func sixNonlandCard(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Creature}}}
}

// sixLandAgent drives Six's "you may put a land card from among them into your
// hand" trigger: it accepts or declines the optional gate, and when a card must
// be chosen it selects the option whose card is wantCardID, falling back to the
// offered default when that card is not among the candidates.
type sixLandAgent struct {
	acceptMay  bool
	wantCardID id.ID
}

func (sixLandAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a sixLandAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		if a.acceptMay {
			return []int{1}
		}
		return []int{0}
	}
	if a.wantCardID != 0 {
		for _, option := range request.Options {
			if option.Card.Exists && option.Card.Val.CardID == a.wantCardID {
				return []int{option.Index}
			}
		}
	}
	return request.DefaultSelection
}

// sixAttacks declares the real Six as an attacker and runs the point at which a
// player would next receive priority, returning whether a triggered ability was
// placed on the stack. It mirrors the engine's own combat detection path used by
// the attacking-trigger tests.
func sixAttacks(g *game.Game, engine *Engine, six *game.Permanent) bool {
	batchID := g.IDGen.Next()
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
		{Attacker: six.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
	}}
	emitEvent(g, game.Event{
		Kind:           game.EventAttackerDeclared,
		Controller:     six.Controller,
		PermanentID:    six.ObjectID,
		Player:         game.Player2,
		SimultaneousID: batchID,
	})
	return engine.putTriggeredAbilitiesOnStack(g)
}

// resolveSixAttackContent resolves Six's real attack ability bound to the given
// Six permanent as its source, so the Mill publishes and the ChooseFromZone
// consumes the same card-scoped linked set the runtime would build in combat.
func resolveSixAttackContent(g *game.Game, six *game.Permanent, agent PlayerAgent) {
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent, game.Player2: agent}
	engine.resolveAbilityContentWithChoices(g, triggeredObjFor(six), sixAttackContent(), agents, &TurnLog{})
}

// TestSixRealCardHasReach proves the registered card carries Reach as an active
// keyword on the battlefield, not merely in its printed text.
func TestSixRealCardHasReach(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	six := addCombatPermanent(g, game.Player1, cards.Six())

	if !hasKeyword(g, six, game.Reach) {
		t.Fatal("real Six does not have Reach active on the battlefield")
	}
}

// TestSixRealCardAttackTriggerFires proves Six's self-source attack trigger is
// detected and placed on the stack when Six is declared as an attacker, with the
// trigger's source and controller pointing back at Six.
func TestSixRealCardAttackTriggerFires(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	six := addCombatPermanent(g, game.Player1, cards.Six())

	if !sixAttacks(g, engine, six) {
		t.Fatal("Six's attack trigger was not placed on the stack when Six attacked")
	}
	if got := g.Stack.Size(); got != 1 {
		t.Fatalf("stack size = %d, want 1 (Six's attack trigger)", got)
	}
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("attack trigger missing from the stack")
	}
	if obj.Kind != game.StackTriggeredAbility {
		t.Fatalf("stack object kind = %v, want a triggered ability", obj.Kind)
	}
	if obj.SourceID != six.ObjectID || obj.SourceCardID != six.CardInstanceID {
		t.Fatalf("trigger source = (obj %v, card %v), want Six (obj %v, card %v)", obj.SourceID, obj.SourceCardID, six.ObjectID, six.CardInstanceID)
	}
	if obj.Controller != game.Player1 {
		t.Fatalf("trigger controller = %v, want Six's controller Player1", obj.Controller)
	}
}

// TestSixRealCardAttackTriggerDoesNotFireForOtherAttacker proves the trigger is
// self-scoped: a different creature attacking while Six is on the battlefield
// does not fire Six's ability.
func TestSixRealCardAttackTriggerDoesNotFireForOtherAttacker(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, cards.Six())
	other := addCombatCreaturePermanent(g, game.Player1)

	if sixAttacks(g, engine, other) {
		t.Fatal("Six's attack trigger fired for a different attacker")
	}
	if got := g.Stack.Size(); got != 0 {
		t.Fatalf("stack size = %d, want 0 (no trigger for a non-Six attacker)", got)
	}
}

// TestSixRealCardAttackMillsThreeAndReturnsChosenLand is the flagship end-to-end
// path: the real Six attacks, its trigger resolves, exactly three cards are
// milled from its controller's library, and the chosen milled land is put into
// hand while the milled nonland cards stay in the graveyard and the rest of the
// library and the opponent's library are untouched.
func TestSixRealCardAttackMillsThreeAndReturnsChosenLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	six := addCombatPermanent(g, game.Player1, cards.Six())

	oppCard := addCardToLibrary(g, game.Player2, sixNonlandCard("Opponent Card"))
	// Library additions prepend, so the last three added are milled first; the
	// kept card, added first, sits below them and must survive the mill.
	keptID := addCardToLibrary(g, game.Player1, sixNonlandCard("Kept On Bottom"))
	landID := addCardToLibrary(g, game.Player1, sixLandCard("Milled Forest"))
	nonlandA := addCardToLibrary(g, game.Player1, sixNonlandCard("Milled Bear"))
	nonlandB := addCardToLibrary(g, game.Player1, sixNonlandCard("Milled Ox"))

	if !sixAttacks(g, engine, six) {
		t.Fatal("Six's attack trigger was not placed on the stack")
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: sixLandAgent{acceptMay: true, wantCardID: landID},
	}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p1 := g.Players[game.Player1]
	if !p1.Hand.Contains(landID) {
		t.Fatal("chosen milled land was not put into hand")
	}
	if p1.Graveyard.Contains(landID) {
		t.Fatal("chosen milled land remained in the graveyard")
	}
	if !p1.Graveyard.Contains(nonlandA) || !p1.Graveyard.Contains(nonlandB) {
		t.Fatal("milled nonland cards are not in the graveyard")
	}
	if !p1.Library.Contains(keptID) || p1.Library.Size() != 1 {
		t.Fatalf("controller library = %d cards, want only the un-milled kept card", p1.Library.Size())
	}
	if got := g.Players[game.Player2].Library.Size(); got != 1 || !g.Players[game.Player2].Library.Contains(oppCard) {
		t.Fatalf("opponent library size = %d, want the opponent's own card untouched", got)
	}
	if got := g.Players[game.Player2].Graveyard.Size(); got != 0 {
		t.Fatalf("opponent graveyard size = %d, want 0 (only the controller mills)", got)
	}
}

// TestSixRealCardMillStopsWhenLibraryHasFewerThanThree proves the mandatory mill
// mills the whole library and does not error when fewer than three cards remain,
// still offering an eligible milled land.
func TestSixRealCardMillStopsWhenLibraryHasFewerThanThree(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	six := addCombatPermanent(g, game.Player1, cards.Six())
	landID := addCardToLibrary(g, game.Player1, sixLandCard("Only Forest"))
	nonland := addCardToLibrary(g, game.Player1, sixNonlandCard("Only Bear"))

	resolveSixAttackContent(g, six, sixLandAgent{acceptMay: true, wantCardID: landID})

	p1 := g.Players[game.Player1]
	if got := p1.Library.Size(); got != 0 {
		t.Fatalf("library size = %d, want 0 (milled the whole two-card library)", got)
	}
	if !p1.Hand.Contains(landID) {
		t.Fatal("the milled land from a short library was not put into hand")
	}
	if !p1.Graveyard.Contains(nonland) {
		t.Fatal("the milled nonland from a short library is not in the graveyard")
	}
}

// TestSixRealCardLandChoiceRestrictedToMilledCards proves the land return is
// limited to the cards milled by this resolution: a same-named land already in
// the graveyard is not offered, so declining-by-identity leaves it untouched
// while the milled land is the only card that can be returned.
func TestSixRealCardLandChoiceRestrictedToMilledCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	six := addCombatPermanent(g, game.Player1, cards.Six())
	preexisting := addCardToGraveyard(g, game.Player1, sixLandCard("Forest"))
	milledLand := addCardToLibrary(g, game.Player1, sixLandCard("Forest"))
	addCardToLibrary(g, game.Player1, sixNonlandCard("Bear"))
	addCardToLibrary(g, game.Player1, sixNonlandCard("Ox"))

	resolveSixAttackContent(g, six, sixLandAgent{acceptMay: true, wantCardID: preexisting})

	p1 := g.Players[game.Player1]
	if p1.Hand.Contains(preexisting) {
		t.Fatal("a same-named land already in the graveyard was offered and returned")
	}
	if !p1.Graveyard.Contains(preexisting) {
		t.Fatal("the pre-existing graveyard land left the graveyard")
	}
	if !p1.Hand.Contains(milledLand) {
		t.Fatal("the milled land (the only eligible card) was not returned to hand")
	}
}

// TestSixRealCardDeclineKeepsMilledLandInGraveyard proves the return is optional:
// declining the "you may" gate mills three cards but returns none, leaving the
// milled land in the graveyard.
func TestSixRealCardDeclineKeepsMilledLandInGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	six := addCombatPermanent(g, game.Player1, cards.Six())
	landID := addCardToLibrary(g, game.Player1, sixLandCard("Milled Forest"))
	addCardToLibrary(g, game.Player1, sixNonlandCard("Milled Bear"))
	addCardToLibrary(g, game.Player1, sixNonlandCard("Milled Ox"))

	resolveSixAttackContent(g, six, sixLandAgent{acceptMay: false})

	p1 := g.Players[game.Player1]
	if p1.Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0 (declined the optional return)", p1.Hand.Size())
	}
	if !p1.Graveyard.Contains(landID) {
		t.Fatal("declined milled land is not in the graveyard")
	}
}

// TestSixRealCardReturnsAtMostOneLand proves the quantity bound: with two lands
// milled the controller may return only one, leaving the other in the graveyard.
func TestSixRealCardReturnsAtMostOneLand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	six := addCombatPermanent(g, game.Player1, cards.Six())
	firstLand := addCardToLibrary(g, game.Player1, sixLandCard("Forest"))
	secondLand := addCardToLibrary(g, game.Player1, sixLandCard("Island"))
	addCardToLibrary(g, game.Player1, sixNonlandCard("Bear"))

	resolveSixAttackContent(g, six, sixLandAgent{acceptMay: true, wantCardID: firstLand})

	p1 := g.Players[game.Player1]
	if !p1.Hand.Contains(firstLand) {
		t.Fatal("the chosen land was not returned to hand")
	}
	if p1.Hand.Contains(secondLand) {
		t.Fatal("a second land was returned; the return is at most one")
	}
	if !p1.Graveyard.Contains(secondLand) {
		t.Fatal("the unchosen second land is not in the graveyard")
	}
}

// TestSixRealCardNoLandMilledIsNoOp proves accepting the optional return when no
// land was milled is a no-op: nothing enters hand even though the controller said
// yes.
func TestSixRealCardNoLandMilledIsNoOp(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	six := addCombatPermanent(g, game.Player1, cards.Six())
	a := addCardToLibrary(g, game.Player1, sixNonlandCard("Bear"))
	b := addCardToLibrary(g, game.Player1, sixNonlandCard("Ox"))
	c := addCardToLibrary(g, game.Player1, sixNonlandCard("Elk"))

	resolveSixAttackContent(g, six, sixLandAgent{acceptMay: true})

	p1 := g.Players[game.Player1]
	if p1.Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0 (no land was milled)", p1.Hand.Size())
	}
	for _, milled := range []id.ID{a, b, c} {
		if !p1.Graveyard.Contains(milled) {
			t.Fatalf("milled card %v is not in the graveyard", milled)
		}
	}
}

// TestSixRealCardLinkedMilledSetIsPerSourceInstance proves the milled-card set
// Six's land return draws from is scoped to the specific Six that milled, with no
// linked-key cross-talk between two Sixes. A land remembered as one Six's milled
// card is invisible to a different Six's real land-return choice (negative), yet
// returnable by the Six that milled it (positive control). Because the linked key
// is derived from the source permanent's card identity, the two Sixes read
// disjoint sets; the positive control rules out the land merely being ineligible.
// Driving each Six's real ChooseFromZone directly against a pre-remembered set
// isolates the key derivation itself, without the mill's clear-before-publish
// step masking a hypothetical shared key.
func TestSixRealCardLinkedMilledSetIsPerSourceInstance(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	sixOne := addCombatPermanent(g, game.Player1, cards.Six())
	sixTwo := addCombatPermanent(g, game.Player1, cards.Six())
	objOne := triggeredObjFor(sixOne)
	objTwo := triggeredObjFor(sixTwo)

	// A land sits in the graveyard, remembered as only the first Six's milled card.
	milledByOne := addCardToGraveyard(g, game.Player1, sixLandCard("Forest"))
	rememberLinkedObject(g, linkedObjectSourceKey(g, objOne, "milled-cards"), game.LinkedObjectRef{CardID: milledByOne})

	// Negative: the second Six's land return cannot see it, so accepting is a no-op.
	resolveChoose(g, objTwo, sixLandAgent{acceptMay: true, wantCardID: milledByOne}, sixLandChoicePrimitive(t))
	p1 := g.Players[game.Player1]
	if p1.Hand.Contains(milledByOne) {
		t.Fatal("a Six returned a land that only a different Six milled (linked-key cross-talk)")
	}
	if !p1.Graveyard.Contains(milledByOne) {
		t.Fatal("the land left the graveyard without being validly chosen")
	}

	// Positive control: the Six that milled the land can return it, proving the
	// negative result is source-identity scoping rather than the land being
	// ineligible on its own.
	resolveChoose(g, objOne, sixLandAgent{acceptMay: true, wantCardID: milledByOne}, sixLandChoicePrimitive(t))
	if !p1.Hand.Contains(milledByOne) {
		t.Fatal("the Six that milled the land could not return it from its own linked set")
	}
}

// TestSixRealCardLandChoiceIgnoresMilledCardThatLeftGraveyard proves an
// intervening zone change is honored: a milled land that leaves the graveyard
// before the land return resolves is no longer a candidate, so only the milled
// land still in the graveyard can be returned. It drives Six's real ChooseFromZone
// against a hand-published linked set standing in for the mill.
func TestSixRealCardLandChoiceIgnoresMilledCardThatLeftGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	six := addCombatPermanent(g, game.Player1, cards.Six())
	obj := triggeredObjFor(six)

	stillHere := addCardToGraveyard(g, game.Player1, sixLandCard("Forest"))
	leaves := addCardToGraveyard(g, game.Player1, sixLandCard("Island"))
	key := linkedObjectSourceKey(g, obj, "milled-cards")
	rememberLinkedObject(g, key, game.LinkedObjectRef{CardID: stillHere})
	rememberLinkedObject(g, key, game.LinkedObjectRef{CardID: leaves})

	// The second milled land leaves the graveyard for exile before the choice.
	g.Players[game.Player1].Graveyard.Remove(leaves)
	g.Players[game.Player1].Exile.Add(leaves)

	resolveChoose(g, obj, defaultChoiceAgent{}, sixLandChoicePrimitive(t))

	p1 := g.Players[game.Player1]
	if !p1.Hand.Contains(stillHere) {
		t.Fatal("the milled land still in the graveyard was not the returned card")
	}
	if p1.Hand.Contains(leaves) {
		t.Fatal("a milled land that had left the graveyard was returned")
	}
	if !p1.Exile.Contains(leaves) {
		t.Fatal("the milled land that left for exile did not stay in exile")
	}
}
