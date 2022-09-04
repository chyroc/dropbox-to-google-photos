package tokenmanager

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/oauth2"
)

type FileRepository struct {
	dir string
}

func NewFileRepository(dir string) (*FileRepository, error) {
	err := os.MkdirAll(dir+"/tokens", 0700)
	if err != nil {
		return nil, err
	}
	return &FileRepository{dir: dir + "/tokens"}, nil
}

// Set stores a token into the OS keyring.
func (r *FileRepository) Set(email string, token *oauth2.Token) error {
	tokenJSONBytes, err := json.Marshal(token)
	if err != nil {
		return ErrInvalidToken
	}

	return ioutil.WriteFile(r.dir+"/"+wrapEmail(email), tokenJSONBytes, 0644)
}

// getToken returns the specified token from the repository.
func (r *FileRepository) Get(email string) (*oauth2.Token, error) {
	var nullToken = &oauth2.Token{}

	bs, err := ioutil.ReadFile(r.dir + "/" + wrapEmail(email))
	if err != nil {
		return nullToken, ErrTokenNotFound
	}

	var token oauth2.Token
	if err := json.Unmarshal(bs, &token); err != nil {
		return nullToken, ErrInvalidToken
	}

	return &token, nil
}

// Close closes the keyring repository.
func (r *FileRepository) Close() error {
	// in this particular implementation we don't need to do anything.
	return nil
}

func wrapEmail(email string) string {
	for _, v := range []string{".", "\"", "@", "-", "//"} {
		email = strings.ReplaceAll(email, v, "_")
	}
	return email
}
