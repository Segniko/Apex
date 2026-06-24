package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"

	apex "github.com/Segniko/Apex/proto"
	"google.golang.org/protobuf/proto"
	_ "modernc.org/sqlite"
)

// Vault handles encrypted local storage of crash reports.
type Vault struct {
	db  *sql.DB
	key []byte // 32 bytes for AES-256
}

// New creates or opens a new Vault.
func New(dbPath string, encryptionKey []byte) (*Vault, error) {
	if len(encryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open vault: %w", err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS crashes (
		id TEXT PRIMARY KEY,
		data BLOB,
		timestamp INTEGER
	);`

	if _, err := db.Exec(schema); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &Vault{db: db, key: encryptionKey}, nil
}

// Save encrypts and stores a CrashReport.
func (v *Vault) Save(report *apex.CrashReport) error {
	data, err := proto.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	// Encrypt the data
	encrypted, err := v.encrypt(data)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	_, err = v.db.Exec(
		"INSERT INTO crashes (id, data, timestamp) VALUES (?, ?, ?)",
		report.ErrorId, encrypted, report.Timestamp,
	)
	return err
}

// FetchAll retrieves and decrypts all stored crashes.
func (v *Vault) FetchAll() ([]*apex.CrashReport, error) {
	rows, err := v.db.Query("SELECT data FROM crashes")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reports []*apex.CrashReport
	for rows.Next() {
		var encrypted []byte
		if err := rows.Scan(&encrypted); err != nil {
			return nil, err
		}

		data, err := v.decrypt(encrypted)
		if err != nil {
			return nil, fmt.Errorf("decryption failed: %w", err)
		}

		var report apex.CrashReport
		if err := proto.Unmarshal(data, &report); err != nil {
			return nil, err
		}
		reports = append(reports, &report)
	}
	return reports, nil
}

// encrypt implements AES-GCM encryption.
func (v *Vault) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// decrypt implements AES-GCM decryption.
func (v *Vault) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(v.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// Cleanup removes reports older than a certain timestamp.
// This prevents the local database from growing infinitely.
func (v *Vault) Cleanup(olderThan int64) error {
	_, err := v.db.Exec("DELETE FROM crashes WHERE timestamp < ?", olderThan)
	return err
}

func (v *Vault) Close() error {
	return v.db.Close()
}
