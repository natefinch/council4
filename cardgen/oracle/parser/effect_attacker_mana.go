package parser

import (
	"reflect"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// triggerIsControlledCreaturesAttack reports whether the ability triggers on
// "Whenever one or more creatures you control attack" with a plain controlled-
// creature subject: an attacker-declared event, the one-or-more plural form, a
// "you control" controller relation, and a subject selection carrying no filter
// beyond the creature card type. The plain-creature requirement guarantees the
// trigger fires on the count of every attacking creature the controller
// controls, so a "that much" mana amount can safely bind to that same all-
// creatures attacker count (recognizeThatMuchCombinationMana). A subject narrowed
// by any extra filter (a subtype, color, keyword, and so on) would make "that
// much" count only the matching attackers, so it fails closed here.
func triggerIsControlledCreaturesAttack(trigger *TriggerClause) bool {
	if trigger == nil || trigger.TriggerEvent == nil {
		return false
	}
	event := trigger.TriggerEvent
	if event.Kind != TriggerEventKindAttack ||
		!event.OneOrMore ||
		event.Controller != ControllerYou ||
		event.Subject.Kind != TriggerEventSubjectSelection {
		return false
	}
	return event.Subject.Selection.equalsPlainCreatures()
}

// equalsPlainCreatures reports whether the trigger selection constrains its
// subject to exactly the creature card type with no other filter. It is the
// fail-closed guard behind an all-attacking-creatures "that much" amount: any
// additional constraint (a subtype, color, keyword, power, and so on) makes the
// selection unequal to the bare creature type, so the amount stays fail-closed.
func (s TriggerSelection) equalsPlainCreatures() bool {
	return reflect.DeepEqual(s, TriggerSelection{
		RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
	})
}

// recognizeThatMuchCombinationMana types the add-mana body "add that much mana
// in any combination of <colors>" (Grand Warlord Radha) whose amount is the
// anaphoric "that much" referring to the attacking-creatures count named by the
// ability's "Whenever one or more creatures you control attack" trigger. The
// parser leaves the body unrecognized because "that much" is not a standalone
// dynamic amount; this credits the amount as the number of attacking creatures
// the controller controls (the same typed count the explicit "where X is the
// number of attacking creatures you control" wording produces) and types the
// freely-split combination output, so the lowerer can add that many mana split
// among the colors. It fires only when the trigger guard holds and the body
// matches exactly, so an unmodeled wording stays fail-closed.
func recognizeThatMuchCombinationMana(ability *Ability) {
	if !triggerIsControlledCreaturesAttack(ability.Trigger) {
		return
	}
	for si := range ability.Sentences {
		for ei := range ability.Sentences[si].Effects {
			effect := &ability.Sentences[si].Effects[ei]
			if effect.Kind != EffectAddMana || effect.Exact {
				continue
			}
			body := manaBodyAfterVerb(effect)
			if len(body) <= 7 ||
				!effectWordsAt(body, 0, "that", "much", "mana", "in", "any", "combination", "of") {
				continue
			}
			colors, ok := combinationManaColorList(body[7:])
			if !ok {
				continue
			}
			amountSpan := shared.SpanOf(body[:2])
			effect.Amount = EffectAmountSyntax{
				Span:        amountSpan,
				Text:        "that much",
				DynamicKind: EffectDynamicAmountCount,
				DynamicForm: EffectDynamicAmountFormWhereX,
				Multiplier:  1,
				Selection: &SelectionSyntax{
					Span:             amountSpan,
					Text:             "that much",
					Kind:             SelectionCreature,
					Controller:       SelectionControllerYou,
					Attacking:        true,
					RequiredTypesAny: []CardType{CardTypeCreature},
				},
			}
			effect.Mana = EffectManaSyntax{
				Span:               shared.SpanOf(body),
				Combination:        true,
				CombinationColors:  colors,
				CombinationDynamic: true,
			}
			effect.Exact = true
		}
	}
}

// isPersistentManaRiderTokens reports whether the sentence tokens are the "Until
// end of turn, you don't lose this mana as steps and phases end." rider (Grand
// Warlord Radha), matched punctuation-agnostically by its word sequence. The
// rider marks the mana added by the preceding add-mana effect as persistent for
// the rest of the turn (it does not empty as steps and phases end).
func isPersistentManaRiderTokens(tokens []shared.Token) bool {
	return slices.Equal(normalizedWords(tokens), []string{
		"until", "end", "of", "turn",
		"you", "don't", "lose", "this", "mana",
		"as", "steps", "and", "phases", "end",
	})
}

// foldPersistentManaRider folds the "Until end of turn, you don't lose this mana
// as steps and phases end." rider sentence onto the ability's lone add-mana
// effect (Grand Warlord Radha, Savage Ventmaw, Brazen Collector). It records that
// the produced mana persists until end of turn plus a coverage span on the
// add-mana, clears the rider sentence's own "lose this mana" effect, and marks the
// rider sentence so reference and coverage scans credit its tokens to the add-mana
// rather than flagging it as an unrecognized sibling. Because the rider is folded
// rather than a distinct ordered effect, it also clears the add-mana's
// RequiresOrderedLowering flag, which was set only because the ability held the
// add-mana and rider sentences as two legacy effects. It credits only when the
// ability's sole effect-bearing sentences are one lone add-mana body and the
// trailing rider, so a body mixed with other effects (add mana and gain life) or
// any other shape stays fail-closed; the add-mana lowering and its persist tagging
// then fail closed on their own for any mana shape that cannot carry the flag.
func foldPersistentManaRider(ability *Ability) {
	sentences := ability.Sentences
	addIndex := -1
	riderIndex := -1
	for si := range sentences {
		if len(sentences[si].Effects) == 0 {
			continue
		}
		if isPersistentManaRiderTokens(semanticEffectTokens(sentences[si].Tokens)) {
			if riderIndex != -1 {
				return
			}
			riderIndex = si
			continue
		}
		if addIndex != -1 ||
			len(sentences[si].Effects) != 1 ||
			sentences[si].Effects[0].Kind != EffectAddMana {
			return
		}
		addIndex = si
	}
	if addIndex == -1 || riderIndex == -1 || riderIndex <= addIndex {
		return
	}
	addMana := &sentences[addIndex].Effects[0]
	addMana.Mana.PersistUntilEndOfTurn = true
	addMana.PersistUntilEndOfTurnRiderSpan = sentences[riderIndex].Span
	addMana.RequiresOrderedLowering = false
	sentences[riderIndex].Effects = nil
	sentences[riderIndex].PersistentManaRider = true
}
