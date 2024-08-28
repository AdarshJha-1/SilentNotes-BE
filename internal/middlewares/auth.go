package middlewares

import (
	"ama/internal/types"
	"ama/internal/utils"
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// func AuthMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		token, err := r.Cookie("token")
// 		if err == http.ErrNoCookie {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			res := types.Response{StatusCode: http.StatusUnauthorized, Message: "Unauthorized", Error: "Unauthorized"}
// 			json.NewEncoder(w).Encode(res)
// 			return
// 		}

// 		claims, err := utils.VerifyJWT(token.Value)
// 		if err != nil {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			res := types.Response{StatusCode: http.StatusUnauthorized, Message: "Unauthorized", Error: err.Error()}
// 			json.NewEncoder(w).Encode(res)
// 			return
// 		}

// 		userId := claims["user_id"].(string)
// 		fmt.Println("Extracted userID:", userId)

// 		ctx := context.WithValue(r.Context(), types.UserIDKey, userId)
// 		// fmt.Println("Context value set:", ctx.Value(config.UserIDKey))
// 		next.ServeHTTP(w, r.WithContext(ctx))
// 	})
// }

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		token, err := r.Cookie("token")
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			res := types.Response{StatusCode: http.StatusUnauthorized, Success: false, Message: "Unauthorized", Error: "Unauthorized"}
			json.NewEncoder(w).Encode(res)
			return
		}

		claims, err := utils.VerifyJWT(token.Value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			res := types.Response{StatusCode: http.StatusUnauthorized, Success: false, Message: "Unauthorized", Error: "Unauthorized"}
			json.NewEncoder(w).Encode(res)
			return
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			res := types.Response{StatusCode: http.StatusUnauthorized, Success: false, Message: "Unauthorized", Error: "Invalid expiration claim"}
			json.NewEncoder(w).Encode(res)
			return
		}

		if time.Now().After(time.Unix(int64(exp), 0)) {
			w.WriteHeader(http.StatusUnauthorized)
			res := types.Response{StatusCode: http.StatusUnauthorized, Success: false, Message: "Unauthorized", Error: "Cookie Expired"}
			json.NewEncoder(w).Encode(res)
			return
		}

		userId := claims["user_id"].(string)
		ctx := context.WithValue(r.Context(), types.UserIDKey, userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
