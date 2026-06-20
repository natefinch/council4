package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDynamicEffectAmountFormulasResolveSemantically(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "First Creature",
		Types: []types.Card{types.Creature},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Second Creature",
		Types: []types.Card{types.Creature},
	}})
	addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "Opponent Creature",
		Types: []types.Card{types.Creature},
	}})
	obj := &game.StackObject{Controller: game.Player1}

	count := game.DynamicAmount{
		Kind:       game.DynamicAmountCountSelector,
		Multiplier: 2,
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Creature},
			Controller:    game.ControllerYou,
		}),
	}
	if got := dynamicAmountValue(g, obj, game.Player1, count); got != 4 {
		t.Fatalf("twice controlled creature count = %d, want 4", got)
	}
	count.Group = game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Land},
		Controller:    game.ControllerYou,
	})
	if got := dynamicAmountValue(g, obj, game.Player1, count); got != 0 {
		t.Fatalf("zero matching lands = %d, want 0", got)
	}
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Dual Land",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Plains, types.Island},
	}})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Forest",
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}})
	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind:       game.DynamicAmountControllerBasicLandTypeCount,
		Multiplier: 2,
	}); got != 6 {
		t.Fatalf("twice controlled basic land type count = %d, want 6", got)
	}

	g.Players[game.Player1].Life = 17
	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind: game.DynamicAmountControllerLife,
	}); got != 17 {
		t.Fatalf("controller life = %d, want 17", got)
	}

	g.Players[game.Player4].Eliminated = true
	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind:       game.DynamicAmountOpponentCount,
		Multiplier: 2,
	}); got != 4 {
		t.Fatalf("twice alive opponent count = %d, want 4", got)
	}
}

func TestDynamicAmountCountsCardsWithCyclingInGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Cycling Creature",
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.W}),
		},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name: "Cycling Land",
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.O(1)}),
		},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "No Cycling"}})
	addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name: "Opponent Cycling",
		ActivatedAbilities: []game.ActivatedAbility{
			game.CyclingActivatedAbility(cost.Mana{cost.U}),
		},
	}})
	obj := &game.StackObject{Controller: game.Player1}
	player := game.ControllerReference()

	got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind:     game.DynamicAmountCountCardsInZone,
		Player:   &player,
		CardZone: zone.Graveyard,
		Selection: &game.Selection{
			Keyword: game.Cycling,
		},
	})

	if got != 2 {
		t.Fatalf("cycling cards in controller graveyard = %d, want 2", got)
	}
}

func TestModalResolutionUsesEachModesOwnTargets(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	content := game.AbilityContent{
		Modes: []game.Mode{
			{
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "player"}},
				Sequence: []game.Instruction{{
					Primitive: game.GainLife{Amount: game.Fixed(1), Player: game.TargetPlayerReference(0)},
				}},
			},
			{
				Targets: []game.TargetSpec{{MinTargets: 1, MaxTargets: 1, Constraint: "player"}},
				Sequence: []game.Instruction{{
					Primitive: game.GainLife{Amount: game.Fixed(2), Player: game.TargetPlayerReference(0)},
				}},
			},
		},
		MinModes: 2,
		MaxModes: 2,
	}
	obj := &game.StackObject{
		Controller:   game.Player1,
		Targets:      []game.Target{game.PlayerTarget(game.Player2), game.PlayerTarget(game.Player3)},
		TargetCounts: []int{1, 1},
		ChosenModes:  []int{0, 1},
	}

	engine.resolveAbilityContentWithChoices(g, obj, content, [game.NumPlayers]PlayerAgent{}, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 41 {
		t.Fatalf("first mode target life = %d, want 41", got)
	}
	if got := g.Players[game.Player3].Life; got != 42 {
		t.Fatalf("second mode target life = %d, want 42", got)
	}
	if len(obj.Targets) != 2 {
		t.Fatalf("stack targets were not restored: %+v", obj.Targets)
	}
}

func TestCantGainLifeRuleEffectStopsLifeGainAndLifelink(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "No Lifegain",
		Types: []types.Card{types.Enchantment},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCantGainLife,
				AffectedPlayer: game.PlayerAny,
			}},
		}}},
	})
	addEffectSpellToStack(g, game.Player1, game.GainLife{
		Amount: game.Fixed(3),
		Player: game.ControllerReference(),
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
		Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountControllerHandSize}),
		Player: game.ControllerReference(),
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
		Recipient: game.AnyTargetDamageRecipient(0),
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
		Amount:    game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountTargetPower, Object: game.TargetPermanentReference(0)}),
		Recipient: game.AnyTargetDamageRecipient(0),
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
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{Primitive: game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)}, PublishResult: "that-much"},
					{Primitive: game.LoseLife{
						Player: game.TargetPlayerReference(0),
						Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "that-much"}),
					}},
				},
			}.Ability())},
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
				Player: game.ControllerReference(),
				Amount: game.Fixed(1),
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
				Player: game.ControllerReference(),
				Amount: game.Fixed(1),
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
		{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}, Optional: true, PublishResult: "choice"},
		{
			Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "choice", Accepted: game.TriTrue, Succeeded: game.TriTrue}),
		},
		{
			Primitive:  game.LoseLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
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
		{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}, PublishResult: "draw"},
		{
			Primitive:  game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(3)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "draw", Succeeded: game.TriTrue}),
		},
		{
			Primitive:  game.LoseLife{Player: game.ControllerReference(), Amount: game.Fixed(2)},
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
			Primitive:     game.GainLife{Player: game.ControllerReference(), Amount: game.Fixed(5)},
			Optional:      true,
			PublishResult: "amount",
		},
		{
			Primitive: game.LoseLife{
				Player: game.ControllerReference(),
				Amount: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountPreviousEffectResult, ResultKey: "amount"}),
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
		Types:         []types.Card{types.Land},
		ManaAbilities: []game.ManaAbility{game.TapManaCommanderIdentityAbility()},
	}}
}

func TestPayLifeCommanderColorIdentityCostLosesLifePerColor(t *testing.T) {
	tests := []struct {
		name     string
		identity color.Identity
		wantLife int
		wantPaid bool
	}{
		{name: "two colors", identity: color.NewIdentity(color.Blue, color.Black), wantLife: 38, wantPaid: true},
		{name: "colorless", identity: color.NewIdentity(), wantLife: 40, wantPaid: true},
		{name: "five colors", identity: color.NewIdentity(color.White, color.Blue, color.Black, color.Red, color.Green), wantLife: 35, wantPaid: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{
				game.Player1: {
					Commander: &game.CardDef{CardFace: game.CardFace{Name: "Test Commander",
						Types: []types.Card{types.Creature}}, ColorIdentity: tt.identity,
					},
				},
			})
			setSorcerySpeedTurn(g, game.Player1)
			g.Players[game.Player1].ManaPool.Add(mana.C, 3)
			land := addCombatPermanent(g, game.Player1, warRoomLikeLand())
			engine := NewEngine(nil)

			got := engine.applyActionWithChoices(g, game.Player1,
				action.ActivateAbility(land.ObjectID, 0, nil, 0), [game.NumPlayers]PlayerAgent{}, &TurnLog{})
			if got != tt.wantPaid {
				t.Fatalf("activate War Room-like draw ability = %v, want %v", got, tt.wantPaid)
			}
			if life := g.Players[game.Player1].Life; life != tt.wantLife {
				t.Fatalf("life = %d, want %d (lost %d per color identity)", life, tt.wantLife, 40-tt.wantLife)
			}
		})
	}
}

func warRoomLikeLand() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: "War Room-like Land",
		Types: []types.Card{types.Land},
		ActivatedAbilities: []game.ActivatedAbility{{
			ManaCost: opt.Val(cost.Mana{cost.O(3)}),
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalTap},
				{
					Kind:          cost.AdditionalPayLife,
					Text:          "Pay life equal to the number of colors in your commanders' color identity",
					AmountDynamic: cost.AdditionalDynamicCommanderColorIdentityCount,
				},
			},
			ZoneOfFunction: zone.Battlefield,
			Content: game.Mode{Sequence: []game.Instruction{{Primitive: game.Draw{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}.Ability(),
		}},
	}}
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
			Primitive:  game.Draw{Amount: game.Fixed(1), Player: game.ControllerReference()},
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

func TestDynamicAmountEventCardCountReadsTriggerBatch(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	// A "one or more" discard trigger coalesces a simultaneous batch into one
	// trigger, retaining the first matching event as TriggerEvent.
	simultaneousID := g.IDGen.Next()
	first := game.Event{Kind: game.EventCardDiscarded, Player: game.Player1, SimultaneousID: simultaneousID, Amount: 1}
	emitEvent(g, first)
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player1, SimultaneousID: simultaneousID, Amount: 1})
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player1, SimultaneousID: simultaneousID, Amount: 1})
	// An unrelated discard by another player must not inflate the count.
	emitEvent(g, game.Event{Kind: game.EventCardDiscarded, Player: game.Player2, SimultaneousID: g.IDGen.Next(), Amount: 1})

	obj := &game.StackObject{Controller: game.Player1, HasTriggerEvent: true, TriggerEvent: first}

	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind: game.DynamicAmountEventCardCount,
	}); got != 3 {
		t.Fatalf("event card count = %d, want 3", got)
	}
	if got := dynamicAmountValue(g, obj, game.Player1, game.DynamicAmount{
		Kind:       game.DynamicAmountEventCardCount,
		Multiplier: 2,
	}); got != 6 {
		t.Fatalf("twice event card count = %d, want 6", got)
	}

	// A single-card discard (no simultaneous batch) counts the one event.
	single := game.Event{Kind: game.EventCardDiscarded, Player: game.Player1, Amount: 1}
	emitEvent(g, single)
	soloObj := &game.StackObject{Controller: game.Player1, HasTriggerEvent: true, TriggerEvent: single}
	if got := dynamicAmountValue(g, soloObj, game.Player1, game.DynamicAmount{
		Kind: game.DynamicAmountEventCardCount,
	}); got != 1 {
		t.Fatalf("single discard event card count = %d, want 1", got)
	}

	// Without a triggering event there is no count.
	if got := dynamicAmountValue(g, &game.StackObject{Controller: game.Player1}, game.Player1, game.DynamicAmount{
		Kind: game.DynamicAmountEventCardCount,
	}); got != 0 {
		t.Fatalf("event card count without trigger = %d, want 0", got)
	}
}

// TestOptionalIfYouDoFlowSkipsGatedEffectWhenDeclined mirrors the exact wiring
// the executable backend emits for "You may discard a card. If you do, draw a
// card." (issue #364): an optional discard that publishes its result and a draw
// gated on that discard succeeding. Declining the discard must skip the draw.
func TestOptionalIfYouDoFlowSkipsGatedEffectWhenDeclined(t *testing.T) {
	sequence := []game.Instruction{
		{
			Primitive:     game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			Optional:      true,
			PublishResult: "if-you-do",
		},
		{
			Primitive:  game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
			ResultGate: opt.Val(game.InstructionResultGate{Key: "if-you-do", Succeeded: game.TriTrue}),
		},
	}

	t.Run("declined skips draw", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "In Hand"}})
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "In Library"}})
		addInstructionSpellToStack(g, sequence)
		agents := [game.NumPlayers]PlayerAgent{
			game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
		}

		engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

		if got := g.Players[game.Player1].Hand.Size(); got != 1 {
			t.Fatalf("hand size = %d, want declined discard kept and gated draw skipped", got)
		}
		if got := g.Players[game.Player1].Library.Size(); got != 1 {
			t.Fatalf("library size = %d, want gated draw to be skipped", got)
		}
	})

	t.Run("accepted performs draw", func(t *testing.T) {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "In Hand"}})
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "In Library"}})
		addInstructionSpellToStack(g, sequence)

		engine.resolveTopOfStack(g, &TurnLog{})

		if got := g.Players[game.Player1].Hand.Size(); got != 1 {
			t.Fatalf("hand size = %d, want discard then gated draw", got)
		}
		if got := g.Players[game.Player1].Library.Size(); got != 0 {
			t.Fatalf("library size = %d, want gated draw to fire", got)
		}
	})
}
