package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

type stream struct {
	url          string
	hash         string
	dir          string
	cmd          *exec.Cmd
	lastAccessed time.Time
}

type Stream interface {
	URL() string
	Hash() string
	Start() error
	Error() error
	Alive() bool
	Stale() bool
	Touch()
	Stop() error
	Dir() string
}

func (s *stream) Start() error {
	err := os.Mkdir(s.dir, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	log.Println(fmt.Sprintf("%s starting stream", s.url))
	cmd := exec.Command("ffmpeg",
		"-i", s.url,
		"-hide_banner",
		"-loglevel", "warning",
		"-abort_on", "empty_output",
		"-err_detect", "explode",
		"-c:v", "copy",
		"-c:a", "copy",
		"-f", "hls",
		"-hls_start_number_source", "datetime",
		"-hls_flags", "omit_endlist+delete_segments+temp_file",
		"-hls_list_size", "10",
		"-hls_time", "6",
		"-hls_base_url", fmt.Sprintf("/_s/%s/", s.hash),
		"index.m3u8")
	cmd.Dir = s.dir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmdErr := cmd.Start()
	if cmdErr != nil {
		return cmdErr
	}
	s.cmd = cmd
	s.Touch()
	// Waitcurl v  for the stream to start
	startTime := time.Now()
	for !s.Alive() && !startTime.Before(time.Now().Add(time.Duration(-10)*time.Second)) {
		os.Stderr.Sync()
		os.Stdout.Sync()
		log.Println(fmt.Sprintf("%s initializing", s.url))
		time.Sleep(1 * time.Second)
	}
	// Check for an error one last time
	err = s.Error()
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("%s started", s.url))
	return nil
}

func (s *stream) Touch() {
	s.lastAccessed = time.Now()
}

func (s *stream) Dir() string {
	return s.cmd.Dir
}

func (s *stream) URL() string {
	return s.url
}

func (s *stream) Hash() string {
	return s.hash
}

func (s *stream) Stale() bool {
	return s.lastAccessed.Before(time.Now().Add(time.Duration(-30) * time.Second))
}

func (s *stream) Stop() error {
	if s.cmd.Process.Pid != -1 {
		log.Println(fmt.Sprintf("killing %d", s.cmd.Process.Pid))

		s.cmd.Process.Kill()
		err := s.cmd.Wait()
		os.RemoveAll(s.Dir())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *stream) Alive() bool {
	err := s.Error()
	return err == nil
}

func (s *stream) Error() error {
	if s.cmd.Process.Pid == -1 {
		return errors.New("process already released")
	}
	process, err := os.FindProcess(s.cmd.Process.Pid)
	if err != nil {
		return err
	}

	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return err
	}

	playlistPath := filepath.Join(s.dir, "index.m3u8")
	_, err = os.Stat(playlistPath)
	if err != nil {
		return err
	}

	return nil
}

type FfmpegHandler struct {
	streams *sync.Map
}

func NewFFmpegHandler() *FfmpegHandler {
	return &FfmpegHandler{
		streams: &sync.Map{},
	}
}

func NewSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:])
}

func (f *FfmpegHandler) ExpireStaleStreams() {
	for {
		numStreams := 0
		f.streams.Range(func(_ interface{}, _ interface{}) bool { numStreams++; return true })
		if numStreams > 0 {
			log.Println(fmt.Sprintf("tracking %d streams", numStreams))
		}
		f.streams.Range(func(key interface{}, streamInterface interface{}) bool {
			stream := streamInterface.(Stream)
			url := stream.URL()
			if stream.Stale() {
				log.Println(fmt.Sprintf("%s is stale", url))
				stopErr := stream.Stop()
				if stopErr != nil {
					log.Println(fmt.Sprintf("%s: %v", url, stopErr))
				}
				f.streams.Delete(key)
			}
			return true
		})
		time.Sleep(10000 * time.Millisecond)
	}
}

func (f *FfmpegHandler) SegmentHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	streamHash := vars["stream"]
	segment := vars["segment"]

	streamInterface, ok := f.streams.Load(streamHash)
	if !ok {
		log.Println(fmt.Sprintf("unknown segment request /_s/%s/%s", streamHash, segment))
		err404(w, r)
		return
	}
	stream := streamInterface.(Stream)

	err := stream.Error()
	if err != nil {
		log.Println(fmt.Sprintf("segment request for stream in error state /_s/%s/%s", streamHash, segment))
		stream.Stop()
		f.streams.Load(streamHash)
		err500(w, r, err)
		return
	}

	segmentPath := filepath.Join(stream.Dir(), segment)
	segmentStat, err := os.Stat(segmentPath)
	if err != nil {
		log.Println(fmt.Sprintf("segment stat failed /_s/%s/%s", streamHash, segment))
		err404(w, r)
		return
	}

	segmentFile, err := os.Open(segmentPath)
	defer segmentFile.Close()
	if err != nil {
		log.Println(fmt.Sprintf("segment missing /_s/%s/%s", streamHash, segment))
		err404(w, r)
		return
	}

	stream.Touch()
	http.ServeContent(w, r, segment, segmentStat.ModTime(), segmentFile)
}

func (f *FfmpegHandler) PlaylistHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	url := vars["url"]

	if os.Getenv("ALLOWED_UPSTREAMS") == "" || !matchWildcard(os.Getenv("ALLOWED_UPSTREAMS"), url) {
		log.Println(fmt.Sprintf("%s unauthorized. add it to ALLOWED_UPSTREAMS", url))
		err403(w, r)
		return
	}

	streamHash := NewSHA256([]byte(url))
	streamInterface, loaded := f.streams.LoadOrStore(streamHash, &stream{
		url:  url,
		hash: streamHash,
		// TODO allow setting this to a memdir
		dir: filepath.Join(os.TempDir(), string(streamHash)),
	})
	stream := streamInterface.(*stream)
	if loaded {
		err := stream.Error()
		if err != nil {
			log.Println(fmt.Sprintf("%s in error state, stopping", stream.URL()))
			stopErr := stream.Stop()
			if stopErr != nil {
				log.Println(fmt.Sprintf("%s: %v", url, stopErr))
			}
			f.streams.Delete(stream.Hash())
			err500(w, r, err)
			return
		}
	} else {
		err := stream.Start()
		if err != nil {
			log.Println(fmt.Sprintf("%s error starting stream: %v", url, err))
			stopErr := stream.Stop()
			if stopErr != nil {
				log.Println(fmt.Sprintf("%s: %v", url, stopErr))
			}
			err500(w, r, err)
			return
		}
	}

	filePath := filepath.Join(stream.Dir(), "index.m3u8")
	fileStat, err := os.Stat(filePath)
	if err != nil {
		log.Println(fmt.Sprintf("%s couldn't stat playlist", stream.URL()))
		err404(w, r)
		return
	}

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		log.Println(fmt.Sprintf("%s couldn't open playlist", stream.URL()))
		err404(w, r)
		return
	}

	http.ServeContent(w, r, "index.m3u8", fileStat.ModTime(), file)
}
