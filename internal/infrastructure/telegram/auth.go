package telegram

import (
	"context"
	"fmt"
	"ndc/ai_bot/internal/entity"
	"ndc/ai_bot/pkg/jwt"
)

func (t *Handler) CreateUser(user entity.Client) error {
	
	userres, err := t.Usecase.CreateClient(context.Background(), user)
	if err != nil {
		return fmt.Errorf("user yaratishda xatolik: %w", err)
	}

	jwtFields := map[string]interface{}{
		"sub":         userres.Id,
		"user_role":   "user",
		// "role_id":     userres.RoleId,
		// "first_email": userres.Email,
		// "user_phone":  userres.PhoneNumber,
	}
	_, err = jwt.GenerateJWT(jwtFields, t.cfg.JWT.Secret)
	if err != nil {
		return fmt.Errorf("JWT yaratishda xatolik: %w", err)
	}
	
	

	return nil
}
