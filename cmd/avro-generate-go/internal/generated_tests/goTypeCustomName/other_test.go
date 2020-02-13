package goTypeCustomName

// Check that the types exist and look correct
var (
	_ customName
	_ customEnum = customEnumA
	_            = customFixed{}[1]
)
