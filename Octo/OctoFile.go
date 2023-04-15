package Octo

import (
	"Octo/Octo/ToOcto"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

type LimiterCloser struct {
	limiter io.Reader
	closer  io.Closer
}

func (lc *LimiterCloser) Read(p []byte) (n int, err error) {
	return lc.limiter.Read(p)
}

func (lc *LimiterCloser) Close() error {
	return lc.closer.Close()
}

type OctoFile struct {
	file             fileDetails
	user             ToOcto.OctoUser
	src_data         io.Reader
	cached_src_chunk *CachedReader
	repo_limiter     SourceLimiter
	path_index       int
	chunk_index      int
	encrypter        EncryptDecrypter
}

func (of *OctoFile) GetName() string {
	return of.file.Name
}

func (of *OctoFile) GetSize() uint64 {
	return of.file.Size
}

func (of *OctoFile) Get() (io.ReadCloser, error) {
	Rdrs := make([]io.ReadCloser, 0)
	enc_dec := newAesEncDec(of.file.Key[:32], of.file.Key[32:])
	for _, repo := range of.file.Paths {
		c, err := getPartCount(of.user, repo, of.file.Name)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, NewMultipartReader(of.user, repo, of.file.Name, int(c), enc_dec))
		if err != nil {
			return nil, err
		}
	}
	return &octoFileReader{readers: Rdrs, read_end: true}, nil
}

func (of *OctoFile) GetBytes(from uint64, to uint64) (io.ReadCloser, error) {
	StartPathIndex := from / of.file.MaxRepoSize
	EndPathIndex := to / of.file.MaxRepoSize
	StartPartNo := from % of.file.MaxRepoSize / of.file.ChunkSize
	StartPartOffset := from % of.file.MaxRepoSize % of.file.ChunkSize
	EndPartNo := to % of.file.MaxRepoSize / of.file.ChunkSize
	EndPartOffset := to % of.file.MaxRepoSize % of.file.ChunkSize

	fmt.Println(StartPathIndex, EndPathIndex, StartPartNo, StartPartOffset, EndPartNo, EndPartOffset)

	enc_dec := newAesEncDec(of.file.Key[:32], of.file.Key[32:])

	Rdrs := make([]io.ReadCloser, 0)
	if StartPathIndex == EndPathIndex && StartPartNo == EndPartNo {
		// Make a http Request to file with start and end range
		//of.user.MakeRequest(,)
		req, err := of.user.MakeRequest(http.MethodGet, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(StartPartNo), nil, true)
		if err != nil {
			return nil, err
		}
		delayed_reader := &delayedReader{req: req, decrypter: enc_dec, ignoreBytes: StartPartOffset}
		limiter_closer := &LimiterCloser{limiter: io.LimitReader(delayed_reader, int64(EndPartOffset-StartPartOffset)), closer: delayed_reader}
		Rdrs = append(Rdrs, limiter_closer)
	} else if StartPathIndex == EndPathIndex {
		// Make a http request to first file with start range
		req, err := of.user.MakeRequest(http.MethodGet, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(StartPartNo), nil, true)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, &delayedReader{req: req, decrypter: enc_dec, ignoreBytes: StartPartOffset})
		// create range reader for intermediate files
		Rdrs = append(Rdrs, NewMultipartRangeReader(of.user, of.file.Paths[StartPathIndex], of.file.Name, int(StartPartNo+1), int(EndPartNo), enc_dec))
		// Make a http request to last file with end range
		req, err = of.user.MakeRequest(http.MethodGet, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(EndPartNo), nil, true)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, &remoteReader{req: req, decrypter: enc_dec})
	} else {
		// Make http request to first file
		req, err := of.user.MakeRequest(http.MethodGet, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(StartPartNo), nil, true)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, &delayedReader{req: req, decrypter: enc_dec, ignoreBytes: StartPartOffset})
		// create range reader from after the first file to last
		partCount, err := getPartCount(of.user, of.file.Paths[StartPathIndex], of.file.Name)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, NewMultipartRangeReader(of.user, of.file.Paths[StartPathIndex], of.file.Name, int(StartPartNo+1), int(partCount), enc_dec))
		// create a loop from 2nd first path to 2nd last path and Range reader
		for i := StartPathIndex + 1; i < EndPathIndex; i++ {
			partCount, err := getPartCount(of.user, of.file.Paths[i], of.file.Name)
			if err != nil {
				return nil, err
			}
			Rdrs = append(Rdrs, NewMultipartReader(of.user, of.file.Paths[i], of.file.Name, int(partCount), enc_dec))
		}
		// create range reader from 0 to before the last file
		Rdrs = append(Rdrs, NewMultipartRangeReader(of.user, of.file.Paths[EndPathIndex], of.file.Name, 0, int(EndPartNo), enc_dec))
		// make http request to the last file
		req, err = of.user.MakeRequest(http.MethodGet, of.file.Paths[StartPathIndex], of.file.Name+"/"+fmt.Sprint(EndPartNo), nil, true)
		if err != nil {
			return nil, err
		}
		remote_reader := &remoteReader{req: req, decrypter: enc_dec}
		Rdrs = append(Rdrs, &LimiterCloser{limiter: io.LimitReader(remote_reader, int64(EndPartOffset)), closer: remote_reader})
	}
	return &octoFileReader{
		readers:  Rdrs,
		read_end: true,
	}, nil
}

func (of *OctoFile) WriteChunk() error {
	if of.src_data != nil {
		println("VALID SOURCE")
		if of.repo_limiter == nil {
			println("CREATING REPOSITORY")
			Repository := ToOcto.RandomString(10)
			status, err := of.user.CreateRepository(Repository, "OCTODRIVE_CONTENTS")
			if err != nil {
				return err
			}
			if status != http.StatusCreated {
				return errors.New("failed to create repository")
			}
			of.repo_limiter = NewSourceLimiter(of.src_data, of.file.MaxRepoSize)
			of.file.Paths = append(of.file.Paths, Repository)
			of.path_index++
			of.chunk_index = 0
		}
		println("SETTING AND TRANSFERING")
		chunked_src := NewSourceLimiter(of.repo_limiter, of.file.ChunkSize)
		if of.cached_src_chunk != nil {
			of.cached_src_chunk.Dispose()
		}
		cached_chunk, err := NewCachedReader(chunked_src)
		of.cached_src_chunk = cached_chunk
		if err != nil {
			return err
		}
		source, err := of.encrypter.Encrypt(cached_chunk)
		if err != nil {
			return err
		}
		status, _, err := of.user.Transfer(of.file.Paths[of.path_index], of.file.Name+"/"+strconv.Itoa(of.chunk_index), source)
		if err != nil {
			io.Copy(io.Discard, cached_chunk)
			of.file.Size += chunked_src.GetCurrentSize()
			if of.repo_limiter.IsEOF() {
				println("THE SOURCE DATA IS EMPTY")
				of.repo_limiter = nil
				of.src_data = nil
			} else if chunked_src.IsEOF() {
				println("THE REPOSITORY IS FULL")
				of.repo_limiter = nil
			}
			return err
		}
		if status != http.StatusCreated {
			io.Copy(io.Discard, cached_chunk)
			of.file.Size += chunked_src.GetCurrentSize()
			if of.repo_limiter.IsEOF() {
				println("THE SOURCE DATA IS EMPTY")
				of.repo_limiter = nil
				of.src_data = nil
			} else if chunked_src.IsEOF() {
				println("THE REPOSITORY IS FULL")
				of.repo_limiter = nil
			}
			return errors.New("failed to upload chunk:" + strconv.Itoa(status))
		}
		of.cached_src_chunk.Dispose()
		of.file.Size += chunked_src.GetCurrentSize()
		if of.repo_limiter.IsEOF() {
			println("THE SOURCE DATA IS EMPTY")
			of.repo_limiter = nil
			of.src_data = nil
		} else if chunked_src.IsEOF() {
			println("THE REPOSITORY IS FULL")
			of.repo_limiter = nil
		}
		of.chunk_index++
	} else {
		return io.EOF
	}
	return nil
}

func (of *OctoFile) RetryWriteChunk() error {
	if of.cached_src_chunk != nil {
		of.cached_src_chunk.Reset()
		source, err := of.encrypter.Encrypt(of.cached_src_chunk)
		if err != nil {
			return err
		}
		status, _, err := of.user.Transfer(of.file.Paths[of.path_index], of.file.Name+"/"+strconv.Itoa(of.chunk_index), source)
		if err != nil {
			return err
		}
		if status != http.StatusCreated {
			return errors.New("failed to upload chunk:" + strconv.Itoa(status))
		}
		of.chunk_index++
	}
	return nil
}

func (of *OctoFile) WriteAll() error {
	for {
		err := of.WriteChunk()
		println("Writing", err)
		if err != nil {
			return err
		}
	}
}
