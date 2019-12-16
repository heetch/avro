package avro

// Schema represents an Avro schema.
Schema :: TypeName | Union | Record | Enum | Map | Array | Fixed

// Add LogicalType to the above disjunction when
// https://github.com/cuelang/cue/issues/224 is
// fixed.

Name :: =~#"^([A-Za-z_][A-Za-z0-9_]*)(\.([A-Za-z_][A-Za-z0-9_]*))*$"#

TypeName :: PrimitiveType | DefinedType
TypeName :: Name

DefinedType :: string

PrimitiveType :: "null" | "boolean" | "int" | "long" | "float" | "double" | "bytes" | "string"

Union :: [... Schema]

Record :: Schema & {
	type:       "record"
	name:       DefinedType
	namespace?: string
	doc?:       string
	aliases?: [... DefinedType]
	fields: [... Field]
}

Field :: {
	name:     string
	doc?:     string
	type:     Schema
	default?: _
	order?:   "ascending" | "descending" | "ignore"
	aliases?: [... DefinedType]
}

FullTypedef :: {
	name:       DefinedType
	namespace?: =~#"^([A-Za-z_][A-Za-z0-9_]*)(\.([A-Za-z_][A-Za-z0-9_]*))*"#
	aliases?: [...string]
	doc?: string
	...
}

Enum :: FullTypedef & {
	type:       "enum"
	name:       string
	namespace?: string
	aliases?: [...string]
	doc?: string
	symbols: [... Name]
	default?: Name // & strings.Contains(symbols)
}

Array :: Schema & {
	type:  "array"
	items: Schema
}

Map :: Schema & {
	type:   "map"
	values: Schema
}

Fixed :: FullTypedef & {
	type:       "fixed"
	name:       DefinedType
	namespace?: string
	aliases?: [... DefinedType]
	size: int
}

LogicalType :: {
	Schema
	logicalType: string
}
LogicalType :: DecimalBytes | DecimalFixed | UUID | Date | TimeMillis | TimeMicros | TimestampMillis | TimestampMicros

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

TimestampDuration :: {
	type:        "fixed"
	logicalType: "duration"
	size:        12
}
