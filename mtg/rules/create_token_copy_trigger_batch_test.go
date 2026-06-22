package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// twilightDivinerTrigger builds the "Whenever one or more other creatures you
// control enter, ... create a token that's a copy of one of them." inline
// trigger used to drive the copy-of-triggering-set runtime tests.
func twilightDivinerTrigger() *game.TriggeredAbility {
	return &game.TriggeredAbility{
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:            game.EventPermanentEnteredBattlefield,
				Controller:       game.TriggerControllerYou,
				ExcludeSelf:      true,
				OneOrMore:        true,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.CreateToken{
				Amount: game.Fixed(1),
				Source: game.TokenCopyOf(game.TokenCopySpec{
					Source: game.TokenCopySourceChosenFromTriggerBatch,
				}),
			},
		}}}.Ability(),
	}
}

func enteringCreature(g *game.Game, controller game.PlayerID, name string, power int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}})
}

// fixedSelectionAgent answers every ChooseChoice with a fixed option selection.
type fixedSelectionAgent struct {
	selection []int
}

func (fixedSelectionAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Action{}
}

func (a fixedSelectionAgent) ChooseChoice(_ PlayerObservation, _ game.ChoiceRequest) []int {
	return a.selection
}

// TestCreateCopyTokenFromTriggerBatchChoosesAmongTriggeringCreatures verifies
// that "create a token that's a copy of one of them." copies a controller-chosen
// member of the triggering event batch, restricting candidates to the creatures
// that triggered the ability (controlled by the source's controller) and
// excluding simultaneously-entering permanents an opponent controls.
func TestCreateCopyTokenFromTriggerBatchChoosesAmongTriggeringCreatures(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Twilight Diviner",
		Types: []types.Card{types.Creature},
	}})
	first := enteringCreature(g, game.Player1, "Alpha Beast", 3)
	second := enteringCreature(g, game.Player1, "Beta Beast", 4)
	opponent := enteringCreature(g, game.Player2, "Enemy Beast", 9)

	batchID := g.IDGen.Next()
	for _, permanent := range []*game.Permanent{first, second, opponent} {
		emitEvent(g, game.Event{
			Kind:           game.EventPermanentEnteredBattlefield,
			PermanentID:    permanent.ObjectID,
			Controller:     permanent.Controller,
			SimultaneousID: batchID,
		})
	}

	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		Controller:      game.Player1,
		SourceID:        source.ObjectID,
		InlineTrigger:   twilightDivinerTrigger(),
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:           game.EventPermanentEnteredBattlefield,
			PermanentID:    first.ObjectID,
			Controller:     game.Player1,
			SimultaneousID: batchID,
		},
	}

	// Agent that always picks the second candidate to copy.
	agents := [game.NumPlayers]PlayerAgent{}
	agents[game.Player1] = fixedSelectionAgent{selection: []int{1}}
	r := &effectResolver{engine: NewEngine(nil), game: g, obj: obj, agents: agents, log: &TurnLog{}}

	candidates := r.triggeringBatchPermanents()
	if len(candidates) != 2 {
		t.Fatalf("triggeringBatchPermanents = %d candidates, want 2 (opponent excluded)", len(candidates))
	}

	resolved := handleCreateToken(r, game.CreateToken{
		Amount: game.Fixed(1),
		Source: game.TokenCopyOf(game.TokenCopySpec{Source: game.TokenCopySourceChosenFromTriggerBatch}),
	})
	if !resolved.succeeded {
		t.Fatal("handleCreateToken did not succeed")
	}

	copies := map[string]int{}
	for _, permanent := range g.Battlefield {
		if permanent.Token && permanent.Controller == game.Player1 && permanent.TokenDef != nil {
			copies[permanent.TokenDef.Name]++
		}
	}
	if copies["Beta Beast"] != 1 {
		t.Fatalf("expected one Beta Beast copy, got %#v", copies)
	}
	if copies["Enemy Beast"] != 0 {
		t.Fatalf("opponent creature must not be a candidate, got %#v", copies)
	}
}
