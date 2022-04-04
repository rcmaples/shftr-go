package app

// import (
// 	"os"

// 	"golang.org/x/oauth2"
// 	"golang.org/x/oauth2/google"
// )

// var googleOauthConfig = &oauth2.Config{
// 	RedirectURL:  "http://localhost:4000/auth/google/callback",
// 	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
// 	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
// 	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
// 	Endpoint:     google.Endpoint,
// }

// const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

// func oauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
// 	oauthState := generateStateOauthCookie(w)
// 	u := googleOauthConfig.AuthCodeURL(oauthState)
// 	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
// }

// func generateStateOauthCookie(w http.ResponseWriter) string {
// 	var exp = time.Now().Add(7 * 24 * time.Hour)
// 	b := make([]byte, 16)
// 	rand.Read(b)
// 	state := b64.URLEncoding.EncodeToString(b)
// 	cookie := http.Cookie{Name: "auth", Value: state, Expires: exp}
// 	http.SetCookie(w, &cookie)
// 	return state
// }

// func oauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
// 	helpers.Logger.Println("calling back...")
// 	oauthState, _ := r.Cookie("auth")

// 	if r.FormValue("state") != oauthState.Value {
// 		helpers.Logger.Println("invalid oauth google state")
// 		http.Redirect(w, r, "/", http.StatusFound)
// 		return
// 	}

// 	data, err := getUserDataFromGoogle(r.FormValue("code"))
// 	if err != nil {
// 		helpers.Logger.Println("error: ", err.Error())
// 		http.Redirect(w, r, "/", http.StatusFound)
// 		return
// 	}

// 	authedUser, err := helpers.GoogleHelper(data)
// 	if err != nil {
// 		helpers.Logger.Println("error: ", err)
// 	}

// 	tokCookie, _ := cookieMaker(authedUser)
// 	w.Header().Add("Access-Control-Expose-Headers", "Set-Cookie")
// 	http.SetCookie(w, &tokCookie)
// 	helpers.Logger.Printf("w:\n%+v\n", w.Header())
// 	http.Redirect(w, r, "/", http.StatusFound)
// }

// func getUserDataFromGoogle(code string) (models.GoogleUser, error) {
// 	var gu models.GoogleUser

// 	token, err := googleOauthConfig.Exchange(context.Background(), code)
// 	if err != nil {
// 		return models.GoogleUser{}, fmt.Errorf("code exchange wrong: %s", err.Error())
// 	}

// 	resp, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
// 	if err != nil {
// 		return models.GoogleUser{}, fmt.Errorf("failed getting user info: %s", err.Error())
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return models.GoogleUser{}, fmt.Errorf("failed read response: %s", err.Error())
// 	}
// 	json.Unmarshal(body, &gu)
// 	return gu, nil
// }

// func generateKey(w http.ResponseWriter, r *http.Request) {
// 	jwtContext := r.Context().Value(contextKey("user")).(jwt.MapClaims)
// 	user := helpers.UnmarshalToken(jwtContext)
// 	org := user.Org
// 	id := user.Id
// 	name := user.Name

// 	dsClient := services.GetDB()
// 	ctx := context.Background()
// 	q := datastore.NewQuery(helpers.Users).Filter("Org =", org).Filter("GoogleId =", id)

// 	c, _ := dsClient.Count(ctx, q)
// 	if c < 1 {
// 		helpers.Logger.Printf("no user found for id: %s", id)
// 		err := errors.New("no user found with id")
// 		errorJson(w, err, http.StatusBadRequest)
// 		return
// 	}

// 	secret := []byte(os.Getenv("JWT_SECRET"))
// 	exp := jwt.NewNumericDate(time.Now().AddDate(99, 0, 0)) // 99 years from now
// 	iat := jwt.NewNumericDate(time.Now())

// 	env := flag.Lookup("env").Value.String()
// 	var clientUrl string
// 	clientUrl = os.Getenv("CLIENT_URL_DEV")
// 	if env != "dev" {
// 		clientUrl = os.Getenv("CLIENT_URL_PROD")
// 	}
// 	domain := strings.Split(clientUrl, "://")[1]
// 	if strings.Contains(domain, ":") {
// 		domain = strings.Split(domain, ":")[0]
// 	}

// 	type customClaims struct {
// 		Id    string `json:"id"`
// 		Name  string `json:"name"`
// 		Org   string `json:"org"`
// 		Email string `json:"email"`
// 		jwt.RegisteredClaims
// 	}

// 	claims := &customClaims{
// 		user.Id,
// 		user.Name,
// 		user.Org,
// 		user.Email,
// 		jwt.RegisteredClaims{
// 			ExpiresAt: exp,
// 			Issuer:    domain,
// 			IssuedAt:  iat,
// 		},
// 	}

// 	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	stok, err := tok.SignedString(secret)
// 	if err != nil {
// 		helpers.Logger.Println("error signing token: ", err)
// 		err = errors.New(err.Error() + errors.New("error signing token").Error())
// 		errorJson(w, err, http.StatusInternalServerError)
// 	}

// 	apikey := models.APIKey{
// 		GoogleId:  id,
// 		UserName:  name,
// 		Org:       org,
// 		Token:     stok,
// 		CreatedAt: time.Now(),
// 	}

// 	ik := datastore.IncompleteKey(helpers.Keys, nil)
// 	_, err = dsClient.Put(ctx, ik, &apikey)
// 	if err != nil {
// 		helpers.Logger.Println("error saving key: ", err)
// 		errorJson(w, err, http.StatusInternalServerError)
// 		return
// 	}

// 	responseJson(w, http.StatusOK, stok, "key")
// }

// func cookieMaker(user models.ShftrUser) (http.Cookie, string) {
// 	// Get the "domain" value based on what env we're in (dev|prod)
// 	env := flag.Lookup("env").Value.String()
// 	var clientUrl string
// 	clientUrl = os.Getenv("CLIENT_URL_DEV")
// 	if env != "dev" {
// 		clientUrl = os.Getenv("CLIENT_URL_PROD")
// 	}
// 	domain := strings.Split(clientUrl, "://")[1]
// 	if strings.Contains(domain, ":") {
// 		domain = strings.Split(domain, ":")[0]
// 	}

// 	secret := []byte(os.Getenv("JWT_SECRET"))
// 	expirationTime := jwt.NewNumericDate(time.Now().AddDate(0, 0, 7)) // 7 days from Now() in ms.
// 	iat := jwt.NewNumericDate(time.Now())

// 	type customClaims struct {
// 		Id    string `json:"id"`
// 		Name  string `json:"name"`
// 		Org   string `json:"org"`
// 		Email string `json:"email"`
// 		jwt.RegisteredClaims
// 	}

// 	claims := &customClaims{
// 		user.GoogleId,
// 		user.Name,
// 		user.Org,
// 		user.Email,
// 		jwt.RegisteredClaims{
// 			ExpiresAt: expirationTime,
// 			Issuer:    domain,
// 			IssuedAt:  iat,
// 		},
// 	}

// 	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	stok, err := tok.SignedString(secret)
// 	if err != nil {
// 		helpers.Logger.Println("error signing token: ", err)
// 	}

// 	cookie := http.Cookie{Name: "token", Value: stok, Expires: time.Now().AddDate(0, 0, 7)}

// 	return cookie, stok
// }
