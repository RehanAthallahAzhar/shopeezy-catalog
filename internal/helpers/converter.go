package helpers

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

func StringToUUID(id string) (uuid.UUID, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid UUID: %w", err)
	}
	return uuidID, nil
}

func IntToNullInt32(val int) sql.NullInt32 {
	return sql.NullInt32{
		Int32: int32(val),
		Valid: true,
	}
}

func StringToNullString(val string) sql.NullString {
	return sql.NullString{
		String: val,
		Valid:  true,
	}
}

func ConvertNullInt32(v reflect.Value) int {
	nullInt, ok := v.Interface().(sql.NullInt32)
	if !ok {
		return int(v.Interface().(int32))
	}

	if nullInt.Valid {
		return int(nullInt.Int32)
	}

	return 0
}

func ConvertNullString(v reflect.Value) string {
	nullInt, ok := v.Interface().(sql.NullString)
	if !ok {
		return v.Interface().(string)
	}

	if nullInt.Valid {
		return nullInt.String
	}

	return ""
}
