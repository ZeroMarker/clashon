// Package profile manages downloaded subscriptions on disk.
package profile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/ZeroMarker/clashon/sdk/subscription"
)

// Profile is one subscription.
type Profile struct {
	Name     string                 `json:"name"`
	URL      string                 `json:"url"`
	Updated  time.Time              `json:"updated"`
	Interval time.Duration          `json:"interval,omitempty"`
	UserInfo *subscription.UserInfo `json:"userinfo,omitempty"`
}

type index struct {
	Current  string     `json:"current"`
	Profiles []*Profile `json:"profiles"`
}

// Store persists profiles under dir with an index file.
type Store struct {
	Dir    string
	Index  string
	Client *http.Client
}

var ErrNotFound = errors.New("profile not found")

var validName = regexp.MustCompile(`^[\w.-]+$`)

// File returns the YAML path of a profile.
func (s *Store) File(name string) string { return filepath.Join(s.Dir, name+".yaml") }

// List returns all profiles sorted by name and the current profile name.
func (s *Store) List() ([]*Profile, string, error) {
	idx, err := s.load()
	if err != nil {
		return nil, "", err
	}
	sort.Slice(idx.Profiles, func(i, j int) bool { return idx.Profiles[i].Name < idx.Profiles[j].Name })
	return idx.Profiles, idx.Current, nil
}

// Get returns the named profile.
func (s *Store) Get(name string) (*Profile, error) {
	idx, err := s.load()
	if err != nil {
		return nil, err
	}
	if p := idx.find(name); p != nil {
		return p, nil
	}
	return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
}

// Current returns the selected profile.
func (s *Store) Current() (*Profile, error) {
	idx, err := s.load()
	if err != nil {
		return nil, err
	}
	if p := idx.find(idx.Current); p != nil {
		return p, nil
	}
	return nil, fmt.Errorf("%w: no current profile, run `clashon sub add`", ErrNotFound)
}

// Add downloads url as a new profile. The first profile becomes current.
func (s *Store) Add(ctx context.Context, name, url string) (*Profile, error) {
	if !validName.MatchString(name) {
		return nil, fmt.Errorf("invalid profile name %q", name)
	}
	idx, err := s.load()
	if err != nil {
		return nil, err
	}
	if idx.find(name) != nil {
		return nil, fmt.Errorf("profile %q already exists", name)
	}
	p := &Profile{Name: name, URL: url}
	if err := s.download(ctx, p); err != nil {
		return nil, err
	}
	idx.Profiles = append(idx.Profiles, p)
	if idx.Current == "" {
		idx.Current = name
	}
	return p, s.save(idx)
}

// Update re-downloads the named profile.
func (s *Store) Update(ctx context.Context, name string) (*Profile, error) {
	idx, err := s.load()
	if err != nil {
		return nil, err
	}
	p := idx.find(name)
	if p == nil {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	if err := s.download(ctx, p); err != nil {
		return nil, err
	}
	return p, s.save(idx)
}

// Use makes name the current profile.
func (s *Store) Use(name string) error {
	idx, err := s.load()
	if err != nil {
		return err
	}
	if idx.find(name) == nil {
		return fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	idx.Current = name
	return s.save(idx)
}

// Remove deletes a profile and its file.
func (s *Store) Remove(name string) error {
	idx, err := s.load()
	if err != nil {
		return err
	}
	for i, p := range idx.Profiles {
		if p.Name == name {
			idx.Profiles = append(idx.Profiles[:i], idx.Profiles[i+1:]...)
			if idx.Current == name {
				idx.Current = ""
			}
			if err := os.Remove(s.File(name)); err != nil && !os.IsNotExist(err) {
				return err
			}
			return s.save(idx)
		}
	}
	return fmt.Errorf("%w: %s", ErrNotFound, name)
}

func (s *Store) download(ctx context.Context, p *Profile) error {
	r, err := subscription.Fetch(ctx, s.Client, p.URL)
	if err != nil {
		return err
	}
	if err := writeFile(s.File(p.Name), r.Body, 0o600); err != nil {
		return err
	}
	p.Updated = time.Now()
	p.UserInfo = r.UserInfo
	p.Interval = r.Interval
	return nil
}

func (idx *index) find(name string) *Profile {
	for _, p := range idx.Profiles {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (s *Store) load() (*index, error) {
	idx := &index{}
	b, err := os.ReadFile(s.Index)
	if os.IsNotExist(err) {
		return idx, nil
	}
	if err != nil {
		return nil, err
	}
	return idx, json.Unmarshal(b, idx)
}

func (s *Store) save(idx *index) error {
	b, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return writeFile(s.Index, b, 0o600)
}

// writeFile writes atomically via a temp file in the same directory.
func writeFile(path string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
