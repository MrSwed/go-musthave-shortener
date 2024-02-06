package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
)

type FileStorage interface {
	Save(data Store) error
	Restore() (Store, error)
}

type FileStorageItem struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorageRepository struct {
	Items    []FileStorageItem
	fileName string
	m        sync.RWMutex
}

func NewFileStorage(f string) *FileStorageRepository {
	return &FileStorageRepository{
		fileName: f,
	}
}

func (f *FileStorageRepository) Save(data Store) error {
	if f.fileName == "" {
		return fmt.Errorf("no storage file provided")
	}
	f.m.Lock()
	defer f.m.Unlock()

	s, err := NewSaver(f.fileName)
	if err != nil {
		return err
	}
	for short, item := range data {
		fItem := FileStorageItem{
			UUID:        item.uuid,
			ShortURL:    short.String(),
			OriginalURL: item.url,
		}
		if err = s.WriteData(&fItem); err != nil {
			return err
		}
	}

	return s.Close()
}

func (f *FileStorageRepository) Restore() (data Store, err error) {
	if f.fileName == "" {
		err = fmt.Errorf("no storage file provided")
		return
	}
	f.m.Lock()
	defer f.m.Unlock()

	data = make(Store)
	var r *Reader
	if r, err = NewReader(f.fileName); err != nil {
		return
	}

	for err == nil {
		var item *FileStorageItem
		if item, err = r.ReadData(); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return
		}
		data[config.ShortKey([]byte(item.ShortURL))] = storeItem{
			uuid: item.UUID,
			url:  item.OriginalURL,
		}
	}
	err = r.Close()
	return
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
