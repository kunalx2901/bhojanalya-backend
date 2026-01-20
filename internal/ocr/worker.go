package ocr

import (
	"log"
	"time"
)

func StartWorker(service *Service) {
	go func() {
		for {
			err := service.ProcessOne()

			if err != nil {
				log.Println("OCR worker:", err)
			}

			// ✅ Always sleep to avoid tight loop
			time.Sleep(2 * time.Second)
		}
	}()
}
