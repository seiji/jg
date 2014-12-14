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

func _TestGenerate(t *testing.T, path string) {
	var err error
	app := newApp()
	path = filepath.Join("http", path)

	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}

	stdin, stdout, stderr := os.Stdin, os.Stdout, os.Stderr

	r, w, err := os.Pipe()
	if err != nil {
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
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

	assert.Nil(t, err)
	if err != nil {
		fmt.Printf("out %+v\n", string(out))
		fmt.Printf("err %+v\n", err)
		os.Exit(1)
	}
}

func TestGithubUsers(t *testing.T) {
	_TestGenerate(t, "api.github.com/users.json")
}

func TestGithubUser(t *testing.T) {
	_TestGenerate(t, "api.github.com/user/seiji.json")
}

func TestGithubUserStarred(t *testing.T) {
	_TestGenerate(t, "api.github.com/user/seiji/starred.json")
}
