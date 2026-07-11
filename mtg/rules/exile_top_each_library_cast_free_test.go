package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// castFirstResolutionAgent casts eagerly: for every resolution choice it selects
// the first offered option, so a repeated "cast any number" loop casts every
// castable card it is offered.
type castFirstResolutionAgent struct{}

func (castFirstResolutionAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (castFirstResolutionAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	if len(request.Options) == 0 {
		return nil
	}
	return []int{request.Options[0].Index}
}

func spellCardDef(name string, cardType types.Card) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:     name,
		Types:    []types.Card{cardType},
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
	}}
}

// etaliInstruction is the exile-top-of-each-library-then-cast-any-number-free
// primitive with a per-library exile count of one, as Etali, Primal Storm lowers.
func etaliInstruction() *game.Instruction {
	return &game.Instruction{Primitive: game.ExileTopEachLibraryCastFree{Amount: game.Fixed(1)}}
}

// TestExileTopEachLibraryCastFreeCastsEveryPlayersTop exiles the top card of all
// four players' libraries and lets the controller cast every castable one for
// free, including cards drawn from opponents' libraries, while the uncastable
// land stays exiled.
func TestExileTopEachLibraryCastFreeCastsEveryPlayersTop(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, spellCardDef("Etali, Primal Storm", types.Creature))
	obj := triggeredObjFor(source)

	ownTop := addCardToLibrary(g, game.Player1, spellCardDef("Own Bolt", types.Instant))
	oppSorcery := addCardToLibrary(g, game.Player2, spellCardDef("Opp Sorcery", types.Sorcery))
	oppLand := addCardToLibrary(g, game.Player3, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opp Forest",
		Types: []types.Card{types.Land},
	}})
	oppInstant := addCardToLibrary(g, game.Player4, spellCardDef("Opp Shock", types.Instant))

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: castFirstResolutionAgent{},
		game.Player2: defaultChoiceAgent{},
		game.Player3: defaultChoiceAgent{},
		game.Player4: defaultChoiceAgent{},
	}
	engine.resolveInstructionWithChoices(g, obj, etaliInstruction(), agents, &TurnLog{})

	if !g.Players[game.Player3].Exile.Contains(oppLand) {
		t.Fatal("uncastable land was not left exiled")
	}
	for _, castID := range []id.ID{ownTop, oppSorcery, oppInstant} {
		if anyZoneStillHolds(g, castID) {
			t.Fatalf("card %d was not cast off the stack-bound exile", castID)
		}
	}
	if g.Stack.Size() != 3 {
		t.Fatalf("stack size = %d, want 3 free-cast spells", g.Stack.Size())
	}
	for _, stackObj := range g.Stack.Objects() {
		if stackObj.Controller != game.Player1 {
			t.Fatalf("stack object %d controller = %v, want Player1 casting every spell", stackObj.SourceID, stackObj.Controller)
		}
	}
}

// TestExileTopEachLibraryCastFreeDeclinedLeavesCardsExiled proves the "you may"
// clause: a controller who casts nothing still exiles every player's top card,
// and each stays exiled with an empty stack.
func TestExileTopEachLibraryCastFreeDeclinedLeavesCardsExiled(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, spellCardDef("Etali, Primal Storm", types.Creature))
	obj := triggeredObjFor(source)

	tops := [game.NumPlayers]id.ID{
		game.Player1: addCardToLibrary(g, game.Player1, spellCardDef("P1 Bolt", types.Instant)),
		game.Player2: addCardToLibrary(g, game.Player2, spellCardDef("P2 Sorcery", types.Sorcery)),
		game.Player3: addCardToLibrary(g, game.Player3, spellCardDef("P3 Bolt", types.Instant)),
		game.Player4: addCardToLibrary(g, game.Player4, spellCardDef("P4 Sorcery", types.Sorcery)),
	}

	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: defaultChoiceAgent{},
		game.Player2: defaultChoiceAgent{},
		game.Player3: defaultChoiceAgent{},
		game.Player4: defaultChoiceAgent{},
	}
	engine.resolveInstructionWithChoices(g, obj, etaliInstruction(), agents, &TurnLog{})

	for player, cardID := range tops {
		if !g.Players[player].Exile.Contains(cardID) {
			t.Fatalf("player %d top card was not exiled", player)
		}
	}
	if g.Stack.Size() != 0 {
		t.Fatalf("stack size = %d, want 0 (controller cast nothing)", g.Stack.Size())
	}
}

// anyZoneStillHolds reports whether cardID rests in any player's library or
// exile, the zones a cast free spell must have left for the stack.
func anyZoneStillHolds(g *game.Game, cardID id.ID) bool {
	for i := range g.Players {
		player := g.Players[i]
		if player.Library.Contains(cardID) || player.Exile.Contains(cardID) {
			return true
		}
	}
	return false
}
