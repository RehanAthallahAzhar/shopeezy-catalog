package helpers

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/RehanAthallahAzhar/shopeezy-inventory-cart/pkg/errors"
)

/*
GenerateNewUserID digunakan agar skalabilitas, desentralisasi, keamanan, dan fleksibilitas terjamin

	skalabilitas -> Menghindari bottleneck saat data tubuh besar
	desentralisasi -> mengurangi otonomi dan meningkatkan coupling antar komponen
	global uniquesness without coordination -> tidak khawatir akan tabrakan ID
	fleksibilitas -> memberikan ID yang dapat dibagikan dan dijamin unik di seluruh ekosistem, tanpa perlu khawatir tentang konflik ID dengan sistem eksternal
*/
func GenerateNewID() string {
	return uuid.New().String()
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}

func GetIDFromPathParam(c echo.Context, key string) (string, error) {
	val := c.Param(key)
	if val == "" || !isValidUUID(val) {
		return "", errors.ErrInvalidRequestPayload
	}
	return val, nil
}

func GetFromPathParam(c echo.Context, key string) (string, error) {
	val := c.Param(key)
	if val == "" {
		return "", errors.ErrInvalidRequestPayload
	}
	return val, nil
}
