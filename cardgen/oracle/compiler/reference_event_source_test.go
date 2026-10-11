package compiler

import "testing"

func TestLibraryGraveyardSourceAttribution(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
		want DamageAttribution
	}{
		{"named self", "When Event Source is put into your graveyard from your library, Event Source deals 3 damage to each opponent.", DamageAttributionResolvingSource},
		{"card self", "When this card is put into your graveyard from your library, Event Source deals 3 damage to each opponent.", DamageAttributionResolvingSource},
		{"owned exile", "When Event Source is put into your graveyard from your library, you may exile it. If you do, Event Source deals 3 damage to each opponent and you gain 3 life.", DamageAttributionResolvingSource},
		{"battlefield watcher", "Whenever a card is put into your graveyard from your library, Event Source deals 3 damage to each opponent.", DamageAttributionOriginalObject},
		{"another watcher", "Whenever another card is put into your graveyard from your library, Event Source deals 3 damage to each opponent.", DamageAttributionOriginalObject},
		{"self death", "When Event Source dies, Event Source deals 3 damage to each opponent.", DamageAttributionOriginalObject},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content := compiledRoleContent(t, tc.text, pipelineContext{CardName: "Event Source"}, 0)
			found := 0
			for _, effect := range content.Effects {
				if effect.Kind != EffectDealDamage {
					continue
				}
				if len(effect.SubjectReferences) != 1 {
					t.Fatalf("damage source references=%d", len(effect.SubjectReferences))
				}
				reference := effect.SubjectReferences[0]
				found++
				if !content.OwnsSubject(reference) || reference.DamageAttribution() != tc.want {
					t.Fatalf("domain=%v lifetime=%v attribution=%v want=%v", reference.SubjectDomain(),
						reference.SubjectLifetime(), reference.DamageAttribution(), tc.want)
				}
				if tc.want == DamageAttributionResolvingSource &&
					(reference.SubjectDomain() != ReferenceSubjectCard ||
						reference.SubjectLifetime() != ReferenceLifetimeCardIncarnation ||
						reference.SupportsUse(ReferenceUseMutation)) {
					t.Fatal("event source card gained permanent mutation authority")
				}
			}
			if found != 1 {
				t.Fatalf("damage clauses=%d want1", found)
			}
		})
	}
}

func TestLibraryGraveyardSourceProofControls(t *testing.T) {
	text := "When Event Source is put into your graveyard from your library, Event Source deals 3 damage to each opponent."
	content := compiledRoleContent(t, text, pipelineContext{CardName: "Event Source"}, 0)
	foreign := compiledRoleContent(t, text, pipelineContext{CardName: "Event Source"}, 0)
	reference := content.Effects[0].SubjectReferences[0]
	if !content.OwnsSubject(reference) || foreign.OwnsSubject(reference) {
		t.Fatal("source proof crossed its owning body")
	}
	for _, mutate := range []func(*CompiledReference){
		func(r *CompiledReference) { r.Subject = ReferenceSubjectProof{} },
		func(r *CompiledReference) { r.NodeID++ },
		func(r *CompiledReference) { r.Binding = ReferenceBindingEventPermanent },
	} {
		changed := reference
		mutate(&changed)
		if changed.DamageAttribution() != DamageAttributionUnsupported {
			t.Fatal("missing or changed source identity retained attribution")
		}
	}
	if _, ok := content.Source.OriginalObjectSubject(); ok {
		t.Fatal("library card was proved as an original permanent")
	}
}

func TestLibraryGraveyardSourceEventRequiresExactSelf(t *testing.T) {
	pattern := TriggerPattern{
		Event: TriggerEventZoneChanged, Source: TriggerSourceSelf, Player: TriggerPlayerYou,
		MatchFromZone: true, FromZone: TriggerZoneLibrary, MatchToZone: true, ToZone: TriggerZoneGraveyard,
	}
	if !exactLibraryGraveyardSourceEvent(pattern) {
		t.Fatal("exact self event refused")
	}
	for _, mutate := range []func(*TriggerPattern){
		func(p *TriggerPattern) { p.Source = TriggerSourceAny },
		func(p *TriggerPattern) { p.ExcludeSelf = true },
		func(p *TriggerPattern) { p.OneOrMore = true },
		func(p *TriggerPattern) { p.SubjectSelectionOrSelf = true },
		func(p *TriggerPattern) { p.UnionEvent = TriggerEventPermanentDied },
		func(p *TriggerPattern) { p.Player = TriggerPlayerOpponent },
		func(p *TriggerPattern) { p.MatchFromZone = false },
		func(p *TriggerPattern) { p.FromZone = TriggerZoneBattlefield },
		func(p *TriggerPattern) { p.ExcludeFromZone = true },
		func(p *TriggerPattern) { p.FromZones = []TriggerZone{TriggerZoneLibrary, TriggerZoneHand} },
		func(p *TriggerPattern) { p.MatchToZone = false },
		func(p *TriggerPattern) { p.ToZone = TriggerZoneExile },
		func(p *TriggerPattern) { p.ExcludeToZone = true },
	} {
		changed := pattern
		mutate(&changed)
		if exactLibraryGraveyardSourceEvent(changed) {
			t.Fatalf("inexact or foreign source event accepted: %+v", changed)
		}
	}
}
