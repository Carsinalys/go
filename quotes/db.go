package quotes

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

type DB struct {
	db *bolt.DB
}

const (
	quoteBucket = "shit"
)

// Open opens the database file at path and returns a DB or an error.
func Open(path string) (*DB, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Open: cannot open DB file "+path)
	}
	return &DB{
		db: db,
	}, nil
}

func (d *DB) Close() error {
	err := d.db.Close()
	if err != nil {
		return errors.Wrap(err, "Close: cannot close database")
	}
	return nil
}

// Create takes a quote and saves it to the database, using the author name
// as the key. If the author already exists, Create returns an error.
func (d *DB) Create(q *Quote) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(quoteBucket))

		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		v := b.Get([]byte(q.Author))

		if v == nil {
			buffer, err := q.Serialize()

			if err != nil {
				return fmt.Errorf("can`t serialize quote: %s", err)
			}
			error := b.Put([]byte(q.Author), buffer)

			if error != nil {
				return fmt.Errorf("put data to bucket: %s", error)
			}
		}

		return errors.New("recodr already exists")
	})

	return err
}

// udate value in DB if it exists
func (d *DB) Update(q *Quote) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(quoteBucket))
		if err != nil {
			return err
		}

		v := bucket.Get([]byte(q.Author))
		if len(v) != 0 {
			buffer, err := q.Serialize()

			if err != nil {
				return fmt.Errorf("can`t serialize quote: %s", err)
			}
			error := bucket.Put([]byte(q.Author), buffer)

			if error != nil {
				return fmt.Errorf("update data to bucket: %s", error)
			}
		}

		return err
	})

	return err
}

// Get takes an author name and retrieves the corresponding quote from the DB.
func (d *DB) Get(author string) (*Quote, error) {
	q := &Quote{}
	err := d.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(quoteBucket))
		if bucket == nil {
			return errors.Errorf("Cannot get %s - bucket %s not found", author, quoteBucket)
		}
		v := bucket.Get([]byte(author))

		if v == nil {
			return errors.New("can`t fin record")
		}

		err := q.Deserialize(v)
		if err != nil {
			return errors.Wrapf(err, "Get: cannot deserialize %s", v)
		}
		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "Get: DB.View() failed")
	}

	return q, nil
}

func (d *DB) Delete(author string) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(quoteBucket))
		if err != nil {
			return err
		}

		bucket.Delete([]byte(author))

		return err
	})

	return err
}

// // List lists all records in the DB.
// func (d *DB) List() ([]*Quote, error) {
// 	// The database returns byte slices that we need to de-serialize
// 	// into Quote structures.
// 	structList := []*Quote{}

// 	// We use a View as we don't update anything.
// 	err := d.db.View(func(tx *bolt.Tx) error {

// 		// TODO:
// 		// Get the bucket from the transaction tx.
// 		//
// 		// Iterate over all elements of the bucket.
// 		// Hint: BoltDB has a ForEach method for this.
// 		//   * For each element, create a new *Quote and deserialize
// 		//     the element value into the *Quote.
// 		//   * Then append the *Quote to structList.
// 		//
// 		// Check and return any errors.
// 	})

// 	// TODO: Check the error returned by d.db.View().
// 	// Return (structList, nil) or (nil, err), respectively.
// }
