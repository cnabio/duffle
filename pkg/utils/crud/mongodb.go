package crud

import "github.com/globalsign/mgo"

// MongoClaimsCollection is the name of the claims collection.
const MongoClaimsCollection = "cnab_claims"

type mongoDBStore struct {
	session    *mgo.Session
	collection *mgo.Collection
}

type doc struct {
	Name string `json:"name"`
	Data []byte `json:"data"`
}

// NewMongoDBStore creates a new storage engine that uses MongoDB
//
// The URL provided must point to a MongoDB server and database.
func NewMongoDBStore(url string) (Store, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	return &mongoDBStore{
		session:    session,
		collection: session.DB("").C(MongoClaimsCollection),
	}, nil
}

func (s *mongoDBStore) List() ([]string, error) {
	res := []doc{}
	if err := s.collection.Find("{}").All(res); err != nil {
		return []string{}, err
	}
	buf := []string{}
	for _, v := range res {
		buf = append(buf, v.Name)
	}
	return buf, nil
}

func (s *mongoDBStore) Store(name string, data []byte) error {
	return s.collection.Insert(doc{name, data})
}
func (s *mongoDBStore) Read(name string) ([]byte, error) {
	res := doc{}
	if err := s.collection.Find(map[string]string{"name": name}).One(&res); err != nil {
		return []byte{}, err
	}
	return res.Data, nil
}
func (s *mongoDBStore) Delete(name string) error {
	return s.collection.Remove(map[string]string{"name": name})
}
