package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"io"
	"os"
)

type FileStorage interface {
	Save(data map[config.ShortKey]string) error
	Restore() (map[config.ShortKey]string, error)
}

type FileStorageItem struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorageRepository struct {
	Items []FileStorageItem
	f     string
}

func NewFileStorage(f string) *FileStorageRepository {
	return &FileStorageRepository{
		f: f,
	}
}

func (f *FileStorageRepository) Save(data map[config.ShortKey]string) error {
	s, err := NewSaver(f.f)
	if err != nil {
		return err
	}
	defer s.Close()
	var ind int
	for short, original := range data {
		ind++
		item := FileStorageItem{
			UUID:        fmt.Sprintf("%d", ind),
			ShortURL:    fmt.Sprintf("%s", short),
			OriginalURL: original,
		}
		if err = s.WriteData(&item); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileStorageRepository) Restore() (map[config.ShortKey]string, error) {
	data := make(map[config.ShortKey]string)
	r, err := NewReader(f.f)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for err == nil {
		var item *FileStorageItem
		if item, err = r.ReadData(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		data[config.ShortKey([]byte(item.ShortURL))] = item.OriginalURL
	}

	return data, nil
}

type Saver struct {
	file    *os.File
	encoder *json.Encoder
}

func NewSaver(filename string) (*Saver, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return &Saver{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (s *Saver) WriteData(data *FileStorageItem) error {
	return s.encoder.Encode(data)
}

func (s *Saver) Close() error {
	return s.file.Close()
}

type Reader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewReader(filename string) (*Reader, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	return &Reader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (r *Reader) ReadData() (e *FileStorageItem, err error) {
	err = r.decoder.Decode(&e)
	return
}

func (r *Reader) Close() error {
	return r.file.Close()
}
