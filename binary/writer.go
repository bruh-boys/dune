package binary

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"

	"github.com/dunelang/dune"
)

func Write(w io.Writer, p *dune.Program) error {
	key := byte(5 + rand.Intn(255-5))

	if err := binary.Write(w, binary.BigEndian, int32(key)); err != nil {
		return err
	}

	if err := writeString(w, header, key); err != nil {
		return err
	}

	if err := writeAttributes(w, p.Attributes, key); err != nil {
		return err
	}

	if err := writeEnums(w, p.Enums, key); err != nil {
		return err
	}

	if err := writeClass(w, p.Classes, key); err != nil {
		return err
	}

	if err := writeFunctions(w, p.Functions, key); err != nil {
		return err
	}

	if err := writeConstants(w, p.Constants, key); err != nil {
		return err
	}

	if err := writeFiles(w, p.Files, key); err != nil {
		return err
	}

	if err := writeResources(w, p.Resources, key); err != nil {
		return err
	}

	if err := writeSection(w, section_EOF, 0); err != nil {
		return err
	}

	return nil
}

func writeResources(w io.Writer, resources map[string][]byte, key byte) error {
	if err := writeSection(w, section_resources, len(resources)); err != nil {
		return err
	}

	for k, v := range resources {
		if err := writeString(w, k, key); err != nil {
			return err
		}
		if err := writeBytes(w, v); err != nil {
			return err
		}
	}

	return nil
}

func writeFiles(w io.Writer, files []string, key byte) error {
	if err := writeSection(w, section_files, len(files)); err != nil {
		return err
	}
	for _, f := range files {
		if err := writeString(w, f, key); err != nil {
			return err
		}
	}
	return nil
}

func writeConstants(w io.Writer, constants []dune.Value, key byte) error {
	if err := writeSection(w, section_constants, len(constants)); err != nil {
		return err
	}

	for _, k := range constants {
		switch k.Type {
		case dune.Int:
			if err := writeSection(w, section_kInt, 0); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, k.ToInt()); err != nil {
				return err
			}

		case dune.Float:
			if err := writeSection(w, section_kFloat, 0); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, k.ToFloat()); err != nil {
				return err
			}

		case dune.Bool:
			if err := writeSection(w, section_kBool, 0); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, k.ToBool()); err != nil {
				return err
			}

		case dune.String:
			b := []byte(k.ToString())
			xor(b, key)
			if err := writeSection(w, section_kString, len(b)); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, b); err != nil {
				return err
			}

		case dune.Null:
			if err := writeSection(w, section_kNull, 0); err != nil {
				return err
			}

		case dune.Undefined:
			if err := writeSection(w, section_kUndefined, 0); err != nil {
				return err
			}

		case dune.Rune:
			if err := writeSection(w, section_kRune, 0); err != nil {
				return err
			}
			if err := binary.Write(w, binary.BigEndian, int64(k.ToRune())); err != nil {
				return err
			}

		default:
			return fmt.Errorf("invalid constant type: %v", k.Type)
		}
	}
	return nil
}

func writeEnums(w io.Writer, enums []*dune.EnumList, key byte) error {
	if err := writeSection(w, section_enums, len(enums)); err != nil {
		return err
	}

	for _, enum := range enums {
		if err := writeString(w, enum.Name, key); err != nil {
			return err
		}
		if err := writeBool(w, enum.Exported); err != nil {
			return err
		}
		if err := writeEnumValues(w, enum.Values, key); err != nil {
			return err
		}
	}

	return nil
}

func writeClass(w io.Writer, classes []*dune.Class, key byte) error {
	if err := writeSection(w, section_classes, len(classes)); err != nil {
		return err
	}

	for _, c := range classes {
		if err := writeAttributes(w, c.Attributes, key); err != nil {
			return err
		}
		if err := writeString(w, c.Name, key); err != nil {
			return err
		}
		if err := writeBool(w, c.Exported); err != nil {
			return err
		}
		if err := writeClassFields(w, c.Fields, key); err != nil {
			return err
		}
		if err := writeClassFunctions(w, c.Functions, key); err != nil {
			return err
		}
	}

	return nil
}

func writeClassFunctions(w io.Writer, funcs []int, key byte) error {
	if err := writeSection(w, section_classFunctions, len(funcs)); err != nil {
		return err
	}

	for _, i := range funcs {
		if err := writeInt32(w, i); err != nil {
			return err
		}
	}

	return nil
}

func writeClassFields(w io.Writer, fields []*dune.Field, key byte) error {
	if err := writeSection(w, section_classFields, len(fields)); err != nil {
		return err
	}

	for _, f := range fields {
		if err := writeString(w, f.Name, key); err != nil {
			return err
		}
		if err := writeBool(w, f.Exported); err != nil {
			return err
		}
	}

	return nil
}

func writeFunctions(w io.Writer, funcs []*dune.Function, key byte) error {
	if err := writeSection(w, section_functions, len(funcs)); err != nil {
		return err
	}

	for _, f := range funcs {
		if err := binary.Write(w, binary.BigEndian, int32(f.Index)); err != nil {
			return err
		}
		if err := writeAttributes(w, f.Attributes, key); err != nil {
			return err
		}
		if err := writeString(w, f.Name, key); err != nil {
			return err
		}
		if err := writeBool(w, f.Variadic); err != nil {
			return err
		}
		if err := writeBool(w, f.Exported); err != nil {
			return err
		}
		if err := writeBool(w, f.Anonimous); err != nil {
			return err
		}
		if err := writeBool(w, f.IsClass); err != nil {
			return err
		}
		if err := writeInt32(w, f.Class); err != nil {
			return err
		}
		if err := writeInt32(w, f.WrapClass); err != nil {
			return err
		}
		if err := writeBool(w, f.IsGlobal); err != nil {
			return err
		}
		if err := writeInt32(w, f.Arguments); err != nil {
			return err
		}
		if err := writeInt32(w, f.OptionalArguments); err != nil {
			return err
		}
		if err := writeInt32(w, f.MaxRegIndex); err != nil {
			return err
		}
		if err := writeRegisters(w, f.Registers, key); err != nil {
			return err
		}
		if err := writeRegisters(w, f.Closures, key); err != nil {
			return err
		}
		if err := writeInstructions(w, f.Instructions, key); err != nil {
			return err
		}
		if err := writePositions(w, f.Positions); err != nil {
			return err
		}
	}

	return nil
}

func writeInstructions(w io.Writer, ins []*dune.Instruction, key byte) error {
	if err := writeSection(w, section_instructions, len(ins)); err != nil {
		return err
	}

	for _, i := range ins {
		if err := binary.Write(w, binary.BigEndian, byte(i.Opcode)); err != nil {
			return err
		}
		if err := writeAddress(w, i.A, key); err != nil {
			return err
		}
		if err := writeAddress(w, i.B, key); err != nil {
			return err
		}
		if err := writeAddress(w, i.C, key); err != nil {
			return err
		}
	}

	return nil
}

func writeAddress(w io.Writer, a *dune.Address, key byte) error {
	if err := binary.Write(w, binary.BigEndian, byte(a.Kind)); err != nil {
		return err
	}

	if a.Kind == dune.AddrNativeFunc {
		f := dune.NativeFuncFromIndex(int(a.Value))
		if err := writeString(w, f.Name, key); err != nil {
			return err
		}
	} else {
		if err := binary.Write(w, binary.BigEndian, a.Value); err != nil {
			return err
		}
	}

	return nil
}

func writePositions(w io.Writer, positions []dune.Position) error {
	if err := writeSection(w, section_positions, len(positions)); err != nil {
		return err
	}
	for _, pos := range positions {
		if err := binary.Write(w, binary.BigEndian, int32(pos.File)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, int32(pos.Line)); err != nil {
			return err
		}
	}
	return nil
}

func writeEnumValues(w io.Writer, values []*dune.EnumValue, key byte) error {
	if err := writeSection(w, section_enumValues, len(values)); err != nil {
		return err
	}

	for _, v := range values {
		if err := writeString(w, v.Name, key); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, int32(v.KIndex)); err != nil {
			return err
		}
	}

	return nil
}

func writeRegisters(w io.Writer, regs []*dune.Register, key byte) error {
	if err := writeSection(w, section_registers, len(regs)); err != nil {
		return err
	}
	for _, r := range regs {
		if err := writeString(w, r.Name, key); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, int32(r.Index)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, int32(r.StartPC)); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, int32(r.EndPC)); err != nil {
			return err
		}
		if err := writeBool(w, r.Exported); err != nil {
			return err
		}
	}
	return nil
}

func writeAttributes(w io.Writer, attributes []string, key byte) error {
	if err := writeSection(w, section_attributes, len(attributes)); err != nil {
		return err
	}
	for _, v := range attributes {
		if err := writeString(w, v, key); err != nil {
			return err
		}
	}

	return nil
}

func writeString(w io.Writer, v string, key byte) error {
	b := []byte(v)

	xor(b, key)

	if err := writeSection(w, section_string, len(b)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, b); err != nil {
		return err
	}
	return nil
}

func writeBytes(w io.Writer, v []byte) error {
	if err := writeSection(w, section_bytes, len(v)); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, v); err != nil {
		return err
	}
	return nil
}

func writeSection(w io.Writer, sType SectionType, v int) error {
	s := newSection(sType, v)
	return binary.Write(w, binary.BigEndian, int64(s))
}

func writeBool(w io.Writer, b bool) error {
	var v int8
	if b {
		v = 1
	}
	return binary.Write(w, binary.BigEndian, v)
}

func writeInt32(w io.Writer, i int) error {
	return binary.Write(w, binary.BigEndian, int32(i))
}

func xor(b []byte, key byte) {
	for i, j := range b {
		b[i] = j ^ key
	}
}
