package preservation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/natefinch/council4/mtg/game"
)

// This tool-only walk includes sealed primitive payloads and does not depend on
// renderer coverage. It encodes values, never pointer addresses or Go functions.
func definitionFingerprint(def *game.CardDef) (string, error) {
	var text strings.Builder
	if err := snapshotValue(&text, reflect.ValueOf(def), make(map[uintptr]bool)); err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(text.String()))
	return hex.EncodeToString(sum[:]), nil
}

func snapshotValue(out *strings.Builder, value reflect.Value, active map[uintptr]bool) error {
	if !value.IsValid() {
		_, _ = out.WriteString("invalid;")
		return nil
	}
	_, _ = out.WriteString(value.Type().PkgPath())
	_ = out.WriteByte(':')
	_, _ = out.WriteString(value.Type().String())
	_ = out.WriteByte('{')
	switch value.Kind() {
	case reflect.Pointer:
		if value.IsNil() {
			_, _ = out.WriteString("nil")
			break
		}
		address := value.Pointer()
		if active[address] {
			return fmt.Errorf("cyclic definition at %s", value.Type())
		}
		active[address] = true
		if err := snapshotValue(out, value.Elem(), active); err != nil {
			return err
		}
		delete(active, address)
	case reflect.Interface:
		if value.IsNil() {
			_, _ = out.WriteString("nil")
		} else if err := snapshotValue(out, value.Elem(), active); err != nil {
			return err
		}
	case reflect.Struct:
		for index := range value.NumField() {
			_, _ = out.WriteString(value.Type().Field(index).Name)
			_ = out.WriteByte('=')
			if err := snapshotValue(out, value.Field(index), active); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice && value.IsNil() {
			_, _ = out.WriteString("nil")
		}
		for index := range value.Len() {
			if err := snapshotValue(out, value.Index(index), active); err != nil {
				return err
			}
		}
	case reflect.Map:
		var entries []string
		for iterator := value.MapRange(); iterator.Next(); {
			var entry strings.Builder
			if err := snapshotValue(&entry, iterator.Key(), active); err != nil {
				return err
			}
			if err := snapshotValue(&entry, iterator.Value(), active); err != nil {
				return err
			}
			entries = append(entries, entry.String())
		}
		slices.Sort(entries)
		if value.IsNil() {
			_, _ = out.WriteString("nil")
		}
		for _, entry := range entries {
			_, _ = out.WriteString(entry)
		}
	case reflect.String:
		_, _ = out.WriteString(strconv.Quote(value.String()))
	case reflect.Bool:
		_, _ = out.WriteString(strconv.FormatBool(value.Bool()))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		_, _ = out.WriteString(strconv.FormatInt(value.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		_, _ = out.WriteString(strconv.FormatUint(value.Uint(), 10))
	case reflect.Float32, reflect.Float64:
		_, _ = out.WriteString(strconv.FormatFloat(value.Float(), 'g', -1, value.Type().Bits()))
	default:
		return fmt.Errorf("unsupported definition value %s", value.Type())
	}
	_, _ = out.WriteString("};")
	return nil
}
