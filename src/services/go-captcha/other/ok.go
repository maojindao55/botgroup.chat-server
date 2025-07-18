package other

import (
	"encoding/json"
	"fmt"
	"net/http"
	"project/src/services/go-captcha/internal/cache"
)

// CheckOk .
func CheckOk(w http.ResponseWriter, r *http.Request) {
	code := 1
	err := r.ParseForm()
	if err != nil {
		bt, _ := json.Marshal(map[string]interface{}{
			"code":    code,
			"message": "parse form data err",
		})
		_, _ = fmt.Fprintf(w, string(bt))
		return
	}

	key := r.Form.Get("key")
	if key == "" {
		bt, _ := json.Marshal(map[string]interface{}{
			"code":    code,
			"message": "key param is empty",
		})
		_, _ = fmt.Fprintf(w, string(bt))
		return
	}

	isOk := cache.HasCacheOk(key)
	if isOk {
		code = 0
	}

	bt, _ := json.Marshal(map[string]interface{}{
		"code": code,
		"ok":   code == 0,
	})
	_, _ = fmt.Fprintf(w, string(bt))

	return
}
