package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	_ "golang.org/x/tools/go/gcimporter"
)

func _TestGenerate(path string) (err error) {
	app := newApp()
	path = filepath.Join("http", path)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		return
	}

	stdin, stdout, stderr := os.Stdin, os.Stdout, os.Stderr

	r, w, err := os.Pipe()
	if err != nil {
		return
	}

	os.Stdin, os.Stdout, os.Stderr = f, w, w
	os.Args = []string{"jg", "--package", "main"}
	app.Run(os.Args)

	ch := make(chan string)
	go func() {
		buf := new(bytes.Buffer)
		io.Copy(buf, r)
		ch <- buf.String()
	}()
	w.Close()
	os.Stdin, os.Stdout, os.Stderr = stdin, stdout, stderr

	user := <-ch
	main := `
package main
import(
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	f, _ := os.Open("%s")
	defer f.Close()

	dec := json.NewDecoder(f)
	user := User{}
	dec.Decode(&user)
	fmt.Printf("user %+v\n", user)
}
	`
	fileNameUser := filepath.Join(os.TempDir(), "user.go")
	fileNameMain := filepath.Join(os.TempDir(), "main.go")

	ioutil.WriteFile(fileNameUser, []byte(user), os.ModePerm)
	ioutil.WriteFile(fileNameMain, []byte(fmt.Sprintf(main, path)), os.ModePerm)
	defer func() {
		os.Remove(fileNameUser)
		os.Remove(fileNameMain)
	}()

	out, err := exec.Command("go", "run", fileNameUser, fileNameMain).Output()
	if err != nil {
		fmt.Printf("out %+v\n", out)
		return
	}
	return
}

func TestGithubUsers(t *testing.T) {
	err := _TestGenerate("api.github.com/users.json")
	assert.Nil(t, err)
}

func TestGithubUser(t *testing.T) {
	err := _TestGenerate("api.github.com/user/seiji.json")
	assert.Nil(t, err)
}

func TestGithubUserStarred(t *testing.T) {
	err := _TestGenerate("api.github.com/user/seiji/starred.json")
	assert.Nil(t, err)
}

