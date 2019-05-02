package manticore

/// aggregate function to apply
type eBsonType byte

const (
	bsonEof = eBsonType(iota)
	bsonInt32
	bsonInt64
	bsonDouble
	bsonString
	bsonStringVector
	bsonInt32Vector
	bsonInt64Vector
	bsonDoubleVector
	bsonMixedVector
	bsonObject
	bsonTrue
	bsonFalse
	bsonNull
	bsonRoot
)

type bsonField = struct {
	etype eBsonType
	blob  []byte
}
