package Octo

import (
	"Octo/Octo/ToOcto"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// closes a SourceLimiter
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

// Represents a file in OctoDrive
type OctoFile struct {
	file             fileDetails
	user             *ToOcto.OctoUser
	src_data         io.Reader
	cached_src_chunk *CachedReader
	repo_limiter     SourceLimiter
	path_index       int
	chunk_index      int
	enc_dec          EncryptDecrypter
}

// sets encryption/decryption for the file
func (of *OctoFile) SetEncDec(enc_dec EncryptDecrypter) {
	of.enc_dec = enc_dec
}

// sets optional user data for the file
func (of *OctoFile) SetUserData(data []byte) {
	of.file.UserData = data
}

// returns optional user data for the file
func (of *OctoFile) GetUserData() []byte {
	return of.file.UserData
}

// returns size of the file
func (of *OctoFile) GetSize() uint64 {
	return of.file.Size
}

// represents ReadSeekCloser for OctoFile
type rd_sk_closer struct {
	current_pos uint64
	file        *OctoFile
	read_closer io.ReadCloser
}

func (rsc *rd_sk_closer) Read(p []byte) (n int, err error) {
	n, err = rsc.read_closer.Read(p)
	rsc.current_pos += uint64(n)
	return
}

// seeks to the specified position
// whence can be io.SeekStart, io.SeekCurrent, io.SeekEnd
// returns the new position and error
func (rsc *rd_sk_closer) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		rsc.current_pos = uint64(offset)
	case io.SeekCurrent:
		rsc.current_pos += uint64(offset)
	case io.SeekEnd:
		rsc.current_pos = rsc.file.file.Size - uint64(offset)
	}
	read_closer, err := rsc.file.GetBytesReader(rsc.current_pos, rsc.file.file.Size)
	rsc.read_closer = read_closer
	return int64(rsc.current_pos), err
}

func (rsc *rd_sk_closer) Close() error {
	return rsc.read_closer.Close()
}

// returns a io.ReadSeekCloser for the file
func (of *OctoFile) GetSeekReader() (io.ReadSeekCloser, error) {
	return &rd_sk_closer{file: of, current_pos: 0}, nil
}

// returns a io.ReadCloser for the file
func (of *OctoFile) GetReader() (io.ReadCloser, error) {
	Rdrs := make([]io.ReadCloser, 0)
	for _, repo := range of.file.Paths {
		c, err := getPartCount(of.user, repo, of.file.Name)
		if err != nil {
			return nil, err
		}
		Rdrs = append(Rdrs, NewMultipartReader(of.user, repo, of.file.Name, int(c), of.enc_dec))
		if err != nil {
			return nil, err
		}
	}
	return &octoFileReader{readers: Rdrs, read_end: true}, nil
}

// returns a io.ReadCloser for the file from the specified position to the specified end position
func (of *OctoFile) GetBytesReader(from uint64, to uint64) (io.ReadCloser, error) {
	StartPathIndex := from / of.file.MaxRepoSize
	EndPathIndex := to / of.file.MaxRepoSize
	StartPartNo := from % of.file.MaxRepoSize / of.file.ChunkSize
	StartPartOffset := from % of.file.MaxRepoSize % of.file.ChunkSize
	EndPartNo := to % of.file.MaxRepoSize / of.file.ChunkSize
	EndPartOffset := to % of.file.MaxRepoSize % of.file.ChunkSize

	enc_dec := of.enc_dec

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

// Writes a chunk of data to the file
func (of *OctoFile) WriteChunk() error {
	if of.src_data != nil {
		println("VALID SOURCE")
		if of.repo_limiter == nil {
			Repository := RandomString(10)
			Err := of.user.CreateRepository(Repository, "OCTODRIVE_CONTENTS")
			if Err != nil {
				return Err
			}
			of.repo_limiter = NewSourceLimiter(of.src_data, of.file.MaxRepoSize)
			of.file.Paths = append(of.file.Paths, Repository)
			of.path_index++
			of.chunk_index = 0
		}
		chunked_src := NewSourceLimiter(of.repo_limiter, of.file.ChunkSize)
		if of.cached_src_chunk != nil {
			of.cached_src_chunk.Dispose()
		}
		cached_chunk, err := NewCachedReader(chunked_src)
		of.cached_src_chunk = cached_chunk
		if err != nil {
			return err
		}
		source, err := of.enc_dec.Encrypt(cached_chunk)
		if err != nil {
			return err
		}
		Err := of.user.Transfer(of.file.Paths[of.path_index], of.file.Name+"/"+strconv.Itoa(of.chunk_index), source)
		if Err != nil {
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
			return Err
		}
		of.cached_src_chunk.Dispose()
		of.file.Size += chunked_src.GetCurrentSize()
		if of.repo_limiter.IsEOF() {
			of.repo_limiter = nil
			of.src_data = nil
		} else if chunked_src.IsEOF() {
			of.repo_limiter = nil
		}
		of.chunk_index++
	} else {
		return io.EOF
	}
	return nil
}

// Retries writing the last chunk of data to the file (if any error occurred during the last write)
func (of *OctoFile) RetryWriteChunk() error {
	var Err *ToOcto.Error
	if of.cached_src_chunk != nil {
		of.cached_src_chunk.Reset()
		source, err := of.enc_dec.Encrypt(of.cached_src_chunk)
		if err != nil {
			return err
		}
		Err = of.user.Transfer(of.file.Paths[of.path_index], of.file.Name+"/"+strconv.Itoa(of.chunk_index), source)
		if Err != nil {
			return Err
		}
		of.chunk_index++
	}
	if Err != nil {
		return Err
	}
	return nil
}

// Writes all the data to the file (if any error occurred during the last write, need to call RetryWriteChunk )
func (of *OctoFile) WriteAll() error {
	for {
		err := of.WriteChunk()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
