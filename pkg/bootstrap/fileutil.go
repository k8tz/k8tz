package bootstrap

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

func copyDirectory(src, dst string, overwrite bool) error {
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dst, entry.Name())

		fileInfo, err := os.Lstat(sourcePath)
		if err != nil {
			return err
		}

		switch fileInfo.Mode() & os.ModeType {
		case os.ModeDir:
			if err := createIfNotExists(destPath, 0755); err != nil {
				return err
			}
			if err := copyDirectory(sourcePath, destPath, overwrite); err != nil {
				return err
			}
		case os.ModeSymlink:
			if err := copySymLinkIfNotExists(sourcePath, destPath, overwrite); err != nil {
				return err
			}
		default:
			exists, err := exists(destPath)
			if err != nil {
				return fmt.Errorf("failed to check existence of file: %s, error: %w", destPath, err)
			}

			if !exists || overwrite {
				if err := copyFile(sourcePath, destPath); err != nil {
					return fmt.Errorf("failed to copy file from '%s' to '%s', error: %w", sourcePath, destPath, err)
				}
			} else {
				fmt.Fprintf(os.Stderr, "skipping file '%s' because it already exists\n", destPath)
			}
		}

		isLink := entry.Mode()&os.ModeSymlink != 0
		if !isLink {
			if err := os.Chmod(destPath, entry.Mode()); err != nil {
				return err
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	fmt.Fprintf(os.Stderr, "Copying '%s' to '%s'\n", src, dst)
	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(src)
	if err != nil {
		return err
	}

	defer in.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func exists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, nil
	} else if err == nil {
		return true, nil
	}

	return false, err
}

func createIfNotExists(dir string, perm os.FileMode) error {
	exists, err := exists(dir)
	if err != nil {
		return fmt.Errorf("failed to check for directory existence: %s error: %w", dir, err)
	}

	if exists {
		fmt.Fprintf(os.Stderr, "not creating '%s' since its already exists\n", dir)
		return nil
	}

	if err := os.MkdirAll(dir, perm); err != nil {
		return fmt.Errorf("failed to create directory: '%s', error: '%w'", dir, err)
	}

	fmt.Fprintf(os.Stderr, "directory created: %s\n", dir)
	return nil
}

func copySymLinkIfNotExists(source, dest string, force bool) error {
	exists, err := exists(dest)
	if err != nil {
		return fmt.Errorf("failed to check for symlink existence: %s error: %w", dest, err)
	}

	if exists {
		if force {
			err = os.Remove(dest)
			if err != nil {
				return fmt.Errorf("failed to remove symlink for overwrite: %s error: %w", dest, err)
			}
		} else {
			fmt.Fprintf(os.Stderr, "skipping symlink '%s' because it already exists\n", dest)
			return nil
		}
	}

	link, err := os.Readlink(source)
	if err != nil {
		return fmt.Errorf("failed to read link: %s, error: %w", source, err)
	}

	err = os.Symlink(link, dest)
	if err != nil {
		return fmt.Errorf("failed to create symlink from '%s' to '%s', error: %w", dest, link, err)
	}

	fmt.Fprintf(os.Stderr, "symlink created: '%s' => '%s'\n", dest, link)
	return nil
}
