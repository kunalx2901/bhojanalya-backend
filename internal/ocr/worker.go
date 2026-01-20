package ocr

import (
	"log"
	"time"
)

func StartWorker(service *Service) {
	go func() {
		for {
			if err := service.processOne(); err != nil {
				log.Println("OCR idle or error:", err)
				time.Sleep(5 * time.Second)
			}
		}
	}()
}
