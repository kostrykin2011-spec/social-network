package handlers

import (
	"net/http"
	"social-network/pkg/models"
	"social-network/pkg/service"

	"encoding/json"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type ProfileHandler interface {
	GetProfile(w http.ResponseWriter, r *http.Request)
	SearchProfile(w http.ResponseWriter, r *http.Request)
}

type profileHandler struct {
	profileService service.ProfileService
}

func InitUserHandler(service service.ProfileService) ProfileHandler {
	return &profileHandler{profileService: service}
}

func (handler *profileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := uuid.Parse(vars["id"])

	if err != nil {
		models.SendSuccessResponse(w, "Невалидные данные", http.StatusOK)
		return
	}

	profile, err := handler.profileService.GetById(userId)

	if err != nil {
		models.SendSuccessResponse(w, "Анкета не найдена", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(models.ProfileResponse{
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		Gender:    profile.Gender,
		Birthdate: profile.Birthdate.Format("2006-01-02 15:04:05"),
		Biography: profile.Biography,
		City:      profile.City,
	})
}

func (handler *profileHandler) SearchProfile(w http.ResponseWriter, r *http.Request) {
	lastName := r.URL.Query().Get("last_name")
	firstName := r.URL.Query().Get("first_name")
	var (
		limit  int = 100
		offset int = 0
	)

	profiles, err := handler.profileService.SearchProfile(lastName, firstName, limit, offset)

	if err != nil {
		models.SendSuccessResponse(w, err.Error(), http.StatusNotFound)
		return
	}

	var profileList []models.ProfileResponse

	for _, profile := range profiles {
		obj := models.ProfileResponse{
			FirstName: profile.FirstName,
			LastName:  profile.LastName,
			Gender:    profile.Gender,
			Birthdate: profile.Birthdate.Format("2006-01-02 15:04:05"),
			Biography: profile.Biography,
			City:      profile.City,
		}
		profileList = append(profileList, obj)
	}

	json.NewEncoder(w).Encode(profileList)
}
