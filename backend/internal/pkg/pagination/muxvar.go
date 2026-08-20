package pagination

import "net/http"

// muxVar supports gorilla/mux style path variables via the request's context.
// The mux router injects vars through mux.VarsRequest; we read them through
// a small adapter to avoid importing mux at the call sites.
func muxVar(r *http.Request, key string) string {
	if v := r.PathValue(key); v != "" {
		return v
	}
	if vs := muxVars(r); vs != nil {
		if v, ok := vs[key]; ok {
			return v
		}
	}
	return ""
}

// muxVarsHook is set by the router package to expose gorilla/mux vars.
var muxVarsHook func(*http.Request) map[string]string

// SetMuxVarsHook registers the gorilla/mux var reader.
// SetMuxVarsHook is called from router init.
func SetMuxVarsHook(fn func(*http.Request) map[string]string) {
	muxVarsHook = fn
}

func muxVars(r *http.Request) map[string]string {
	if muxVarsHook != nil {
		return muxVarsHook(r)
	}
	return nil
}
