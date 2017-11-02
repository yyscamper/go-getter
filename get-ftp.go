package getter

import (
	"bytes"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/dutchcoders/goftp"
)

type FtpDetector struct{}

func (d *FtpDetector) Detect(src, _ string) (string, bool, error) {
	return "", false, nil
}

type FtpGetter struct{}

func (g *FtpGetter) ClientMode(u *url.URL) (ClientMode, error) {
	client, err := g.createClient(u)
	if err != nil {
		return ClientModeInvalid, err
	}
	defer client.Close()

	stat, err := client.Stat(u.Path)
	if err != nil {
		return ClientModeInvalid, err
	}

	//We can simply determine the ClientMode by just checking the `stat`
	//if it is file, the length has to be 1
	//if it is a directory, the length has to >= 2, as at least the '.', '..' directory will be showed in `stat`
	if len(stat) == 1 {
		return ClientModeFile, nil
	} else if len(stat) > 1 {
		return ClientModeDir, nil
	} else {
		return ClientModeInvalid, nil
	}
}

func (g *FtpGetter) Get(dst string, u *url.URL) error {
	return g.GetFile(dst, u)
}

func (g *FtpGetter) GetFile(dst string, u *url.URL) error {
	client, err := g.createClient(u)
	if err != nil {
		return err
	}
	defer client.Close()

	log.Printf("FTP Downloading remote %s to local %s", u.Path, dst)

	buf := bytes.NewBuffer(nil)
	_, err = client.Retr(u.Path, func(r io.Reader) error {
		if _, err = io.Copy(buf, r); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	log.Printf("FTP Download total size: %d", buf.Len())

	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = buf.WriteTo(file)
	if err != nil {
		return err
	}

	return nil
}

func (g *FtpGetter) createClient(u *url.URL) (*goftp.FTP, error) {
	port := u.Port()
	if port == "" {
		port = "21"
	}
	client, err := goftp.Connect(u.Hostname() + ":" + port)
	if err != nil {
		return nil, err
	}

	var username, password string
	if u.User != nil {
		username = u.User.Username()
		if pw, ok := u.User.Password(); ok {
			password = pw
		}
	}
	if username == "" {
		username = u.Query().Get("user")
	}
	if password == "" {
		password = u.Query().Get("password")
	}

	if username == "" {
		username = "anonymous"
	}
	if password == "" {
		password = "anonymous"
	}

	err = client.Login(username, password)
	if err != nil {
		return nil, err
	}

	return client, nil
}
