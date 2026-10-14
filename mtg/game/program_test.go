package game

import (
	"encoding/json"
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestProgramDeclarationsSerializeAsDataAndValidate(t *testing.T) {
	t.Parallel()
	ref := ProgramReference{ID: "test/body", Version: "1"}
	def := &CardDef{CardFace: CardFace{Name: "Program", Types: []types.Card{types.Sorcery},
		SpellAbility: opt.Val(Mode{Program: ref}.Ability())}}
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatal(err)
	}
	var restored CardDef
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.ProgramReferences()[0] != ref || len(ValidateCardDef(&restored)) != 0 {
		t.Fatal("binding data did not round-trip")
	}
	for _, mode := range []Mode{
		{Program: ProgramReference{ID: ref.ID}},
		{Program: ref, Sequence: []Instruction{{Primitive: Draw{Amount: Fixed(1), Player: ControllerReference()}}}},
	} {
		def.SpellAbility = opt.Val(mode.Ability())
		if len(ValidateCardDef(def)) == 0 {
			t.Fatal("malformed/mixed program declaration was accepted")
		}
	}
}

func TestProgramReferencesIncludeDelayedAndGrantedBodies(t *testing.T) {
	t.Parallel()
	ref := ProgramReference{ID: "test/nested", Version: "1"}
	body := Mode{Program: ref}.Ability()
	def := &CardDef{CardFace: CardFace{Name: "Program", StaticAbilities: []StaticAbility{{
		ContinuousEffects: []ContinuousEffect{{AddAbilities: []Ability{&TriggeredAbility{Content: body}}}},
		KeywordAbilities:  []KeywordAbility{KickerKeyword{BonusContent: body}},
	}}}}
	if len(def.ProgramReferences()) != 2 {
		t.Fatal("granted or kicker code was omitted from preflight")
	}
	content := Mode{Sequence: []Instruction{{Primitive: CreateReflexiveTrigger{
		Trigger: ReflexiveTriggerDef{Content: body},
	}}}}.Ability()
	if content.ProgramReferences()[0] != ref {
		t.Fatal("nested code was omitted from preflight")
	}
	token := &CardDef{CardFace: CardFace{Name: "Token", TriggeredAbilities: []TriggeredAbility{{Content: body}}}}
	creators := []Primitive{
		CreateToken{Amount: Fixed(1), Source: TokenDef(token)},
		PutOnBattlefield{ContinuousEffects: []ContinuousEffect{{AddAbilities: []Ability{&TriggeredAbility{Content: body}}}}},
	}
	for _, primitive := range creators {
		content := Mode{Sequence: []Instruction{{Primitive: primitive}}}.Ability()
		if len(content.ProgramReferences()) != 1 || content.ProgramReferences()[0] != ref {
			t.Fatal("introduced ability code was omitted from preflight")
		}
	}
	def.StaticAbilities = nil
	def.ReplacementAbilities = []ReplacementAbility{{Replacement: ReplacementEffect{
		EntersAsCopyAddAbilities: []Ability{&TriggeredAbility{Content: body}},
	}}}
	if len(def.ProgramReferences()) != 1 {
		t.Fatal("copy-entry code was omitted from preflight")
	}
}
