package main

import (
	"context"
	"log"
	"math/rand"
	"time"

	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/SevereCloud/vksdk/v2/api/params"
	"github.com/SevereCloud/vksdk/v2/events"
	"github.com/SevereCloud/vksdk/v2/longpoll-bot"

	"bufio"
	"fmt"
	"os"
)

// readLines reads a whole file into memory
// and returns a slice of its lines.
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// writeLines writes the lines to the given file.
func writeLines(lines []string, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}

func main() {
	insults, err := readLines("./insults")
	if err != nil {
		log.Fatalf("readLines: %s", err)
	}

	rand.Seed(time.Now().UTC().UnixNano())

	token := os.Getenv("ZZ_TOKEN")
	vk := api.NewVK(token)

	// Получаем информацию о группе
	group, err := vk.GroupsGetByID(api.Params{})
	if err != nil {
		log.Fatal(err)
	}

	// Инициализируем longpoll
	lp, err := longpoll.NewLongPoll(vk, group[0].ID)
	if err != nil {
		log.Fatal(err)
	}

	// Событие нового сообщения
	lp.MessageNew(func(ctx context.Context, obj events.MessageNewObject) {
		if os.Getenv("LOG_ALL_MSGS") == "1" {
			log.Printf("%d: %s", obj.Message.PeerID, obj.Message.Text)
		}

		// Если сообщение не из беседы
		if obj.Message.FromID == obj.Message.PeerID ||
			obj.Message.ID != 0 ||
			obj.Message.PeerID < 2000000000 {
			return
		}

		// Если это ГС
		amAtt := obj.Message.Attachments[:0]
		for _, x := range obj.Message.Attachments {
			if x.Type == "audio_message" {
				amAtt = append(amAtt, x)
			}
		}

		if len(amAtt) > 0 {
			users, err := vk.UsersGet(api.Params{
				"user_ids": obj.Message.FromID,
			})
			if err != nil {
				log.Fatal(err)
			}

			b := params.NewMessagesSendBuilder()
			text := fmt.Sprintf("%s%d%s%s%s%s", "[id", obj.Message.FromID, "|", users[0].FirstName, "], ", insults[rand.Intn(len(insults))])
			b.Message(text)
			b.RandomID(0)
			b.PeerID(obj.Message.PeerID)

			log.Printf(text)

			_, err = vk.MessagesSend(b.Params)
			if err != nil {
				log.Fatal(err)
			}
		}
	})

	// Запускаем Bots Longpoll
	log.Println("Start longpoll")
	if err := lp.Run(); err != nil {
		log.Fatal(err)
	}
}
