package game

import (
	"errors"
	"strings"
)

// ProgramReference is a serializable binding to installed, versioned Go code.
// It contains no function, live game pointer or execution continuation.
type ProgramReference struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// IsZero reports that this mode uses ordinary declarative instructions.
func (p ProgramReference) IsZero() bool {
	return p == (ProgramReference{})
}

// Validate requires an explicit, unambiguous code identity.
func (p ProgramReference) Validate() error {
	if p.ID == "" || p.Version == "" || strings.ContainsAny(p.ID+p.Version, " \t\r\n") {
		return errors.New("program binding requires an ID and version without whitespace")
	}
	return nil
}

// ProgramReferences reports code required by a resolving body.
func (m *AbilityContent) ProgramReferences() []ProgramReference {
	var references []ProgramReference
	for _, mode := range m.Modes {
		if !mode.Program.IsZero() {
			references = append(references, mode.Program)
		}
		references = append(references, instructionPrograms(mode.Sequence)...)
	}
	return references
}

func instructionPrograms(sequence []Instruction) []ProgramReference {
	var references []ProgramReference
	for _, instruction := range sequence {
		if nested, ok := instruction.Primitive.(interface{ ProgramReferences() []ProgramReference }); ok {
			references = append(references, nested.ProgramReferences()...)
		}
		references = append(references, instructionPrograms(instruction.TemptingOfferBody)...)
	}
	return references
}

// ProgramReferences reports a repeated body's dependencies.
func (p RepeatProcess) ProgramReferences() []ProgramReference {
	return p.Body.ProgramReferences()
}

// ProgramReferences reports a delayed body's dependencies.
func (p CreateDelayedTrigger) ProgramReferences() []ProgramReference {
	return p.Trigger.Content.ProgramReferences()
}

// ProgramReferences reports a reflexive body's dependencies.
func (p CreateReflexiveTrigger) ProgramReferences() []ProgramReference {
	return p.Trigger.Content.ProgramReferences()
}

// ProgramReferences reports an emblem's dependencies.
func (p CreateEmblem) ProgramReferences() []ProgramReference {
	return abilityPrograms(p.EmblemAbilities)
}

func keywordPrograms(keywords []KeywordAbility) []ProgramReference {
	var references []ProgramReference
	for _, keyword := range keywords {
		if body, ok := keyword.(interface{ ProgramReferences() []ProgramReference }); ok {
			references = append(references, body.ProgramReferences()...)
		}
	}
	return references
}

// ProgramReferences reports kicker bonus code.
func (ability KickerKeyword) ProgramReferences() []ProgramReference {
	return ability.BonusContent.ProgramReferences()
}

// ProgramReferences reports gift delivery code.
func (ability GiftKeyword) ProgramReferences() []ProgramReference {
	return ability.Delivery.ProgramReferences()
}

func continuousPrograms(effects []ContinuousEffect) []ProgramReference {
	var references []ProgramReference
	for _, effect := range effects {
		references = append(references, abilityPrograms(effect.AddAbilities)...)
		if effect.CopyValues.Exists {
			references = append(references, abilityPrograms(effect.CopyValues.Val.Abilities)...)
		}
	}
	return references
}

// ProgramReferences reports code introduced by a resolving ability grant.
func (p ApplyContinuous) ProgramReferences() []ProgramReference {
	return continuousPrograms(p.ContinuousEffects)
}

// ProgramReferences reports code granted as a permanent enters.
func (p PutOnBattlefield) ProgramReferences() []ProgramReference {
	return continuousPrograms(p.ContinuousEffects)
}

// ProgramReferences reports code on a concrete token definition.
func (s TokenSource) ProgramReferences() []ProgramReference {
	if def, exists := s.TokenDefRef(); exists {
		return def.ProgramReferences()
	}
	return nil
}

// ProgramReferences reports code required by created token bodies.
func (p CreateToken) ProgramReferences() []ProgramReference { return p.Source.ProgramReferences() }

// ProgramReferences reports code required by dice-created token bodies.
func (p RollDiceCreateTokens) ProgramReferences() []ProgramReference {
	return p.Source.ProgramReferences()
}

// ProgramReferences reports code required by destruction-correlated tokens.
func (p CreateTokenForEachDestroyed) ProgramReferences() []ProgramReference {
	return p.Source.ProgramReferences()
}

// ProgramReferences reports the action and code granted by a replacement.
func (r *ReplacementEffect) ProgramReferences() []ProgramReference {
	references := abilityPrograms(r.EntersAsCopyAddAbilities)
	if r.Program != (ProgramReference{}) {
		references = append(references, r.Program)
	}
	return references
}

// ProgramReferences reports code required by a replacement ability.
func (a *ReplacementAbility) ProgramReferences() []ProgramReference {
	return a.Replacement.ProgramReferences()
}

// ProgramReferences reports code in a resolving replacement declaration.
func (p CreateReplacement) ProgramReferences() []ProgramReference {
	if p.Replacement != nil {
		return p.Replacement.ProgramReferences()
	}
	return nil
}

func abilityPrograms(abilities []Ability) []ProgramReference {
	var references []ProgramReference
	for _, ability := range abilities {
		if body, ok := ability.(interface{ ProgramReferences() []ProgramReference }); ok {
			references = append(references, body.ProgramReferences()...)
		}
	}
	return references
}

// ProgramReferences reports an activation's code dependencies.
func (a *ActivatedAbility) ProgramReferences() []ProgramReference {
	return append(a.Content.ProgramReferences(), keywordPrograms(a.KeywordAbilities)...)
}

// ProgramReferences reports a mana ability's code dependencies.
func (a *ManaAbility) ProgramReferences() []ProgramReference { return a.Content.ProgramReferences() }

// ProgramReferences reports a loyalty ability's code dependencies.
func (a *LoyaltyAbility) ProgramReferences() []ProgramReference { return a.Content.ProgramReferences() }

// ProgramReferences reports a triggered ability's code dependencies.
func (a *TriggeredAbility) ProgramReferences() []ProgramReference {
	return append(a.Content.ProgramReferences(), keywordPrograms(a.KeywordAbilities)...)
}

// ProgramReferences reports a chapter's code dependencies.
func (a *ChapterAbility) ProgramReferences() []ProgramReference { return a.Content.ProgramReferences() }

// ProgramReferences reports code in granted resolving abilities.
func (a *StaticAbility) ProgramReferences() []ProgramReference {
	return append(continuousPrograms(a.ContinuousEffects), keywordPrograms(a.KeywordAbilities)...)
}

// ProgramReferences reports code used by the printed faces and overload body.
func (c *CardDef) ProgramReferences() []ProgramReference {
	var references []ProgramReference
	for _, face := range []*CardFace{&c.CardFace, faceIfPresent(c.Back.Exists, &c.Back.Val), faceIfPresent(c.Alternate.Exists, &c.Alternate.Val)} {
		if face == nil {
			continue
		}
		for index := range face.AbilityCount() {
			if body, ok := face.BodyAt(index).(interface{ ProgramReferences() []ProgramReference }); ok {
				references = append(references, body.ProgramReferences()...)
			}
		}
		if face.Overload.Exists {
			references = append(references, face.Overload.Val.SpellAbility.ProgramReferences()...)
		}
	}
	return references
}

func faceIfPresent(exists bool, face *CardFace) *CardFace {
	if exists {
		return face
	}
	return nil
}

// ValidatePrimitive checks an authored operation against its declared target scope.
// The same sealed primitive validators are used by declarative Card Definitions.
func ValidatePrimitive(primitive Primitive, targets []TargetSpec) error {
	if primitive == nil {
		return errors.New("authored operation has no primitive")
	}
	return primitive.validatePrimitive(targets, true)
}
