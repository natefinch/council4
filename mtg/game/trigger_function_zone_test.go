package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

func TestValidateTriggeredFunctionZone(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*TriggeredAbility)
		valid  bool
	}{
		{"graveyard upkeep", func(*TriggeredAbility) {}, true},
		{"graveyard combat", func(a *TriggeredAbility) { a.Trigger.Pattern.Step = StepBeginningOfCombat }, true},
		{"graveyard end", func(a *TriggeredAbility) { a.Trigger.Pattern.Step = StepEnd }, true},
		{"battlefield default", func(a *TriggeredAbility) { a.ZoneOfFunction = zone.None }, true},
		{"explicit battlefield", func(a *TriggeredAbility) { a.ZoneOfFunction = zone.Battlefield }, true},
		{"wrong zone", func(a *TriggeredAbility) { a.ZoneOfFunction = zone.Hand }, false},
		{"wrong event", func(a *TriggeredAbility) { a.Trigger.Pattern.Event = EventPermanentDied }, false},
		{"wrong trigger", func(a *TriggeredAbility) { a.Trigger.Type = TriggerWhenever }, false},
		{"unproved step", func(a *TriggeredAbility) { a.Trigger.Pattern.Step = StepDraw }, false},
		{"permanent source", func(a *TriggeredAbility) { a.Trigger.Pattern.Source = TriggerSourceSelf }, false},
		{"attached source", func(a *TriggeredAbility) { a.Trigger.Pattern.Source = TriggerSourceAttachedPermanent }, false},
		{"enchanted player", func(a *TriggeredAbility) { a.Trigger.Pattern.StepPlayerIsSourceEnchantedPlayer = true }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ability := TriggeredAbility{
				ZoneOfFunction: zone.Graveyard,
				Trigger:        TriggerCondition{Type: TriggerAt, Pattern: TriggerPattern{Event: EventBeginningOfStep, Step: StepUpkeep}},
				Content:        Mode{Sequence: []Instruction{{Primitive: GainLife{Amount: Fixed(1), Player: ControllerReference()}}}}.Ability(),
			}
			tc.mutate(&ability)
			card := &CardDef{CardFace: CardFace{Name: "Zone Proof", TriggeredAbilities: []TriggeredAbility{ability}}}
			issues := ValidateCardDef(card)
			found := hasCardDefIssue(issues, CardDefIssueInvalidAbilityBody)
			if found == tc.valid {
				t.Fatalf("valid=%v issues=%+v", tc.valid, issues)
			}
		})
	}
}
