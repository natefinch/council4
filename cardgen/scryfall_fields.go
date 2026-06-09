package cardgen

import "strings"

type scryfallFaceFields struct {
	Name       string
	Layout     string
	ManaCost   string
	TypeLine   string
	OracleText string
	Colors     []string

	ColorIdentity []string

	Power     *string
	Toughness *string
	Loyalty   *string
	Defense   *string
}

func rootFields(card *ScryfallCard) scryfallFaceFields {
	if len(card.CardFaces) > 0 && faceLayoutUsesFrontAsRoot(card.Layout) {
		root := fieldsFromFace(card.CardFaces[0])
		root.Layout = card.Layout
		root.ColorIdentity = append([]string(nil), card.ColorIdentity...)
		if layoutEmitsAlternate(card.Layout) {
			root.Colors = colorsFromManaCost(root.ManaCost)
		}
		return root
	}
	return scryfallFaceFields{
		Name:          card.Name,
		Layout:        card.Layout,
		ManaCost:      card.ManaCost,
		TypeLine:      card.TypeLine,
		OracleText:    card.OracleText,
		Colors:        append([]string(nil), card.Colors...),
		ColorIdentity: append([]string(nil), card.ColorIdentity...),
		Power:         card.Power,
		Toughness:     card.Toughness,
		Loyalty:       card.Loyalty,
		Defense:       card.Defense,
	}
}

func fieldsFromFace(face ScryfallCardFace) scryfallFaceFields {
	return scryfallFaceFields{
		Name:       face.Name,
		ManaCost:   face.ManaCost,
		TypeLine:   face.TypeLine,
		OracleText: face.OracleText,
		Colors:     append([]string(nil), face.Colors...),
		Power:      face.Power,
		Toughness:  face.Toughness,
		Loyalty:    face.Loyalty,
		Defense:    face.Defense,
	}
}

func generatedFaces(card *ScryfallCard) []scryfallFaceFields {
	if len(card.CardFaces) < 2 || !layoutEmitsFaces(card.Layout) {
		return nil
	}
	faces := make([]scryfallFaceFields, 0, len(card.CardFaces)-1)
	for _, face := range card.CardFaces[1:] {
		faces = append(faces, fieldsFromFace(face))
	}
	return faces
}

func alternateFields(card *ScryfallCard) []scryfallFaceFields {
	if len(card.CardFaces) < 2 || !layoutEmitsAlternate(card.Layout) {
		return nil
	}
	faces := make([]scryfallFaceFields, 0, len(card.CardFaces)-1)
	for _, face := range card.CardFaces[1:] {
		fields := fieldsFromFace(face)
		fields.Colors = colorsFromManaCost(fields.ManaCost)
		faces = append(faces, fields)
	}
	return faces
}

func facesFromAllCardFaces(card *ScryfallCard) []scryfallFaceFields {
	faces := make([]scryfallFaceFields, 0, len(card.CardFaces))
	for _, face := range card.CardFaces {
		fields := fieldsFromFace(face)
		fields.Layout = card.Layout
		fields.ColorIdentity = append([]string(nil), card.ColorIdentity...)
		faces = append(faces, fields)
	}
	return faces
}

func faceLayoutUsesFrontAsRoot(layout string) bool {
	switch layout {
	case "transform", "modal_dfc", "meld", "double_faced_token", "reversible_card", "adventure", "split", "prepare":
		return true
	default:
		return false
	}
}

func layoutEmitsFaces(layout string) bool {
	switch layout {
	case "transform", "modal_dfc", "double_faced_token":
		return true
	default:
		return false
	}
}

func layoutEmitsAlternate(layout string) bool {
	switch layout {
	case "adventure", "split", "prepare":
		return true
	default:
		return false
	}
}

func colorsFromManaCost(manaCost string) []string {
	seen := map[string]bool{}
	for _, match := range manaSymbolRe.FindAllStringSubmatch(strings.ToUpper(manaCost), -1) {
		if len(match) < 2 {
			continue
		}
		symbol := match[1]
		for _, color := range []string{"W", "U", "B", "R", "G"} {
			if strings.Contains(symbol, color) {
				seen[color] = true
			}
		}
	}
	var colors []string
	for _, color := range []string{"W", "U", "B", "R", "G"} {
		if seen[color] {
			colors = append(colors, color)
		}
	}
	return colors
}
