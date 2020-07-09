// CLIENT ///
    package main

import (
    //"bytes"
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
    "time"
    //"encoding/hex"
)

const BUFFERSIZE = 1024

func main() {

    //get port and ip address to dial

    if len(os.Args) != 5 {
        fmt.Println("use example: run client.go 127.0.0.1 9000 send \"t*t.txt\"(pattern)")
        fmt.Println("use example: run client.go 127.0.0.1 9000 get item_name")
        return
    }

    var ip string = os.Args[1]
    var port string = os.Args[2]
    var command string = os.Args[3]
    var name_pattern string = os.Args[4]

    names, err := ioutil.ReadDir("../Downloads/")
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
    if command == "get" {
        names_array = append(names_array, name_pattern)
    }
    fmt.Println(names_array)
    

    for len(names_array) > 0 {
        // fmt.Println(names_array[0])
        connection, err := net.Dial("tcp", ip+":"+port)
        if err != nil {
           fmt.Println("There was an error making a connection")
           return
        }
        if command == "get" {
            connection.Write([]byte("get " + names_array[0]))  // name_pattern in get command is item name
            names_array = names_array[1:]
            getFileFromServer(connection)

        } else if command == "send" {
            sendFileToServer(names_array[0], connection)
            names_array = names_array[1:]
        } else {
            fmt.Println("Bad Command")
        }
	time.Sleep(800 * time.Millisecond)  //This helps not overshooting the server with too quick consecutive requests and may cause bugs
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
    file, err := os.Open("../Downloads/"+fileName)
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
    sent_byte_count := 0
    for {
	_, err = file.Read(sendBuffer)
	if err == io.EOF {
	    break
	}
	connection.Write(sendBuffer)
        sent_byte_count = sent_byte_count + BUFFERSIZE
        fmt.Println("Sending process:", sent_byte_count, fileInfo.Size(), (float64(sent_byte_count) / float64(fileInfo.Size())))
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



func getFileFromServer(connection net.Conn) {

    defer connection.Close()
    fmt.Println("Connected to server, start receiving the file name and file size")
    bufferFileName := make([]byte, 64)
    bufferFileSize := make([]byte, 10)
    server_response := make([]byte, BUFFERSIZE)
    connection.Read(server_response)
    fmt.Println(string(server_response))
    
    // Send ready signal
    connection.Write([]byte("Ready to receive"))

    connection.Read(bufferFileSize)
    fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

    connection.Read(bufferFileName)
    fileName := strings.Trim(string(bufferFileName), ":")
    
    newFile, err := os.Create("./test/"+fileName)

    if err != nil {
	panic(err)
    }
    defer newFile.Close()
    var receivedBytes int64

    for {
        //fmt.Println("Size:", fileSize, receivedBytes)
	if (fileSize - receivedBytes) != 0 && (fileSize - receivedBytes) < BUFFERSIZE {
	    io.CopyN(newFile, connection, (fileSize - receivedBytes))
	    connection.Read(make([]byte, (receivedBytes+BUFFERSIZE)-fileSize))
	    break
	} else if (fileSize - receivedBytes) == 0 {	// This will break the loop when the received byte is equal to size. not hang on connection Read() when (fileSize - receivedBytes) == 0
            break
        }
	io.CopyN(newFile, connection, BUFFERSIZE)
	receivedBytes += BUFFERSIZE
    }
    fmt.Println("Received file completely!", fileName)

}
// END CLIENT //
