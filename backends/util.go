package backends

import (
	"path"
	"path/filepath"
)

// ReadDirRecursively recursively walks a given storage from a given path with
// no restrictions in depth.
func ReadDirRecursively(storage Storage, relpath string, cb filepath.WalkFunc) error {
	return ReadDirRecursivelyLimited(storage, relpath, 0, cb)
}

// ReadDirRecursivelyLimited recursively walks a given storage from a given path.
// If depth is 0 or less, recursion will be unlimited.
func ReadDirRecursivelyLimited(
	storage Storage,
	relpath string,
	depth int,
	cb filepath.WalkFunc,
) error {
	pwds := [][]string{} // item: ["depth1","depth2","depth3"]
	pwd := []string{relpath}
	for {
		spwd := path.Join(pwd...)
		files, err := storage.ReadDir(spwd)
		if err != nil {
			return cb(spwd, nil, err)
		}

		for _, f := range files {
			err := cb(spwd, f, err)
			if err != nil {
				return err
			}
			if f.IsDir() {
				if depth != 0 && depth == len(pwd) {
					// Can't go deeper since we reached the limit
					continue
				}
				pwds = append(pwds, append(pwd, f.Name()))
			}
		}

		// If no paths are left, return
		if len(pwds) == 0 {
			break
		}

		// Get next directory path from stack
		pwd = pwds[0]
		pwds = pwds[1:]
	}

	return nil
}
