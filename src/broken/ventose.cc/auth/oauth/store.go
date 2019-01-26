package oauth

import (
	"github.com/boltdb/bolt"
	"ventose.cc/tools"
)

const DBFOLDER = "db"
const DBFILE = "oauthstore.db"
const CLIENTBUCKET = "OAUTHCLIENTS"

type Store struct {
	clientSerializeReader chan Client
	clientsCache map[string]*Client
	db *bolt.DB
}

func getStore() (*Store, error ){
	db, err := bolt.Open(DBFOLDER + tools.GetDirSeperator() + DBFILE, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	s := &Store{
		clientSerializeReader:make(chan Client),
		db: db,
	}
	defer s.Close()

	return s, nil
}

func (s *Store) Close() error {
	return nil
}

func (s *Store) SetSession(ses *Session) error {
	return nil
}

func (s *Store) GetSession(sessId string) (*Session, error) {
	return nil, nil
}

func (s *Store) GetClients() (clients map[string]*Client, err error) {

	for client := range s.clientSerializeReader {
		clients[client.ClientId] = &client
	}

	return
}