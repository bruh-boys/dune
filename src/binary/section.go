//go:generate stringer -type=SectionType

package binary

const header = "DUNE v1"

type SectionType int

const (
	section_attributes SectionType = iota
	section_build
	section_enums
	section_enumValues
	section_classes
	section_classFunctions
	section_classFields
	section_functions
	section_dynamicCalls
	section_registers
	section_instructions
	section_constants
	section_positions
	section_files
	section_resources
	section_sources
	section_sourceLines
	section_string
	section_bytes
	section_kInt
	section_kFloat
	section_kBool
	section_kString
	section_kNull
	section_kUndefined
	section_kRune
	section_EOF
)

type section uint64

func newSection(sType SectionType, v int) section {
	return section(int64(sType) | int64(v)<<5)
}

func (s section) values() (SectionType, int64) {
	t := SectionType(int((s >> 0) & ((1 << 5) - 1)))
	v := int64((int64(s) >> 5) & ((1 << 59) - 1))
	return t, v
}
