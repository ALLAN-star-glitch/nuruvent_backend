package redis

import (
	"errors"
	"fmt"
	"math/rand"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/modules/auth/domain"
	"github.com/ALLAN-star-glitch/nuruvent-backend/internal/shared/redis"
	"time"
)

type OTPService struct{}

func NewOTPService() domain.OTPService {
    return &OTPService{}
}

func (s *OTPService) GenerateOTP() string {
    return fmt.Sprintf("%06d", rand.Intn(1000000))
}

func (s *OTPService) StoreOTP(email, otp string) error {
    return redis.Set("otp:"+email, otp, 5*time.Minute)
}

func (s *OTPService) GetOTP(email string) (string, error) {
    otp, exists, err := redis.Get("otp:" + email)
    if err != nil {
        return "", err
    }
    if !exists {
        return "", domain.ErrOTPNotFound
    }
    return otp, nil
}

func (s *OTPService) DeleteOTP(email string) error {
    return redis.Delete("otp:" + email)
}

func (s *OTPService) StoreTwoFactorOTP(email, otp string) error {
    return redis.Set("2fa:"+email, otp, 5*time.Minute)
}

func (s *OTPService) GetTwoFactorOTP(email string) (string, error) {
    otp, exists, err := redis.Get("2fa:" + email)
    if err != nil {
        return "", err
    }
    if !exists {
        return "", errors.New("2FA OTP not found or expired")
    }
    return otp, nil
}

func (s *OTPService) DeleteTwoFactorOTP(email string) error {
    return redis.Delete("2fa:" + email)
}

func (s *OTPService) StoreUserData(email string, data map[string]interface{}) error {
    key := "user:data:" + email
    if err := redis.HSet(key, data); err != nil {
        return err
    }
    return redis.Expire(key, 5*time.Minute)
}

func (s *OTPService) GetUserData(email string) (map[string]string, error) {
    result, exists, err := redis.HGetAll("user:data:" + email)
    if err != nil {
        return nil, err
    }
    if !exists || len(result) == 0 {
        return nil, errors.New("user data not found")
    }
    return result, nil
}

func (s *OTPService) DeleteUserData(email string) error {
    return redis.Delete("user:data:" + email)
}

func (s *OTPService) StoreResetData(email, otp, newPassword string) error {
    key := "reset:" + email
    data := map[string]interface{}{
        "otp":          otp,
        "new_password": newPassword,
    }
    if err := redis.HSet(key, data); err != nil {
        return err
    }
    return redis.Expire(key, 5*time.Minute)
}

func (s *OTPService) GetResetData(email string) (map[string]string, error) {
    result, exists, err := redis.HGetAll("reset:" + email)
    if err != nil {
        return nil, err
    }
    if !exists || len(result) == 0 {
        return nil, errors.New("reset data not found")
    }
    return result, nil
}

func (s *OTPService) DeleteResetData(email string) error {
    return redis.Delete("reset:" + email)
}

func (s *OTPService) SendOTPEmail(to, name, otp string) error {
    return nil
}

func (s *OTPService) SendTwoFactorOTP(to, name, otp, ipAddress, userAgent string) error {
    return nil
}

func (s *OTPService) SendPasswordResetOTP(to, name, otp string) error {
    return nil
}