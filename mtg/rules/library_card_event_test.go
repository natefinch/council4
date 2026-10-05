package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestLibraryCardEventConsumerNeverUsesObservedCardOrTargetDecoyAsAttacker(t *testing.T) {
	for _, forest := range []bool{false, true} {
		t.Run(map[bool]string{false: "nonmatching", true: "matching"}[forest], func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatCreaturePermanentWithPower(g, game.Player1, 4)
			target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
			attacker := addCombatCreaturePermanentWithPower(g, game.Player3, 6)
			g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{
				{Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
				{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player1}},
			}}
			below := addCardToLibrary(g, game.Player1, vanillaCreature("New Top", 99, 99))
			def := &game.CardDef{CardFace: game.CardFace{Name: "Observed", Types: []types.Card{types.Land}}}
			if forest {
				def.Subtypes = []types.Sub{types.Forest}
			}
			cardID := addCardToLibrary(g, game.Player1, def)
			g.CardInstances[cardID].ZoneVersion = 7
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1, SourceID: source.ObjectID,
				Targets:         []game.Target{game.PlayerTarget(game.Player2), game.PermanentTarget(target.ObjectID)},
				HasTriggerEvent: true, TriggerEvent: game.Event{PermanentID: attacker.ObjectID},
			}
			resolver := newEffectResolver(NewEngine(nil), g, obj, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			resolver.resolveInstruction(&game.Instruction{Primitive: game.Reveal{
				Player: game.ControllerReference(), Amount: game.Fixed(1), PublishLinked: "observed",
			}})
			resolver.resolveInstruction(&game.Instruction{
				Primitive: game.RemoveFromCombat{Object: game.EventPermanentReference()},
				Condition: opt.Val(game.EffectCondition{Condition: opt.Val(game.Condition{
					Object:        opt.Val(game.LinkedObjectReference("observed")),
					ObjectMatches: opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Forest}}),
				})}),
			})
			resolver.resolveInstruction(&game.Instruction{Primitive: game.MoveCard{
				Card:     game.CardReference{Kind: game.CardReferenceLinked, LinkID: "observed"},
				FromZone: zone.Library, Destination: zone.Library, DestinationBottom: true,
			}})
			attacks := map[id.ID]bool{}
			for _, declaration := range g.Combat.Attackers {
				attacks[declaration.Attacker] = true
			}
			top, _ := g.Players[game.Player1].Library.Top()
			if attacks[attacker.ObjectID] == forest || !attacks[target.ObjectID] ||
				top != below || g.CardInstances[cardID].ZoneVersion != 7 {
				t.Fatal("card predicate, event attacker, target decoy, or unconditional placement became aliased")
			}
		})
	}
}
