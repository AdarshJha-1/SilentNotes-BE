package server

import (
	"ama/internal/models"
	"ama/internal/types"
	"ama/internal/utils"
	"ama/internal/utils/email"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Server) SignUp(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup

	var user models.UserModel
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "invalid input", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}
	defer r.Body.Close()

	var validate = validator.New()
	err = validate.Struct(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "validation failed", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	existingUser := s.db.CheckExistingUser(user.Username, user.Email)
	if existingUser {
		w.WriteHeader(http.StatusConflict)
		res := types.Response{StatusCode: http.StatusConflict, Success: false, Message: "username/email already taken"}
		json.NewEncoder(w).Encode(res)
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "internal server error", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	user.ID = primitive.NewObjectID()
	user.Password = string(hashedPassword)
	user.IsAcceptingMessages = true
	user.IsVerified = false
	user.VerifyCode = utils.GenerateVerifyCode()
	user.VerifyCodeExpiry = utils.VerifyCodeExpiry()

	wg.Add(1)

	var emailError error
	go func() {
		defer wg.Done()
		emailError = email.SendVerificationEmail(user.Username, user.Email, user.VerifyCode)
	}()

	wg.Wait()

	if emailError != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "error sending email verification code", Error: emailError.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	userId, err := s.db.CreateUser(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "error creating user", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}
	res := types.Response{StatusCode: http.StatusCreated, Success: true, Message: "user signed up successfully", Data: map[string]interface{}{"userId": userId}}
	json.NewEncoder(w).Encode(res)

}
func (s *Server) SignIn(w http.ResponseWriter, r *http.Request) {

	var wg sync.WaitGroup

	var user models.SingInModel

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "invalid input", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}
	defer r.Body.Close()

	var validate = validator.New()
	err = validate.Struct(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "validation failed", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	dbUser := s.db.GetUser(user.Identifier, "")

	if dbUser == nil {
		w.WriteHeader(http.StatusNotFound)
		res := types.Response{StatusCode: http.StatusNotFound, Success: false, Message: "user not found"}
		json.NewEncoder(w).Encode(res)
		return
	}

	isPasswordCorrect := utils.CheckPassword(user.Password, dbUser.Password)
	if !isPasswordCorrect {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "wrong credentials", Error: "wrong credentials"}
		json.NewEncoder(w).Encode(res)
		return
	}

	if !dbUser.IsVerified {
		verifyCode := utils.GenerateVerifyCode()
		verifyCodeExpiry := utils.VerifyCodeExpiry()
		s.db.ReVerifyCode(dbUser.ID, verifyCode, verifyCodeExpiry)

		wg.Add(1)

		var emailError error
		go func() {
			defer wg.Done()
			emailError = email.SendVerificationEmail(dbUser.Username, dbUser.Email, verifyCode)
		}()

		wg.Wait()

		if emailError != nil {
			w.WriteHeader(http.StatusInternalServerError)
			res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "error sending email verification code", Error: emailError.Error()}
			json.NewEncoder(w).Encode(res)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "please verify your account before singing up, check email"}
		json.NewEncoder(w).Encode(res)
		return
	}

	token := utils.CreateJWT(dbUser.ID.Hex())

	if token == nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "error creating jwt token"}
		json.NewEncoder(w).Encode(res)
		return
	}

	cookie := &http.Cookie{
		Name:     "token",
		Value:    token.(string),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   86400,
	}
	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	res := types.Response{StatusCode: http.StatusOK, Success: true, Message: "user signed in successfully", Data: map[string]interface{}{"token": token}}
	json.NewEncoder(w).Encode(res)
}

func (s *Server) SignOut(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	cookie := http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusOK)
	res := types.Response{StatusCode: http.StatusOK, Success: true, Message: "user signed out successfully"}
	json.NewEncoder(w).Encode(res)
}

func (s *Server) VerifyUser(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	verifyCode := r.URL.Query().Get("code")

	defer r.Body.Close()
	if username == "" || verifyCode == "" {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "invalid inputs"}
		json.NewEncoder(w).Encode(res)
		return
	}

	verifyCodeInt, err := strconv.Atoi(verifyCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "internal server error", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	user := s.db.GetUser(username, "password")
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		res := types.Response{StatusCode: http.StatusNotFound, Success: false, Message: "user not found"}
		json.NewEncoder(w).Encode(res)
		return
	}

	isVerifyCodeExpired := time.Now().After(user.VerifyCodeExpiry)
	if isVerifyCodeExpired {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "verify code expired"}
		json.NewEncoder(w).Encode(res)
		return
	}

	isCorrectVerifyCode := verifyCodeInt == user.VerifyCode

	if !isCorrectVerifyCode {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "invalid verify code"}
		json.NewEncoder(w).Encode(res)
		return
	}
	userId, err := s.db.VerifyUser(username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "error verifying user", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	res := types.Response{StatusCode: http.StatusOK, Success: true, Message: "user verified successfully", Data: map[string]interface{}{"userId": userId}}
	json.NewEncoder(w).Encode(res)

}

func (s *Server) AcceptMessages(w http.ResponseWriter, r *http.Request) {

	userId := r.Context().Value(types.UserIDKey).(string)
	userIdObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "internal server error", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}
	isAcceptingMessagesQuery := r.URL.Query().Get("is_accepting_messages")
	defer r.Body.Close()
	if isAcceptingMessagesQuery == "" {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "invalid inputs"}
		json.NewEncoder(w).Encode(res)
		return
	}
	isAcceptingMessages := isAcceptingMessagesQuery == "true"

	result := s.db.ToggleAcceptMessages(isAcceptingMessages, userIdObjectId)
	if !result {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "failed to toggle accept messages", Error: "failed to toggle accept messages"}
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusOK)
	res := types.Response{StatusCode: http.StatusOK, Success: true, Message: "accept message status updated successfully"}
	json.NewEncoder(w).Encode(res)
}

func (s *Server) SendMessage(w http.ResponseWriter, r *http.Request) {

	var sendMessageData types.SendMessageType
	err := json.NewDecoder(r.Body).Decode(&sendMessageData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "invalid input", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}
	defer r.Body.Close()

	var validate = validator.New()
	err = validate.Struct(sendMessageData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "validation failed", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	user := s.db.GetUser(sendMessageData.Identifier, "password")
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		res := types.Response{StatusCode: http.StatusNotFound, Success: false, Message: "user not found"}
		json.NewEncoder(w).Encode(res)
		return
	}

	if !user.IsAcceptingMessages {
		w.WriteHeader(http.StatusForbidden)
		res := types.Response{StatusCode: http.StatusForbidden, Success: false, Message: "user is not accepting messages"}
		json.NewEncoder(w).Encode(res)
		return
	}

	message := models.Message{
		ID:        primitive.NewObjectID(),
		Content:   sendMessageData.Content,
		CreatedAt: time.Time{},
	}

	err = s.db.AddMessage(sendMessageData.Identifier, message)
	if err != nil {
		if err == fmt.Errorf("user not found") {
			w.WriteHeader(http.StatusNotFound)
			res := types.Response{StatusCode: http.StatusBadRequest, Success: false, Message: "user not found", Error: err.Error()}
			json.NewEncoder(w).Encode(res)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "internal server error", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}
	w.WriteHeader(http.StatusCreated)
	res := types.Response{StatusCode: http.StatusCreated, Success: true, Message: "message sent successfully"}
	json.NewEncoder(w).Encode(res)
}

func (s *Server) GetMessages(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(types.UserIDKey).(string)
	userIdObjectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "internal server error", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	messages, err := s.db.GetMessages(userIdObjectId)
	if err == nil && messages == nil {
		w.WriteHeader(http.StatusNotFound)
		res := types.Response{StatusCode: http.StatusNotFound, Success: false, Message: "no messages currently"}
		json.NewEncoder(w).Encode(res)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "internal server error", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	w.WriteHeader(http.StatusOK)
	res := types.Response{StatusCode: http.StatusOK, Success: true, Message: "current messages", Messages: map[string][]models.Message{"messages": messages}}
	json.NewEncoder(w).Encode(res)
}
