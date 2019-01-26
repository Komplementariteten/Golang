package ctrl

import (
	"net/http"
)

type StorageHandle struct {
	StorageOnline bool
}

func (s *StorageHandle) CheckStorage() {
	go func(s *StorageHandle) {

		//serverold.Portal.StorageWaiter.Wait()
		s.StorageOnline = true
	}(s)
}

func NewStorageHandle() *StorageHandle {
	s := new(StorageHandle)
	s.StorageOnline = false
	s.CheckStorage()
	return s
}

func (h *StorageHandle) ControllStorage(r *Request, w http.ResponseWriter) {

	if !h.StorageOnline {
		logAndError(w, "Storage not Online")
		return
	}

	switch r.Command {
	default:
		http.Error(w, "Command Not Found", http.StatusNotFound)
	case "ping":
		if this.Portal == nil {
			logAndError(w, "Storage not initialized for Ping")
			return
		}
	}
}

func (h StorageHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
