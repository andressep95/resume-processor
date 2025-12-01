package middleware

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type AuthMiddleware struct {
	jwksURL string
	cache   *jwk.Cache
}

func NewAuthMiddleware(jwksURL string) *AuthMiddleware {
	// Crear cache con refresh automático cada 10 minutos
	cache := jwk.NewCache(context.Background())
	cache.Register(jwksURL, jwk.WithMinRefreshInterval(10*time.Minute))

	return &AuthMiddleware{
		jwksURL: jwksURL,
		cache:   cache,
	}
}

func (a *AuthMiddleware) ValidateJWT() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Obtener token del header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token de autorización requerido",
			})
		}

		// Verificar formato "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Formato de token inválido",
			})
		}

		tokenString := parts[1]

		// Obtener JWKS del cache (descarga automáticamente si es necesario)
		ctx := context.Background()
		keySet, err := a.cache.Get(ctx, a.jwksURL)
		if err != nil {
			log.Printf("❌ Error obteniendo JWKS: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error al obtener claves de validación",
			})
		}

		// Validar y parsear el token
		// Si el token no tiene 'kid', intentar con todas las claves
		var token jwt.Token
		var parseErr error
		
		// Primero intentar con el keyset completo
		token, parseErr = jwt.Parse(
			[]byte(tokenString),
			jwt.WithKeySet(keySet),
			jwt.WithValidate(true),
		)
		
		// Si falla por falta de 'kid', intentar con cada clave individualmente
		if parseErr != nil && strings.Contains(parseErr.Error(), "no key ID") {
			log.Printf("⚠️  Token sin 'kid', intentando con todas las claves...")
			iter := keySet.Keys(ctx)
			for iter.Next(ctx) {
				pair := iter.Pair()
				key := pair.Value.(jwk.Key)
				
				token, err = jwt.Parse(
					[]byte(tokenString),
					jwt.WithKey(key.Algorithm(), key),
					jwt.WithValidate(true),
				)
				if err == nil {
					log.Printf("✅ Token validado con clave sin kid")
					break
				}
			}
			if err != nil {
				log.Printf("❌ Error validando token con todas las claves: %v", err)
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
					"error": "Token inválido",
					"details": err.Error(),
				})
			}
		} else if parseErr != nil {
			log.Printf("❌ Error validando token: %v", parseErr)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Token inválido",
				"details": parseErr.Error(),
			})
		}

		log.Printf("✅ Token validado - Subject: %s, Issuer: %s", token.Subject(), token.Issuer())

		// Guardar información del token en el contexto para uso posterior
		c.Locals("user", token)
		c.Locals("user_id", token.Subject())

		return c.Next()
	}
}