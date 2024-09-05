package application

import (
	"fmt"
	"math/rand"
	"time"
)

type ResilliencyService struct {
}

func NewResilliencyService() *ResilliencyService {
	return &ResilliencyService{}
}

func (r *ResilliencyService) GenerateResilliency(minDelaySec int32, maxDelaySec int32, statusCodes []uint32) (string, uint32) {
	delay := rand.Intn(int(maxDelaySec-minDelaySec)) + int(1+minDelaySec)
	delaySecond := time.Duration(delay) * time.Second
	time.Sleep(delaySecond)

	idx := rand.Intn(len(statusCodes))
	str := fmt.Sprintf("The time now is %v, execution delayed for %v seconds",
		time.Now().Format("15:04:05.000"), delay)

	return str, statusCodes[idx]
}
