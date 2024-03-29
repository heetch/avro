package avro

// Schema represents an Avro schema.
// Note: we use defaults for all alternatives except LogicalType
// so that they take precedence over LogicalType which overlaps with
// other types.
Schema :: *TypeName | *Union | *Record | *Enum | *Map | *Array | *Fixed | LogicalType

Name :: =~#"^([A-Za-z_][A-Za-z0-9_]*)(\.([A-Za-z_][A-Za-z0-9_]*))*$"#

TypeName :: PrimitiveType | DefinedType
TypeName :: Name

DefinedType :: string

PrimitiveType :: "null" | "boolean" | "int" | "long" | "float" | "double" | "bytes" | "string"

Union :: [... Schema]

// Metadata can be used to allow extra metadata fields.
// The specification allows extra fields by default, but
// this schema disallows them unless explicitly added
// to this definition so that spelling mistakes aren't just ignored.
//
// For example, to allow all metadata fields as per the spec:
//
//	import (
//		avroPkg "github.com/heetch/cue-schema/avro"
//	)
//
//	avro :: avroPkg
//	avro :: Metadata :: [_]: _
//	mySchema: avro.Schema & {...}
Metadata :: {}

Definition :: {
	type: string
	name:       DefinedType
	namespace?: =~#"^([A-Za-z_][A-Za-z0-9_]*)(\.([A-Za-z_][A-Za-z0-9_]*))*"#
	aliases?: [...string]
	doc?: string
	Metadata
}

Record :: {
	Definition
	type:       "record"
	doc?:       string
	fields: [... Field]
}

Field :: {
	Metadata
	name:     string
	doc?:     string
	type:     Schema
	default?: _
	order?:   "ascending" | "descending" | "ignore"
	aliases?: [... DefinedType]
}

Fixed :: {
	Definition
	type:       "fixed"
	name:       DefinedType
	size: int
}

Enum :: {
	Definition
	type:       "enum"
	name:       string
	symbols: [... Name]
	default?: Name // & strings.Contains(symbols)
}

Array :: {
	Metadata
	type:  "array"
	items: Schema
}

Map :: {
	Metadata
	type:   "map"
	values: Schema
}

LogicalType :: {
	Schema
	logicalType: string
}

LogicalType :: DecimalBytes | DecimalFixed | UUID | Date | *TimeMillis | *TimeMicros | TimestampMillis | TimestampMicros | DurationNanos

DecimalBytes :: {
	type:        "bytes"
	logicalType: "decimal"
	precision:   >0
	scale?:      *0 | (>=0 & <=precision)
}

DecimalFixed :: {
	Fixed
	logicalType: "decimal"
	precision:   >0
	// TODO constrain
	// Can't yet do bitwise shifts in Cue. https://github.com/cuelang/cue/issues/156
	// precision: <= math.Floor(math.Log10(math.Pow(2, 8 * (n - 1))))
	scale?: *0 | (>=0 & <=precision)
}

UUID :: {
	type:        "string"
	logicalType: "uuid"
}

Date :: {
	type:        "int"
	logicalType: "date"
}

TimeMillis :: {
	type:        "int"
	logicalType: "time-millis"
}

TimeMicros :: {
	type:        "long"
	logicalType: "time-micros"
}

TimestampMillis :: {
	type:        "long"
	logicalType: "timestamp-millis"
}

TimestampMicros :: {
	type:        "long"
	logicalType: "timestamp-micros"
}

DurationNanos :: {
	type:        "long"
	logicalType: "duration-nanos"
}
