package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDrawEffectDrawsRequestedCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Draw{
		Amount:      game.Fixed(2),
		TargetIndex: game.TargetIndexController,
	}, nil)
	firstDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	secondDraw := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if !g.Players[game.Player1].Hand.Contains(firstDraw) {
		t.Fatal("first card was not drawn")
	}
	if !g.Players[game.Player1].Hand.Contains(secondDraw) {
		t.Fatal("second card was not drawn")
	}
	if len(log.Draws) != 2 {
		t.Fatalf("draw logs = %d, want 2", len(log.Draws))
	}
	if log.Resolves[0].SourceID != sourceID {
		t.Fatalf("resolve source = %v, want %v", log.Resolves[0].SourceID, sourceID)
	}
}

func TestGainLifeEffectIncreasesTargetLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.GainLife{
		Amount:      game.Fixed(3),
		TargetIndex: 0,
	}, []game.Target{game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Life != 43 {
		t.Fatalf("player 2 life = %d, want 43", g.Players[game.Player2].Life)
	}
}

func TestCantGainLifeRuleEffectStopsLifeGainAndLifelink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "No Lifegain",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbilityBody{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCantGainLife,
				AffectedPlayer: game.PlayerAny,
			}},
		}}},
	})
	addEffectSpellToStack(g, game.Player1, game.GainLife{
		Amount:      game.Fixed(3),
		TargetIndex: game.TargetIndexController,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("life = %d, want gain prevented", got)
	}

	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2, game.Lifelink)
	g.Combat = &game.CombatState{Attackers: []game.AttackDeclaration{{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}}}}
	engine.resolveCombatDamage(g, &TurnLog{})
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("life after lifelink = %d, want gain prevented", got)
	}
}

func TestDynamicAmountUsesControllerHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	addEffectSpellToStack(g, game.Player1, game.GainLife{
		Amount:      game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountControllerHandSize}),
		TargetIndex: game.TargetIndexController,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("life = %d, want 42", got)
	}
}

func TestDynamicAmountUsesXValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountX}),
		Recipient: game.TargetRecipient(0),
	}, []game.Target{game.PlayerTarget(game.Player2)})
	obj, ok := g.Stack.Peek()
	if !ok {
		t.Fatal("stack is empty")
	}
	obj.XValue = 4

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 36 {
		t.Fatalf("target life = %d, want 36", got)
	}
}

func TestDynamicAmountUsesTargetPower(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 5)
	addEffectSpellToStack(g, game.Player1, game.Damage{
		Amount:    game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountTargetPower, TargetIndex: 0}),
		Recipient: game.TargetRecipient(0),
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := target.MarkedDamage; got != 5 {
		t.Fatalf("marked damage = %d, want 5", got)
	}
}

func TestDynamicAmountCanUsePreviousInstructionResult(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Linked Amount Spell",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.SpellAbilityBody{
				Content: game.PlainAbilityContent{
					Sequence: []game.Instruction{
						{Primitive: game.GainLife{TargetIndex: game.TargetIndexController, Amount: game.Fixed(3)}, PublishResult: "that-much"},
						{Primitive: game.LoseLife{
							TargetIndex: 0,
							Amount:      game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "that-much"}),
						}},
					},
				},
			})},
		},
		Owner: game.Player1,
	}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: game.Player1,
		Targets:    []game.Target{game.PlayerTarget(game.Player2)},
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 43 {
		t.Fatalf("controller life = %d, want 43", got)
	}
	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("target life = %d, want 37", got)
	}
}

func TestOptionalEffectCanBeAcceptedOrDeclined(t *testing.T) {
	t.Run("accepted by fallback", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
			Primitive: game.Draw{
				TargetIndex: game.TargetIndexController,
				Amount:      game.Fixed(1),
			},
			Optional: true,
		}}, nil)
		log := TurnLog{}

		engine.resolveTopOfStack(g, &log)

		if got := g.Players[game.Player1].Hand.Size(); got != 1 {
			t.Fatalf("hand size = %d, want optional draw accepted", got)
		}
		if len(log.Choices) != 1 || log.Choices[0].Selected[0] != 1 {
			t.Fatalf("choices = %+v, want accepted optional effect", log.Choices)
		}
	})
	t.Run("declined by agent", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
			Primitive: game.Draw{
				TargetIndex: game.TargetIndexController,
				Amount:      game.Fixed(1),
			},
			Optional: true,
		}}, nil)
		agents := [game.NumPlayers]PlayerAgent{
			game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
		}
		log := TurnLog{}

		engine.resolveTopOfStackWithChoices(g, agents, &log)

		if got := g.Players[game.Player1].Hand.Size(); got != 0 {
			t.Fatalf("hand size = %d, want optional draw declined", got)
		}
		if len(log.Choices) != 1 || log.Choices[0].Selected[0] != 0 {
			t.Fatalf("choices = %+v, want declined optional effect", log.Choices)
		}
	})
}

func TestInstructionResultGateBranchesOnIfYouDoAndDont(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	sourceID := addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.Draw{TargetIndex: game.TargetIndexController, Amount: game.Fixed(1)}, Optional: true, PublishResult: "choice"},
		{
			Primitive:  game.GainLife{TargetIndex: game.TargetIndexController, Amount: game.Fixed(3)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "choice", Accepted: game.TriTrue, Succeeded: game.TriTrue}),
		},
		{
			Primitive:  game.LoseLife{TargetIndex: game.TargetIndexController, Amount: game.Fixed(3)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "choice", Accepted: game.TriFalse}),
		},
	})
	if sourceID == 0 {
		t.Fatal("missing source id")
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 43 {
		t.Fatalf("life = %d, want if-you-do branch only", got)
	}
}

func TestInstructionResultGateRequiresActualSuccess(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.Draw{TargetIndex: game.TargetIndexController, Amount: game.Fixed(1)}, PublishResult: "draw"},
		{
			Primitive:  game.GainLife{TargetIndex: game.TargetIndexController, Amount: game.Fixed(3)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "draw", Succeeded: game.TriTrue}),
		},
		{
			Primitive:  game.LoseLife{TargetIndex: game.TargetIndexController, Amount: game.Fixed(2)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "draw", Succeeded: game.TriFalse}),
		},
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 38 {
		t.Fatalf("life = %d, want failed-draw branch", got)
	}
}

func TestDeclinedOptionalEffectDoesNotPublishPreviousAmount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, []game.Instruction{
		{
			Primitive:     game.GainLife{TargetIndex: game.TargetIndexController, Amount: game.Fixed(5)},
			Optional:      true,
			PublishResult: "amount",
		},
		{
			Primitive: game.LoseLife{
				TargetIndex: game.TargetIndexController,
				Amount:      game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "amount"}),
			},
		},
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("life = %d, want declined linked amount to be unavailable", got)
	}
}

func TestResolutionChoiceCanChooseManaColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addInstructionSpellToStack(g, []game.Instruction{
		{
			Primitive: game.Choose{
				Choice:        game.ResolutionChoice{Kind: game.ResolutionChoiceMana},
				PublishChoice: "chosen-color",
			},
		},
		{
			Primitive: game.AddMana{
				Amount:     game.Fixed(1),
				ChoiceFrom: "chosen-color",
			},
		},
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{3}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Total(); got != 1 {
		t.Fatalf("mana pool total = %d, want one chosen mana", got)
	}
	if !g.Players[game.Player1].ManaPool.Spend(mana.R, 1) {
		t.Fatal("chosen mana was not red")
	}
}

func TestCommanderIdentityColorChoiceFeedsManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{
		game.Player1: {
			Commander: &game.CardDef{CardFace: game.CardFace{Name: "Dimir Commander",
				Types: []types.Card{types.Creature}}, ColorIdentity: color.NewIdentity(color.Blue, color.Black),
			},
		},
	})
	setSorcerySpeedTurn(g, game.Player1)
	tower := addCombatPermanent(g, game.Player1, commandTowerLikeLand())
	engine := NewEngine(nil)
	log := &TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	if !engine.applyActionWithChoices(g, game.Player1, action.ActivateAbility(tower.ObjectID, 0, nil, 0), agents, log) {
		t.Fatal("activate Command Tower-like ability failed")
	}

	if !tower.Tapped {
		t.Fatal("Command Tower-like permanent was not tapped for its mana ability")
	}
	if len(log.Choices) != 1 {
		t.Fatalf("choices = %+v, want one commander-color choice", log.Choices)
	}
	choice := log.Choices[0]
	if choice.Request.Kind != game.ChoiceResolution || len(choice.Request.Options) != 2 || choice.Request.Options[0].Label != "U" || choice.Request.Options[1].Label != "B" {
		t.Fatalf("choice request = %+v, want only U/B commander identity options", choice.Request)
	}
	if len(choice.Selected) != 1 || choice.Selected[0] != 1 {
		t.Fatalf("selected = %+v, want black option", choice.Selected)
	}
	if got := g.Players[game.Player1].ManaPool.Total(); got != 1 {
		t.Fatalf("mana pool total = %d, want one chosen mana", got)
	}
	if !g.Players[game.Player1].ManaPool.Spend(mana.B, 1) {
		t.Fatal("chosen mana was not black")
	}
}

func TestCommanderIdentityColorChoiceUnavailableWithoutColors(t *testing.T) {
	tests := []struct {
		name    string
		configs [game.NumPlayers]game.PlayerConfig
	}{
		{name: "no commander"},
		{
			name: "colorless commander",
			configs: [game.NumPlayers]game.PlayerConfig{
				game.Player1: {
					Commander: &game.CardDef{CardFace: game.CardFace{Name: "Colorless Commander",
						Types: []types.Card{types.Creature}}, ColorIdentity: color.NewIdentity(),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame(tt.configs)
			setSorcerySpeedTurn(g, game.Player1)
			tower := addCombatPermanent(g, game.Player1, commandTowerLikeLand())
			card, ok := permanentCardDef(g, tower)
			if !ok {
				t.Fatal("permanent card definition not found")
			}
			if len(card.ManaAbilities) == 0 {
				t.Fatal("no abilities on card")
			}

			if canActivateManaAbility(g, game.Player1, tower, &card.ManaAbilities[0], 0) {
				t.Fatal("canActivateManaAbility() = true, want false without commander color options")
			}
			if got := NewEngine(nil).legalActivateAbilityActions(g, game.Player1); len(got) != 0 {
				t.Fatalf("legal activation actions = %+v, want none", got)
			}
		})
	}
}

func commandTowerLikeLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "Command Tower-like Land",
		Types: []types.Card{types.Land},
		ManaAbilities: []game.ManaAbilityBody{{
			Text:            "{T}: Add one mana of any color in your commander's color identity.",
			AdditionalCosts: cost.Tap,
			Content: game.PlainAbilityContent{
				Sequence: []game.Instruction{
					{
						Primitive: game.Choose{
							Choice: game.ResolutionChoice{
								Kind:        game.ResolutionChoiceMana,
								Prompt:      "Choose a color in your commander's color identity",
								ColorSource: game.ResolutionChoiceColorSourceCommanderIdentity,
							},
							PublishChoice: "commander-color",
						},
					},
					{
						Primitive: game.AddMana{
							Amount:     game.Fixed(1),
							ChoiceFrom: "commander-color",
						},
					},
				},
			},
		}}},
	}
}

func TestResolutionPaymentCanGateIfYouDoBranch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
	manaCost := cost.Mana{cost.G}
	addInstructionSpellToStack(g, []game.Instruction{
		{
			Primitive:     game.Pay{Payment: game.ResolutionPayment{Prompt: "Pay {G}?", ManaCost: opt.Val(manaCost)}},
			PublishResult: "paid",
		},
		{
			Primitive:  game.Draw{Amount: game.Fixed(1), TargetIndex: game.TargetIndexController},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "paid", Accepted: game.TriTrue, Succeeded: game.TriTrue}),
		},
	})
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("hand size = %d, want payment branch to draw", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceMay {
		t.Fatalf("choices = %+v, want payment may choice", log.Choices)
	}
}

func TestDamageAndLoseLifeEffectsCanEliminatePlayers(t *testing.T) {
	tests := []struct {
		name      string
		primitive game.Primitive
	}{
		{name: "damage", primitive: game.Damage{Amount: game.Fixed(3), Recipient: game.TargetRecipient(0)}},
		{name: "lose life", primitive: game.LoseLife{Amount: game.Fixed(3), TargetIndex: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Players[game.Player2].Life = 3
			addEffectSpellToStack(g, game.Player1, tt.primitive, []game.Target{game.PlayerTarget(game.Player2)})

			engine.resolveTopOfStack(g, &TurnLog{})
			losses := engine.applyStateBasedActions(g)

			if len(losses) != 1 {
				t.Fatalf("losses = %d, want 1", len(losses))
			}
			if losses[0].Player != game.Player2 {
				t.Fatalf("loss player = %v, want %v", losses[0].Player, game.Player2)
			}
			if !g.Players[game.Player2].Eliminated {
				t.Fatal("player 2 was not eliminated")
			}
		})
	}
}

func TestFailedDrawEffectLogsAndEliminatesPlayer(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Draw{
		Amount:      game.Fixed(1),
		TargetIndex: game.TargetIndexController,
	}, nil)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)
	losses := engine.applyStateBasedActions(g)
	log.Losses = append(log.Losses, losses...)

	if len(log.Draws) != 1 {
		t.Fatalf("draw logs = %d, want 1", len(log.Draws))
	}
	if !log.Draws[0].Failed {
		t.Fatal("draw log did not record failed draw")
	}
	if len(log.Losses) != 1 {
		t.Fatalf("loss logs = %d, want 1", len(log.Losses))
	}
	if log.Losses[0].Player != game.Player1 || log.Losses[0].Reason != LossReasonEmptyLibraryDraw {
		t.Fatalf("loss log = %+v, want player %v reason %q", log.Losses[0], game.Player1, LossReasonEmptyLibraryDraw)
	}
	if !g.Players[game.Player1].Eliminated {
		t.Fatal("player 1 was not eliminated")
	}
}

func TestMillScryAndSurveilLibraryEffectsUseDeterministicFallback(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	third := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Third"}})
	addEffectSpellToStack(g, game.Player1, game.Scry{Amount: game.Fixed(2), TargetIndex: game.TargetIndexController}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Library.All(); len(got) < 3 || got[0] != third || got[1] != second || got[2] != top {
		t.Fatalf("library after scry = %+v, want deterministic keep-top order", got)
	}

	addEffectSpellToStack(g, game.Player1, game.Surveil{Amount: game.Fixed(2), TargetIndex: game.TargetIndexController}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Library.All(); len(got) < 3 || got[0] != third || got[1] != second || got[2] != top {
		t.Fatalf("library after surveil = %+v, want deterministic keep-top order", got)
	}

	addEffectSpellToStack(g, game.Player1, game.Mill{Amount: game.Fixed(2), TargetIndex: game.TargetIndexController}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if !g.Players[game.Player1].Graveyard.Contains(third) || !g.Players[game.Player1].Graveyard.Contains(second) {
		t.Fatal("mill did not move top two cards to graveyard")
	}
	if got := g.Players[game.Player1].Library.All(); len(got) != 1 || got[0] != top {
		t.Fatalf("library after mill = %+v, want only original bottom card", got)
	}
}

func TestCounterEffectCountersTargetStackObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	targetID := g.IDGen.Next()
	g.CardInstances[targetID] = &game.CardInstance{
		ID: targetID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Target Spell",
			Types: []types.Card{types.Sorcery}},
		},
		Owner: game.Player2,
	}
	targetObj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   targetID,
		Controller: game.Player2,
	}
	g.Stack.Push(targetObj)
	addEffectSpellToStack(g, game.Player1, game.CounterObject{TargetIndex: 0}, []game.Target{game.StackObjectTarget(targetObj.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, targetObj.ID); ok {
		t.Fatal("target stack object remained after counter effect")
	}
	if !g.Players[game.Player2].Graveyard.Contains(targetID) {
		t.Fatal("countered spell did not move to graveyard")
	}
}

func TestCounterEffectCannotCounterProtectedCreatureSpell(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Counter Shield",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbilityBody{{
			RuleEffects: []game.RuleEffect{{
				Kind:               game.RuleEffectCantBeCountered,
				AffectedController: game.ControllerYou,
				SpellTypes:         []types.Card{types.Creature},
			}},
		}}},
	})
	targetID := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Protected Creature",
		Types: []types.Card{types.Creature}},
	})
	g.Players[game.Player2].Hand.Remove(targetID)
	targetObj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   targetID,
		Controller: game.Player2,
	}
	g.Stack.Push(targetObj)
	addEffectSpellToStack(g, game.Player1, game.CounterObject{TargetIndex: 0}, []game.Target{game.StackObjectTarget(targetObj.ID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := stackObjectByID(g, targetObj.ID); !ok {
		t.Fatal("creature spell was countered despite can't-be-countered rule effect")
	}
	if g.Players[game.Player2].Graveyard.Contains(targetID) {
		t.Fatal("protected spell moved to graveyard")
	}
}

func TestExcessDamageCanFeedLaterEffectAmount(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Small Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2})},
	})
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive:     game.Damage{Amount: game.Fixed(5), Recipient: game.TargetRecipient(0)},
			PublishResult: "damage",
		},
		{
			Primitive: game.Damage{
				Recipient: game.TargetRecipient(1),
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:      game.DynamicAmountPreviousEffectExcessDamage,
					ResultKey: "damage",
				}),
			},
		},
	}, []game.Target{game.PermanentTarget(target.ObjectID), game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("player 2 life = %d, want 37 from excess damage", got)
	}
}

func TestZeroExcessDamageDoesNotSatisfySuccessCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3})},
	})
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{
			Primitive:     game.Damage{Amount: game.Fixed(2), Recipient: game.TargetRecipient(0), ResultAmountKind: game.EffectResultAmountExcessDamage},
			PublishResult: "excess",
		},
		{
			Primitive:  game.GainLife{TargetIndex: game.TargetIndexController, Amount: game.Fixed(5)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "excess", Succeeded: game.TriTrue}),
		},
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("player 1 life = %d, want no gain from zero excess damage", got)
	}
}

func TestDiscardEffectDiscardsDeterministicHandCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	bottom := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	top := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, game.Discard{Amount: game.Fixed(1), TargetIndex: 0}, []game.Target{game.PlayerTarget(game.Player2)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if g.Players[game.Player2].Hand.Contains(top) || !g.Players[game.Player2].Graveyard.Contains(top) {
		t.Fatal("discard effect did not discard deterministic top hand card")
	}
	if !g.Players[game.Player2].Hand.Contains(bottom) {
		t.Fatal("discard effect discarded more cards than requested")
	}
}

func TestSearchRevealAndInvestigateKeywordActions(t *testing.T) {
	t.Run("search library to hand with reveal", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		creature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Creature", Types: []types.Card{types.Creature}}})
		_ = addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Instant", Types: []types.Card{types.Instant}}})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount:      game.Fixed(1),
			TargetIndex: game.TargetIndexController,
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Creature),
				Reveal:      true,
			},
		}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if !g.Players[game.Player1].Hand.Contains(creature) || g.Players[game.Player1].Library.Contains(creature) {
			t.Fatal("search effect did not move matching card library -> hand")
		}
		if !hasEvent(g, game.EventCardRevealed) {
			t.Fatalf("events = %+v, want reveal event for searched card", g.Events)
		}
	})

	t.Run("search can require a basic land", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		basic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
			Supertypes: []types.Super{types.Basic},
			Types:      []types.Card{types.Land}},
		})
		nonbasic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Nonbasic Land",
			Types: []types.Card{types.Land}},
		})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount:      game.Fixed(1),
			TargetIndex: game.TargetIndexController,
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Land),
				Supertype:   opt.Val(types.Basic),
				Reveal:      true,
			},
		}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if !g.Players[game.Player1].Hand.Contains(basic) || g.Players[game.Player1].Library.Contains(basic) {
			t.Fatal("search effect did not move matching basic land library -> hand")
		}
		if !g.Players[game.Player1].Library.Contains(nonbasic) {
			t.Fatal("search effect moved nonbasic land despite basic filter")
		}
		if !hasEvent(g, game.EventCardRevealed) {
			t.Fatalf("events = %+v, want reveal event for searched basic land", g.Events)
		}
	})

	t.Run("search without supertype filter still matches nonbasic lands", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		nonbasic := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Nonbasic Land",
			Types: []types.Card{types.Land}},
		})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount:      game.Fixed(1),
			TargetIndex: game.TargetIndexController,
			Spec: game.SearchSpec{
				SourceZone:  zone.Library,
				Destination: zone.Hand,
				CardType:    opt.Val(types.Land),
			},
		}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if !g.Players[game.Player1].Hand.Contains(nonbasic) || g.Players[game.Player1].Library.Contains(nonbasic) {
			t.Fatal("search effect did not move nonbasic land without a supertype filter")
		}
	})

	t.Run("search can put subtype-matching land onto battlefield tapped", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		forest := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest}},
		})
		_ = addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Wastes",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Desert}},
		})
		addEffectSpellToStack(g, game.Player1, game.Search{
			Amount:      game.Fixed(1),
			TargetIndex: game.TargetIndexController,
			Spec: game.SearchSpec{
				SourceZone:   zone.Library,
				Destination:  zone.Battlefield,
				CardType:     opt.Val(types.Land),
				SubtypesAny:  []types.Sub{types.Forest},
				EntersTapped: true,
			},
		}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if g.Players[game.Player1].Library.Contains(forest) {
			t.Fatal("search effect left matching land in library")
		}
		permanent := permanentByCardID(g, forest)
		if permanent == nil {
			t.Fatal("search effect did not put matching land onto battlefield")
		}
		if !permanent.Tapped {
			t.Fatal("searched land entered untapped, want tapped")
		}
	})

	t.Run("reveal top library card", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
		addEffectSpellToStack(g, game.Player1, game.Reveal{Amount: game.Fixed(1), TargetIndex: game.TargetIndexController}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if !g.Players[game.Player1].Library.Contains(cardID) {
			t.Fatal("reveal effect moved the card")
		}
		if !hasEvent(g, game.EventCardRevealed) {
			t.Fatalf("events = %+v, want reveal event", g.Events)
		}
	})

	t.Run("investigate creates clue token", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		drawn := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn"}})
		addEffectSpellToStack(g, game.Player1, game.Investigate{Amount: game.Fixed(2)}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if len(g.Battlefield) != 2 {
			t.Fatalf("battlefield size = %d, want 2 clues", len(g.Battlefield))
		}
		clue := g.Battlefield[0]
		if !clue.Token || clue.TokenDef == nil || clue.TokenDef.Name != "Clue Token" || !clue.TokenDef.HasSubtype(types.Clue) {
			t.Fatalf("clue token = %+v def=%+v", clue, clue.TokenDef)
		}
		if clue.TokenDef.AbilityCount() != 1 || len(clue.TokenDef.ActivatedAbilities) != 1 {
			t.Fatalf("clue abilities = count=%d activated=%d, want activated draw ability", clue.TokenDef.AbilityCount(), len(clue.TokenDef.ActivatedAbilities))
		}
		g.Players[game.Player1].ManaPool.Add(mana.C, 2)
		if !engine.applyAction(g, game.Player1, actionBuild.activateAbility(clue.ObjectID, 0, nil, 0)) {
			t.Fatal("clue activation failed")
		}
		if _, ok := permanentByObjectID(g, clue.ObjectID); ok {
			t.Fatal("clue activation did not sacrifice its source")
		}
		engine.resolveTopOfStack(g, &TurnLog{})
		if !g.Players[game.Player1].Hand.Contains(drawn) {
			t.Fatal("clue activation did not draw a card")
		}
	})
}

func TestStartEnginesAndSpeedIncreasesOncePerTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.ActivePlayer = game.Player1
	g.Turn.TurnNumber = 1

	if !startEngines(g, game.Player1) {
		t.Fatal("startEngines failed")
	}
	if got := g.Players[game.Player1].Speed; got != 1 {
		t.Fatalf("speed = %d, want 1", got)
	}
	loseLife(g, game.Player2, 1)
	loseLife(g, game.Player3, 1)
	if got := g.Players[game.Player1].Speed; got != 2 {
		t.Fatalf("speed = %d, want one increase to 2 this turn", got)
	}
	g.Turn.TurnNumber = 2
	loseLife(g, game.Player2, 1)
	if got := g.Players[game.Player1].Speed; got != 3 {
		t.Fatalf("speed = %d, want 3 after next-turn opponent life loss", got)
	}
}

func TestMonstrosityEffectAddsCountersOnlyOnce(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Monster",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	})
	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.Monstrosity{Amount: game.Fixed(5), TargetIndex: game.TargetIndexSourcePermanent}, &TurnLog{})
	resolveInstruction(engine, g, obj, game.Monstrosity{Amount: game.Fixed(5), TargetIndex: game.TargetIndexSourcePermanent}, &TurnLog{})

	if !source.Monstrous {
		t.Fatal("source did not become monstrous")
	}
	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 5 {
		t.Fatalf("+1/+1 counters = %d, want 5 after repeated monstrosity resolutions", got)
	}
}

func TestSetClassLevelEffectAndClassInitialLevel(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	cardID := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Class",
		Types:    []types.Card{types.Enchantment},
		Subtypes: []types.Sub{types.Class}},
	})
	card := g.CardInstances[cardID]
	source, ok := createCardPermanent(g, card, game.Player1, zone.Hand)
	if !ok {
		t.Fatal("createCardPermanent failed")
	}
	if got := source.ClassLevel; got != 1 {
		t.Fatalf("initial class level = %d, want 1", got)
	}
	obj := &game.StackObject{
		Kind:         game.StackActivatedAbility,
		SourceID:     source.ObjectID,
		SourceCardID: source.CardInstanceID,
		Controller:   game.Player1,
	}

	resolveInstruction(engine, g, obj, game.SetClassLevel{Amount: game.Fixed(2), TargetIndex: game.TargetIndexSourcePermanent}, &TurnLog{})
	resolveInstruction(engine, g, obj, game.SetClassLevel{Amount: game.Fixed(1), TargetIndex: game.TargetIndexSourcePermanent}, &TurnLog{})

	if got := source.ClassLevel; got != 2 {
		t.Fatalf("class level = %d, want upgraded and not downgraded level 2", got)
	}
}

func TestRuleEffectCantBeBlockedBindsAffectedObject(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	otherAttacker := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		Kind:             game.RuleEffectCantBeBlocked,
		AffectedObjectID: attacker.ObjectID,
	})

	if canBlockAttacker(g, blocker, attacker) {
		t.Fatal("blocker could block creature affected by can't-be-blocked rule effect")
	}
	if !canBlockAttacker(g, blocker, otherAttacker) {
		t.Fatal("can't-be-blocked rule effect affected the wrong attacker")
	}
}

func TestProliferateAddsOneChosenCounterKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	permanent.Counters.Add(counter.PlusOnePlusOne, 1)
	permanent.Counters.Add(counter.Charge, 1)
	g.Players[game.Player2].PoisonCounters = 1
	addEffectSpellToStack(g, game.Player1, game.Proliferate{}, nil)
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{0}, {0}}},
	}
	log := TurnLog{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("+1/+1 counters = %d, want chosen counter incremented", got)
	}
	if got := permanent.Counters.Get(counter.Charge); got != 1 {
		t.Fatalf("charge counters = %d, want unchosen counter unchanged", got)
	}
	if got := g.Players[game.Player2].PoisonCounters; got != 2 {
		t.Fatalf("poison counters = %d, want proliferated player counter", got)
	}
	if len(log.Choices) != 2 || log.Choices[0].Request.Kind != game.ChoiceProliferate {
		t.Fatalf("choices = %+v, want proliferate choices", log.Choices)
	}
}

func TestGoadEffectExpiresOnGoadingPlayersNextTurn(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	addEffectSpellToStack(g, game.Player1, game.Goad{TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !wasGoadedBy(target, game.Player1) {
		t.Fatal("target was not goaded")
	}
	g.Turn.TurnNumber = 5
	g.Turn.ActivePlayer = game.Player1
	engine.runBeginningPhase(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	if wasGoadedBy(target, game.Player1) {
		t.Fatal("goad did not expire on goading player's next turn")
	}
}

func TestScryAndSurveilUseChoiceAgent(t *testing.T) {
	t.Run("scry bottom", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		bottom := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
		top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
		addEffectSpellToStack(g, game.Player1, game.Scry{Amount: game.Fixed(1), TargetIndex: game.TargetIndexController}, nil)
		log := TurnLog{}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

		engine.resolveTopOfStackWithChoices(g, agents, &log)

		if got := g.Players[game.Player1].Library.All(); len(got) != 2 || got[0] != bottom || got[1] != top {
			t.Fatalf("library after scry = %+v, want chosen card on bottom", got)
		}
		if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceScry || log.Choices[0].UsedFallback {
			t.Fatalf("choices = %+v, want non-fallback scry choice", log.Choices)
		}
	})
	t.Run("surveil graveyard", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		top := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
		addEffectSpellToStack(g, game.Player1, game.Surveil{Amount: game.Fixed(1), TargetIndex: game.TargetIndexController}, nil)
		log := TurnLog{}
		agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

		engine.resolveTopOfStackWithChoices(g, agents, &log)

		if g.Players[game.Player1].Library.Contains(top) || !g.Players[game.Player1].Graveyard.Contains(top) {
			t.Fatal("surveil choice did not move card to graveyard")
		}
		if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceSurveil || log.Choices[0].UsedFallback {
			t.Fatalf("choices = %+v, want non-fallback surveil choice", log.Choices)
		}
	})
}

func TestDestroyEffectMovesPermanentToGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Destroy{TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("destroyed permanent remained on battlefield")
	}
	if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("destroyed card was not in owner's graveyard")
	}
}

func TestExileAndBounceEffectsMovePermanentsToOwnerZones(t *testing.T) {
	tests := []struct {
		name      string
		primitive game.Primitive
	}{
		{name: "exile", primitive: game.Exile{TargetIndex: 0}},
		{name: "bounce", primitive: game.Bounce{TargetIndex: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			target := addCreaturePermanent(g, game.Player2)
			addEffectSpellToStack(g, game.Player1, tt.primitive, []game.Target{game.PermanentTarget(target.ObjectID)})

			engine.resolveTopOfStack(g, &TurnLog{})

			if _, ok := permanentByObjectID(g, target.ObjectID); ok {
				t.Fatal("moved permanent remained on battlefield")
			}
			var z *zone.Zone
			switch tt.name {
			case "exile":
				z = &g.Players[game.Player2].Exile
			case "bounce":
				z = &g.Players[game.Player2].Hand
			default:
			}
			if z == nil || !z.Contains(target.CardInstanceID) {
				t.Fatalf("card was not moved to expected zone for %s", tt.name)
			}
		})
	}
}

func TestSacrificeEffectMovesControllerPermanentThroughGraveyardIgnoringIndestructible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player1, game.Indestructible)
	addEffectSpellToStack(g, game.Player1, game.Sacrifice{TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("sacrificed permanent remained on battlefield")
	}
	if !g.Players[game.Player1].Graveyard.Contains(target.CardInstanceID) {
		t.Fatal("sacrificed permanent did not move to graveyard")
	}
}

func TestTapAndUntapEffectsChangeTappedState(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCreaturePermanent(g, game.Player2)
	addEffectSpellToStack(g, game.Player1, game.Tap{TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !target.Tapped {
		t.Fatal("tap effect did not tap permanent")
	}

	addEffectSpellToStack(g, game.Player1, game.Untap{TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if target.Tapped {
		t.Fatal("untap effect did not untap permanent")
	}
}

func TestDamageToPermanentEffectCanCauseLethalSBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addEffectSpellToStack(g, game.Player1, game.Damage{Amount: game.Fixed(3), Recipient: game.TargetRecipient(0)}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})

	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 1 {
		t.Fatalf("deaths = %d, want 1", len(deaths))
	}
	if _, ok := permanentByObjectID(g, target.ObjectID); ok {
		t.Fatal("lethally damaged permanent remained on battlefield")
	}
}

func TestMassDestroyCreaturesUsesSnapshotAndRespectsIndestructible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCreaturePermanent(g, game.Player1)
	creature2 := addCreaturePermanent(g, game.Player2)
	indestructible := addCombatCreaturePermanent(g, game.Player3, game.Indestructible)
	artifact := addCombatPermanent(g, game.Player4, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		TargetIndex: game.TargetIndexController,
		Selector:    game.EffectSelectorAllCreatures,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, creature1.ObjectID); ok {
		t.Fatal("first creature survived mass destroy")
	}
	if _, ok := permanentByObjectID(g, creature2.ObjectID); ok {
		t.Fatal("second creature survived mass destroy")
	}
	if _, ok := permanentByObjectID(g, indestructible.ObjectID); !ok {
		t.Fatal("indestructible creature did not survive mass destroy")
	}
	if _, ok := permanentByObjectID(g, artifact.ObjectID); !ok {
		t.Fatal("noncreature artifact did not survive mass destroy")
	}
}

func TestMassDestroyNonlandPermanentsLeavesLands(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island",
		Types: []types.Card{types.Land}},
	})
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	enchantment := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Aura",
		Types: []types.Card{types.Enchantment}},
	})
	addEffectSpellToStack(g, game.Player1, game.Destroy{
		TargetIndex: game.TargetIndexController,
		Selector:    game.EffectSelectorAllNonlandPermanents,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if _, ok := permanentByObjectID(g, land.ObjectID); !ok {
		t.Fatal("land did not survive nonland permanent wipe")
	}
	if _, ok := permanentByObjectID(g, artifact.ObjectID); ok {
		t.Fatal("artifact survived nonland permanent wipe")
	}
	if _, ok := permanentByObjectID(g, enchantment.ObjectID); ok {
		t.Fatal("enchantment survived nonland permanent wipe")
	}
}

func TestSelectorOtherCreaturesDefendingPlayerControlsUsesTriggerRecipientController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	damagedBlocker := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	otherDefenderCreature := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
	defenderArtifact := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Defender Relic",
		Types: []types.Card{types.Artifact}},
	})
	attackerCreature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	obj := &game.StackObject{
		ID:              g.IDGen.Next(),
		Controller:      game.Player1,
		HasTriggerEvent: true,
		TriggerEvent: game.GameEvent{
			Kind:        game.EventDamageDealt,
			PermanentID: damagedBlocker.ObjectID,
		},
	}

	resolveInstruction(engine, g, obj, game.Destroy{
		Selector: game.EffectSelectorOtherCreaturesDefendingPlayerControls,
	}, &TurnLog{})

	if _, ok := permanentByObjectID(g, otherDefenderCreature.ObjectID); ok {
		t.Fatal("other defender creature survived selector destroy")
	}
	for _, permanent := range []*game.Permanent{damagedBlocker, defenderArtifact, attackerCreature} {
		if _, ok := permanentByObjectID(g, permanent.ObjectID); !ok {
			t.Fatalf("selector destroyed permanent %v unexpectedly", permanent.ObjectID)
		}
	}
}

func TestMassDamageDeathsAreLoggedTogetherBySBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature2 := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	artifact := addCombatPermanent(g, game.Player3, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.Damage{
			Amount:    game.Fixed(3),
			Recipient: game.SelectorRecipient(game.EffectSelectorAllCreatures),
		},
	}}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if len(deaths) != 2 {
		t.Fatalf("deaths = %d, want 2", len(deaths))
	}
	if _, ok := permanentByObjectID(g, creature1.ObjectID); ok {
		t.Fatal("first damaged creature survived SBA")
	}
	if _, ok := permanentByObjectID(g, creature2.ObjectID); ok {
		t.Fatal("second damaged creature survived SBA")
	}
	if _, ok := permanentByObjectID(g, artifact.ObjectID); !ok {
		t.Fatal("noncreature artifact was affected by creature mass damage")
	}
}

func TestTemporaryPTModifierChangesCombatDamageAndLethalThreshold(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	blocker := addCombatCreaturePermanentWithPower(g, game.Player2, 4)
	addEffectSpellToStack(g, game.Player1, game.ModifyPT{
		TargetIndex:    0,
		PowerDelta:     game.Fixed(3),
		ToughnessDelta: game.Fixed(3),
		Duration:       game.DurationUntilEndOfTurn,
	}, []game.Target{game.PermanentTarget(creature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: creature.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: blocker.ObjectID, Blocking: creature.ObjectID},
		},
	}
	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if blocker.MarkedDamage != 5 {
		t.Fatalf("blocker marked damage = %d, want 5", blocker.MarkedDamage)
	}
	if _, ok := permanentByObjectID(g, blocker.ObjectID); ok {
		t.Fatal("blocker survived pumped combat damage")
	}
	if _, ok := permanentByObjectID(g, creature.ObjectID); !ok {
		t.Fatal("pumped creature died despite increased toughness")
	}
	if len(deaths) != 1 || deaths[0].Permanent != blocker.ObjectID {
		t.Fatalf("deaths = %+v, want blocker death only", deaths)
	}
}

func TestTemporaryPTModifiersStackDeterministically(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	for _, primitive := range []game.Primitive{
		game.ModifyPT{TargetIndex: 0, PowerDelta: game.Fixed(1), ToughnessDelta: game.Fixed(2), Duration: game.DurationUntilEndOfTurn},
		game.ModifyPT{TargetIndex: 0, PowerDelta: game.Fixed(-2), ToughnessDelta: game.Fixed(-1), Duration: game.DurationUntilEndOfTurn},
	} {
		addEffectSpellToStack(g, game.Player1, primitive, []game.Target{game.PermanentTarget(creature.ObjectID)})
		engine.resolveTopOfStack(g, &TurnLog{})
	}

	if got := effectivePower(g, creature); got != 1 {
		t.Fatalf("effective power = %d, want 1", got)
	}
	if got, ok := effectiveToughness(g, creature); !ok || got != 3 {
		t.Fatalf("effective toughness = %d ok=%v, want 3 true", got, ok)
	}
}

func TestAddCounterEffectAddsCountersToTargetPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	addEffectSpellToStack(g, game.Player1, game.AddCounter{
		TargetIndex: 0,
		Amount:      game.Fixed(3),
		CounterKind: counter.PlusOnePlusOne,
	}, []game.Target{game.PermanentTarget(artifact.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := artifact.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("+1/+1 counters = %d, want 3", got)
	}
}

func TestMoveCountersEffectMovesCountersBetweenTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Source Relic",
		Types: []types.Card{types.Artifact}},
	})
	destination := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Destination Relic",
		Types: []types.Card{types.Artifact}},
	})
	source.Counters.Add(counter.PlusOnePlusOne, 2)
	source.Counters.Add(counter.Charge, 1)
	addEffectSpellToStack(g, game.Player1, game.MoveCounters{
		TargetIndex: 1,
		Source: game.CounterSourceSpec{
			Kind:        game.CounterSourceTarget,
			TargetIndex: 0,
		},
	}, []game.Target{
		game.PermanentTarget(source.ObjectID),
		game.PermanentTarget(destination.ObjectID),
	})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := source.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("source +1/+1 counters = %d, want 0", got)
	}
	if got := source.Counters.Get(counter.Charge); got != 0 {
		t.Fatalf("source charge counters = %d, want 0", got)
	}
	if got := destination.Counters.Get(counter.PlusOnePlusOne); got != 2 {
		t.Fatalf("destination +1/+1 counters = %d, want 2", got)
	}
	if got := destination.Counters.Get(counter.Charge); got != 1 {
		t.Fatalf("destination charge counters = %d, want 1", got)
	}
}

func TestConditionalContinuousEffectAnimatesNonCreatureArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Relic",
		Types: []types.Card{types.Artifact}},
	})
	zero := game.PT{Value: 0}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.ApplyContinuous{
			TargetIndex: 0,
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:       game.LayerType,
					AddTypes:    []types.Card{types.Creature},
					AddSubtypes: []types.Sub{types.Robot},
				},
				{
					Layer:        game.LayerPowerToughnessSet,
					SetPower:     opt.Val(zero),
					SetToughness: opt.Val(zero),
				},
			},
		},
		Condition: opt.Val(game.EffectCondition{Text: "it isn't a creature", TargetIndex: 0, PermanentType: opt.Val(types.Creature), Negate: true}),
	}}, []game.Target{game.PermanentTarget(artifact.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if !permanentHasType(g, artifact, types.Creature) {
		t.Fatal("noncreature artifact did not become a creature")
	}
	if !permanentHasSubtype(g, artifact, types.Robot) {
		t.Fatal("noncreature artifact did not gain Robot subtype")
	}
	if got := effectivePower(g, artifact); got != 0 {
		t.Fatalf("effective power = %d, want 0", got)
	}
	if got, ok := effectiveToughness(g, artifact); !ok || got != 0 {
		t.Fatalf("effective toughness = %d ok=%v, want 0 true", got, ok)
	}
}

func TestConditionalContinuousEffectSkipsCreatureArtifact(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	two := game.PT{Value: 2}
	artifactCreature := addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     opt.Val(two),
		Toughness: opt.Val(two)},
	})
	zero := game.PT{Value: 0}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{
		Primitive: game.ApplyContinuous{
			TargetIndex: 0,
			ContinuousEffects: []game.ContinuousEffect{
				{
					Layer:       game.LayerType,
					AddTypes:    []types.Card{types.Creature},
					AddSubtypes: []types.Sub{types.Robot},
				},
				{
					Layer:        game.LayerPowerToughnessSet,
					SetPower:     opt.Val(zero),
					SetToughness: opt.Val(zero),
				},
			},
		},
		Condition: opt.Val(game.EffectCondition{Text: "it isn't a creature", TargetIndex: 0, PermanentType: opt.Val(types.Creature), Negate: true}),
	}}, []game.Target{game.PermanentTarget(artifactCreature.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if permanentHasSubtype(g, artifactCreature, types.Robot) {
		t.Fatal("creature artifact incorrectly gained Robot subtype")
	}
	if got := effectivePower(g, artifactCreature); got != 2 {
		t.Fatalf("effective power = %d, want 2", got)
	}
	if got, ok := effectiveToughness(g, artifactCreature); !ok || got != 2 {
		t.Fatalf("effective toughness = %d ok=%v, want 2 true", got, ok)
	}
}

func TestTemporaryPTModifierExpiresDuringCleanup(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature.TemporaryPowerModifier = 3
	creature.TemporaryToughnessModifier = 3

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if creature.TemporaryPowerModifier != 0 || creature.TemporaryToughnessModifier != 0 {
		t.Fatalf("temporary modifiers = +%d/+%d, want 0/0", creature.TemporaryPowerModifier, creature.TemporaryToughnessModifier)
	}
	if got := effectivePower(g, creature); got != 2 {
		t.Fatalf("effective power after cleanup = %d, want 2", got)
	}
}

func TestCreateTokenEffectCreatesTokenPermanent(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	token := &game.CardDef{CardFace: game.CardFace{Name: "Soldier Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1})},
	}
	addEffectSpellToStack(g, game.Player1, game.CreateToken{Amount: game.Fixed(2), Source: game.TokenDef(token)}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	var tokens []*game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			tokens = append(tokens, permanent)
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("tokens = %d, want 2", len(tokens))
	}
	for _, permanent := range tokens {
		if permanent.TokenDef != token {
			t.Fatalf("token def = %p, want %p", permanent.TokenDef, token)
		}
		if permanent.Controller != game.Player1 || permanent.Owner != game.Player1 {
			t.Fatalf("token owner/controller = %v/%v, want %v", permanent.Owner, permanent.Controller, game.Player1)
		}
		if !permanent.SummoningSick {
			t.Fatal("token did not enter summoning sick")
		}
	}
}

func TestCreateTokenPermanentAppliesReplacementAbilities(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	token := &game.CardDef{CardFace: game.CardFace{Name: "Modified Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		ReplacementAbilities: []game.ReplacementAbilityBody{
			game.EntersTappedReplacement("This token enters tapped."),
			game.EntersWithCountersReplacement("This token enters with a +1/+1 counter.", game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
		}},
	}

	permanent, ok := createTokenPermanent(g, game.Player1, token)

	if !ok {
		t.Fatal("token was not created")
	}
	if !permanent.Tapped {
		t.Fatal("token did not enter tapped")
	}
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 1 {
		t.Fatalf("+1/+1 counters = %d, want 1", got)
	}
}

func TestCreateTokenCanCopySourceCardWithModifications(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Fanatic Source",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Druid},
			ManaCost:  opt.Val(cost.Mana{cost.G}),
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4})},
		},
		Owner: game.Player1,
	}
	g.Stack.Push(&game.StackObject{
		ID:           g.IDGen.Next(),
		Kind:         game.StackActivatedAbility,
		SourceID:     sourceID,
		SourceCardID: sourceID,
		Controller:   game.Player1,
		AbilityIndex: 0,
	})
	g.CardInstances[sourceID].Def.ActivatedAbilities = []game.ActivatedAbilityBody{
		game.EternalizeActivatedBody(cost.Mana{cost.O(0)}, types.Snake, types.Druid),
	}

	engine.resolveTopOfStack(g, &TurnLog{})

	var token *game.Permanent
	for _, permanent := range g.Battlefield {
		if permanent.Token {
			token = permanent
			break
		}
	}
	if token == nil || token.TokenDef == nil {
		t.Fatal("copy token was not created")
	}
	if token.TokenDef.ManaCost.Exists || token.TokenDef.ManaValue() != 0 {
		t.Fatalf("token mana cost/value = %+v/%d, want no cost and mana value 0", token.TokenDef.ManaCost, token.TokenDef.ManaValue())
	}
	if got := token.TokenDef.Subtypes; !slices.Equal(got, []types.Sub{types.Zombie, types.Snake, types.Druid}) {
		t.Fatalf("token subtypes = %+v, want Zombie Snake Druid", got)
	}
	if got := token.TokenDef.Colors; !slices.Equal(got, []color.Color{color.Black}) {
		t.Fatalf("token colors = %+v, want black", got)
	}
	if got := effectivePower(g, token); got != 4 {
		t.Fatalf("token power = %d, want 4", got)
	}
	if got, ok := effectiveToughness(g, token); !ok || got != 4 {
		t.Fatalf("token toughness = %d ok=%v, want 4 true", got, ok)
	}
}

func TestCopyCardDefPreservesCategorizedAbilitiesWithoutDuplication(t *testing.T) {
	source := &game.CardDef{CardFace: game.CardFace{
		Name: "Categorized Source",
		StaticAbilities: []game.StaticAbilityBody{{
			Text:             "Flying",
			KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Flying}},
		}},
	}}

	copied := copyCardDef(source)

	if copied.AbilityCount() != 1 {
		t.Fatalf("copied abilities = %d, want one categorized ability without duplication", copied.AbilityCount())
	}
	if !copied.HasKeyword(game.Flying) {
		t.Fatal("copied categorized keyword ability was not preserved")
	}
}

func TestClearCardFaceAbilitiesClearsCategorizedAbilities(t *testing.T) {
	card := &game.CardDef{CardFace: game.CardFace{
		StaticAbilities: []game.StaticAbilityBody{{
			Text:             "Flying",
			KeywordAbilities: []game.KeywordAbility{game.SimpleKeyword{Kind: game.Flying}},
		}},
	}}
	face := card.CardFace

	clearCardFaceAbilities(&face)

	if face.AbilityCount() != 0 {
		t.Fatalf("abilities = %d, want categorized abilities cleared", face.AbilityCount())
	}
	face.StaticAbilities = []game.StaticAbilityBody{game.FlyingStaticBody}
	if !face.HasKeyword(game.Flying) {
		t.Fatal("ability cache remained stale after clearing and adding a categorized ability")
	}
}

func TestTokenCanBlockTakeCombatDamageAndDie(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pt := game.PT{Value: 2}
	token, ok := createTokenPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Bear Token",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt)},
	})
	if !ok {
		t.Fatal("token was not created")
	}
	attacker := addCombatCreaturePermanentWithPower(g, game.Player1, 3)
	g.Combat = &game.CombatState{
		Attackers: []game.AttackDeclaration{
			{Attacker: attacker.ObjectID, Target: game.AttackTarget{Player: game.Player2}},
		},
		Blockers: []game.BlockDeclaration{
			{Blocker: token.ObjectID, Blocking: attacker.ObjectID},
		},
	}

	engine.resolveCombatDamage(g, &TurnLog{})
	_, deaths := engine.applyStateBasedActionsWithDeaths(g)

	if _, ok := permanentByObjectID(g, token.ObjectID); ok {
		t.Fatal("lethally damaged token remained on battlefield")
	}
	if g.Players[game.Player2].Graveyard.Contains(token.ObjectID) {
		t.Fatal("dead token did not cease to exist from graveyard")
	}
	if len(deaths) != 1 || deaths[0].Permanent != token.ObjectID || deaths[0].TokenName != "Bear Token" {
		t.Fatalf("death logs = %+v, want readable token death", deaths)
	}
}

func addEffectSpellToStack(g *game.Game, controller game.PlayerID, primitive game.Primitive, targets []game.Target) id.ID {
	return addInstructionSpellToStackForController(g, controller, []game.Instruction{{Primitive: primitive}}, targets)
}

func addInstructionSpellToStack(g *game.Game, instructions []game.Instruction) id.ID {
	return addInstructionSpellToStackForController(g, game.Player1, instructions, nil)
}

func addInstructionSpellToStackForController(g *game.Game, controller game.PlayerID, instructions []game.Instruction, targets []game.Target) id.ID {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{CardFace: game.CardFace{Name: "Effect Spell",
			Types: []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.SpellAbilityBody{
				Content: game.PlainAbilityContent{
					Sequence: append([]game.Instruction(nil), instructions...),
				},
			})},
		},
		Owner: controller,
	}
	g.Stack.Push(&game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   sourceID,
		Controller: controller,
		Targets:    targets,
	})
	return sourceID
}

func resolveInstruction(engine *Engine, g *game.Game, obj *game.StackObject, primitive game.Primitive, log *TurnLog) {
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{Primitive: primitive}, [game.NumPlayers]PlayerAgent{}, log)
}
