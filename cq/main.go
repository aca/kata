package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/xtdlib/sqlitex"
)

var SHELL = "zsh"

func must(err error) {
	if err != nil {
		panic(err)
	}
}

var db *sqlitex.DB

type Command struct {
	ID          int
	Command     string
	Stdout      string
	Stderr      string
	Code        int
	Pid         int
	Workdir     string
	Environment []string
	CreatedAt   string
}

func main() {
	cachedir, err := os.UserCacheDir()
	if err != nil {
		log.Printf("Error getting cache dir: %v", err)
		cachedir = "."
	}

	dbdir := filepath.Join(cachedir, "cq")
	err = os.MkdirAll(dbdir, 0777)
	if err != nil {
		log.Fatal(err)
	}
	dbpath := filepath.Join(dbdir, "cq.db")
	log.Printf("Using dbpath: %s", dbpath)

	db = sqlitex.New(dbpath)
	db.MustExec(`
CREATE TABLE IF NOT EXISTS cq (
	id INTEGER PRIMARY KEY,
	command TEXT NOT NULL,
	stdout TEXT,
	stderr TEXT,
	code INTEGER,
	pid INTEGER,
	workdir TEXT,
	environment TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)
`)

	var id int
	var command string
	// envs := os.Environ()

	rows, err := db.NamedQuery(`INSERT INTO cq (command) values (:command) RETURNING id`, command)
	if rows.Next() {
		rows.Scan(&id)
	}

}

func RunCommand(id int) {

	db.Get()
	cmd := exec.Command(SHELL, "/dev/stdin")
	stdinpipe, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	cmd.Dir = workdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// cmd.Env = envs

	_, err = stdinpipe.Write([]byte(command))
	must(err)
	must(stdinpipe.Close())
	// must(cmd.Run())
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			_ = exitError
			// return exitError.ExitCode()
		}
	}
}
