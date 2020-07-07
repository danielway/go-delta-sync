package value

// Refers to an element inside of a map
type MapElement struct {
	valueType
	mapValue Any
	key      Any
}

// THe map that contains the value
func (m MapElement) Map() Any {
	return m.mapValue
}

// The key of the element that is being referred to
func (m MapElement) Key() Any {
	return m.key
}

func NewMap(mapValue, key Any) MapElement {
	return MapElement{
		mapValue: mapValue,
		key:      key,
	}
}
