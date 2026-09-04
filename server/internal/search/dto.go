package search

type ObjectDTO struct {
	OID  int64  `json:"oid"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (o Object) ToDTO() ObjectDTO {
	return ObjectDTO{
		OID:  o.OID,
		Name: o.Identifier,
		Type: o.ObjectType,
	}
}

func (dto ObjectDTO) ToDomain() Object {
	return Object{
		OID:        dto.OID,
		Identifier: dto.Name,
		ObjectType: dto.Type,
	}
}
