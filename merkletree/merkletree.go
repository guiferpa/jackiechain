package merkletree

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func generateLeafHash(hl string, hr string) string {
	payload := fmt.Sprintf("%s%s", hl, hr)
	h := sha256.New()
	h.Write([]byte(payload))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateRootHash(hs []string) string {
	if len(hs) == 1 {
		return hs[0]
	}
	nhs := make([]string, 0)
	i := 0
	for i < len(hs) {
		if i+1 < len(hs) {
			l := hs[i]
			r := hs[i+1]
			nhs = append(nhs, generateLeafHash(l, r))
			i += 2
			continue
		}
		l := hs[i]
		nhs = append(nhs, l)
		i++
	}
	return GenerateRootHash(nhs)
}
