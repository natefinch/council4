package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// selectAllSearchAgent accepts any may-choice and takes every card a search
// offers, up to the request's maximum. It drives a multi-card controlled search
// where the deciding player finds more than one card at once.
type selectAllSearchAgent struct{}

func (selectAllSearchAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (selectAllSearchAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if request.Kind == game.ChoiceMay {
		return []int{1}
	}
	indices := make([]int, 0, len(request.Options))
	for _, option := range request.Options {
		indices = append(indices, option.Index)
	}
	return indices
}

// oppositionAgentDef returns a creature carrying Opposition Agent's two static
// abilities: RuleEffectControlOpponentSearches (its controller makes every choice
// while an opponent searches their library) and RuleEffectExileOpponentSearchFinds
// (every card such a search finds is exiled instead, and its controller may play
// the exiled card spending mana of any color for as long as it stays exiled). The
// two effects live on separate static abilities to mirror the lowered card.
func oppositionAgentDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Opposition Agent",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{
			{RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectControlOpponentSearches}}},
			{RuleEffects: []game.RuleEffect{{Kind: game.RuleEffectExileOpponentSearchFinds}}},
		},
	}}
}

// bearDef and elephantDef are distinct creature cards used to prove which card a
// controlled search actually took.
func bearDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}}
}

func elephantDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Elephant", Types: []types.Card{types.Creature}}}
}

// searchInstruction builds a single mandatory library search whose searcher is the
// resolving spell's controller, so the controller of the spell on the stack is the
// player whose library is searched.
func searchInstruction(spec game.SearchSpec) []game.Instruction {
	return []game.Instruction{{
		Primitive: game.Search{
			Amount: game.Fixed(1),
			Player: game.ControllerReference(),
			Spec:   spec,
		},
	}}
}

// playFromExilePermissionFor returns the lasting play-from-exile permission that a
// controlled search granted a beneficiary for a specific exiled card, if any.
func playFromExilePermissionFor(g *game.Game, cardID id.ID) (game.RuleEffect, bool) {
	for _, effect := range g.RuleEffects {
		if effect.Kind == game.RuleEffectPlayFromZone &&
			effect.AffectedCardID == cardID &&
			effect.CastFromZone == zone.Exile {
			return effect, true
		}
	}
	return game.RuleEffect{}, false
}

// TestOppositionAgentExilesEveryCardOfMultiCardSearch verifies that a controlled
// search finding more than one card exiles every found card and grants the Agent
// controller a play permission for each.
func TestOppositionAgentExilesEveryCardOfMultiCardSearch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, oppositionAgentDef())
	bear := addCardToLibrary(g, game.Player2, bearDef())
	elephant := addCardToLibrary(g, game.Player2, elephantDef())
	addInstructionSpellToStackForController(g, game.Player2, []game.Instruction{{
		Primitive: game.Search{
			Amount: game.Fixed(2),
			Player: game.ControllerReference(),
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
		},
	}}, nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: selectAllSearchAgent{}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p2 := g.Players[game.Player2]
	for _, card := range []id.ID{bear, elephant} {
		if !p2.Exile.Contains(card) {
			t.Fatalf("card %d found by the multi-card search was not exiled", card)
		}
		if p2.Hand.Contains(card) {
			t.Fatalf("card %d entered the searching opponent's hand instead of exile", card)
		}
		if _, ok := playFromExilePermissionFor(g, card); !ok {
			t.Fatalf("no play-from-exile permission was granted for card %d", card)
		}
	}
	castable := foreignExileCastableCards(g, game.Player1)
	if !slices.Contains(castable, bear) || !slices.Contains(castable, elephant) {
		t.Fatalf("controller should be able to play every exiled card, got %v", castable)
	}
}

// TestOppositionAgentExilesOpponentTutorToHand covers the core behavior: while a
// controller has Opposition Agent, an opponent's tutor-to-hand exiles the found
// card to that opponent's exile instead of their hand, and the controller gains a
// lasting any-color play permission for it.
func TestOppositionAgentExilesOpponentTutorToHand(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, oppositionAgentDef())
	bear := addCardToLibrary(g, game.Player2, bearDef())
	addInstructionSpellToStackForController(g, game.Player2, searchInstruction(game.SearchSpec{
		SourceZone:  zone.Library,
		Destination: zone.Hand,
	}), nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Bear"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p2 := g.Players[game.Player2]
	if p2.Hand.Contains(bear) {
		t.Fatal("found card entered the searching opponent's hand instead of exile")
	}
	if p2.Library.Contains(bear) {
		t.Fatal("found card was left in the library; the search did not complete")
	}
	if !p2.Exile.Contains(bear) {
		t.Fatal("found card was not exiled to its owner")
	}
	effect, ok := playFromExilePermissionFor(g, bear)
	if !ok {
		t.Fatal("no play-from-exile permission was granted for the exiled card")
	}
	if effect.Controller != game.Player1 {
		t.Fatalf("permission controller = %d, want Player1", effect.Controller)
	}
	if !effect.SpendAnyMana {
		t.Fatal("permission does not allow spending mana as any color")
	}
	if effect.Duration != game.DurationPermanent {
		t.Fatalf("permission duration = %v, want DurationPermanent", effect.Duration)
	}
	if !slices.Contains(foreignExileCastableCards(g, game.Player1), bear) {
		t.Fatal("Opposition Agent controller cannot play the exiled card")
	}
	if !castFromZoneAllowsAnyMana(g, game.Player1, bear, zone.Exile, game.FaceFront) {
		t.Fatal("exiled card cannot be cast spending mana of any color")
	}
	if slices.Contains(foreignExileCastableCards(g, game.Player2), bear) {
		t.Fatal("the searching opponent must not gain permission to play their exiled card")
	}
}

// TestOppositionAgentControllerMakesSearchDecision proves the Agent controller,
// not the searching opponent, chooses which card a controlled search finds: only
// the controller has an agent, and it selects the second of two matching cards.
func TestOppositionAgentControllerMakesSearchDecision(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, oppositionAgentDef())
	bear := addCardToLibrary(g, game.Player2, bearDef())
	elephant := addCardToLibrary(g, game.Player2, elephantDef())
	addInstructionSpellToStackForController(g, game.Player2, searchInstruction(game.SearchSpec{
		SourceZone:  zone.Library,
		Destination: zone.Hand,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}), nil)
	// Only the Agent controller has an agent; it wants the Elephant. The searcher
	// has none, so a default choice would take the first candidate (the Bear).
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Elephant"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p2 := g.Players[game.Player2]
	if !p2.Exile.Contains(elephant) {
		t.Fatal("controller's chosen card (Elephant) was not the one exiled")
	}
	if p2.Exile.Contains(bear) {
		t.Fatal("a card the controller did not choose was exiled")
	}
	if !p2.Library.Contains(bear) {
		t.Fatal("the unchosen card should remain in the library")
	}
}

// TestOppositionAgentExilesBattlefieldTutorFind covers a search that would put a
// card onto the battlefield (a fetch land): under Opposition Agent the card is
// exiled instead of entering play.
func TestOppositionAgentExilesBattlefieldTutorFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, oppositionAgentDef())
	forest := addCardToLibrary(g, game.Player2, basicForestDef())
	addInstructionSpellToStackForController(g, game.Player2, searchInstruction(game.SearchSpec{
		SourceZone:  zone.Library,
		Destination: zone.Battlefield,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Land}},
	}), nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Forest"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	for _, perm := range g.Battlefield {
		if perm.CardInstanceID == forest {
			t.Fatal("found land entered the battlefield instead of being exiled")
		}
	}
	if !g.Players[game.Player2].Exile.Contains(forest) {
		t.Fatal("found land was not exiled to its owner")
	}
	if _, ok := playFromExilePermissionFor(g, forest); !ok {
		t.Fatal("no play-from-exile permission was granted for the exiled land")
	}
}

// TestOppositionAgentPreservesFailureToFind verifies that a controlled search that
// may legally fail to find still can: the Agent controller declines, nothing is
// exiled, and the search completes leaving the library intact.
func TestOppositionAgentPreservesFailureToFind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, oppositionAgentDef())
	bear := addCardToLibrary(g, game.Player2, bearDef())
	addInstructionSpellToStackForController(g, game.Player2, searchInstruction(game.SearchSpec{
		SourceZone:  zone.Library,
		Destination: zone.Hand,
		Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}},
	}), nil)
	// A qualified search ("for a creature card") may fail to find; the controller
	// chooses nothing by wanting a card that is not present.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Absent"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p2 := g.Players[game.Player2]
	if !p2.Library.Contains(bear) {
		t.Fatal("failing to find must leave the matching card in the library")
	}
	if p2.Exile.Contains(bear) {
		t.Fatal("nothing should be exiled when the controlled search finds no card")
	}
	if _, ok := playFromExilePermissionFor(g, bear); ok {
		t.Fatal("no play permission should be granted when nothing is found")
	}
}

// TestOppositionAgentDoesNotAffectControllerOwnSearch verifies the Agent controller
// searching their own library is unaffected: the found card enters their hand
// normally and no exile permission is created.
func TestOppositionAgentDoesNotAffectControllerOwnSearch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, oppositionAgentDef())
	bear := addCardToLibrary(g, game.Player1, bearDef())
	addInstructionSpellToStackForController(g, game.Player1, searchInstruction(game.SearchSpec{
		SourceZone:  zone.Library,
		Destination: zone.Hand,
	}), nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Bear"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p1 := g.Players[game.Player1]
	if !p1.Hand.Contains(bear) {
		t.Fatal("controller's own search should put the found card into their hand")
	}
	if p1.Exile.Contains(bear) {
		t.Fatal("controller's own search should not exile the found card")
	}
	if _, ok := playFromExilePermissionFor(g, bear); ok {
		t.Fatal("controller's own search should not create an exile play permission")
	}
}

// TestOppositionAgentMostRecentAgentControls verifies deterministic resolution
// when several opponents could control the same search: the controller of the
// Opposition Agent that most recently entered the battlefield controls it and
// becomes the beneficiary (per the card's ruling). Player3's Agent enters first
// and Player1's Agent second, so Player1 controls the search even though Player3
// is the nearer opponent to the searcher in turn order — proving entry order, not
// seating, decides.
func TestOppositionAgentMostRecentAgentControls(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player3, oppositionAgentDef())
	addCombatPermanent(g, game.Player1, oppositionAgentDef())
	bear := addCardToLibrary(g, game.Player2, bearDef())
	addInstructionSpellToStackForController(g, game.Player2, searchInstruction(game.SearchSpec{
		SourceZone:  zone.Library,
		Destination: zone.Hand,
	}), nil)
	// Player1's Agent entered most recently, so Player1 (not the nearer Player3)
	// makes the decision and gains the permission.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Bear"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player2].Exile.Contains(bear) {
		t.Fatal("found card was not exiled")
	}
	effect, ok := playFromExilePermissionFor(g, bear)
	if !ok {
		t.Fatal("no play-from-exile permission was granted")
	}
	if effect.Controller != game.Player1 {
		t.Fatalf("permission controller = %d, want Player1 (most recent Agent)", effect.Controller)
	}
	if slices.Contains(foreignExileCastableCards(g, game.Player3), bear) {
		t.Fatal("an older Agent's controller must not gain the permission")
	}
	if !slices.Contains(foreignExileCastableCards(g, game.Player1), bear) {
		t.Fatal("the most recent Agent's controller should be able to play the exiled card")
	}
}

// TestOppositionAgentPermissionPersistsAfterAgentLeaves verifies the granted play
// permission lasts for as long as the card stays exiled even after the Opposition
// Agent leaves the battlefield (CR 610.3b).
func TestOppositionAgentPermissionPersistsAfterAgentLeaves(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	agent := addCombatPermanent(g, game.Player1, oppositionAgentDef())
	bear := addCardToLibrary(g, game.Player2, bearDef())
	addInstructionSpellToStackForController(g, game.Player2, searchInstruction(game.SearchSpec{
		SourceZone:  zone.Library,
		Destination: zone.Hand,
	}), nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalSearchAgent{accept: true, wanted: "Bear"}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !slices.Contains(foreignExileCastableCards(g, game.Player1), bear) {
		t.Fatal("controller cannot play the exiled card while the Agent is present")
	}
	g.Battlefield = removePermanent(g.Battlefield, agent)
	if !slices.Contains(foreignExileCastableCards(g, game.Player1), bear) {
		t.Fatal("permission must persist while the card remains exiled after the Agent leaves")
	}
}

// TestOppositionAgentRedirectsRevealOnlySearch covers the reveal-only search path
// (Scholar of New Horizons): an opponent's search reveals a card and a following
// ConditionalDestinationPlace would route it to the battlefield or hand. Under
// Opposition Agent the controller makes the choice, the revealed card is exiled
// face up to its owner with a lasting play permission for the controller, and the
// follow-up placement finds no published card and is a safe no-op — the card
// reaches neither the searcher's hand nor the battlefield.
func TestOppositionAgentRedirectsRevealOnlySearch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, oppositionAgentDef())
	plains := addCardToLibrary(g, game.Player2, plainsCard())
	addInstructionSpellToStackForController(g, game.Player2, scholarSequence(), nil)
	// Only the Agent controller has an agent; it makes the reveal-only choice and
	// picks the Plains. The searcher, Player2, makes no choices here.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: conditionalDestinationAgent{wanted: "Plains", acceptPut: true}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p2 := g.Players[game.Player2]
	if !p2.Exile.Contains(plains) {
		t.Fatal("reveal-only search find was not exiled to its owner")
	}
	if p2.Exile.IsFaceDown(plains) {
		t.Fatal("Opposition Agent exiles the reveal-only find face up, not face down")
	}
	if p2.Hand.Contains(plains) {
		t.Fatal("the reveal-only follow-up must not route the redirected card into the searcher's hand")
	}
	for _, perm := range g.Battlefield {
		if perm.CardInstanceID == plains {
			t.Fatal("the redirected reveal-only find must not enter the battlefield")
		}
	}
	if p2.Library.Contains(plains) {
		t.Fatal("the found card must have left the searched library")
	}
	effect, ok := playFromExilePermissionFor(g, plains)
	if !ok {
		t.Fatal("no play-from-exile permission was granted for the reveal-only find")
	}
	if effect.Controller != game.Player1 || !effect.SpendAnyMana {
		t.Fatalf("permission = {controller:%d anyMana:%v}, want Player1 with any-color spending", effect.Controller, effect.SpendAnyMana)
	}
	if !slices.Contains(foreignExileCastableCards(g, game.Player1), plains) {
		t.Fatal("Agent controller cannot play the exiled reveal-only find")
	}
}

// TestOppositionAgentRedirectsExileFaceDownSearch covers the exile-face-down
// search path (Beseech the Mirror): an opponent's search exiles a found card face
// down and a following linked free cast or move-to-hand would use it. Under
// Opposition Agent the controller makes the choice, the found card is exiled face
// up (overriding the face-down exile) to its owner with a lasting play permission
// for the controller, and no card is published, so the searcher's linked free cast
// and move-to-hand fallback both find nothing and are safe no-ops.
func TestOppositionAgentRedirectsExileFaceDownSearch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, oppositionAgentDef())
	spell := addCardToLibrary(g, game.Player2, instantDef("Cheap Bolt", 2))
	// Player2 casts a bargained Beseech-style search/exile/conditional-cast payoff.
	pushBeseechPayoff(g, game.Player2, true, false)
	// The Agent controller (Player1) makes the search choice; Player2 would accept
	// the linked free cast, proving the redirect (not a decline) makes it a no-op.
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: beseechAgent{acceptCast: true},
		game.Player2: beseechAgent{acceptCast: true},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	p2 := g.Players[game.Player2]
	if !p2.Exile.Contains(spell) {
		t.Fatal("exile-face-down search find was not exiled to its owner")
	}
	if p2.Exile.IsFaceDown(spell) {
		t.Fatal("Opposition Agent exiles the find face up, overriding the searcher's face-down exile")
	}
	if p2.Hand.Contains(spell) {
		t.Fatal("the searcher's linked move-to-hand fallback must find nothing (no-op)")
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0: the searcher's linked free cast must find nothing to cast", g.Stack.Size())
	}
	if p2.Library.Contains(spell) {
		t.Fatal("the found card must have left the searched library")
	}
	effect, ok := playFromExilePermissionFor(g, spell)
	if !ok {
		t.Fatal("no play-from-exile permission was granted for the exile-face-down find")
	}
	if effect.Controller != game.Player1 || !effect.SpendAnyMana {
		t.Fatalf("permission = {controller:%d anyMana:%v}, want Player1 with any-color spending", effect.Controller, effect.SpendAnyMana)
	}
	if !slices.Contains(foreignExileCastableCards(g, game.Player1), spell) {
		t.Fatal("Agent controller cannot play the exiled find")
	}
}
