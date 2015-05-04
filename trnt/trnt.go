package trnt

import (
	"bytes"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"os"
)

const privateKeyPath = "/Users/markmoniz/.ssh/id_rsa"

type TRNT struct {
	id, size_sent, size_total int
	filename, from, to        string
	created_at, updated_at    []byte
	io.Reader
}

func Torrents(query string) func() map[int]*TRNT {
	/*Open Database*/
	db, err := sql.Open("mysql", os.Getenv("FOLLICLE_DB"))
	if err != nil {
		panic(err.Error())
	}

	/*Query Database

	  | id(int) | from(string) | filename(string) | to(string) | size_sent(int) |
	      | size_total(int) | created_at([]byte) | updated_at([]byte) |

	  INFO: 'size_sent' and 'size_total' are set to -1 if they have not been transfered yet
	  TODO: Change 'created_at' and 'updated_at' type from []byte to time.TIME
	*/
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}

	return func() map[int]*TRNT {
		/*
		   Closure

		   INFO: Makes a map of torrents from the provided query
		   RETURN: map[int]*TNRT
		*/
		defer db.Close()
		defer rows.Close()
		m := make(map[int]*TRNT)
		var filename, from, to string
		var id, size_sent, size_total int
		var created_at, updated_at []byte
		for rows.Next() {
			if err := rows.Scan(&id, &from, &filename, &to, &size_sent, &size_total, &created_at, &updated_at); err != nil {
				log.Fatal(err)
			}
			tor := new(TRNT)
			tor.WriteTorrent(id, size_sent, size_total, filename, from, to, created_at, updated_at)
			m[id] = tor
		}
		return m
	}
}

func (t *TRNT) WriteTorrent(id, size_sent, size_total int, filename, from, to string, created_at, updated_at []byte) error {
	/*
	   Write Torrent

	   INFO: Overwrites values for a given torrent
	   RETURN: error
	*/
	t.id = id
	t.filename = filename
	t.from = from
	t.to = to
	t.size_sent = size_sent
	t.size_total = size_total
	t.created_at = created_at
	t.updated_at = updated_at
	return nil
}

func (t *TRNT) printTorrent() error {
	fmt.Println(t.id)
	fmt.Println(t.filename)
	fmt.Println(t.from)
	fmt.Println(t.to)
	fmt.Println(t.size_sent)
	fmt.Println(t.size_total)
	fmt.Println(t.created_at)
	fmt.Println(t.updated_at)
	return nil
}

//Overloads io Read
func (t *TRNT) Read(p []byte) (int, error) {
	n, err := t.Reader.Read(p)
	if err != nil {
		log.Print(err)
	}
	t.size_sent += n

	t.UpdateTransfer("UPDATE transfers SET size_sent=? WHERE id=?")

	return n, err
}

func (t *TRNT) UpdateTransfer(exec string) error {
	db, err := sql.Open("mysql", os.Getenv("FOLLICLE_DB"))
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec(exec, t.size_sent, t.id)
	if err != nil {
		panic(err.Error())
	}
	return nil
}

func (t *TRNT) TransferTorrent() error {
	//Return []byte of file contents of private key
	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		log.Fatal(err)
	}

	clientConfig := &ssh.ClientConfig{
		User: "home",
		Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
	}

	client, err := ssh.Dial("tcp", "192.168.1.170:22", clientConfig)
	if err != nil {
		log.Fatal(err)
	}

	session, err := client.NewSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	sftpclient, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftpclient.Close()

	localFile, err := ioutil.ReadFile(t.from)
	if err != nil {
		log.Fatal(err)
	}
	t.Reader = bytes.NewBuffer(localFile)

	remoteFile, err := sftpclient.Create(t.to + "/" + t.filename)
	if err != nil {
		log.Fatal(err)
	}
	defer remoteFile.Close()

	t.size_sent = 0
	_, err = io.Copy(remoteFile, t)
	return nil
}

func InsertDB() error {
	db, err := sql.Open("mysql", os.Getenv("FOLLICLE_DB"))
	if err != nil {
		panic(err.Error())
	}

	_, err = db.Exec("INSERT ")
	if err != nil {
		panic(err.Error())
	}

	return nil
}

func main() {

	//allTorrents := Torrents("SELECT * FROM transfers")
	//all := allTorrents()
	//unsentTorrents := Torrents("SELECT * FROM transfers WHERE size_sent = -1")
	//unsent := unsentTorrents()
	myTorrent := Torrents("SELECT * FROM transfers WHERE id=136")
	my := myTorrent()

	//print the filename of each row
	for _, v := range my {
		v.TransferTorrent()
		break
	}

}
