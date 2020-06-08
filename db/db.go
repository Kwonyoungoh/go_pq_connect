package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/ssh"
)

// 변수 
const (
	sshHost string = "hostIP"
	sshUser string = "hostUser"
	sshPort int = 22
	sshPath string = "./key/path/if/u/got/a/key"
	dbUser string = "dbUser"
	dbPass string = "dbPassword"
	dbHost string = "dbIP:port"
	dbName string = "dbName"
)

// ViaSSHDialer 구조체 선언 구조체 필드(맴버) 정의
type ViaSSHDialer struct {
	// client 라는 맴버는 ssh.Clinet형 포인터 type을 가진다. * 참조연산자(역참조)
	// 쉬운 말로 ViaSSHDialer의 client는 ssh.Client를 가리킨다. ssh.Client를 역참조하겠따.
	client *ssh.Client
}

// Open 은 ViaSShDialer 구조체에 연결되어(*ViaSSHDialer) 리시버 변수 v로 받는 메소드 _(driver.Conn 타입) 과 err(error타입)을 리턴한다.
//객체( ex)구조체 )에 연결되어야만 메소드 라고 할 수있다.
func (v *ViaSSHDialer) Open(s string) (_ driver.Conn, err error) {
	return pq.DialOpen(v, s)
}

// Dial 함수는 ViaSSHDialer포인터(*ViaSSHDialer) v를 리시버로 한다. 
func (v *ViaSSHDialer) Dial(network, address string) (net.Conn, error) {
	return v.client.Dial(network, address)
}

// DialTimeout 함수는 ViaSSHDialer포인터(*ViaSSHDialer) v를 리시버로 한다.
func (v *ViaSSHDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return v.client.Dial(network, address)
}

// PublicKeyFile 함수는 file을 string으로 인자로 받고 ssh.AuthMethod를 리턴한다.
func PublicKeyFile(file string) ssh.AuthMethod {
	buffer, err := ioutil.ReadFile(file)

	if err != nil {
		fmt.Printf("%s\n",err.Error())
		return nil
	}

	key, err := ssh.ParsePrivateKey(buffer)
	if err != nil {
		fmt.Printf("%s\n",err.Error())
		return nil
	}
	return ssh.PublicKeys(key)
}

// CreateDB 는 리시버는 없고 인자도 업고 두 값만을 리턴한다.
func CreateDB() (*sql.DB, error) {

	// The client configuration with configuration option to use the ssh-agent
	sshConfig := &ssh.ClientConfig{
		User: sshUser,
		Auth: []ssh.AuthMethod{
		PublicKeyFile(sshPath)},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error { return nil }),
	}

	sshcon, ssherr := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshHost, sshPort), sshConfig)
	
	if ssherr !=nil{
		fmt.Printf("Failed to connect to the ssh\n")
		return nil, ssherr
	}
		// Connect to the SSH Server
		// if sshcon, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", sshHost, sshPort), sshConfig); err == nil {
		// defer sshcon.Close()

	// Now we register the ViaSSHDialer with the ssh connection as a parameter
	// sshcon을 맴버로 가지는 구조체 ViaSSHDialer의 주소를 파라미터로 가진다.
	sql.Register("postgres+ssh", &ViaSSHDialer{sshcon})

	// And now we can use our new driver with the regular postgres connection string tunneled through the SSH connection
	db, err := sql.Open("postgres+ssh", fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, dbHost, dbName))

	if err != nil {
		fmt.Printf("Failed to connect to the db\n")
    	return nil, err
    }
	
	err = db.Ping()
	
	if err != nil {
		fmt.Printf("Failed to connect to the db connection\n")
		return nil, err
	}
     
	fmt.Printf("Successfully connected to the db\n")
	
	return db,nil
	// db query test

	// if rows, err := db.Query("SELECT csp,namespace_name FROM kube_services ORDER BY csp"); err == nil {
	// 	for rows.Next() {
	// 		var id string
	// 		var name string
	// 		rows.Scan(&id, &name)
	// 		fmt.Printf("CSP: %s  NameSpace: %s\n", id, name)
	// 	}
	// 	rows.Close()
	// }else{
	// 	fmt.Printf(err.Error())
	// }

}