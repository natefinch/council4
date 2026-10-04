package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func resolvingSourceCreature(name string, power int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: power}),
		Toughness: opt.Val(game.PT{Value: power}),
	}}
}

func resolvingSourceGate(condition game.Condition) opt.V[game.EffectCondition] {
	return opt.Val(game.EffectCondition{Condition: opt.Val(condition)})
}

func TestEffectConditionSourceExclusion(t *testing.T) {
	t.Parallel()
	selection := game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		ExcludeSource: true,
		Power:         opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
	}
	gate := resolvingSourceGate(game.Condition{
		ControlsMatching: opt.Val(game.SelectionCount{Selection: selection, MinCount: 1}),
	})
	tests := []struct {
		name    string
		prepare func(*testing.T, *game.Game, *game.StackObject, *game.Permanent)
		want    bool
	}{
		{name: "source only"},
		{
			name: "another qualifying creature",
			prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
				addCombatPermanent(g, game.Player1, resolvingSourceCreature("Other", 4))
			},
			want: true,
		},
		{
			name: "wrong controller",
			prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
				addCombatPermanent(g, game.Player2, resolvingSourceCreature("Opponent", 4))
			},
		},
		{
			name: "wrong type",
			prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
				addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
					Name: "Artifact", Types: []types.Card{types.Artifact},
				}})
			},
		},
		{
			name: "wrong power",
			prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
				addCombatPermanent(g, game.Player1, resolvingSourceCreature("Small", 1))
			},
		},
		{
			name: "token source only",
			prepare: func(_ *testing.T, g *game.Game, obj *game.StackObject, _ *game.Permanent) {
				g.Battlefield = []*game.Permanent{{
					ObjectID: g.IDGen.Next(), Owner: game.Player1, Controller: game.Player1,
					Token: true, TokenDef: resolvingSourceCreature("Token Source", 5),
				}}
				obj.SourceID = g.Battlefield[0].ObjectID
				obj.SourceCardID = 0
			},
		},
		{
			name: "source control changed",
			prepare: func(_ *testing.T, _ *game.Game, _ *game.StackObject, source *game.Permanent) {
				source.Controller = game.Player2
			},
		},
		{
			name: "captured controller still controls another creature",
			prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
				source.Controller = game.Player2
				addCombatPermanent(g, game.Player1, resolvingSourceCreature("Other", 4))
			},
			want: true,
		},
		{
			name: "departed source only",
			prepare: func(t *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
				if !movePermanentToZone(g, source, zone.Graveyard) {
					t.Fatal("source did not leave the battlefield")
				}
			},
		},
		{
			name: "departed source with another creature",
			prepare: func(t *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
				if !movePermanentToZone(g, source, zone.Graveyard) {
					t.Fatal("source did not leave the battlefield")
				}
				addCombatPermanent(g, game.Player1, resolvingSourceCreature("Other", 4))
			},
			want: true,
		},
		{
			name: "phased source only",
			prepare: func(_ *testing.T, _ *game.Game, _ *game.StackObject, source *game.Permanent) {
				source.PhasedOut = true
			},
		},
		{
			name: "phased source with another creature",
			prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
				source.PhasedOut = true
				addCombatPermanent(g, game.Player1, resolvingSourceCreature("Other", 4))
			},
			want: true,
		},
		{
			name: "same card returned as a new object",
			prepare: func(t *testing.T, g *game.Game, obj *game.StackObject, source *game.Permanent) {
				engine := NewEngine(nil)
				for _, instruction := range selfBlinkInstructions() {
					engine.resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				}
				if len(g.Battlefield) != 1 || g.Battlefield[0].ObjectID == source.ObjectID ||
					g.Battlefield[0].CardInstanceID != source.CardInstanceID {
					t.Fatal("blink did not preserve the card and replace the object identity")
				}
			},
			want: true,
		},
		{
			name: "source identity unavailable",
			prepare: func(_ *testing.T, _ *game.Game, obj *game.StackObject, _ *game.Permanent) {
				obj.SourceID = 0
				obj.SourceCardID = 0
			},
			want: true,
		},
	}
	for _, kind := range []game.StackObjectKind{game.StackTriggeredAbility, game.StackActivatedAbility} {
		for _, tt := range tests {
			t.Run(fmtStackKind(kind)+"/"+tt.name, func(t *testing.T) {
				t.Parallel()
				g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
				source := addCombatPermanent(g, game.Player1, resolvingSourceCreature("Source", 5))
				obj := &game.StackObject{
					Kind: kind, SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
					Controller: game.Player1,
				}
				if tt.prepare != nil {
					tt.prepare(t, g, obj, source)
				}
				if got := effectConditionSatisfied(g, obj, gate); got != tt.want {
					t.Errorf("source-excluding gate = %v, want %v", got, tt.want)
				}
				before := g.Players[game.Player1].Life
				instruction := game.Instruction{
					Condition: gate,
					Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
				}
				NewEngine(nil).resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
				wantLife := before
				if tt.want {
					wantLife += 2
				}
				if got := g.Players[game.Player1].Life; got != wantLife {
					t.Errorf("gated resolution life = %d, want %d", got, wantLife)
				}
			})
		}
	}
}

func fmtStackKind(kind game.StackObjectKind) string {
	if kind == game.StackSpell {
		return "spell"
	}
	if kind == game.StackTriggeredAbility {
		return "triggered"
	}
	return "activated"
}

func TestEffectConditionSourceExclusionCountScopes(t *testing.T) {
	t.Parallel()
	selection := game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true}
	count := game.SelectionCount{Selection: selection, MinCount: 1}
	tests := []struct {
		name       string
		controller game.PlayerID
		condition  game.Condition
		other      bool
		want       bool
	}{
		{
			name: "any opponent excludes stolen source", controller: game.Player2,
			condition: game.Condition{AnyOpponentControls: opt.Val(count)},
		},
		{
			name: "opponents exclude stolen source", controller: game.Player2,
			condition: game.Condition{OpponentsControl: opt.Val(count)},
		},
		{
			name: "any opponent counts another creature", controller: game.Player2, other: true, want: true,
			condition: game.Condition{AnyOpponentControls: opt.Val(count)},
		},
		{
			name: "opponents count another creature", controller: game.Player2, other: true, want: true,
			condition: game.Condition{OpponentsControl: opt.Val(count)},
		},
		{
			name: "total power excludes source", controller: game.Player1,
			condition: game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
				Selection: selection, MinCount: 1,
				TotalPower: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4}),
			})},
		},
		{
			name: "distinct names exclude source", controller: game.Player1,
			condition: game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
				Selection: selection, MinCount: 1,
				DistinctNames: opt.Val(compare.Int{Op: compare.Equal, Value: 1}),
			})},
		},
		{
			name: "cross-player counts exclude your source", controller: game.Player1, other: true, want: true,
			condition: game.Condition{ControlComparison: opt.Val(game.ControlCountComparison{
				Selection: selection, Left: game.ControlPlayerAnyOpponent,
				Right: game.ControlPlayerController, Op: compare.GreaterThan,
			})},
		},
		{
			name: "cross-player counts exclude stolen source", controller: game.Player2,
			condition: game.Condition{ControlComparison: opt.Val(game.ControlCountComparison{
				Selection: selection, Left: game.ControlPlayerAnyOpponent,
				Right: game.ControlPlayerController, Op: compare.GreaterThan,
			})},
		},
		{
			name: "negated control gate", controller: game.Player1, want: true,
			condition: game.Condition{ControlsMatching: opt.Val(count), Negate: true},
		},
		{
			name: "nonexcluding control gate unchanged", controller: game.Player1, want: true,
			condition: game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
				Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}}, MinCount: 1,
			})},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, resolvingSourceCreature("Source", 5))
			source.Controller = tt.controller
			obj := &game.StackObject{
				Kind: game.StackTriggeredAbility, SourceID: source.ObjectID,
				SourceCardID: source.CardInstanceID, Controller: game.Player1,
			}
			if tt.other {
				addCombatPermanent(g, game.Player2, resolvingSourceCreature("Other", 4))
			}
			if got := effectConditionSatisfied(g, obj, resolvingSourceGate(tt.condition)); got != tt.want {
				t.Errorf("source-excluding count gate = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEffectConditionSourceCardIdentityIsNotPermanentIdentity(t *testing.T) {
	t.Parallel()
	for _, kind := range []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility} {
		t.Run(fmtStackKind(kind), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			permanent := addCombatPermanent(g, game.Player1, resolvingSourceCreature("New incarnation", 4))
			obj := &game.StackObject{
				Kind: kind, SourceID: permanent.CardInstanceID, SourceCardID: permanent.CardInstanceID,
				SourceZone: zone.Hand, Controller: game.Player1,
			}
			gate := resolvingSourceGate(game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
				Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true},
				MinCount:  1,
			})})
			if !effectConditionSatisfied(g, obj, gate) {
				t.Fatal("spell/hand source card identity excluded a distinct battlefield object")
			}
		})
	}
}

func TestEffectConditionSourceReferenceState(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		prepare func(*testing.T, *game.Game, *game.StackObject, *game.Permanent)
		want    bool
	}{
		{name: "live", want: true},
		{
			name: "departed uses existing LKI", want: true,
			prepare: func(t *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
				if !movePermanentToZone(g, source, zone.Graveyard) {
					t.Fatal("source did not leave the battlefield")
				}
			},
		},
		{
			name: "phased uses existing snapshot", want: true,
			prepare: func(_ *testing.T, _ *game.Game, _ *game.StackObject, source *game.Permanent) {
				source.PhasedOut = true
			},
		},
		{
			name: "missing source still fails",
			prepare: func(_ *testing.T, g *game.Game, obj *game.StackObject, _ *game.Permanent) {
				obj.SourceID = g.IDGen.Next()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, resolvingSourceCreature("Source", 5))
			obj := &game.StackObject{
				Kind: game.StackTriggeredAbility, SourceID: source.ObjectID,
				SourceCardID: source.CardInstanceID, Controller: game.Player1,
			}
			if tt.prepare != nil {
				tt.prepare(t, g, obj, source)
			}
			gate := resolvingSourceGate(game.Condition{
				Object: opt.Val(game.SourcePermanentReference()),
				ObjectMatches: opt.Val(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Power:         opt.Val(compare.Int{Op: compare.Equal, Value: 5}),
				}),
			})
			if got := effectConditionSatisfied(g, obj, gate); got != tt.want {
				t.Errorf("source-characteristic reference gate = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEffectConditionSourceExclusionSelectionConsumers(t *testing.T) {
	t.Parallel()
	selection := game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true}
	tests := []struct {
		name      string
		condition game.Condition
		want      bool
	}{
		{
			name: "source object selection",
			condition: game.Condition{
				Object: opt.Val(game.SourcePermanentReference()), ObjectMatches: opt.Val(selection),
			},
		},
		{
			name: "greatest mana value group",
			condition: game.Condition{
				ControlsGreatestManaValueInGroup: opt.Val(selection),
			},
		},
		{
			name: "source exclusion inside alternatives",
			condition: game.Condition{
				Object: opt.Val(game.SourcePermanentReference()),
				ObjectMatches: opt.Val(game.Selection{
					AnyOf: []game.Selection{selection},
				}),
			},
		},
		{
			name: "source exclusion inside nested alternatives",
			condition: game.Condition{
				Object: opt.Val(game.SourcePermanentReference()),
				ObjectMatches: opt.Val(game.Selection{
					AnyOf: []game.Selection{{AnyOf: []game.Selection{selection}}},
				}),
			},
		},
		{
			name: "nonexcluding alternative still matches",
			condition: game.Condition{
				Object: opt.Val(game.SourcePermanentReference()),
				ObjectMatches: opt.Val(game.Selection{
					AnyOf: []game.Selection{selection, {RequiredTypes: []types.Card{types.Creature}}},
				}),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			for _, state := range []string{"live", "departed", "phased"} {
				t.Run(state, func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					source := addCombatPermanent(g, game.Player1, resolvingSourceCreature("Source", 5))
					obj := &game.StackObject{
						Kind: game.StackTriggeredAbility, SourceID: source.ObjectID,
						SourceCardID: source.CardInstanceID, Controller: game.Player1,
					}
					if state == "departed" && !movePermanentToZone(g, source, zone.Graveyard) {
						t.Fatal("source did not leave the battlefield")
					}
					if state == "phased" {
						source.PhasedOut = true
					}
					gate := resolvingSourceGate(tt.condition)
					if got := effectConditionSatisfied(g, obj, gate); got != tt.want {
						t.Errorf("source-object selection gate = %v, want %v", got, tt.want)
					}
					before := g.Players[game.Player1].Life
					instruction := game.Instruction{
						Condition: gate,
						Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
					}
					NewEngine(nil).resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
					wantLife := before
					if tt.want {
						wantLife += 2
					}
					if got := g.Players[game.Player1].Life; got != wantLife {
						t.Errorf("source-object instruction life = %d, want %d", got, wantLife)
					}
				})
			}
		})
	}
}

func TestConditionalDestinationPlaceSourceExclusion(t *testing.T) {
	t.Parallel()
	for _, other := range []bool{false, true} {
		name := "source only"
		if other {
			name = "another creature"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, resolvingSourceCreature("Source", 5))
			obj := &game.StackObject{
				Kind: game.StackActivatedAbility, SourceID: source.ObjectID,
				SourceCardID: source.CardInstanceID, Controller: game.Player1,
			}
			if other {
				addCombatPermanent(g, game.Player1, resolvingSourceCreature("Other", 4))
			}
			cardID := addCardToLibrary(g, game.Player1, resolvingSourceCreature("Looked-at card", 1))
			link := game.CardReference{Kind: game.CardReferenceLinked, LinkID: "source-gate-card"}
			rememberLinkedObject(g, linkedObjectSourceKey(g, obj, link.LinkID), game.LinkedObjectRef{CardID: cardID})
			primitive := game.ConditionalDestinationPlace{
				Card: link, FromZone: zone.Library, Then: zone.Hand, ThenMandatory: true, Else: zone.Graveyard,
				Condition: resolvingSourceGate(game.Condition{ControlsMatching: opt.Val(game.SelectionCount{
					Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true},
					MinCount:  1,
				})}),
			}
			resolveInstruction(NewEngine(nil), g, obj, primitive, &TurnLog{})
			if g.Players[game.Player1].Hand.Contains(cardID) != other ||
				g.Players[game.Player1].Graveyard.Contains(cardID) == other ||
				g.Players[game.Player1].Library.Contains(cardID) {
				t.Fatal("conditional destination did not route the card according to the source-excluding gate")
			}
		})
	}
}

func TestConditionSnapshotSourceIdentityDomain(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		subject selectionSubject
		want    bool
	}{
		{
			name: "live source", want: true,
			subject: selectionSubject{
				kind: subjectPermanent, permanent: &game.Permanent{ObjectID: 17}, sourceObjectID: 17,
			},
		},
		{
			name: "source snapshot", want: true,
			subject: selectionSubject{
				kind: subjectEventPermanent, snapshotObjectID: 17, sourceObjectID: 17,
			},
		},
		{
			name: "other snapshot",
			subject: selectionSubject{
				kind: subjectEventPermanent, snapshotObjectID: 18, sourceObjectID: 17,
			},
		},
		{
			name: "ordinary event matching is unchanged",
			subject: selectionSubject{
				kind: subjectEventPermanent, event: game.Event{PermanentID: 17}, sourceObjectID: 17,
			},
		},
		{
			name: "missing source identity",
			subject: selectionSubject{
				kind: subjectEventPermanent, snapshotObjectID: 17,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.subject.isSource(); got != tt.want {
				t.Errorf("source identity = %v, want %v", got, tt.want)
			}
		})
	}
}
