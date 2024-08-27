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

	var user *models.UserModel
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

	userId, err := s.db.CreateUser(*user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "error creating user", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}
	res := types.Response{StatusCode: http.StatusCreated, Success: true, Message: "user signed up successfully", Data: map[string]interface{}{"userId": userId}}
	json.NewEncoder(w).Encode(res)

}

func (s *Server) VerifyUser(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	verifyCode := r.URL.Query().Get("code")
	verifyCodeInt, err := strconv.Atoi(verifyCode)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res := types.Response{StatusCode: http.StatusInternalServerError, Success: false, Message: "something went wrong", Error: err.Error()}
		json.NewEncoder(w).Encode(res)
		return
	}

	user := s.db.GetUser(username)
	fmt.Println(user)
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

	isVerifyCodeCorrect := verifyCodeInt == user.VerifyCode
	fmt.Println(isVerifyCodeCorrect, verifyCode, user.VerifyCode)
	if !isVerifyCodeCorrect {
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
	fmt.Println(userId)
	res := types.Response{StatusCode: http.StatusOK, Success: true, Message: "user verified successfully", Data: map[string]interface{}{"userId": userId}}
	json.NewEncoder(w).Encode(res)

}
