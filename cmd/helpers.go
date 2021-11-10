package cmd

import (
	"crypto/sha256"
	"crypto/subtle"
	"log"
	"net/http"
	"oneserve/utils"
	"strings"
)

func basicAuth(next http.Handler, actualUsername string, actualPassword string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if ok {
			uHash := sha256.Sum256([]byte(username))
			pHash := sha256.Sum256([]byte(password))
			auHash := sha256.Sum256([]byte(actualUsername))
			apHash := sha256.Sum256([]byte(actualPassword))

			uMatch := (subtle.ConstantTimeCompare(uHash[:], auHash[:]) == 1)
			pMatch := (subtle.ConstantTimeCompare(pHash[:], apHash[:]) == 1)

			if uMatch && pMatch {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Printf("%s | %s | %s %s | %s\n", utils.Colorise(strings.Split(r.RemoteAddr, ":")[0]), utils.ColourStatusCode(http.StatusUnauthorized), utils.ColourHTTPMethod(r.Method), r.URL.String(), "-")
	})
}
