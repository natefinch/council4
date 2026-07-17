package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const flamerushTokenLink = game.LinkedKey("flamerush-rider-token")

func flamerushRiderDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Flamerush Rider",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Warrior},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:  game.EventAttackerDeclared,
					Source: game.TriggerSourceSelf,
				},
			},
			Content: game.Mode{
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						ExcludeSource: true,
						CombatState:   game.CombatStateAttacking,
					}),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.CreateToken{
						Amount: game.Fixed(1),
						Source: game.TokenCopyOf(game.TokenCopySpec{
							Source: game.TokenCopySourceObject,
							Object: game.TargetPermanentReference(0),
						}),
						EntryTapped:        true,
						AttackSameAsObject: opt.Val(game.TargetPermanentReference(0)),
						PublishLinked:      flamerushTokenLink,
					}},
					{Primitive: game.CreateDelayedTrigger{Trigger: game.DelayedTriggerDef{
						Timing:              game.DelayedAtEndOfCombat,
						CapturedObjectGroup: opt.Val(game.LinkedObjectReference(string(flamerushTokenLink))),
						Content: game.Mode{Sequence: []game.Instruction{{
							Primitive: game.Exile{Group: game.CapturedObjectsGroup()},
						}}}.Ability(),
					}}},
				},
			}.Ability(),
		}},
	}}
}

func flamerushCopyTargetDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Copied Attacker",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dragon},
		Power:     opt.Val(game.PT{Value: 7}),
		Toughness: opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		TriggeredAbilities: []game.TriggeredAbility{{
			Trigger: game.TriggerCondition{
				Type: game.TriggerWhenever,
				Pattern: game.TriggerPattern{
					Event:  game.EventAttackerDeclared,
					Source: game.TriggerSourceSelf,
				},
			},
			Content: game.Mode{}.Ability(),
		}},
	}}
}

func TestFlamerushRiderCopiesTargetAgainstSameDefenderAndCapturesDoubledBatch(t *testing.T) {
	for _, defender := range []struct {
		name string
		make func(*game.Game) game.AttackTarget
	}{
		{name: "player", make: func(*game.Game) game.AttackTarget {
			return game.AttackTarget{Player: game.Player2}
		}},
		{name: "planeswalker", make: func(g *game.Game) game.AttackTarget {
			permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Defended Walker", Types: []types.Card{types.Planeswalker},
			}})
			return game.AttackTarget{Player: game.Player2, PlaneswalkerID: permanent.ObjectID}
		}},
		{name: "battle", make: func(g *game.Game) game.AttackTarget {
			permanent := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Defended Battle", Types: []types.Card{types.Battle},
			}})
			return game.AttackTarget{Player: game.Player2, BattleID: permanent.ObjectID}
		}},
	} {
		t.Run(defender.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			rider := addCombatPermanent(g, game.Player1, flamerushRiderDef())
			target := addCombatPermanent(g, game.Player1, flamerushCopyTargetDef())
			addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
			riderTarget := game.AttackTarget{Player: game.Player3}
			copyTarget := defender.make(g)
			g.Turn.ActivePlayer = game.Player1
			g.Turn.Phase = game.PhaseCombat
			g.Combat = &game.CombatState{
				Attackers: []game.AttackDeclaration{
					{Attacker: rider.ObjectID, Target: riderTarget},
					{Attacker: target.ObjectID, Target: copyTarget},
				},
				AttackersDeclared: true,
			}
			stackObject := flamerushTriggerObject(g, rider, target)
			rider.Controller = game.Player4
			if !movePermanentToZone(g, rider, zone.Graveyard) {
				t.Fatal("could not remove Rider before resolution")
			}
			g.Stack.Push(stackObject)
			engine.resolveTopOfStack(g, &TurnLog{})

			tokens := flamerushTokens(g)
			if len(tokens) != 2 {
				t.Fatalf("created %d tokens, want 2 after doubling", len(tokens))
			}
			for _, token := range tokens {
				if token.Controller != game.Player1 || !token.Tapped {
					t.Fatalf("token = %+v, want tapped under trigger controller", token)
				}
				face, ok := permanentFaceDef(g, token)
				if !ok || face.Name != "Copied Attacker" ||
					len(face.StaticAbilities) != 1 ||
					len(face.TriggeredAbilities) != 1 {
					t.Fatalf("token copy characteristics = %+v", face)
				}
				declaration, ok := attackerDeclarationFor(g, token.ObjectID)
				if !ok || declaration.Target != copyTarget {
					t.Fatalf("token attack = %+v, want target attack %+v", declaration, copyTarget)
				}
			}
			if len(g.DelayedTriggers) != 1 ||
				g.DelayedTriggers[0].Timing != game.DelayedAtEndOfCombat ||
				len(g.DelayedTriggers[0].CapturedObjectIDs) != 2 {
				t.Fatalf("delayed triggers = %+v", g.DelayedTriggers)
			}
		})
	}
}

func TestFlamerushRiderTargetMustStillBeAttackingAtResolution(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	rider := addCombatPermanent(g, game.Player1, flamerushRiderDef())
	target := addCombatPermanent(g, game.Player1, flamerushCopyTargetDef())
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: rider.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		},
		AttackersDeclared: true,
	}
	g.Stack.Push(flamerushTriggerObject(g, rider, target))
	g.Combat.Attackers = g.Combat.Attackers[:1]

	NewEngine(nil).resolveTopOfStack(g, &TurnLog{})

	if tokens := flamerushTokens(g); len(tokens) != 0 {
		t.Fatalf("created %d tokens for a nonattacking target", len(tokens))
	}
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("scheduled delayed trigger after target became illegal: %+v", g.DelayedTriggers)
	}
}

func TestFlamerushRiderDelayedExileKeepsTriggerBatchesIndependent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	rider := addCombatPermanent(g, game.Player1, flamerushRiderDef())
	target := addCombatPermanent(g, game.Player1, flamerushCopyTargetDef())
	addReplacementPermanent(t, g, game.Player1, tokenDoublingReplacementCardDef())
	g.Turn.ActivePlayer = game.Player1
	g.Turn.Phase = game.PhaseCombat
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: rider.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
			{Attacker: target.ObjectID, Target: game.AttackTarget{Player: game.Player3}},
		},
		AttackersDeclared: true,
	}
	for range 2 {
		g.Stack.Push(flamerushTriggerObject(g, rider, target))
		engine.resolveTopOfStack(g, &TurnLog{})
	}
	if len(g.DelayedTriggers) != 2 {
		t.Fatalf("delayed triggers = %d, want 2", len(g.DelayedTriggers))
	}
	first := append([]id.ID(nil), g.DelayedTriggers[0].CapturedObjectIDs...)
	second := append([]id.ID(nil), g.DelayedTriggers[1].CapturedObjectIDs...)
	if len(first) != 2 || len(second) != 2 {
		t.Fatalf("captured batches = %v and %v, want two doubled pairs", first, second)
	}

	fireDelayedTrigger(engine, g, g.DelayedTriggers[0])
	for _, objectID := range first {
		if _, ok := permanentByObjectID(g, objectID); ok {
			t.Fatalf("first-batch token %d survived exile", objectID)
		}
	}
	for _, objectID := range second {
		if _, ok := permanentByObjectID(g, objectID); !ok {
			t.Fatalf("second-batch token %d was cross-contaminated", objectID)
		}
	}
	fireDelayedTrigger(engine, g, g.DelayedTriggers[1])
	for _, objectID := range second {
		if _, ok := permanentByObjectID(g, objectID); ok {
			t.Fatalf("second-batch token %d survived exile", objectID)
		}
	}
}

func flamerushTriggerObject(g *game.Game, rider, target *game.Permanent) *game.StackObject {
	trigger := flamerushRiderDef().TriggeredAbilities[0]
	return &game.StackObject{
		ID:              g.IDGen.Next(),
		Kind:            game.StackTriggeredAbility,
		Controller:      game.Player1,
		SourceID:        rider.ObjectID,
		SourceCardID:    rider.CardInstanceID,
		InlineTrigger:   &trigger,
		Targets:         []game.Target{game.PermanentTarget(target.ObjectID)},
		TargetCounts:    []int{1},
		HasTriggerEvent: true,
		TriggerEvent: game.Event{
			Kind:           game.EventAttackerDeclared,
			SourceObjectID: rider.ObjectID,
			Player:         game.Player3,
			AttackTarget:   game.AttackTarget{Player: game.Player3},
		},
	}
}

func flamerushTokens(g *game.Game) []*game.Permanent {
	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens = append(tokens, permanent)
		}
	}
	return tokens
}
