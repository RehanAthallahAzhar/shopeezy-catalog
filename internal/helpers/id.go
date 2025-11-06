package helpers

import (
	"github.com/google/uuid"
)

/*
GenerateNewUserID digunakan agar skalabilitas, desentralisasi, keamanan, dan fleksibilitas terjamin

	skalabilitas -> Menghindari bottleneck saat data tubuh besar
	desentralisasi -> mengurangi otonomi dan meningkatkan coupling antar komponen
	global uniquesness without coordination -> tidak khawatir akan tabrakan ID
	fleksibilitas -> memberikan ID yang dapat dibagikan dan dijamin unik di seluruh ekosistem, tanpa perlu khawatir tentang konflik ID dengan sistem eksternal
*/
func GenerateNewID() uuid.UUID {
	return uuid.New()
}

func IsValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
