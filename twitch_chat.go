package main

import (
	"bufio"
	"log"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func connect(wg sync.WaitGroup, nick string, oauth string, channel string) {
	defer wg.Done()

	// Connect to the server
	conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
	if err != nil {
		panic(err)
	}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Log into the server
	conn.Write([]byte("PASS " + oauth + "\r\n"))
	conn.Write([]byte("NICK " + nick + "\r\n"))
	conn.Write([]byte("CAP REQ :twitch.tv/tags\r\n"))
	// TODO: Is it worthwile to support RECONNECT events?
	conn.Write([]byte("CAP REQ :twitch.tv/commands\r\n"))
	conn.Write([]byte("JOIN " + channel + "\r\n"))

	reader := bufio.NewReaderSize(conn, 2048)
	tp := textproto.NewReader(reader)

	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Printf("[%s] Disconnected\n", channel)

			reconnected := false
			sleepTime := 2
			attemptNum := 1

			for !reconnected {
				conn.Close()

				log.Printf("[%s] Reconnecting in %d seconds\n", channel, sleepTime)
				time.Sleep(time.Duration(sleepTime) * time.Second)

				// Cap sleepTime at 32 seconds
				if sleepTime < 32 {
					sleepTime *= 2
				}

				conn, err := net.Dial("tcp", "irc.chat.twitch.tv:6667")
				if err != nil {
					log.Printf("[%s] Reconnection attempt #%d failed\n", channel, attemptNum)
					attemptNum++
				} else {
					reconnected = true

					conn.Write([]byte("PASS " + oauth + "\r\n"))
					conn.Write([]byte("NICK " + nick + "\r\n"))
					conn.Write([]byte("CAP REQ :twitch.tv/tags\r\n"))
					conn.Write([]byte("CAP REQ :twitch.tv/commands\r\n"))
					conn.Write([]byte("JOIN " + channel + "\r\n"))

					reader = bufio.NewReaderSize(conn, 2048)
					tp = textproto.NewReader(reader)
				}
			}
		}

		// DEBUG
		// log.Println(line)

		if len(line) > 0 && line[0] == '@' {
			lineSplit := strings.SplitN(line, " :", 3)

			middleSplit := strings.Split(lineSplit[1], " ")

			switch middleSplit[1] {
			case "PRIVMSG":
				// Parse the "tmi-sent-ts" value from the tags string
				// Remove the '@' prefix
				tagsSplit := strings.Split(lineSplit[0][1:], ";")

				var timestamp int
				for idx := range tagsSplit {
					if strings.HasPrefix(tagsSplit[idx], "tmi-sent-ts") {
						timestamp, err = strconv.Atoi(tagsSplit[idx][12:])
						if err != nil {
							panic(err)
						}

						break
					}
				}

				// This takes the Unix timestamp in milliseconds for when
				// the given message was sent and calculates the timestamp
				// in seconds for when the corresponding day started
				// (00:00:00). For example, a message sent at 1535726924446
				// will map to 1535673600.
				// NOTE: 0x5265C00 = 8.64e+7 (number of milliseconds in a day)
				timestampDay := (timestamp - (timestamp % 0x5265C00)) / 1000

				timestampDayStr := strconv.Itoa(timestampDay)

				outputPath := filepath.Join("./logs/", middleSplit[2][1:], timestampDayStr+".txt")

				// Open and write the line to file
				logFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
				if err != nil {
					log.Fatal(err)
				}

				if _, err = logFile.WriteString(line + "\n"); err != nil {
					log.Fatal(err)
				}

				logFile.Close()
			case "USERNOTICE":
				tagsSplit := strings.Split(lineSplit[0], ";")

				var timestamp int
				for idx := range tagsSplit {
					if strings.HasPrefix(tagsSplit[idx], "tmi-sent-ts") {
						timestamp, err = strconv.Atoi(tagsSplit[idx][12:])
						if err != nil {
							panic(err)
						}

						break
					}
				}

				timestampDay := (timestamp - (timestamp % 0x5265C00)) / 1000

				timestampDayStr := strconv.Itoa(timestampDay)

				outputPath := filepath.Join("./logs/", middleSplit[2][1:], timestampDayStr+"_usernotice.txt")

				// Open and write the line to file
				logFile, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
				if err != nil {
					log.Fatal(err)
				}

				if _, err = logFile.WriteString(line + "\n"); err != nil {
					log.Fatal(err)
				}

				logFile.Close()
			}
		} else if len(line) > 0 && line[0] == ':' {
			var lineSplit []string
			cursor := 1
			maxSize := 2

			for idx := range line {
				if line[idx] == ' ' {
					lineSplit = append(lineSplit, line[cursor:cursor+(idx-cursor)])
					cursor += (idx - cursor) + 1
				}

				if len(lineSplit) == 2 {
					switch lineSplit[1] {
					case "CAP":
						maxSize = 5
					case "JOIN":
						maxSize = 3
					}
				}

				// Break when the maxSize is hit, ensure the last portion of
				// the line is added
				if len(lineSplit) == maxSize {
					break
				} else if cursor+(idx-cursor)+1 == len(line) {
					lineSplit = append(lineSplit, line[cursor:cursor+(idx-cursor)+1])
				}
			}

			switch lineSplit[1] {
			case "001":
				log.Printf("[%s] Connected\n", channel)
			case "NOTICE":
				// The only time this should catch is when login authentication
				// fails, but maybe I'm wrong...
				log.Printf("[%s] Notice: %s\n", channel, line[25:])
			case "JOIN":
				// Parse the sender from a given ':sender!sender@tmi.twitch.tv' string
				var sender string
				for idx := range lineSplit[0][1:] {
					if lineSplit[0][idx] == '!' {
						sender = lineSplit[0][:idx]
						break
					}
				}

				if sender == nick {
					log.Printf("[%s] Joined\n", channel)
				}
			}
		} else if line == "PING :tmi.twitch.tv" {
			conn.Write([]byte("PONG :tmi.twitch.tv\r\n"))
		}

		conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
	}
}
