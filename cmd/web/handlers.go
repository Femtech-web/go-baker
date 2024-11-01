package main

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/femtech-web/baker/internal/models"
	"github.com/femtech-web/baker/internal/validator"
)

func (app *application) getHome(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, http.StatusOK, "home.tmpl", data)
}

// Handle user signup
type SignupForm struct {
	Name                string `form:"name"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) getSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = SignupForm{}

	app.render(w, http.StatusOK, "signup.tmpl", data)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	var form SignupForm

	err := app.decodeForm(r, "form", &form)
	if err != nil {
		fmt.Printf("%v:", err)
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		return
	}

	err = app.users.Insert(form.Name, form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrDuplicateEmail) {
			form.AddFieldError("email", "email already exists")
			data := app.newTemplateData(r)
			data.Form = form

			app.render(w, http.StatusUnprocessableEntity, "signup.tmpl", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	app.sessionManager.Put(r.Context(), "flash", "user signup successfully. Please login!")

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Handle user login
type LoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

func (app *application) getLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = LoginForm{}

	app.render(w, http.StatusOK, "login.tmpl", data)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	var form LoginForm

	err := app.decodeForm(r, "form", &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form

		app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		return
	}

	user, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrInvalidCredentials) {
			form.AddNonFieldError("email or password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form

			app.render(w, http.StatusUnprocessableEntity, "login.tmpl", data)
		} else {
			app.serverError(w, err)
		}

		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", user.ID)
	initialPath := app.sessionManager.PopString(r.Context(), "redirectPathAfterLogin")

	if initialPath != "" {
		http.Redirect(w, r, initialPath, http.StatusSeeOther)
		return
	}

	if !user.IsActive {
		http.Redirect(w, r, "/features", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/predict", http.StatusSeeOther)
}

// Handle logout
func (app *application) userLogout(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "you logged out successfully")

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Handle Adding features on first time
type FeaturesData struct {
	Features []string `form:"features"`
}

func (app *application) getFeatures(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, http.StatusOK, "features.tmpl", data)
}

func (app *application) addFeatures(w http.ResponseWriter, r *http.Request) {
	var formData FeaturesData

	exUser, err := app.users.Get(app.currentUser(r))
	if err != nil {
		app.serverError(w, err)
		return
	}

	if exUser.Features != nil {
		app.clientError(w, http.StatusConflict)
		return
	}

	err = app.decodeForm(r, "multipart", &formData)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	var features []string
	err = json.Unmarshal([]byte(formData.Features[0]), &features)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	err = app.users.AddFeatures(app.currentUser(r), features)
	if err != nil {
		app.serverError(w, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusSeeOther)
	json.NewEncoder(w).Encode(map[string]string{"redirect": "/import"})

	fmt.Printf("formData: %v", features)
}

// Handle importing production data on first time
func (app *application) getImport(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, http.StatusOK, "import.tmpl", data)
}

// func (app *application) savePastData(w http.ResponseWriter, r *http.Request) {
// 	file, _, err := r.FormFile("file")
// 	if err != nil {
// 		app.clientError(w, http.StatusBadRequest)
// 		return
// 	}

// 	defer file.Close()

// 	reader := csv.NewReader(file)
// 	records, err := reader.ReadAll()
// 	fmt.Printf("%v", records)
// 	if err != nil {
// 		app.serverError(w, err)
// 		return
// 	}

// 	columns, err := app.features.GetColumns()
// 	fmt.Printf("columns: %v", columns)
// 	if err != nil {
// 		app.serverError(w, err)
// 		return
// 	}

// 	header := records[0]

// 	csvToDBMap, features := app.csvMap(header, columns)
// 	fmt.Printf("csvMap: %v", csvToDBMap)

// 	for _, row := range records[1:] {
// 		dateStr := row[csvToDBMap["date"]]
// 		date, err := time.Parse("2006-01-02", dateStr)
// 		if err != nil {
// 			app.serverError(w, err)
// 			return
// 		}

// 		loavesStr := row[csvToDBMap["loaves"]]
// 		loaves, err := strconv.Atoi(loavesStr)
// 		if err != nil {
// 			app.serverError(w, err)
// 			return
// 		}

// 		feature1Str := row[csvToDBMap["feature1"]]
// 		feature1, err := strconv.Atoi(feature1Str)
// 		if err != nil {
// 			app.serverError(w, err)
// 			return
// 		}

// 		feature2Str := row[csvToDBMap["feature2"]]
// 		feature2, err := strconv.Atoi(feature2Str)
// 		if err != nil {
// 			app.serverError(w, err)
// 			return
// 		}

// 		feature3Str := row[csvToDBMap["feature3"]]
// 		feature3, err := strconv.Atoi(feature3Str)
// 		if err != nil {
// 			app.serverError(w, err)
// 			return
// 		}

// 		app.features.Insert(date, loaves, feature1, feature2, feature3, features)
// 	}
// 	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
// }

func (app *application) savePastData(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		app.serverError(w, err)
		return
	}

	columns, err := app.features.GetColumns()
	if err != nil {
		app.serverError(w, err)
		return
	}

	header := records[0]
	csvToDBMap, features := app.csvMap(header, columns)

	err = app.features.InsertRecords(records[1:], csvToDBMap, features)
	if err != nil {
		app.serverError(w, err)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

// New function to abstract record insertion

// Handle Predict
func (app *application) getPredict(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)

	app.render(w, http.StatusOK, "predict.tmpl", data)
}
