package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestManaVaultUpkeepPaymentChoicesAndOrdering(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		canPay          bool
		accepts         bool
		wantUntapped    bool
		wantMana        int
		wantChoiceCount int
	}{
		{name: "pays then untaps", canPay: true, accepts: true, wantUntapped: true, wantChoiceCount: 1},
		{name: "declines and stays tapped", canPay: true, wantMana: 4, wantChoiceCount: 1},
		{name: "cannot pay and stays tapped", wantChoiceCount: 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			vault := addCombatPermanent(g, game.Player1, manaVaultCardDef())
			vault.Tapped = true
			if tc.canPay {
				g.Players[game.Player1].ManaPool.Add(mana.C, 4)
			}
			emitEvent(g, beginningStepEvent(game.Player1, game.StepUpkeep))
			if !engine.putTriggeredAbilitiesOnStack(g) {
				t.Fatal("upkeep trigger was not put on the stack")
			}
			choice := []int{0}
			if tc.accepts {
				choice = []int{1}
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{choice}},
			}
			log := TurnLog{}
			engine.resolveTopOfStackWithChoices(g, agents, &log)

			if got := !vault.Tapped; got != tc.wantUntapped {
				t.Fatalf("untapped = %v, want %v", got, tc.wantUntapped)
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != tc.wantMana {
				t.Fatalf("mana = %d, want %d", got, tc.wantMana)
			}
			if len(log.Choices) != tc.wantChoiceCount {
				t.Fatalf("choices = %+v, want %d", log.Choices, tc.wantChoiceCount)
			}
			if len(log.Choices) != 0 && log.Choices[0].Request.Player != game.Player1 {
				t.Fatalf("choice player = %v, want Player1", log.Choices[0].Request.Player)
			}
		})
	}
}

func TestManaVaultUpkeepTriggerUsesCapturedControllerAndSourceLKI(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vault := addCombatPermanent(g, game.Player1, manaVaultCardDef())
	vault.Tapped = true
	g.Players[game.Player1].ManaPool.Add(mana.C, 4)

	emitEvent(g, beginningStepEvent(game.Player1, game.StepUpkeep))
	if !movePermanentToZone(g, vault, zone.Graveyard) {
		t.Fatal("moving Mana Vault before trigger stacking failed")
	}
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("captured upkeep trigger was lost with its source")
	}
	agents := [game.NumPlayers]PlayerAgent{
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)

	if got := g.Players[game.Player1].ManaPool.Total(); got != 0 {
		t.Fatalf("mana = %d, want payment by captured controller", got)
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Player != game.Player1 {
		t.Fatalf("choices = %+v, want original controller payment choice", log.Choices)
	}
}

func TestManaVaultDrawTriggerChecksTappedStateTwice(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vault := addCombatPermanent(g, game.Player1, manaVaultCardDef())

	emitEvent(g, beginningStepEvent(game.Player1, game.StepDraw))
	if engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("untapped Mana Vault triggered at the beginning of the draw step")
	}

	vault.Tapped = true
	emitEvent(g, beginningStepEvent(game.Player1, game.StepDraw))
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("tapped Mana Vault did not trigger")
	}
	vault.Tapped = false
	log := TurnLog{}
	engine.resolveTopOfStack(g, &log)
	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("life = %d, want no damage after Mana Vault became untapped", got)
	}
	if len(log.Resolves) != 1 || log.Resolves[0].Result != "intervening if false" {
		t.Fatalf("resolve log = %+v, want intervening-if false", log.Resolves)
	}
}

func TestManaVaultDamageUsesAbilityControllerAndPermanentSource(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	vault := addCombatPermanent(g, game.Player1, manaVaultCardDef())
	vault.Tapped = true

	emitEvent(g, beginningStepEvent(game.Player1, game.StepDraw))
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("draw-step trigger was not put on the stack")
	}
	vault.Controller = game.Player2
	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 39 {
		t.Fatalf("original ability controller life = %d, want 39", got)
	}
	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("new permanent controller life = %d, want 40", got)
	}
	assertEvent(t, g.Events, game.EventDamageDealt, func(event game.Event) bool {
		return event.Player == game.Player1 &&
			event.Amount == 1 &&
			event.SourceID == vault.CardInstanceID &&
			event.SourceObjectID == vault.ObjectID
	})
}

func beginningStepEvent(player game.PlayerID, step game.Step) game.Event {
	return game.Event{
		Kind:       game.EventBeginningOfStep,
		Controller: player,
		Player:     player,
		Step:       step,
	}
}

func manaVaultCardDef() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Mana Vault",
		Types: []types.Card{types.Artifact},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectDoesntUntap,
				AffectedSource: true,
			}},
		}},
		TriggeredAbilities: []game.TriggeredAbility{
			{
				Trigger: game.TriggerCondition{
					Type: game.TriggerAt,
					Pattern: game.TriggerPattern{
						Event:      game.EventBeginningOfStep,
						Controller: game.TriggerControllerYou,
						Step:       game.StepUpkeep,
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{
					{
						Primitive: game.Pay{Payment: game.ResolutionPayment{
							Prompt:   "Pay {4}?",
							ManaCost: opt.Val(cost.Mana{cost.O(4)}),
						}},
						PublishResult: "controller-paid",
					},
					{
						Primitive: game.Untap{Object: game.SourcePermanentReference()},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "controller-paid",
							Succeeded: game.TriTrue,
						}),
					},
				}}.Ability(),
			},
			{
				Trigger: game.TriggerCondition{
					Type: game.TriggerAt,
					Pattern: game.TriggerPattern{
						Event:      game.EventBeginningOfStep,
						Controller: game.TriggerControllerYou,
						Step:       game.StepDraw,
					},
					InterveningCondition: opt.Val(game.Condition{
						Object: opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{
							RequiredTypes: []types.Card{types.Artifact},
							Tapped:        game.TriTrue,
						}),
					}),
				},
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.Damage{
						Amount:       game.Fixed(1),
						Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
						DamageSource: opt.Val(game.SourcePermanentReference()),
					},
				}}}.Ability(),
			},
		},
	}}
}
