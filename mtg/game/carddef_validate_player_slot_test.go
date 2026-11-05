package game

import (
	"testing"

	"github.com/natefinch/council4/opt"
)

// TestValidateCardDefPlayerSlotTargets requires per-player target slots
// ("up to one target ... that player controls") to name a real seat, target
// only permanents, be chosen by the controller and hold at most one target,
// and requires a targeted ForEachPlayer to take every target from such slots
// rather than a resolution-time choice.
func TestValidateCardDefPlayerSlotTargets(t *testing.T) {
	targeted := ForEachPlayer{
		Scope:     AllPlayersReference(),
		Removal:   DistributiveRemovalExile,
		LinkedKey: "exiled",
		Targeted:  true,
	}
	slots := PlayerSlotTargetSpecs(Selection{})
	badSeat := PlayerSlotTargetSpecs(Selection{})
	badSeat[0].ControlledBy = opt.Val(PlayerID(NumPlayers))
	twoTargets := PlayerSlotTargetSpecs(Selection{})
	twoTargets[1].MaxTargets = 2
	playerAllowed := PlayerSlotTargetSpecs(Selection{})
	playerAllowed[2].Allow = TargetAllowPermanent | TargetAllowPlayer
	plain := []TargetSpec{{MinTargets: 0, MaxTargets: 1, Allow: TargetAllowPermanent, Selection: opt.Val(Selection{})}}
	withChooser := targeted
	withChooser.Chooser = ControllerReference()

	tests := []struct {
		name    string
		targets []TargetSpec
		prim    ForEachPlayer
		valid   bool
	}{
		{"player slots", slots, targeted, true},
		{"unknown seat", badSeat, targeted, false},
		{"two targets in a slot", twoTargets, targeted, false},
		{"slot allows players", playerAllowed, targeted, false},
		{"targeted without slots", plain, targeted, false},
		{"targeted with chooser", slots, withChooser, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			card := &CardDef{CardFace: CardFace{
				Name: "Player Slots",
				TriggeredAbilities: []TriggeredAbility{{
					Content: Mode{
						Targets:  test.targets,
						Sequence: []Instruction{{Primitive: test.prim}},
					}.Ability(),
					Trigger: TriggerCondition{Pattern: TriggerPattern{Event: EventPermanentEnteredBattlefield, Source: TriggerSourceSelf}},
				}},
			}}
			issues := ValidateCardDef(card)
			if test.valid && len(issues) != 0 {
				t.Fatalf("issues = %+v, want none", issues)
			}
			if !test.valid && len(issues) == 0 {
				t.Fatal("issues = none, want a rejection")
			}
		})
	}
}
