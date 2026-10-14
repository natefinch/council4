package game

import (
	"errors"
	"reflect"

	"github.com/natefinch/council4/opt"
)

// ValidateProgram rejects unsupported or mixed authored replacement actions.
func (r *ReplacementEffect) ValidateProgram() error {
	if r.Program == (ProgramReference{}) {
		if r.Optional {
			return errors.New("optional replacement requires a Program")
		}
		return nil
	}
	if err := r.Program.Validate(); err != nil {
		return err
	}
	if r.MatchEvent != EventCardDrawn {
		return errors.New("replacement programs currently require EventCardDrawn")
	}
	// Applicability is data; every other action field must remain unset.
	residue := *r
	residue.ID, residue.SourceObjectID, residue.SourceCardID = 0, 0, 0
	residue.Controller, residue.CreatedTurn = 0, 0
	residue.Description, residue.Duration = "", 0
	residue.MatchEvent, residue.ControllerFilter = 0, 0
	residue.Program = ProgramReference{}
	residue.Optional, residue.SelfReplacement = false, false
	residue.DrawCardExceptFirstInDrawStep = false
	residue.FromSourceAbility = false
	residue.Condition = opt.V[Condition]{}
	residue.MatchFromZone, residue.MatchToZone = false, false
	residue.FromZone, residue.ToZone = 0, 0
	if !reflect.DeepEqual(residue, ReplacementEffect{}) {
		return errors.New("replacement Program cannot also declare another replacement action")
	}
	return nil
}
