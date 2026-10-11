package rules

import (
	"strconv"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestStunUntapSteps(t *testing.T) {
	for _, otherStep := range []bool{false, true} {
		for _, stun := range []int{0, 1, 2} {
			t.Run(strconv.FormatBool(otherStep)+"/"+strconv.Itoa(stun), func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				target := addCreaturePermanent(g, game.Player1)
				target.Tapped = true
				target.SummoningSick = true
				target.Counters.Add(counter.Stun, stun)
				target.Counters.Add(counter.Charge, 2)
				opponent := addCreaturePermanent(g, game.Player3)
				opponent.Tapped = true
				opponent.Counters.Add(counter.Stun, 2)
				active := game.Player1
				if otherStep {
					active = game.Player2
					// Overlapping grants are one untap event, not two replacements.
					for range 2 {
						addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
							Name: "Untap Grant", Types: []types.Card{types.Enchantment},
							StaticAbilities: []game.StaticAbility{seedbornStaticBody()},
						}})
					}
				}
				g.Turn.ActivePlayer = active
				g.Turn.PriorityPlayer = active
				addCardToLibrary(g, active, &game.CardDef{CardFace: game.CardFace{Name: "Draw"}})
				NewEngine(nil).runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if target.Tapped != (stun > 0) || target.Counters.Get(counter.Stun) != max(0, stun-1) ||
					target.Counters.Get(counter.Charge) != 2 || target.SummoningSick != otherStep {
					t.Fatalf("step state = tapped %v, stun %d, sick %v; want tapped %v, stun %d, sick %v",
						target.Tapped, target.Counters.Get(counter.Stun), target.SummoningSick, stun > 0, max(0, stun-1), otherStep)
				}
				if !opponent.Tapped || opponent.Counters.Get(counter.Stun) != 2 {
					t.Fatal("untap step changed an unrelated opponent's permanent")
				}
				events := stunUntappedEvents(g)
				if stun == 0 {
					if len(events) != 1 || events[0].PermanentID != target.ObjectID {
						t.Fatalf("step events = %+v, want actual target untap", events)
					}
				} else if len(events) != 0 {
					t.Fatalf("stun-replaced step emitted untap events: %+v", events)
				}
			})
		}
	}
}

func TestStunUntapStepRestrictionsPrecedeReplacement(t *testing.T) {
	for _, otherStep := range []bool{false, true} {
		for _, restriction := range []string{"doesnt-untap", "exerted", "phased-out"} {
			t.Run(strconv.FormatBool(otherStep)+"/"+restriction, func(t *testing.T) {
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				def := &game.CardDef{CardFace: game.CardFace{Name: "Restricted", Types: []types.Card{types.Creature}}}
				if restriction == "doesnt-untap" {
					def.StaticAbilities = []game.StaticAbility{game.DoesntUntapStaticBody}
				}
				target := addCombatPermanent(g, game.Player1, def)
				target.Tapped = true
				target.Counters.Add(counter.Stun, 2)
				target.Exerted = restriction == "exerted"
				active := game.Player1
				if otherStep {
					active = game.Player2
					addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
						Name: "Untap Grant", Types: []types.Card{types.Enchantment},
						StaticAbilities: []game.StaticAbility{seedbornStaticBody()},
					}})
				}
				if restriction == "phased-out" {
					if !phaseOutPermanentTree(g, target, game.Player1, make(map[game.ObjectID]bool)) {
						t.Fatal("phase out failed")
					}
				}
				g.Turn.ActivePlayer = active
				g.Turn.PriorityPlayer = active
				addCardToLibrary(g, active, &game.CardDef{CardFace: game.CardFace{Name: "Draw"}})
				NewEngine(nil).runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				attempted := (restriction == "exerted" && otherStep) || (restriction == "phased-out" && !otherStep)
				wantStun := 2
				if attempted {
					wantStun = 1
				}
				if !target.Tapped || target.Counters.Get(counter.Stun) != wantStun || len(stunUntappedEvents(g)) != 0 {
					t.Fatalf("restricted step = tapped %v, stun %d; want true, %d", target.Tapped, target.Counters.Get(counter.Stun), wantStun)
				}
				if restriction == "exerted" && target.Exerted != otherStep {
					t.Fatal("exertion was not consumed only in the controller's untap step")
				}
				if restriction == "doesnt-untap" {
					// The static forbids untapping during the step, not by effects.
					obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.PermanentTarget(target.ObjectID)}}
					NewEngine(nil).resolveInstructionWithChoices(g, obj, &game.Instruction{
						Primitive: game.Untap{Object: game.TargetPermanentReference(0)},
					}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
					if !target.Tapped || target.Counters.Get(counter.Stun) != 1 {
						t.Fatal("step-only restriction incorrectly blocked ability-based stun replacement")
					}
				}
			})
		}
	}
}
