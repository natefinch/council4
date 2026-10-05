package rules

import (
	"fmt"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestPastStackConditionInformationIdentity(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		setup func(*game.Game, *game.StackObject, *game.StackObject)
		want  bool
	}{
		{name: "exact countered target", want: true},
		{name: "removed externally", setup: func(g *game.Game, target, obj *game.StackObject) {
			g.Stack.RemoveByID(target.ID)
			delete(obj.TargetManaValueLKI, 1)
			delete(obj.TargetManaValueLKIObjectIDs, 1)
		}},
		{name: "reassigned missing target", setup: func(g *game.Game, _, obj *game.StackObject) { obj.Targets[1] = game.StackObjectTarget(g.IDGen.Next()) }},
		{name: "reassigned known live target", setup: func(g *game.Game, _, obj *game.StackObject) {
			replacement := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(8)})
			obj.Targets[1] = game.StackObjectTarget(replacement.ID)
		}},
		{name: "reassigned qualifying live target", setup: func(g *game.Game, _, obj *game.StackObject) {
			replacement := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(2)})
			obj.Targets[1] = game.StackObjectTarget(replacement.ID)
		}},
		{name: "absent identity", setup: func(_ *game.Game, _, obj *game.StackObject) { obj.TargetManaValueLKIObjectIDs = nil }},
		{name: "missing target slot", setup: func(_ *game.Game, _, obj *game.StackObject) { obj.Targets = obj.Targets[:1] }},
		{name: "nonspell recapture resets numeric", setup: func(g *game.Game, _, obj *game.StackObject) {
			target := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(3)})
			target.Kind = game.StackActivatedAbility
			obj.Targets[1] = game.StackObjectTarget(target.ID)
			counterTargetStackObject(g, obj, 1, false, game.CounteredSpellGraveyard)
		}},
		{name: "unavailable recapture resets numeric", setup: func(g *game.Game, _, obj *game.StackObject) {
			target := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(3)})
			delete(g.CardInstances, target.SourceID)
			obj.Targets[1] = game.StackObjectTarget(target.ID)
			target.Copy = true
			counterTargetStackObject(g, obj, 1, false, game.CounteredSpellGraveyard)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			decoy := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(3)})
			target := spellWithManaValue(g, game.Player3, cost.Mana{cost.O(3)})
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.StackObjectTarget(decoy.ID), game.StackObjectTarget(target.ID)}}
			if !counterTargetStackObject(g, obj, 1, false, game.CounteredSpellGraveyard) {
				t.Fatal("counter failed")
			}
			if tt.setup != nil {
				tt.setup(g, target, obj)
			}
			condition := stackSelectionCondition(1, game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})})
			condition.Condition.Val.UseCounteredSpellManaValue = true
			if got := conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition.Condition); got != tt.want {
				t.Fatalf("past comparison=%t, want %t", got, tt.want)
			}
			if _, live := stackObjectByID(g, decoy.ID); !live {
				t.Fatal("decoy countered")
			}
		})
	}
}

func TestReassignedPastStackConditionCannotUseLiveComplement(t *testing.T) {
	t.Parallel()
	for _, value := range []int{2, 8} {
		for _, negate := range []bool{false, true} {
			t.Run(fmt.Sprintf("value=%d/negate=%t", value, negate), func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				target := spellWithManaValue(g, game.Player2, cost.Mana{cost.O(3)})
				obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.StackObjectTarget(target.ID)}}
				if !counterTargetStackObject(g, obj, 0, false, game.CounteredSpellGraveyard) {
					t.Fatal("counter failed")
				}
				replacement := spellWithManaValue(g, game.Player3, cost.Mana{cost.O(value)})
				obj.Targets[0] = game.StackObjectTarget(replacement.ID)
				condition := stackSelectionCondition(0, game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})})
				condition.Condition.Val.UseCounteredSpellManaValue = true
				condition.Condition.Val.Negate = negate
				if conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition.Condition) {
					t.Fatal("reassigned live target supplied information or a false complement for the captured spell")
				}
				NewEngine(nil).resolveInstructionSequence(g, obj, []game.Instruction{
					{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
						Condition: opt.Val(condition), PublishCondition: "past"},
					{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(1)},
						ConditionGate: "past", ConditionGateNegate: true},
				}, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				if g.Players[game.Player1].Life != 40 {
					t.Fatal("reassigned capture published a cached truth or false complement")
				}
			})
		}
	}
}

func TestPastStackConditionRetainsSpellCharacteristics(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name                      string
		cost                      cost.Mana
		x, want                   int
		copy, alternate, faceDown bool
	}{
		{"multiple X", cost.Mana{cost.X, cost.X, cost.U}, 3, 7, false, false, false},
		{"copy", cost.Mana{cost.X, cost.U}, 3, 4, true, false, false},
		{"alternate face", cost.Mana{cost.X, cost.R, cost.R}, 3, 5, false, true, false},
		{"face down known zero", cost.Mana{cost.X, cost.O(8)}, 3, 0, false, false, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			target := spellWithManaValue(g, game.Player2, tt.cost)
			target.XValue, target.Copy, target.FaceDown = tt.x, tt.copy, tt.faceDown
			if tt.alternate {
				def := g.CardInstances[target.SourceID].Def
				def.Alternate = opt.Val(def.CardFace)
				def.ManaCost = opt.Val(cost.Mana{cost.O(1)})
				target.Face = game.FaceAlternate
			}
			obj := &game.StackObject{Controller: game.Player1, Targets: []game.Target{game.StackObjectTarget(target.ID)}}
			if !counterTargetStackObject(g, obj, 0, false, game.CounteredSpellGraveyard) {
				t.Fatal("counter failed")
			}
			for _, value := range []int{tt.want, tt.want + 1} {
				condition := stackSelectionCondition(0, game.Selection{ManaValue: opt.Val(compare.Int{Op: compare.Equal, Value: value})})
				condition.Condition.Val.UseCounteredSpellManaValue = true
				if got := conditionSatisfied(g, conditionContext{controller: game.Player1, obj: obj}, condition.Condition); got != (value == tt.want) {
					t.Fatalf("value=%d got=%t", value, got)
				}
			}
		})
	}
}

func TestCompiledEventSpellManaValueIdentityAndDeparture(t *testing.T) {
	t.Parallel()
	def := compileUnlessCard(t, cardgen.ScryfallCard{Name: "Event Numeric", Layout: "normal", TypeLine: "Creature",
		Power: new("1"), Toughness: new("1"),
		OracleText: "Whenever you cast or copy an instant or sorcery spell, you gain 1 life. If that spell has mana value 5 or greater, you gain 2 life."})
	for _, removed := range []bool{false, true} {
		for _, value := range []int{4, 5, 6} {
			t.Run(fmt.Sprintf("removed=%t/value=%d", removed, value), func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				source := addCombatPermanent(g, game.Player1, def)
				spell := spellWithManaValue(g, game.Player1, cost.Mana{cost.O(value)})
				event := game.Event{Kind: game.EventSpellCopied, StackObjectID: spell.ID, Controller: game.Player1,
					ManaValue: stackObjectKnownManaValue(g, spell), CardID: spell.SourceID, ToZone: zone.Stack}
				if removed {
					g.Stack.RemoveByID(spell.ID)
					delete(g.CardInstances, spell.SourceID)
				}
				obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1,
					SourceID: source.ObjectID, SourceCardID: source.CardInstanceID, HasTriggerEvent: true, TriggerEvent: event}
				NewEngine(nil).resolveAbilityContentWithChoices(g, obj, def.TriggeredAbilities[0].Content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				want := 41
				if value >= 5 {
					want += 2
				}
				if g.Players[game.Player1].Life != want {
					t.Fatalf("life=%d, want %d; ability source has no spell value", g.Players[game.Player1].Life, want)
				}
			})
		}
	}
}

func TestCopiedEventManaValueUsesActualSpellProjection(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name                          string
		faceDown, copyValues, missing bool
		want                          int
		known                         bool
	}{
		{"multiple X", false, false, false, 7, true},
		{"copied characteristics", false, true, false, 4, true},
		{"face down", true, false, false, 0, true},
		{"unavailable", false, false, true, 0, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			spell := spellWithManaValue(g, game.Player1, cost.Mana{cost.X, cost.X, cost.U})
			spell.XValue = 3
			spell.Copy = true
			spell.FaceDown = tt.faceDown
			def := g.CardInstances[spell.SourceID].Def
			if tt.copyValues {
				spell.CopyValues = opt.Val(game.CopyableValues{ManaCost: opt.Val(cost.Mana{cost.X, cost.U})})
			}

			if tt.missing {
				delete(g.CardInstances, spell.SourceID)
			}
			emitSpellCopiedEvent(g, spell, def)
			event := g.Events[len(g.Events)-1]
			if event.ManaValue.Exists != tt.known || tt.known && event.ManaValue.Val != tt.want {
				t.Fatalf("event mana value=%#v, want known=%t value=%d", event.ManaValue, tt.known, tt.want)
			}
		})
	}
}
