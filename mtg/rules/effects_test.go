package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDrawEffectDrawsRequestedCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDraw,
		Amount:      2,
		TargetIndex: -1,
	}, nil)
	firstDraw := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "First"})
	secondDraw := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Second"})
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

func TestUnsupportedEffectsAreLogged(t *testing.T) {
	tests := []game.EffectType{
		game.EffectGainControl,
		game.EffectCopy,
		game.EffectAttach,
	}
	for _, effectType := range tests {
		t.Run(effectTypeName(effectType), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			sourceID := addEffectSpellToStack(g, game.Player1, game.Effect{
				Type:        effectType,
				Description: "unsupported test effect",
			}, nil)
			log := TurnLog{}

			engine.resolveTopOfStack(g, &log)

			if IsEffectTypeExecuted(effectType) {
				t.Fatalf("%v reported supported unexpectedly", effectType)
			}
			if len(log.Unsupported) != 1 {
				t.Fatalf("unsupported logs = %d, want 1", len(log.Unsupported))
			}
			if log.Unsupported[0].EffectType != effectType || log.Unsupported[0].SourceID != sourceID {
				t.Fatalf("unsupported log = %+v, want type %v source %v", log.Unsupported[0], effectType, sourceID)
			}
		})
	}
}

func effectTypeName(effectType game.EffectType) string {
	switch effectType {
	case game.EffectCounter:
		return "counter"
	case game.EffectDiscard:
		return "discard"
	case game.EffectSearch:
		return "search"
	case game.EffectReveal:
		return "reveal"
	case game.EffectInvestigate:
		return "investigate"
	case game.EffectGainControl:
		return "gain-control"
	case game.EffectCopy:
		return "copy"
	case game.EffectAttach:
		return "attach"
	case game.EffectReplace:
		return "replace"
	default:
		return "unknown"
	}
}

func TestGainLifeEffectIncreasesTargetLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectGainLife,
		Amount:      3,
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
	addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "No Lifegain",
		Types: []types.Card{types.Enchantment},
		Abilities: []game.AbilityDef{{
			Kind: game.StaticAbility,
			Effects: []game.Effect{{
				Type: game.EffectApplyRule,
				RuleEffects: []game.RuleEffect{{
					Kind:           game.RuleEffectCantGainLife,
					AffectedPlayer: game.PlayerAny,
				}},
			}},
		}},
	})
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectGainLife,
		Amount:      3,
		TargetIndex: -1,
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
	addCardToHand(g, game.Player1, &game.CardDef{Name: "First"})
	addCardToHand(g, game.Player1, &game.CardDef{Name: "Second"})
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectGainLife,
		TargetIndex: -1,
		DynamicAmount: optDynamicAmount(game.DynamicAmount{
			Kind: game.DynamicAmountControllerHandSize,
		}),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 42 {
		t.Fatalf("life = %d, want 42", got)
	}
}

func TestDynamicAmountUsesXValue(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDamage,
		TargetIndex: 0,
		DynamicAmount: optDynamicAmount(game.DynamicAmount{
			Kind: game.DynamicAmountX,
		}),
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
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDamage,
		TargetIndex: 0,
		DynamicAmount: optDynamicAmount(game.DynamicAmount{
			Kind:        game.DynamicAmountTargetPower,
			TargetIndex: 0,
		}),
	}, []game.Target{game.PermanentTarget(target.ObjectID)})

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := target.MarkedDamage; got != 5 {
		t.Fatalf("marked damage = %d, want 5", got)
	}
}

func TestDynamicAmountCanUsePreviousEffectResult(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{
			Name:  "Linked Amount Spell",
			Types: []types.Card{types.Sorcery},
			Abilities: []game.AbilityDef{
				{
					Kind: game.SpellAbility,
					Effects: []game.Effect{
						{Type: game.EffectGainLife, TargetIndex: -1, Amount: 3, LinkID: "that-much"},
						{
							Type:        game.EffectLoseLife,
							TargetIndex: 0,
							DynamicAmount: optDynamicAmount(game.DynamicAmount{
								Kind:   game.DynamicAmountPreviousEffectResult,
								LinkID: "that-much",
							}),
						},
					},
				},
			},
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
		addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
		addEffectSpellToStack(g, game.Player1, game.Effect{
			Type:        game.EffectDraw,
			TargetIndex: -1,
			Amount:      1,
			Optional:    true,
		}, nil)
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
		addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
		addEffectSpellToStack(g, game.Player1, game.Effect{
			Type:        game.EffectDraw,
			TargetIndex: -1,
			Amount:      1,
			Optional:    true,
		}, nil)
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

func TestEffectResultConditionBranchesOnIfYouDoAndDont(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	sourceID := addLinkedResultSpellToStack(g, []game.Effect{
		{Type: game.EffectDraw, TargetIndex: -1, Amount: 1, Optional: true, LinkID: "choice"},
		{
			Type:        game.EffectGainLife,
			TargetIndex: -1,
			Amount:      3,
			ResultCondition: optEffectResultCondition(game.EffectResultCondition{
				LinkID:    "choice",
				Accepted:  game.TriTrue,
				Succeeded: game.TriTrue,
			}),
		},
		{
			Type:        game.EffectLoseLife,
			TargetIndex: -1,
			Amount:      3,
			ResultCondition: optEffectResultCondition(game.EffectResultCondition{
				LinkID:   "choice",
				Accepted: game.TriFalse,
			}),
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

func TestEffectResultConditionRequiresActualSuccess(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addLinkedResultSpellToStack(g, []game.Effect{
		{Type: game.EffectDraw, TargetIndex: -1, Amount: 1, LinkID: "draw"},
		{
			Type:        game.EffectGainLife,
			TargetIndex: -1,
			Amount:      3,
			ResultCondition: optEffectResultCondition(game.EffectResultCondition{
				LinkID:    "draw",
				Succeeded: game.TriTrue,
			}),
		},
		{
			Type:        game.EffectLoseLife,
			TargetIndex: -1,
			Amount:      2,
			ResultCondition: optEffectResultCondition(game.EffectResultCondition{
				LinkID:    "draw",
				Succeeded: game.TriFalse,
			}),
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
	addLinkedResultSpellToStack(g, []game.Effect{
		{
			Type:        game.EffectGainLife,
			TargetIndex: -1,
			Amount:      5,
			Optional:    true,
			LinkID:      "amount",
		},
		{
			Type:        game.EffectLoseLife,
			TargetIndex: -1,
			DynamicAmount: optDynamicAmount(game.DynamicAmount{
				Kind:   game.DynamicAmountPreviousEffectResult,
				LinkID: "amount",
			}),
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

func TestResolutionChoiceCanFeedLaterEffect(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addLinkedResultSpellToStack(g, []game.Effect{
		{
			Type:   game.EffectChoose,
			LinkID: "chosen-player",
			Choice: optResolutionChoice(game.ResolutionChoice{
				Kind:           game.ResolutionChoicePlayer,
				PlayerRelation: game.PlayerOpponent,
			}),
		},
		{
			Type:         game.EffectLoseLife,
			Amount:       3,
			ChoiceLinkID: "chosen-player",
		},
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("player 2 life = %d, want unchosen opponent unchanged", got)
	}
	if got := g.Players[game.Player3].Life; got != 37 {
		t.Fatalf("player 3 life = %d, want chosen opponent to lose life", got)
	}
}

func TestResolutionChoiceCanChooseManaColor(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addLinkedResultSpellToStack(g, []game.Effect{
		{
			Type:   game.EffectChoose,
			LinkID: "chosen-color",
			Choice: optResolutionChoice(game.ResolutionChoice{
				Kind: game.ResolutionChoiceColor,
			}),
		},
		{
			Type:         game.EffectAddMana,
			Amount:       1,
			ChoiceLinkID: "chosen-color",
		},
	})
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{3}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].ManaPool.Total(); got != 1 {
		t.Fatalf("mana pool total = %d, want one chosen mana", got)
	}
	if !g.Players[game.Player1].ManaPool.Spend(mana.Red, 1) {
		t.Fatal("chosen mana was not red")
	}
}

func TestCommanderIdentityColorChoiceFeedsManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{
		game.Player1: {
			Commander: &game.CardDef{
				Name:          "Dimir Commander",
				Types:         []types.Card{types.Creature},
				ColorIdentity: mana.NewColorIdentity(mana.Blue, mana.Black),
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
	if !g.Players[game.Player1].ManaPool.Spend(mana.Black, 1) {
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
					Commander: &game.CardDef{
						Name:          "Colorless Commander",
						Types:         []types.Card{types.Creature},
						ColorIdentity: mana.NewColorIdentity(),
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

			if canActivateManaAbility(g, game.Player1, tower, &card.Abilities[0], 0) {
				t.Fatal("canActivateManaAbility() = true, want false without commander color options")
			}
			if got := NewEngine(nil).legalActivateAbilityActions(g, game.Player1); len(got) != 0 {
				t.Fatalf("legal activation actions = %+v, want none", got)
			}
		})
	}
}

func commandTowerLikeLand() *game.CardDef {
	return &game.CardDef{
		Name:  "Command Tower-like Land",
		Types: []types.Card{types.Land},
		Abilities: []game.AbilityDef{{
			Kind:          game.ActivatedAbility,
			Text:          "{T}: Add one mana of any color in your commander's color identity.",
			IsManaAbility: true,
			AdditionalCosts: []game.AdditionalCost{
				{Kind: game.AdditionalCostTap},
			},
			Effects: []game.Effect{
				{
					Type: game.EffectChoose,
					Choice: opt.Val(game.ResolutionChoice{
						Kind:        game.ResolutionChoiceColor,
						Prompt:      "Choose a color in your commander's color identity",
						ColorSource: game.ResolutionChoiceColorSourceCommanderIdentity,
					}),
					LinkID: "commander-color",
				},
				{
					Type:         game.EffectAddMana,
					Amount:       1,
					ChoiceLinkID: "commander-color",
				},
			},
		}},
	}
}

func TestResolutionPaymentCanGateIfYouDoBranch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Forest)
	addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
	cost := mana.Cost{mana.ColoredMana(mana.Green)}
	addLinkedResultSpellToStack(g, []game.Effect{
		{
			Type:   game.EffectPay,
			LinkID: "paid",
			Payment: optResolutionPayment(game.ResolutionPayment{
				Prompt:   "Pay {G}?",
				ManaCost: optCost(cost),
			}),
		},
		{
			Type:        game.EffectDraw,
			Amount:      1,
			TargetIndex: -1,
			ResultCondition: optEffectResultCondition(game.EffectResultCondition{
				LinkID:    "paid",
				Accepted:  game.TriTrue,
				Succeeded: game.TriTrue,
			}),
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
		name       string
		effectType game.EffectType
	}{
		{name: "damage", effectType: game.EffectDamage},
		{name: "lose life", effectType: game.EffectLoseLife},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			g.Players[game.Player2].Life = 3
			addEffectSpellToStack(g, game.Player1, game.Effect{
				Type:        tt.effectType,
				Amount:      3,
				TargetIndex: 0,
			}, []game.Target{game.PlayerTarget(game.Player2)})

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
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDraw,
		Amount:      1,
		TargetIndex: -1,
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
	top := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Top"})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Second"})
	third := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Third"})
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectScry, Amount: 2, TargetIndex: -1}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Library.All(); len(got) < 3 || got[0] != third || got[1] != second || got[2] != top {
		t.Fatalf("library after scry = %+v, want deterministic keep-top order", got)
	}

	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectSurveil, Amount: 2, TargetIndex: -1}, nil)
	engine.resolveTopOfStack(g, &TurnLog{})
	if got := g.Players[game.Player1].Library.All(); len(got) < 3 || got[0] != third || got[1] != second || got[2] != top {
		t.Fatalf("library after surveil = %+v, want deterministic keep-top order", got)
	}

	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectMill, Amount: 2, TargetIndex: -1}, nil)
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
		Def: &game.CardDef{
			Name:  "Target Spell",
			Types: []types.Card{types.Sorcery},
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
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectCounter, TargetIndex: 0}, []game.Target{game.StackObjectTarget(targetObj.ID)})

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
	addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:  "Counter Shield",
		Types: []types.Card{types.Enchantment},
		Abilities: []game.AbilityDef{{
			Kind: game.StaticAbility,
			Effects: []game.Effect{{
				Type: game.EffectApplyRule,
				RuleEffects: []game.RuleEffect{{
					Kind:               game.RuleEffectCantBeCountered,
					AffectedController: game.ControllerYou,
					SpellTypes:         []types.Card{types.Creature},
				}},
			}},
		}},
	})
	targetID := addCardToHand(g, game.Player2, &game.CardDef{
		Name:  "Protected Creature",
		Types: []types.Card{types.Creature},
	})
	g.Players[game.Player2].Hand.Remove(targetID)
	targetObj := &game.StackObject{
		ID:         g.IDGen.Next(),
		Kind:       game.StackSpell,
		SourceID:   targetID,
		Controller: game.Player2,
	}
	g.Stack.Push(targetObj)
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectCounter, TargetIndex: 0}, []game.Target{game.StackObjectTarget(targetObj.ID)})

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
	target := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:      "Small Creature",
		Types:     []types.Card{types.Creature},
		Power:     optPT(game.PT{Value: 2}),
		Toughness: optPT(game.PT{Value: 2}),
	})
	addLinkedResultSpellToStackForController(g, game.Player1, []game.Effect{
		{
			Type:        game.EffectDamage,
			Amount:      5,
			TargetIndex: 0,
			LinkID:      "damage",
		},
		{
			Type:        game.EffectDamage,
			TargetIndex: 1,
			DynamicAmount: opt.Val(game.DynamicAmount{
				Kind:   game.DynamicAmountPreviousEffectExcessDamage,
				LinkID: "damage",
			}),
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
	target := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:      "Large Creature",
		Types:     []types.Card{types.Creature},
		Power:     optPT(game.PT{Value: 3}),
		Toughness: optPT(game.PT{Value: 3}),
	})
	addLinkedResultSpellToStackForController(g, game.Player1, []game.Effect{
		{
			Type:         game.EffectDamage,
			Amount:       2,
			TargetIndex:  0,
			ResultAmount: game.EffectResultAmountExcessDamage,
			LinkID:       "excess",
		},
		{
			Type:        game.EffectGainLife,
			TargetIndex: -1,
			Amount:      5,
			ResultCondition: optEffectResultCondition(game.EffectResultCondition{
				LinkID:    "excess",
				Succeeded: game.TriTrue,
			}),
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
	bottom := addCardToHand(g, game.Player2, &game.CardDef{Name: "Bottom"})
	top := addCardToHand(g, game.Player2, &game.CardDef{Name: "Top"})
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectDiscard, Amount: 1, TargetIndex: 0}, []game.Target{game.PlayerTarget(game.Player2)})

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
		creature := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Creature", Types: []types.Card{types.Creature}})
		_ = addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Instant", Types: []types.Card{types.Instant}})
		addEffectSpellToStack(g, game.Player1, game.Effect{
			Type:        game.EffectSearch,
			Amount:      1,
			TargetIndex: -1,
			Search: opt.Val(game.SearchSpec{
				SourceZone:    game.ZoneLibrary,
				Destination:   game.ZoneHand,
				MatchCardType: true,
				CardType:      types.Creature,
				Reveal:        true,
				Shuffle:       true,
			}),
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
		basic := addCardToLibrary(g, game.Player1, &game.CardDef{
			Name:       "Forest",
			Supertypes: []types.Super{types.Basic},
			Types:      []types.Card{types.Land},
		})
		nonbasic := addCardToLibrary(g, game.Player1, &game.CardDef{
			Name:  "Nonbasic Land",
			Types: []types.Card{types.Land},
		})
		addEffectSpellToStack(g, game.Player1, game.Effect{
			Type:        game.EffectSearch,
			Amount:      1,
			TargetIndex: -1,
			Search: opt.Val(game.SearchSpec{
				SourceZone:     game.ZoneLibrary,
				Destination:    game.ZoneHand,
				MatchCardType:  true,
				CardType:       types.Land,
				MatchSupertype: true,
				Supertype:      types.Basic,
				Reveal:         true,
			}),
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
		nonbasic := addCardToLibrary(g, game.Player1, &game.CardDef{
			Name:  "Nonbasic Land",
			Types: []types.Card{types.Land},
		})
		addEffectSpellToStack(g, game.Player1, game.Effect{
			Type:        game.EffectSearch,
			Amount:      1,
			TargetIndex: -1,
			Search: opt.Val(game.SearchSpec{
				SourceZone:    game.ZoneLibrary,
				Destination:   game.ZoneHand,
				MatchCardType: true,
				CardType:      types.Land,
			}),
		}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if !g.Players[game.Player1].Hand.Contains(nonbasic) || g.Players[game.Player1].Library.Contains(nonbasic) {
			t.Fatal("search effect did not move nonbasic land without a supertype filter")
		}
	})

	t.Run("reveal top library card", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		cardID := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Top"})
		addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectReveal, Amount: 1, TargetIndex: -1}, nil)

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
		drawn := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Drawn"})
		addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectInvestigate, Amount: 2, TargetIndex: -1}, nil)

		engine.resolveTopOfStack(g, &TurnLog{})

		if len(g.Battlefield) != 2 {
			t.Fatalf("battlefield size = %d, want 2 clues", len(g.Battlefield))
		}
		clue := g.Battlefield[0]
		if !clue.Token || clue.TokenDef == nil || clue.TokenDef.Name != "Clue Token" || !clue.TokenDef.HasSubtype(types.Clue) {
			t.Fatalf("clue token = %+v def=%+v", clue, clue.TokenDef)
		}
		if len(clue.TokenDef.Abilities) != 1 {
			t.Fatalf("clue abilities = %d, want activated draw ability", len(clue.TokenDef.Abilities))
		}
		g.Players[game.Player1].ManaPool.Add(mana.Colorless, 2)
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

func TestUnsupportedSearchSpecIsLogged(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectSearch,
		TargetIndex: -1,
		Search: opt.Val(game.SearchSpec{
			SourceZone:  game.ZoneLibrary,
			Destination: game.ZoneExile,
		}),
		Description: "unsupported search destination",
	}, nil)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	if len(log.Unsupported) != 1 || log.Unsupported[0].EffectType != game.EffectSearch {
		t.Fatalf("unsupported logs = %+v, want EffectSearch unsupported log", log.Unsupported)
	}
}

func TestProliferateAddsOneChosenCounterKind(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	permanent := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	permanent.Counters.Add(counter.PlusOnePlusOne, 1)
	permanent.Counters.Add(counter.Charge, 1)
	g.Players[game.Player2].PoisonCounters = 1
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectProliferate}, nil)
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
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectGoad, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

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
		bottom := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Bottom"})
		top := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Top"})
		addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectScry, Amount: 1, TargetIndex: -1}, nil)
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
		top := addCardToLibrary(g, game.Player1, &game.CardDef{Name: "Top"})
		addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectSurveil, Amount: 1, TargetIndex: -1}, nil)
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
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectDestroy, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

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
		name        string
		effectType  game.EffectType
		destination *game.Zone
	}{
		{name: "exile", effectType: game.EffectExile, destination: nil},
		{name: "bounce", effectType: game.EffectBounce, destination: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			target := addCreaturePermanent(g, game.Player2)
			addEffectSpellToStack(g, game.Player1, game.Effect{Type: tt.effectType, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

			engine.resolveTopOfStack(g, &TurnLog{})

			if _, ok := permanentByObjectID(g, target.ObjectID); ok {
				t.Fatal("moved permanent remained on battlefield")
			}
			var zone *game.Zone
			switch tt.effectType {
			case game.EffectExile:
				zone = &g.Players[game.Player2].Exile
			case game.EffectBounce:
				zone = &g.Players[game.Player2].Hand
			}
			if zone == nil || !zone.Contains(target.CardInstanceID) {
				t.Fatalf("card was not moved to expected zone for %s", tt.name)
			}
		})
	}
}

func TestSacrificeEffectMovesControllerPermanentThroughGraveyardIgnoringIndestructible(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanent(g, game.Player1, game.Indestructible)
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectSacrifice, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})

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
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectTap, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if !target.Tapped {
		t.Fatal("tap effect did not tap permanent")
	}

	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectUntap, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if target.Tapped {
		t.Fatal("untap effect did not untap permanent")
	}
}

func TestDamageToPermanentEffectCanCauseLethalSBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectDamage, Amount: 3, TargetIndex: 0}, []game.Target{game.PermanentTarget(target.ObjectID)})
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
	artifact := addCombatPermanent(g, game.Player4, &game.CardDef{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	})
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDestroy,
		TargetIndex: -1,
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
	land := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Island",
		Types: []types.Card{types.Land},
	})
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	})
	enchantment := addCombatPermanent(g, game.Player2, &game.CardDef{
		Name:  "Aura",
		Types: []types.Card{types.Enchantment},
	})
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDestroy,
		TargetIndex: -1,
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

func TestMassDamageDeathsAreLoggedTogetherBySBA(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	creature1 := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
	creature2 := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
	artifact := addCombatPermanent(g, game.Player3, &game.CardDef{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	})
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectDamage,
		Amount:      3,
		TargetIndex: -1,
		Selector:    game.EffectSelectorAllCreatures,
	}, nil)

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
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:           game.EffectModifyPT,
		TargetIndex:    0,
		PowerDelta:     3,
		ToughnessDelta: 3,
		UntilEndOfTurn: true,
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
	for _, effect := range []game.Effect{
		{Type: game.EffectModifyPT, TargetIndex: 0, PowerDelta: 1, ToughnessDelta: 2, UntilEndOfTurn: true},
		{Type: game.EffectModifyPT, TargetIndex: 0, PowerDelta: -2, ToughnessDelta: -1, UntilEndOfTurn: true},
	} {
		addEffectSpellToStack(g, game.Player1, effect, []game.Target{game.PermanentTarget(creature.ObjectID)})
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
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	})
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectAddCounter,
		TargetIndex: 0,
		Amount:      3,
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
	source := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Source Relic",
		Types: []types.Card{types.Artifact},
	})
	destination := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Destination Relic",
		Types: []types.Card{types.Artifact},
	})
	source.Counters.Add(counter.PlusOnePlusOne, 2)
	source.Counters.Add(counter.Charge, 1)
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectMoveCounters,
		TargetIndex: 1,
		CounterSource: game.CounterSourceSpec{
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
	artifact := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:  "Relic",
		Types: []types.Card{types.Artifact},
	})
	zero := game.PT{Value: 0}
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectApplyContinuous,
		TargetIndex: 0,
		Condition: optEffectCondition(game.EffectCondition{
			Text:               "it isn't a creature",
			TargetIndex:        0,
			MatchPermanentType: true,
			PermanentType:      types.Creature,
			Negate:             true,
		}),
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer:       game.LayerType,
				AddTypes:    []types.Card{types.Creature},
				AddSubtypes: []types.Sub{types.Robot},
			},
			{
				Layer:        game.LayerPowerToughnessSet,
				SetPower:     optPT(zero),
				SetToughness: optPT(zero),
			},
		},
	}, []game.Target{game.PermanentTarget(artifact.ObjectID)})

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
	artifactCreature := addCombatPermanent(g, game.Player1, &game.CardDef{
		Name:      "Construct",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Construct},
		Power:     optPT(two),
		Toughness: optPT(two),
	})
	zero := game.PT{Value: 0}
	addEffectSpellToStack(g, game.Player1, game.Effect{
		Type:        game.EffectApplyContinuous,
		TargetIndex: 0,
		Condition: optEffectCondition(game.EffectCondition{
			Text:               "it isn't a creature",
			TargetIndex:        0,
			MatchPermanentType: true,
			PermanentType:      types.Creature,
			Negate:             true,
		}),
		ContinuousEffects: []game.ContinuousEffect{
			{
				Layer:       game.LayerType,
				AddTypes:    []types.Card{types.Creature},
				AddSubtypes: []types.Sub{types.Robot},
			},
			{
				Layer:        game.LayerPowerToughnessSet,
				SetPower:     optPT(zero),
				SetToughness: optPT(zero),
			},
		},
	}, []game.Target{game.PermanentTarget(artifactCreature.ObjectID)})

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
	token := &game.CardDef{
		Name:      "Soldier Token",
		Types:     []types.Card{types.Creature},
		Power:     optPT(game.PT{Value: 1}),
		Toughness: optPT(game.PT{Value: 1}),
	}
	addEffectSpellToStack(g, game.Player1, game.Effect{Type: game.EffectCreateToken, Amount: 2, TargetIndex: -1, Token: optToken(token)}, nil)

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

func TestCreateTokenCanCopySourceCardWithModifications(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{
			Name:      "Fanatic Source",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake, types.Druid},
			ManaCost:  optCost(mana.Cost{mana.ColoredMana(mana.Green)}),
			ManaValue: 1,
			Power:     optPT(game.PT{Value: 1}),
			Toughness: optPT(game.PT{Value: 4}),
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
	g.CardInstances[sourceID].Def.Abilities = []game.AbilityDef{
		game.EternalizeAbility(mana.Cost{mana.GenericMana(0)}, types.Snake, types.Druid),
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
	if token.TokenDef.ManaCost.Exists || token.TokenDef.ManaValue != 0 {
		t.Fatalf("token mana cost/value = %+v/%d, want no cost and mana value 0", token.TokenDef.ManaCost, token.TokenDef.ManaValue)
	}
	if got := token.TokenDef.Subtypes; !slices.Equal(got, []types.Sub{types.Zombie, types.Snake, types.Druid}) {
		t.Fatalf("token subtypes = %+v, want Zombie Snake Druid", got)
	}
	if got := token.TokenDef.Colors; !slices.Equal(got, []mana.Color{mana.Black}) {
		t.Fatalf("token colors = %+v, want black", got)
	}
	if got := effectivePower(g, token); got != 4 {
		t.Fatalf("token power = %d, want 4", got)
	}
	if got, ok := effectiveToughness(g, token); !ok || got != 4 {
		t.Fatalf("token toughness = %d ok=%v, want 4 true", got, ok)
	}
}

func TestTokenCanBlockTakeCombatDamageAndDie(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	pt := game.PT{Value: 2}
	token, ok := createTokenPermanent(g, game.Player2, &game.CardDef{
		Name:      "Bear Token",
		Types:     []types.Card{types.Creature},
		Power:     optPT(pt),
		Toughness: optPT(pt),
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

func addEffectSpellToStack(g *game.Game, controller game.PlayerID, effect game.Effect, targets []game.Target) id.ID {
	return addLinkedResultSpellToStackForController(g, controller, []game.Effect{effect}, targets)
}

func addLinkedResultSpellToStack(g *game.Game, effects []game.Effect) id.ID {
	return addLinkedResultSpellToStackForController(g, game.Player1, effects, nil)
}

func addLinkedResultSpellToStackForController(g *game.Game, controller game.PlayerID, effects []game.Effect, targets []game.Target) id.ID {
	sourceID := g.IDGen.Next()
	g.CardInstances[sourceID] = &game.CardInstance{
		ID: sourceID,
		Def: &game.CardDef{
			Name:  "Effect Spell",
			Types: []types.Card{types.Sorcery},
			Abilities: []game.AbilityDef{
				{
					Kind:    game.SpellAbility,
					Effects: append([]game.Effect(nil), effects...),
				},
			},
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
