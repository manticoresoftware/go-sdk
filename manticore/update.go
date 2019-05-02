package manticore
/*
EUpdateType is values for `vtype` of UpdateAttributes() call, which determines meaning of `values` param of this function.

UpdateInt

This is the default value. `values` hash holds documents IDs as keys and a plain arrays of new attribute values.

UpdateMva

Points that MVA attributes are being updated. In this case the `values` must be a hash with document IDs as keys
and array of arrays of int values (new MVA attribute values).

UpdateString

Points that string attributes are being updated. `values` must be a hash with document IDs as keys and array of strings as values.

UpdateJson

Works the same as `UpdateString`, but for JSON attribute updates.
 */
type EUpdateType uint32

const (
	UpdateInt EUpdateType = iota
	UpdateMva
	UpdateString
	UpdateJson
)

func buildUpdateRequest(index string, attrs []string, values map[DocID][]interface{},
	vtype EUpdateType, ignorenonexistent bool) func(*apibuf) {
	nattrs := len(attrs)
	nvalues := len(values)
	return func(buf *apibuf) {
		buf.putString(index)
		buf.putLen(nattrs)
		buf.putBoolDword(ignorenonexistent)

		for j := 0; j < nattrs; j++ {
			buf.putString(attrs[j])
			buf.putDword(uint32(vtype))
		}

		buf.putLen(nvalues)
		for key, value := range values {
			buf.putDocid(key)
			switch vtype {
			case UpdateInt:
				for j := 0; j < nattrs; j++ {
					buf.putInt(int32(value[j].(int)))
				}
			case UpdateMva:
				for j := 0; j < nattrs; j++ {
					foo := value[j].([]uint32)
					buf.putLen(len(foo))
					for k := 0; k<len(foo); k++ {
						buf.putDword(foo[k])
					}
				}

			case UpdateString:
			case UpdateJson:
				for j := 0; j < nattrs; j++ {
					buf.putString(value[j].(string))
				}
			}
		}
	}
}