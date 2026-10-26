package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
)

const (
	// Files are stored in KV entries of this size, so no request or RPC call
	// ever carries a whole file.
	chunkSize = 4 << 20

	metaPrefix = "m_"
	kvRetries  = 5

	// Every KV entry also expires on its own a day after its message, in case
	// the sweeper misses it.
	expiryMargin = 24 * time.Hour
)

var errNotFound = errors.New("message not found")

func metaKey(id string) string {
	return metaPrefix + id
}

func chunkKey(id string, file, chunk int) string {
	return fmt.Sprintf("c_%s_%d_%d", id, file, chunk)
}

func ttlSeconds(expiresAt int64) int64 {
	return max(1, (expiresAt-model.GetMillis())/1000+int64(expiryMargin.Seconds()))
}

func (p *Plugin) loadMessage(id string) (*message, []byte, error) {
	raw, appErr := p.API.KVGet(metaKey(id))
	if appErr != nil {
		return nil, nil, appErr
	}
	if raw == nil {
		return nil, nil, errNotFound
	}
	var m message
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, nil, fmt.Errorf("decode message %s: %w", id, err)
	}
	return &m, raw, nil
}

func (p *Plugin) saveMessage(m *message, old []byte) (bool, error) {
	raw, err := json.Marshal(m)
	if err != nil {
		return false, err
	}
	ok, appErr := p.API.KVSetWithOptions(metaKey(m.ID), raw, model.PluginKVSetOptions{
		Atomic:          true,
		OldValue:        old,
		ExpireInSeconds: ttlSeconds(m.ExpiresAt),
	})
	if appErr != nil {
		return false, appErr
	}
	return ok, nil
}

func (p *Plugin) createMessage(m *message) error {
	ok, err := p.saveMessage(m, nil)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("message %s already exists", m.ID)
	}
	return nil
}

// updateMessage is a compare-and-set loop: several viewers can open or close
// the same message at once.
func (p *Plugin) updateMessage(id string, change func(*message) error) (*message, error) {
	for range kvRetries {
		m, old, err := p.loadMessage(id)
		if err != nil {
			return nil, err
		}
		if err := change(m); err != nil {
			return nil, err
		}
		ok, err := p.saveMessage(m, old)
		if err != nil {
			return nil, err
		}
		if ok {
			return m, nil
		}
	}
	return nil, errors.New("message kept changing while saving it")
}

// storeChunks reads a file into chunk entries and returns its size and chunk
// count. More than limit bytes is an error, and what was written is removed.
func (p *Plugin) storeChunks(m *message, file int, r io.Reader, limit int64) (int64, int, error) {
	buf := make([]byte, chunkSize)
	var size int64
	chunks := 0
	for {
		n, err := io.ReadFull(r, buf)
		if n > 0 {
			size += int64(n)
			if size > limit {
				p.deleteChunks(m.ID, file, chunks)
				return 0, 0, errTooLarge
			}
			if appErr := p.API.KVSetWithExpiry(chunkKey(m.ID, file, chunks), buf[:n], ttlSeconds(m.ExpiresAt)); appErr != nil {
				p.deleteChunks(m.ID, file, chunks)
				return 0, 0, appErr
			}
			chunks++
		}
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return size, chunks, nil
		}
		if err != nil {
			p.deleteChunks(m.ID, file, chunks)
			return 0, 0, err
		}
	}
}

var errTooLarge = errors.New("file too large")

func (p *Plugin) writeChunks(w io.Writer, id string, file int, chunks int) error {
	for i := range chunks {
		data, appErr := p.API.KVGet(chunkKey(id, file, i))
		if appErr != nil {
			return appErr
		}
		if data == nil {
			return fmt.Errorf("chunk %d of file %d of %s is missing", i, file, id)
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func (p *Plugin) deleteChunks(id string, file, chunks int) {
	for i := range chunks {
		if appErr := p.API.KVDelete(chunkKey(id, file, i)); appErr != nil {
			p.API.LogWarn("Failed to delete a chunk", "id", id, "err", appErr.Error())
		}
	}
}

// purge deletes a message for good: file chunks first and the message last,
// so a failure leaves the message around for the sweeper to retry. The post
// goes too, unless it is already being deleted.
func (p *Plugin) purge(id string, deletePost bool) {
	m, _, err := p.loadMessage(id)
	if errors.Is(err, errNotFound) {
		return
	}
	if err != nil {
		p.API.LogError("Failed to load a message to delete it", "id", id, "err", err.Error())
		return
	}
	// An upload still in progress deletes its own chunks when it cannot save them.
	for i, f := range m.Files {
		p.deleteChunks(id, i, f.Chunks)
	}
	if appErr := p.API.KVDelete(metaKey(id)); appErr != nil {
		p.API.LogError("Failed to delete a message", "id", id, "err", appErr.Error())
		return
	}
	if deletePost && m.PostID != "" {
		// Deleting the post runs MessageHasBeenDeleted, which finds nothing left.
		if appErr := p.API.DeletePost(m.PostID); appErr != nil {
			p.API.LogWarn("Failed to delete the post of a message", "id", id, "post_id", m.PostID, "err", appErr.Error())
		}
	}
	p.notifyChanged(m, "")
}

// messageIDs lists every stored message by paging through the plugin's keys.
func (p *Plugin) messageIDs() ([]string, error) {
	const perPage = 1000
	var ids []string
	for page := 0; ; page++ {
		keys, appErr := p.API.KVList(page, perPage)
		if appErr != nil {
			return nil, appErr
		}
		for _, key := range keys {
			if id, ok := strings.CutPrefix(key, metaPrefix); ok {
				ids = append(ids, id)
			}
		}
		if len(keys) < perPage {
			return ids, nil
		}
	}
}
