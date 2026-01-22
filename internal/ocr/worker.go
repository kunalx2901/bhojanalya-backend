package ocr

import "log"

// StartWorkers starts both OCR and LLM workers
func StartWorkers(service *Service) {
	log.Println("[WORKER] Starting OCR and LLM workers")

	go service.StartOCRWorker()
	go service.StartLLMWorker()
}
