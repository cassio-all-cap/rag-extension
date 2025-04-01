package oauth

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"net/http"
)

// Service provides endpoints to allow this agent to be authorized.
type Service struct {
	conf *oauth2.Config
}

func NewService(clientID, clientSecret, callback string) *Service {
	return &Service{
		conf: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  callback,
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://github.com/login/oauth/authorize",
				TokenURL: "https://github.com/login/oauth/access_token",
			},
		},
	}
}

const (
	STATE_COOKIE = "oauth_state"
)

// PreAuth is the landing page that the user arrives at when they first attempt
// to use the agent while unauthorized.  You can do anything you want here,
// including making sure the user has an account on your side.  At some point,
// you'll probably want to make a call to the authorize endpoint to authorize
// the app.
func (s *Service) PreAuth(w http.ResponseWriter, r *http.Request) {
	state := uuid.New().String()

	// Configurar o cookie mais permissivo possível para testes
	stateCookie := &http.Cookie{
		Name:     STATE_COOKIE,
		Value:    state,
		MaxAge:   10 * 60,
		Path:     "/",
		Secure:   false, // true em produção
		HttpOnly: false, // permitir leitura no dev, se quiser
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, stateCookie)

	// Redirecionar com o mesmo state
	authURL := s.conf.AuthCodeURL(state)
	http.Redirect(w, r, authURL, http.StatusFound)
}


// PostAuth is the landing page where the user lads after authorizing.  As
// above, you can do anything you want here.  A common thing you might do is
// get the user information and then perform some sort of account linking in
// your database.
func (s *Service) PostAuth(w http.ResponseWriter, r *http.Request) {
	// ⚠️ Ignorando validação de state para ambiente de desenvolvimento
	fmt.Println("⚠️ Ignorando validação de state (somente dev)")

	code := r.URL.Query().Get("code")
	if code == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("código de autorização ausente"))
		return
	}

	token, err := s.conf.Exchange(r.Context(), code)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("erro ao trocar código por token: %v", err)))
		return
	}

	// ✅ Aqui está o token real
	fmt.Println("🔐 GitHub Token:", token.AccessToken)

	// Opcional: salvar em cookie
	http.SetCookie(w, &http.Cookie{
		Name:  "github_token",
		Value: token.AccessToken,
		Path:  "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login finalizado com sucesso! Pode fechar esta aba."))
}


