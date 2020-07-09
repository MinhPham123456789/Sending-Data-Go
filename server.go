// SERVER //
package main

import (
    //"crypto/sha256"
    "fmt"
    "io"
    "net"
    "os"
    "strings"
    "strconv"
    "bytes"
)

const BUFFERSIZE = 1024
const PORT = "8090"

func main() {

    fmt.Println("start listening")

    server, error := net.Listen("tcp", "localhost:"+PORT)
    if error != nil {
        fmt.Println("There was an error starting the server" + error.Error())
        return
    }

    //infinite loop
    for {

        connection, error := server.Accept()
        if error != nil {
            fmt.Println("There was an error with the connection" + error.Error())
            return
        }
        fmt.Println("connected")
        //handle the connection, on it's own thread, per connection
        go connectionHandler(connection)
    }
}

func connectionHandler(connection net.Conn) {
    buffer := make([]byte, BUFFERSIZE)

    _, error := connection.Read(buffer)
    if error != nil {
        fmt.Println("There is an error reading in command connection", error.Error())
        return
    }
    fmt.Println("command received: " + string(buffer))
    connection.Write([]byte("Command received"))

    //loop until disconntect

    cleanedBuffer := bytes.Trim(buffer, "\x00")
    cleanedInputCommandString := strings.TrimSpace(string(cleanedBuffer))
    arrayOfCommands := strings.Split(cleanedInputCommandString, " ")

    fmt.Println(arrayOfCommands[0])
    if arrayOfCommands[0] == "get" {
        sendFileToClient(arrayOfCommands[1], connection)
    } else if arrayOfCommands[0] == "send" {
        fmt.Println("getting a file")
        getFileFromClient(connection)

    } else {
        _, error = connection.Write([]byte("bad command"))
    }

}

func getFileFromClient(connection net.Conn){
    defer connection.Close()
    fmt.Println("Connected to server, start receiving the file name and file size")
    bufferFileName := make([]byte, 64)
    bufferFileSize := make([]byte, 10)

    connection.Read(bufferFileSize)
    fileSize, _ := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)

    connection.Read(bufferFileName)
    fileName := strings.Trim(string(bufferFileName), ":")
    
    newFile, err := os.Create(fileName)

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
    
    //Hash computing
/*
    connection.Write([]byte("Received data!"))
    response_buffer := make([]byte, BUFFERSIZE)
    newFile, err = os.Open(fileName)
    hasher := sha256.New()
    if _,err :=io.Copy(hasher, newFile); err != nil {
        fmt.Println("calculate hash err:", err)
    }
    new_sha256_sum := hasher.Sum(nil)


    _, err = connection.Read(response_buffer)
    fmt.Println(new_sha256_sum)
    fmt.Println(response_buffer)
    if err != nil {
        fmt.Println("Complete receiving hash err:", err)
    }
    hash_comparison_result := bytes.Compare(new_sha256_sum, response_buffer)
    if hash_comparison_result == 0 {
        connection.Write([]byte("Clean"))
    } else {
        connection.Write([]byte("Not Clean"))
    }
*/ 
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


func sendFileToClient(fileName string, connection net.Conn){
    fmt.Println("Testing Server send")
    response_buffer := make([]byte, BUFFERSIZE)
    _, error := connection.Read(response_buffer)

    if error != nil {
        fmt.Println("There is an error reading response in connection", error.Error())
        return
    }

    defer connection.Close()
    fmt.Println(string(response_buffer))
    fmt.Println("Server is going to send item!")
    
    file, err := os.Open("./"+fileName)
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
}

// END SERVER //
