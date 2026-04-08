package main
import(
"log"
"net/http"
"fmt"
"sync/atomic"
"encoding/json"
"strings"
"database/sql"
"os"
"time"
"errors"
"sort"

"github.com/joho/godotenv"
"github.com/shvarshini/chirpy/internal/database"
"github.com/shvarshini/chirpy/internal/auth"
"github.com/google/uuid"
_ "github.com/lib/pq"
)



type apiConfig struct {
       fileserverHits atomic.Int32
       DB *database.Queries
       Platform string
       jwtSecret string
       polkaKey string
    }

type parameters struct {
    Body string `json:"body"`
 }

type UserRequest struct {
 Password string `json:"password"`
 Email string `json:"email"`
}


type UserResponse struct {
ID uuid.UUID `json:"id"`
CreatedAt time.Time  `json:"created_at"`
UpdatedAt time.Time   `json:"updated_at"`
Email string `json:"email"`
IsChirpyRed bool `json:"is_chirpy_red"`
}

type ChirpRequest struct{
 Body string `json:"body"`
}

type ChirpResponse struct {
ID uuid.UUID  `json:"id"`
CreatedAt time.Time  `json:"created_at"`
UpdatedAt time.Time   `json:"updated_at"`
Body string `json:"body"`
UserID uuid.UUID `json:"user_id"`
}

type LoginRequest struct {
 Password string `json:"password"`
 Email string `json:"email"`
 ExpiresInSeconds int `json:"expires_in_seconds"`
}

func ( cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request){
           decoder := json.NewDecoder(r.Body)
           params := UserRequest{}
           err := decoder.Decode(&params)

           if err!= nil {
              respondWithError(w,http.StatusBadRequest,"Could'nt decode parameters")
              return
           }
      hashedPassword, err := auth.HashPassword(params.Password)
      if err!= nil {
                    respondWithError(w,http.StatusInternalServerError,"Could'nt hash password")
                    return
                 }

     user, err :=  cfg.DB.CreateUser(r.Context(), database.CreateUserParams{
        ID : uuid.New(),
        CreatedAt : time.Now().UTC(),
        UpdatedAt: time.Now().UTC(),
        Email: params.Email,
        HashedPassword: hashedPassword,
       })

   if err != nil {
      respondWithError(w,http.StatusInternalServerError, "Could'nt create user")
      return
   }
   respondWithJSON(w,http.StatusOK,UserResponse{
   ID : user.ID,
   CreatedAt: user.CreatedAt,
   UpdatedAt : user.UpdatedAt,
   Email : user.Email,
   IsChirpyRed : user.IsChirpyRed,
   })
}

func(cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request){
    tokenString, err := auth.GetBearerToken(r.Header)
              if err!= nil {
                    respondWithError(w,http.StatusUnauthorized,"Missing Token")
                    return
              }
          secret := os.Getenv("JWT_SECRET")
             userID, err := auth.ValidateJWT(tokenString,secret)
              if err!= nil {
                             respondWithError(w,http.StatusUnauthorized,"Invalid Token")
                             return
                       }
    decoder := json.NewDecoder(r.Body)
           params := UserRequest{}
           err = decoder.Decode(&params)
    if err!= nil {
               respondWithError(w,http.StatusBadRequest,"Could'nt decode parameters")
               return
                      }
   hashedPassword, err := auth.HashPassword(params.Password)
    if err!= nil {
                  respondWithError(w,http.StatusInternalServerError,"Could'nt hash password")
                  return
                         }
     user, err := cfg.DB.UpdateUser(r.Context(), database.UpdateUserParams{
            Email: params.Email,
            HashedPassword: hashedPassword,
            UpdatedAt: time.Now().UTC(),
            ID: userID,
     })
    if err!= nil {
                     respondWithError(w,http.StatusInternalServerError,"Could'nt update user")
                     return
                            }
     respondWithJSON(w,http.StatusOK,UserResponse{
        ID: userID,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        Email: user.Email,
        IsChirpyRed : user.IsChirpyRed,
     })
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request){
          tokenString, err := auth.GetBearerToken(r.Header)
          if err!= nil {
                respondWithError(w,http.StatusUnauthorized,"Missing Token")
                return
          }
      secret := os.Getenv("JWT_SECRET")
         userID, err := auth.ValidateJWT(tokenString,secret)
          if err!= nil {
                         respondWithError(w,http.StatusUnauthorized,"Invalid Token")
                         return
                   }
        decoder := json.NewDecoder(r.Body)
        params := ChirpRequest{}
        err = decoder.Decode(&params)

         if err!= nil {
            respondWithError(w,http.StatusBadRequest,"Could'nt decode parameters")
            return
                   }
          trimmedBody :=   strings.TrimSpace(params.Body)
        if trimmedBody == "" {
        		respondWithError(w, http.StatusBadRequest, "Chirp cannot be empty")
        		return
        	}

         if len(trimmedBody) > 140 {
               respondWithError(w,http.StatusBadRequest,"The chirp is too long")
                return
            }
         cleaned := getCleanedBody(trimmedBody)

      chirp, err :=   cfg.DB.CreateChirp(r.Context(), database.CreateChirpParams{
                ID: uuid.New(),
                CreatedAt: time.Now().UTC(),
                UpdatedAt: time.Now().UTC(),
                Body: cleaned,
                UserID:userID,
         })
     if err != nil {
        respondWithError(w,http.StatusInternalServerError,"Could'nt create chirp")
        return
     }
    respondWithJSON(w,http.StatusCreated,ChirpResponse{
     ID : chirp.ID,
     CreatedAt: chirp.CreatedAt,
     UpdatedAt: chirp.UpdatedAt,
     Body: chirp.Body,
     UserID: chirp.UserID,
    })

}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request){
    authorIDString := r.URL.Query().Get("author_id")
    sortDirection := r.URL.Query().Get("sort")

    if sortDirection != "desc"{
        sortDirection = "asc"
    }
    var dbChirps []database.Chirp
    var err error

    if authorIDString != ""{
        authorID, err := uuid.Parse(authorIDString)
        if err != nil {
            respondWithError(w,http.StatusBadRequest,"Invalid author ID")
            return
        }
        dbChirps, err = cfg.DB.GetChirpsByAuthor(r.Context(),authorID)
    }else {
        dbChirps, err = cfg.DB.GetChirps(r.Context())
    }

    if err != nil {
     respondWithError(w,http.StatusInternalServerError,"Could'nt retrieve chirps")
     return
    }
    chirps := []ChirpResponse{}
    for _,dbChirp := range dbChirps{
        chirps = append(chirps, ChirpResponse{
            ID: dbChirp.ID,
            CreatedAt: dbChirp.CreatedAt,
            UpdatedAt: dbChirp.UpdatedAt,
            Body: dbChirp.Body,
            UserID: dbChirp.UserID,
        })
    }

    if sortDirection == "desc"{
        sort.Slice(chirps, func(i,j int) bool{
            return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
        })
    }
    respondWithJSON(w,http.StatusOK,chirps)
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request){
      chirpIDString := r.PathValue("chirpID")
      chirpID, err := uuid.Parse(chirpIDString)
       if err!= nil {
                          respondWithError(w,http.StatusBadRequest,"Invalid chirp ID")
                          return
                    }

     tokenString, err := auth.GetBearerToken(r.Header)
              if err!= nil {
                    respondWithError(w,http.StatusUnauthorized,"Missing Token")
                    return
              }
          secret := os.Getenv("JWT_SECRET")
             userID, err := auth.ValidateJWT(tokenString,secret)
              if err!= nil {
                             respondWithError(w,http.StatusUnauthorized,"Invalid Token")
                             return
                       }
      chirp, err := cfg.DB.GetChirp(r.Context(),chirpID)
      if err != nil {
        if err == sql.ErrNoRows {
          respondWithError(w,http.StatusNotFound, "Chirp not found")
          return
        }
            respondWithError(w,http.StatusInternalServerError,"Could'nt fetch chirp")
            return
      }
     if chirp.UserID != userID {
        respondWithError(w,http.StatusForbidden,"You are not authorized to delete this chirp")
        return
     }
     err = cfg.DB.DeleteChirp(r.Context(),chirpID)
     if err != nil {
            respondWithError(w,http.StatusInternalServerError,"Could'nt delete chirp")
            return
     }
    w.WriteHeader(http.StatusNoContent)
}

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request){

         decoder := json.NewDecoder(r.Body)
         params := LoginRequest{}
         err := decoder.Decode(&params)
         if err!= nil {
                 respondWithError(w,http.StatusBadRequest,"Could'nt decode parameters")
                 return
                      }
        user, err := cfg.DB.GetUsersByEmail(r.Context(),params.Email)
        if err != nil {
           respondWithError(w,http.StatusUnauthorized,"Incorrect email or password")
           return
        }

       match,err := auth.CheckPasswordHash(params.Password,user.HashedPassword)
       if err != nil {
            respondWithError(w,http.StatusInternalServerError,"Error checking Password")
            return
       }
        if !match {
            respondWithError(w,http.StatusUnauthorized,"Incorrect email or password")
            return
        }
        defaultExpiration := time.Hour
        expiration := defaultExpiration
        if params.ExpiresInSeconds !=0 && time.Duration(params.ExpiresInSeconds) * time.Second < defaultExpiration{
            expiration = time.Duration(params.ExpiresInSeconds) * time.Second
        }
       token, err := auth.MakeJWT(user.ID,cfg.jwtSecret,expiration)
       if err != nil {
                  respondWithError(w,http.StatusInternalServerError,"Could'nt create access token")
                  return
               }
        respondWithJSON(w,http.StatusOK,struct{
            ID uuid.UUID `json:"id"`
            CreatedAt time.Time `json:"created_at"`
            UpdatedAt time.Time `json:"updated_at"`
            Email string `json:"email"`
            IsChirpyRed bool `json:"is_chirpy_red"`
            Token string `json:"token"`
        }{
            ID: user.ID,
            CreatedAt: user.CreatedAt,
            UpdatedAt : user.UpdatedAt,
            Email: user.Email,
            IsChirpyRed: user.IsChirpyRed,
            Token: token,
        })
}

func respondWithJSON(w http.ResponseWriter,code int, payload interface{}){
 data, err := json.Marshal(payload)
 if err != nil {
    log.Printf("Error marshaling JSON: %S",err)
    w.WriteHeader(500)
    return
 }
w.Header().Set("Content-Type","application/json")
  w.WriteHeader(code)
  w.Write(data)
}

func respondWithError(w http.ResponseWriter , code int, msg string){
if code > 499 {
log.Printf("Responding with 5XX error: %s",msg)
}
 type errorResponse struct {
    Error string `json:"error"`
 }
 respondWithJSON(w,code, errorResponse{
 Error: msg})
}

func getCleanedBody(body string) string {
    badWords := map[string]struct{}{
     "kerfuffle": {},
     "sharbert" : {},
     "fornax" : {},
     }

     words := strings.Split(body, " ")

     for i, word := range words {
      loweredWord := strings.ToLower(word)
      if _, ok :=  badWords[loweredWord]; ok{
         words[i] = "****"
      }
     }
  return strings.Join(words," ")
}


 func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
      return http.HandlerFunc(func(w http.ResponseWriter ,r *http.Request){
         cfg.fileserverHits.Add(1)
         next.ServeHTTP(w,r)
      })
    }

 func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter ,r *http.Request){
      w.Header().Add("Content-Type" ,"text/html; charset= utf-8")
      w.WriteHeader(http.StatusOK)
      template := `<html>
                     <body>
                       <h1>Welcome, Chirpy Admin</h1>
                       <p>Chirpy has been visited %d times!</p>
                     </body>
                   </html>`
      fmt.Fprintf(w,fmt.Sprintf(template,cfg.fileserverHits.Load()))

    }

 func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request){
         if cfg.Platform != "dev"{
          respondWithError(w,http.StatusForbidden,"Forbidden: Reset is only allowed in dev mode")
          return
         }
     err := cfg.DB.DeleteUsers(r.Context())
     if err != nil {
       respondWithError(w,http.StatusInternalServerError,"Could'nt delete users")
       return
     }

         cfg.fileserverHits.Store(0)
         w.Header().Add("Content-Type", "text/plain; charset= utf-8")
         w.WriteHeader(http.StatusOK)
         w.Write([]byte("Hits reset  and database cleared"))
    }

func (cfg *apiConfig) handlerWebhook(w http.ResponseWriter, r *http.Request){
    apiKey, err := auth.GetAPIKey(r.Header)
    if err != nil || apiKey != cfg.polkaKey{
        respondWithError(w,http.StatusUnauthorized,"Invalid API Key")
        return
    }
    type parameters struct {
        Event string `json:"event"`
        Data  struct {
            UserID string `json:"user_id"`
        } `json:"data"`
    }
    decoder := json.NewDecoder(r.Body)
    params := parameters{}
    err = decoder.Decode(&params)
    if err!= nil {
        respondWithError(w,http.StatusBadRequest,"Could'nt decode parameters")
        return
                 }
     if params.Event != "user.upgraded"{
     w.WriteHeader(http.StatusNoContent)
     return
     }
    userID, err := uuid.Parse(params.Data.UserID)
    if err != nil{
    respondWithError(w,http.StatusBadRequest,"Invalid user ID format")
    return
    }
    err = cfg.DB.UpgradeUserToChirpyRed(r.Context(),database.UpgradeUserToChirpyRedParams{
        ID : userID,
        UpdatedAt : time.Now().UTC(),
    })
    if err != nil {
        if errors.Is(err, sql.ErrNoRows){
            respondWithError(w,http.StatusNotFound,"User not found")
            return
        }
        respondWithError(w,http.StatusInternalServerError,"Could'nt upgrade user")
        return
    }
    w.WriteHeader(http.StatusNoContent)
}



func main(){

    err := godotenv.Load()
    if err!= nil {
     log.Fatal("Error loading .env file")
    }

    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == ""{
        log.Fatal("JWT_SECRET environment variable is not set")
    }

    dbURL := os.Getenv("DB_URL")
     db, err := sql.Open("postgres", dbURL)
     dbQueries := database.New(db)

     polkaKey := os.Getenv("POLKA_KEY")
      if polkaKey == ""{
             log.Fatal("POLKA_KEY environment variable is not set")
         }

    apiCfg := &apiConfig{
    DB : dbQueries,
    Platform : os.Getenv("PLATFORM"),
    jwtSecret : jwtSecret,
    polkaKey: polkaKey,
    }

    mux:= http.NewServeMux()

    fileServer := http.FileServer(http.Dir("."))
    mux.Handle("/",apiCfg.middlewareMetricsInc(fileServer))
    mux.HandleFunc("GET /admin/metrics",apiCfg.handlerMetrics)
    mux.HandleFunc("POST /admin/reset",apiCfg.handlerReset)
    mux.HandleFunc("POST /api/login",apiCfg.handlerLogin)
    mux.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
    mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
    mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
    mux.HandleFunc("POST /api/users",apiCfg.handlerCreateUser)
    mux.HandleFunc("PUT /api/users",apiCfg.handlerUpdateUser)
    mux.HandleFunc("POST /api/polka/webhooks",apiCfg.handlerWebhook)



   srv := &http.Server{
        Addr: ":8080",
        Handler : mux,
    }

   err = srv.ListenAndServe()
   if err != nil {
   log.Fatal(err);
   }
}