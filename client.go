// CLIENT ///
    package main

import (
    "bytes"
    //"crypto/sha256"
    "regexp"
    "fmt"
    "log"
    "io"
    "net"
    "os"
    "strings"
    "strconv"
    "io/ioutil"
    //"encoding/hex"
)

const BUFFERSIZE = 1024

func main() {

    //get port and ip address to dial

    if len(os.Args) != 5 {
        fmt.Println("use example: tcpClient 127.0.0.1 7005 send \"t*t.txt\"(pattern)")
        return
    }

    var ip string = os.Args[1]
    var port string = os.Args[2]
    var command string = os.Args[3]
    var name_pattern string = os.Args[4]

    names, err := ioutil.ReadDir("./")
    if err != nil {
        log.Fatal(err)
    }

    names_array := make([]string, 0)
    
    for _, n := range names {
        matched, err := regexp.MatchString(name_pattern, n.Name())
        //fmt.Println(matched, n.Name())
        if err != nil{
            fmt.Println("Regexp err: ", err)
        }
        if matched {
            names_array = append(names_array, n.Name())
        }
    }
    //fmt.Println(names_array)
    

    for len(names_array) > 0 {
        // fmt.Println(names_array[0])
        connection, err := net.Dial("tcp", ip+":"+port)
        if err != nil {
           fmt.Println("There was an error making a connection")
           return
        }
        if command == "get" {
            getFileFromServer(names_array[0], connection)

        } else if command == "send" {
            sendFileToServer(names_array[0], connection)
            names_array = names_array[1:]
        } else {
            fmt.Println("Bad Command")
        }
    }
    

}

func sendFileToServer(fileName string, connection net.Conn) {
    response_buffer := make([]byte, BUFFERSIZE)
    connection.Write([]byte("send " + fileName))
    _, error := connection.Read(response_buffer)

    //fmt.Println("ok1")
    if error != nil {
        fmt.Println("There is an error reading response in connection", error.Error())
        return
    }

    defer connection.Close()
    fmt.Println(string(response_buffer))
    fmt.Println("A client has connected!")
    //fmt.Println(os.Stat("/home/neuralnet/test.txt"))
    file, err := os.Open(fileName)
    if err != nil {
	fmt.Println(err)
	return
    }
    fmt.Println(fileName)
    fileInfo, err := file.Stat()
    if err != nil {
	fmt.Println(err)
	return
    }
    fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
    fileSendName := fillString(fileInfo.Name(), 64)
    fmt.Println("Sending filename and filesize!")
    connection.Write([]byte(fileSize))
    connection.Write([]byte(fileSendName))
    sendBuffer := make([]byte, BUFFERSIZE)
    fmt.Println("Start sending file!")
    for {
	_, err = file.Read(sendBuffer)
	if err == io.EOF {
	    break
	}
	connection.Write(sendBuffer)
    }
    fmt.Println("File has been sent, closing connection!")

    //Hash computing
    /*_, err = connection.Read(response_buffer)
    fmt.Println(string(response_buffer))
    if err != nil {
        fmt.Println("Complete sending response sending data err:", err)
    }
    fmt.Println("Send SHA256 sum")
    hasher := sha256.New()
    if _,err :=io.Copy(hasher, file); err != nil {
        fmt.Println("calculate hash err:", err)
    }
    sha256_sum := hasher.Sum(nil)
    fmt.Println("hash:", hex.EncodeToString(sha256_sum))
    connection.Write(sha256_sum)
    _, err = connection.Read(response_buffer)
    if err != nil {
        fmt.Println("Complete sending response sending hash err:", err)
    }
    _, err = connection.Read(response_buffer)
    fmt.Println(string(response_buffer))*/
    return
}


func fillString(retunString string, toLength int) string {
    for {
        lengtString := len(retunString)
	if lengtString < toLength {
	    retunString = retunString + ":"
	    continue
	}
	break
    }
    return retunString
}



func getFileFromServer(fileName string, connection net.Conn) {

    var currentByte int64 = 0

    fileBuffer := make([]byte, BUFFERSIZE)

    var err error
    file, err := os.Create(strings.TrimSpace(fileName))
    if err != nil {
        log.Fatal(err)
    }
    connection.Write([]byte("get " + fileName))
    for {

        connection.Read(fileBuffer)
        cleanedFileBuffer := bytes.Trim(fileBuffer, "\x00")

        _, err = file.WriteAt(cleanedFileBuffer, currentByte)

        currentByte += BUFFERSIZE

        if err == io.EOF {
            break
        }

    }

    file.Close()
    return

}
// END CLIENT //
