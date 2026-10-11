package cardgen

import (
	"reflect"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/mtg/rules"
	"github.com/natefinch/council4/opt"
)

type repeatQuantityPassAgent struct{}

func (repeatQuantityPassAgent) ChooseAction(rules.PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (repeatQuantityPassAgent) ChooseChoice(rules.PlayerObservation, game.ChoiceRequest) []int {
	return nil
}

func TestRepeatBodyProjectedDrawAndMillUseActualRemappedPlayerCount(t *testing.T) {
	t.Parallel()
	for _, mill := range []bool{false, true} {
		for _, stackPrefix := range []bool{false, true} {
			name := "draw"
			if mill {
				name = "mill"
			}
			if stackPrefix {
				name += "/stack prefix"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				player := game.TargetPlayerReference(0)
				amount := game.Dynamic(game.DynamicAmount{
					Kind: game.DynamicAmountCountCardsInZone, Player: &player, CardZone: zone.Hand, Multiplier: 1,
					Selection: new(game.Selection),
				})
				var action game.Primitive = game.Draw{Player: player, Amount: amount}
				if mill {
					action = game.MoveTopOfLibrary{Player: player, Amount: amount, Destination: zone.Graveyard}
				}
				original := game.RepeatProcess{Times: game.Fixed(2), Body: game.Mode{
					Sequence: []game.Instruction{{Primitive: action}},
				}.Ability()}
				projected, ok := rebaseTargetedPrimitive(original, 2, 1)
				if !ok {
					t.Fatal("nested draw/mill projection failed")
				}
				def := &game.CardDef{CardFace: game.CardFace{Name: "Projected Repeat", Types: []types.Card{types.Sorcery},
					SpellAbility: opt.Val(game.Mode{
						Targets: []game.TargetSpec{
							{Allow: game.TargetAllowPlayer, MinTargets: 1, MaxTargets: 1},
							{Allow: game.TargetAllowCard, TargetZone: zone.Graveyard, MinTargets: 1, MaxTargets: 1},
							{Allow: game.TargetAllowPlayer, MinTargets: 1, MaxTargets: 1},
						},
						Sequence: []game.Instruction{{Primitive: projected}},
					}.Ability()),
				}}
				if stackPrefix {
					def.SpellAbility.Val.Modes[0].Targets[0].Allow = game.TargetAllowStackObject
					def.SpellAbility.Val.Modes[0].Targets[0].Predicate.StackObjectKinds = []game.StackObjectKind{game.StackSpell}
				}
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				g.Turn.ActivePlayer, g.Turn.PriorityPlayer = game.Player4, game.Player4
				g.Turn.Phase, g.Turn.Step = game.PhasePrecombatMain, game.StepNone
				add := func(owner game.PlayerID, def *game.CardDef) id.ID {
					cardID := g.IDGen.Next()
					g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Owner: owner, Def: def}
					return cardID
				}
				vanilla := &game.CardDef{CardFace: game.CardFace{Name: "Counted card", Types: []types.Card{types.Creature}}}
				g.Players[game.Player1].Hand.Add(add(game.Player1, vanilla))
				for range 3 {
					g.Players[game.Player2].Hand.Add(add(game.Player2, vanilla))
				}
				for range 12 {
					g.Players[game.Player1].Library.Add(add(game.Player1, vanilla))
					g.Players[game.Player2].Library.Add(add(game.Player2, vanilla))
				}
				prefixCard := add(game.Player3, vanilla)
				g.Players[game.Player3].Graveyard.Add(prefixCard)
				firstTarget := game.PlayerTarget(game.Player1)
				if stackPrefix {
					prefixDef := &game.CardDef{CardFace: game.CardFace{Name: "Preceding stack target", Types: []types.Card{types.Sorcery},
						SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
							Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
						}}}.Ability()),
					}}
					prefixSource := add(game.Player1, prefixDef)
					prefixObject := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell,
						SourceID: prefixSource, SourceCardID: prefixSource, Controller: game.Player1}
					g.Stack.Push(prefixObject)
					firstTarget = game.StackObjectTarget(prefixObject.ID)
				}
				source := add(game.Player4, def)
				obj := &game.StackObject{ID: g.IDGen.Next(), Kind: game.StackSpell,
					SourceID: source, SourceCardID: source, Controller: game.Player4,
					Targets: []game.Target{firstTarget, game.CardTarget(prefixCard), game.PlayerTarget(game.Player2)},
				}
				g.Stack.Push(obj)
				definitionBefore := appendPaths(nil, "CardDef", reflect.ValueOf(def))
				var policies [game.NumPlayers]rules.PlayerAgent
				for i := range policies {
					policies[i] = repeatQuantityPassAgent{}
				}
				resolved := rules.NewEngine(nil).Simulator().ResolvePriority(g, policies)
				wantHand, wantGraveyard, wantLibrary := 12, 0, 3
				if mill {
					wantHand, wantGraveyard, wantLibrary = 3, 6, 6
				}
				if resolved.Players[game.Player2].Hand.Size() != wantHand ||
					resolved.Players[game.Player2].Graveyard.Size() != wantGraveyard ||
					resolved.Players[game.Player2].Library.Size() != wantLibrary {
					t.Fatalf("remapped player2 hand/graveyard/library=%d/%d/%d, want %d/%d/%d",
						resolved.Players[game.Player2].Hand.Size(), resolved.Players[game.Player2].Graveyard.Size(),
						resolved.Players[game.Player2].Library.Size(), wantHand, wantGraveyard, wantLibrary)
				}
				if resolved.Players[game.Player1].Hand.Size() != 1 || resolved.Players[game.Player1].Library.Size() != 12 ||
					g.Players[game.Player2].Hand.Size() != 3 || g.Players[game.Player2].Library.Size() != 12 {
					t.Fatal("repeat mutated the neighboring target or caller's cloned game")
				}
				before := amount.DynamicAmount().Val
				if *before.Player != game.TargetPlayerReference(0) {
					t.Fatal("runtime or projection mutated the reused original quantity")
				}
				if !slices.Equal(definitionBefore, appendPaths(nil, "CardDef", reflect.ValueOf(def))) {
					t.Fatal("resolving a cloned stack changed the reused CardDef")
				}
			})
		}
	}
}
