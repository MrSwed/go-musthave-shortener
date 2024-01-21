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
	ID          string `json:"id"`
	ShortUrl    string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
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
			ID:          fmt.Sprintf("%d", ind),
			ShortUrl:    fmt.Sprintf("%s", short),
			OriginalUrl: original,
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
		data[config.ShortKey([]byte(item.ShortUrl))] = item.OriginalUrl
	}

	return data, nil
}

type Saver struct {
	file    *os.File
	encoder *json.Encoder
}

func NewSaver(filename string) (*Saver, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &Saver{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Saver) WriteData(data *FileStorageItem) error {
	return p.encoder.Encode(data)
}

func (p *Saver) Close() error {
	return p.file.Close()
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

func (c *Reader) ReadData() (e *FileStorageItem, err error) {
	err = c.decoder.Decode(&e)
	return
}

func (c *Reader) Close() error {
	return c.file.Close()
}
