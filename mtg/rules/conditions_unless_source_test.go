package rules

import (
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestUnlessSourceExclusionCompiledResolution(t *testing.T) {
	t.Parallel()
	for _, card := range []cardgen.ScryfallCard{
		{
			Name: "Fathom Fleet Boarder", Layout: "normal", ManaCost: "{2}{B}", Colors: []string{"B"},
			TypeLine:   "Creature — Orc Pirate",
			OracleText: "When this creature enters, you lose 2 life unless you control another Pirate.",
			Power:      new("3"), Toughness: new("3"),
		},
		{
			Name: "Reaver Drone", Layout: "normal", ManaCost: "{B}",
			TypeLine:   "Creature — Eldrazi Drone",
			OracleText: "Devoid (This card has no color.)\nAt the beginning of your upkeep, you lose 1 life unless you control another colorless creature.",
			Power:      new("2"), Toughness: new("1"),
		},
	} {
		t.Run(card.Name, func(t *testing.T) {
			t.Parallel()
			def := compileUnlessCard(t, card)
			sequence := def.TriggeredAbilities[0].Content.Modes[0].Sequence
			if len(sequence) != 1 || !sequence[0].Condition.Val.Condition.Val.Negate ||
				!sequence[0].Condition.Val.Condition.Val.ControlsMatching.Val.Selection.ExcludeSource {
				t.Fatal("compiled body must retain its negated source-excluding gate")
			}
			loss, ok := sequence[0].Primitive.(game.LoseLife)
			if !ok {
				t.Fatal("compiled body must lose life")
			}
			otherDef := resolvingSourceCreature("Other Pirate", 3)
			otherDef.Subtypes = []types.Sub{types.Pirate}
			nonmatchingDef := resolvingSourceCreature("White Human", 3)
			nonmatchingDef.Subtypes = []types.Sub{types.Human}
			nonmatchingDef.Colors = []color.Color{color.White}
			changeCharacteristics := func(g *game.Game, permanent *game.Permanent, qualifies bool) {
				subtypes, colors := []types.Sub{types.Human}, []color.Color{color.White}
				if qualifies {
					subtypes, colors = []types.Sub{types.Pirate}, nil
				}
				g.ContinuousEffects = append(g.ContinuousEffects,
					game.ContinuousEffect{Layer: game.LayerType, AffectedObjectID: permanent.ObjectID, SetSubtypes: subtypes},
					game.ContinuousEffect{Layer: game.LayerColor, AffectedObjectID: permanent.ObjectID, SetColors: colors, SetColorless: qualifies},
				)
			}
			for _, tt := range []struct {
				name    string
				prepare func(*testing.T, *game.Game, *game.StackObject, *game.Permanent)
				prevent bool
			}{
				{name: "source only"},
				{
					name: "another controlled object", prevent: true,
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
						addCombatPermanent(g, game.Player1, otherDef)
					},
				},
				{
					name: "same name is another object", prevent: true,
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
						addCombatPermanent(g, game.Player1, def)
					},
				},
				{
					name: "another token", prevent: true,
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
						g.Battlefield = append(g.Battlefield, &game.Permanent{
							ObjectID: g.IDGen.Next(), Owner: game.Player2, Controller: game.Player1,
							Token: true, TokenDef: otherDef,
						})
					},
				},
				{
					name: "token source only",
					prepare: func(_ *testing.T, g *game.Game, obj *game.StackObject, _ *game.Permanent) {
						source := &game.Permanent{
							ObjectID: g.IDGen.Next(), Owner: game.Player1, Controller: game.Player1,
							Token: true, TokenDef: def,
						}
						g.Battlefield = []*game.Permanent{source}
						obj.SourceID, obj.SourceCardID, obj.SourceTokenDef = source.ObjectID, 0, def
					},
				},
				{
					name: "other controller",
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
						addCombatPermanent(g, game.Player2, otherDef)
					},
				},
				{
					name: "source controller changed",
					prepare: func(_ *testing.T, _ *game.Game, _ *game.StackObject, source *game.Permanent) {
						source.Controller = game.Player2
					},
				},
				{
					name: "captured controller still controls another", prevent: true,
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
						source.Controller = game.Player2
						other := addCombatPermanent(g, game.Player1, otherDef)
						other.Owner = game.Player2
					},
				},
				{
					name: "source current controller controls another",
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
						source.Controller = game.Player2
						addCombatPermanent(g, game.Player2, otherDef)
					},
				},
				{
					name: "nonmatching other",
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
						addCombatPermanent(g, game.Player1, nonmatchingDef)
					},
				},
				{
					name: "other loses matching characteristics",
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
						changeCharacteristics(g, addCombatPermanent(g, game.Player1, otherDef), false)
					},
				},
				{
					name: "other gains matching characteristics", prevent: true,
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
						changeCharacteristics(g, addCombatPermanent(g, game.Player1, nonmatchingDef), true)
					},
				},
				{
					name: "artifact becomes matching creature", prevent: true,
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
						other := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
							Name: "Animated artifact", Types: []types.Card{types.Artifact},
						}})
						g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
							Layer: game.LayerType, AffectedObjectID: other.ObjectID, SetTypes: []types.Card{types.Creature},
						})
						changeCharacteristics(g, other, true)
					},
				},
				{
					name: "source loses matching characteristics",
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
						changeCharacteristics(g, source, false)
					},
				},
				{
					name: "source gains matching characteristics but stays excluded",
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
						changeCharacteristics(g, source, false)
						changeCharacteristics(g, source, true)
					},
				},
				{
					name: "source becomes artifact but stays excluded",
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
						g.ContinuousEffects = append(g.ContinuousEffects, game.ContinuousEffect{
							Layer: game.LayerType, AffectedObjectID: source.ObjectID, SetTypes: []types.Card{types.Artifact},
						})
					},
				},
				{
					name: "departed source only",
					prepare: func(t *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
						if !movePermanentToZone(g, source, zone.Graveyard) {
							t.Fatal("source did not depart")
						}
					},
				},
				{
					name: "phased source with another", prevent: true,
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
						source.PhasedOut = true
						addCombatPermanent(g, game.Player1, otherDef)
					},
				},
				{
					name: "departed source with another", prevent: true,
					prepare: func(t *testing.T, g *game.Game, _ *game.StackObject, source *game.Permanent) {
						if !movePermanentToZone(g, source, zone.Graveyard) {
							t.Fatal("source did not depart")
						}
						addCombatPermanent(g, game.Player1, otherDef)
					},
				},
				{
					name: "phased source only",
					prepare: func(_ *testing.T, _ *game.Game, _ *game.StackObject, source *game.Permanent) {
						source.PhasedOut = true
					},
				},
				{
					name: "phased other",
					prepare: func(_ *testing.T, g *game.Game, _ *game.StackObject, _ *game.Permanent) {
						addCombatPermanent(g, game.Player1, otherDef).PhasedOut = true
					},
				},
				{
					name: "blink replacement is another object", prevent: true,
					prepare: func(t *testing.T, g *game.Game, obj *game.StackObject, source *game.Permanent) {
						engine := NewEngine(nil)
						for _, instruction := range selfBlinkInstructions() {
							engine.resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
						}
						if len(g.Battlefield) != 1 || g.Battlefield[0].ObjectID == source.ObjectID ||
							g.Battlefield[0].CardInstanceID != source.CardInstanceID {
							t.Fatal("blink must replace the object, not the card")
						}
					},
				},
				{
					name: "ordinary event subject is not source", prevent: true,
					prepare: func(_ *testing.T, g *game.Game, obj *game.StackObject, _ *game.Permanent) {
						other := addCombatPermanent(g, game.Player1, otherDef)
						obj.TriggerEvent = game.Event{
							Kind: game.EventPermanentEnteredBattlefield, PermanentID: other.ObjectID, Controller: game.Player1,
						}
					},
				},
				{
					name: "unrelated target slots do not identify source",
					prepare: func(_ *testing.T, g *game.Game, obj *game.StackObject, _ *game.Permanent) {
						first := addCombatPermanent(g, game.Player2, otherDef)
						second := addCombatPermanent(g, game.Player2, otherDef)
						obj.Targets = []game.Target{game.PermanentTarget(first.ObjectID), game.PermanentTarget(second.ObjectID)}
					},
				},
				{
					name: "card identity is not permanent identity", prevent: true,
					prepare: func(_ *testing.T, _ *game.Game, obj *game.StackObject, source *game.Permanent) {
						obj.SourceID = source.CardInstanceID
						obj.SourceZone = zone.Hand
					},
				},
				{
					name: "source identity unavailable counts all candidates", prevent: true,
					prepare: func(_ *testing.T, _ *game.Game, obj *game.StackObject, _ *game.Permanent) {
						obj.SourceID = 0
					},
				},
			} {
				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()
					g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
					source := addCombatPermanent(g, game.Player1, def)
					obj := &game.StackObject{
						ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
						SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
					}
					if tt.prepare != nil {
						tt.prepare(t, g, obj, source)
					}
					before := g.Players[game.Player1].Life
					opponentLife := g.Players[game.Player2].Life
					g.Stack.Push(obj)
					log := &TurnLog{}
					NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, log)
					want := before
					if !tt.prevent {
						want -= loss.Amount.Value()
					}
					if got := g.Players[game.Player1].Life; got != want {
						t.Fatalf("resolving Unless life=%d, want %d", got, want)
					}
					if g.Players[game.Player2].Life != opponentLife || len(log.Resolves) != 1 || log.Resolves[0].Result != "resolved" {
						t.Fatal("ability must resolve for its captured controller")
					}
				})
			}
		})
	}
}

func TestUnlessSourceExclusionNestedSelections(t *testing.T) {
	t.Parallel()
	excluding := game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true}
	for _, tt := range []struct {
		name      string
		selection game.Selection
		wantLoss  int
	}{
		{"nested excluding", game.Selection{AnyOf: []game.Selection{{AnyOf: []game.Selection{excluding}}}}, 2},
		{"mixed alternatives", game.Selection{AnyOf: []game.Selection{
			excluding, {RequiredTypes: []types.Card{types.Creature}},
		}}, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, resolvingSourceCreature("Source", 2))
			obj := &game.StackObject{Kind: game.StackTriggeredAbility, Controller: game.Player1, SourceID: source.ObjectID}
			instruction := game.Instruction{
				Primitive: game.LoseLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
				Condition: resolvingSourceGate(game.Condition{
					Negate: true, ControlsMatching: opt.Val(game.SelectionCount{Selection: tt.selection}),
				}),
			}
			before := g.Players[game.Player1].Life
			NewEngine(nil).resolveInstructionWithChoices(g, obj, &instruction, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got := before - g.Players[game.Player1].Life; got != tt.wantLoss {
				t.Fatalf("negated nested selection lost %d life, want %d", got, tt.wantLoss)
			}
		})
	}
}

func TestUnlessSourceExclusionCompiledScopeTiming(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, text string
		wantSecond int
	}{
		{
			name: "expanded clause evaluates once", wantSecond: 1,
			text: "Unless you control another creature with power 2 or greater, put a +1/+1 counter on each of up to two target creatures you control.",
		},
		{
			name: "independent clauses see mutation",
			text: "Put a +1/+1 counter on target creature you control unless you control another creature with power 2 or greater. Put a +1/+1 counter on target creature you control unless you control another creature with power 2 or greater.",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			def := compileUnlessCard(t, cardgen.ScryfallCard{
				Name: "Unless Source Scope", Layout: "normal", TypeLine: "Creature — Pirate",
				Power: new("2"), Toughness: new("2"),
				OracleText: "When this creature enters, tap target artifact. " + tt.text,
			})
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			source := addCombatPermanent(g, game.Player1, def)
			artifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
				Name: "Unrelated target", Types: []types.Card{types.Artifact},
			}})
			first := addCombatPermanent(g, game.Player1, resolvingSourceCreature("First", 1))
			second := addCombatPermanent(g, game.Player1, resolvingSourceCreature("Second", 1))
			obj := &game.StackObject{
				ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, Controller: game.Player1,
				SourceID: source.ObjectID, SourceCardID: source.CardInstanceID,
				Targets: []game.Target{
					game.PermanentTarget(artifact.ObjectID),
					game.PermanentTarget(first.ObjectID),
					game.PermanentTarget(second.ObjectID),
				},
				TargetCounts: []int{1, 1, 1},
			}
			g.Stack.Push(obj)
			NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if !artifact.Tapped || artifact.Counters.Get(counter.PlusOnePlusOne) != 0 ||
				source.Counters.Get(counter.PlusOnePlusOne) != 0 {
				t.Fatal("unrelated target and resolving source must remain isolated")
			}
			if first.Counters.Get(counter.PlusOnePlusOne) != 1 || second.Counters.Get(counter.PlusOnePlusOne) != tt.wantSecond {
				t.Fatalf("counters=%d/%d, want 1/%d", first.Counters.Get(counter.PlusOnePlusOne),
					second.Counters.Get(counter.PlusOnePlusOne), tt.wantSecond)
			}
		})
	}
}
