package auth

import (
 "github.com/alexedwards/argon2id"
 "github.com/golang-jwt/jwt/v5"
 "github.com/google/uuid"

 "time"
 "strings"
 "net/http"
 "fmt"
 "errors"
)

func HashPassword(password string) (string,error){
    hash, err := argon2id.CreateHash(password,argon2id.DefaultParams)
    if err != nil {
     return "", err
    }
    return hash, nil
}

func CheckPasswordHash(password , hash string) (bool, error){
     match, err :=   argon2id.ComparePasswordAndHash(password,hash)
     if err != nil {
        return false, err
     }
    return match, nil
}

func MakeJWT(userID uuid.UUID, tokenSecret string,expiresIn time.Duration) (string, error){
  signingKey := []byte(tokenSecret)
  currentTime := time.Now().UTC()

  claims := jwt.RegisteredClaims{
                Issuer: "chirpy-access",
                IssuedAt: jwt.NewNumericDate(currentTime),
                ExpiresAt: jwt.NewNumericDate(currentTime.Add(expiresIn)),
                Subject: userID.String(),
                }

    token :=  jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
     signedToken, err := token.SignedString(signingKey)
     if err != nil {
        return "", err
     }
    return signedToken, err
}

func ValidateJWT(tokenString, tokenSecret string)(uuid.UUID, error){
     claimsStruct := &jwt.RegisteredClaims{}
        token, err :=  jwt.ParseWithClaims(
             tokenString,
             claimsStruct,
             func(token *jwt.Token) (interface{}, error){
                if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok{
                   return nil, fmt.Errorf("unexpected signing method: %v",token.Header["alg"])
                }
                return []byte(tokenSecret), nil
             },
            )
        if err != nil {
         return uuid.Nil, err
        }
      userIDString, err :=  token.Claims.GetSubject()
      if err != nil {
            return uuid.Nil, err
      }
    userID, err := uuid.Parse(userIDString)
    if err != nil {
        return uuid.Nil, fmt.Errorf("invalid user ID in token: %w",err)
    }
    return userID, nil
}


var ErrNoAuthHeaderPayload = errors.New("auth header not found or malformed")
func GetBearerToken(headers http.Header) (string, error){
      authHeader :=  headers.Get("Authorization")
      if authHeader == "" {
            return "", ErrNoAuthHeaderPayload
      }
   splitData := strings.Split(authHeader," ")
   if len(splitData) < 2 ||  splitData[0] != "Bearer"{
    return "",ErrNoAuthHeaderPayload
   }
    return strings.TrimSpace(splitData[1]), nil

}

func GetAPIKey(headers http.Header) (string, error){
    val := headers.Get("Authorization")
    if val == ""{
        return "", errors.New("no authorization header included")
    }
    vals := strings.Split(val, " ")
    if len(vals) !=2 || vals[0] != "ApiKey"{
        return "", errors.New("malformed authorization header")
    }
    return vals[1], nil
}
