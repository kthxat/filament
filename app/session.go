package app

import (
	"log"
	"sync"
	"time"

	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"

	"git.kthx.at/icedream/filament/backends"
	"git.kthx.at/icedream/filament/config"
)

var sessionTimeout = 5 * time.Minute

var activeBackendInstances = []backends.Backend{}

var sessions = map[string]*Session{}
var sessionsMutex sync.Mutex

/*func GC() {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	unsyncedGC()
}*/

func unsyncedGC() {
	for sid, session := range sessions {
		if !session.IsActive() {
			delete(sessions, sid)
		}
	}
}

func GetSessionById(id string) *Session {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	unsyncedGC()

	if session, ok := sessions[id]; ok {
		return session
	}

	return nil
}

func constructBackend(descriptor *backends.BackendDescriptor) (backend backends.Backend, err error) {
	backend, err = descriptor.New(&backends.BackendConstructionParams{
		Config: config.GetBackendConfig(descriptor.Id),
	})
	return
}

func Authenticate(username, password string) (sid string) {
	if sid = GetSessionByAccount(username, password); len(sid) > 0 {
		return
	}

	for _, backendDescriptor := range backends.GetAll() {
		backend, err := constructBackend(backendDescriptor)
		if err != nil {
			log.Printf("Construction of authenticator %s threw an error: %s",
				backendDescriptor.Id, err.Error())
			continue
		}

		authenticator, ok := backend.(backends.Authenticator)
		if !ok {
			backend.Close()
		}
		ok, err = authenticator.Authenticate(username, password)
		if err != nil {
			log.Printf("Authenticator %s threw an error: %s",
				backendDescriptor.Id, err.Error())
			continue
		}
		if !ok {
			continue
		}

		storage, ok := backend.(backends.Storage)
		if !ok {
			backend.Close()
			log.Printf("Backend %s is not authenticator and storage at the same time. This is not yet supported.",
				backendDescriptor.Id)
			continue
		}

		sessionsMutex.Lock()
		defer sessionsMutex.Unlock()

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
		session := &Session{
			updateChan:    make(chan interface{}),
			username:      username,
			passwordHash:  passwordHash,
			isActive:      true,
			authenticator: authenticator,
			storage:       storage,
		}
		sid = xid.New().String()
		sessions[sid] = session
		go session.timeoutLoop()
		return
	}
	return
}

func GetSessionByAccount(username, password string) (id string) {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()

	unsyncedGC()

	// Find an existing active session to reuse
	for sessionID, session := range sessions {
		if session.username == username &&
			session.VerifyPassword(password) {
			// Reset timeout of this session
			session.updateChan <- nil
			id = sessionID
			return
		}
	}

	return
}

type Session struct {
	mutex      sync.Mutex
	updateChan chan interface{}

	username      string
	passwordHash  []byte
	authenticator backends.Authenticator
	storage       backends.Storage

	isActive bool
}

func (s *Session) IsActive() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.isActive
}

func (s *Session) Username() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.username
}

func (s *Session) VerifyPassword(password string) bool {
	return nil == bcrypt.CompareHashAndPassword(s.passwordHash, []byte(password))
}

func (s *Session) Authenticator() backends.Authenticator {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.authenticator
}

func (s *Session) Storage() backends.Storage {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.storage
}

func (s *Session) SetStorage(storage backends.Storage) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.storage = storage
}

func (s *Session) SetAuthenticator(authenticator backends.Authenticator) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.authenticator = authenticator
}

func (s *Session) timeoutLoop() {
	for {
		select {
		case <-s.updateChan:
			continue
		case <-time.After(sessionTimeout):
			s.mutex.Lock()
			defer s.mutex.Unlock()
			s.isActive = false
			if s.storage != nil {
				s.storage.Close()
			}
			if s.authenticator != nil {
				s.authenticator.Close()
			}
			s.storage = nil
			s.authenticator = nil
		}
		break
	}
}
