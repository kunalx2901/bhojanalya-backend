package ocr

import "log"

func StartWorker(service *Service) {
	go func() {
		log.Println("OCR worker started")
		if err := service.Start(); err != nil {
			log.Fatal("OCR worker crashed:", err)
		}
	}()
}
